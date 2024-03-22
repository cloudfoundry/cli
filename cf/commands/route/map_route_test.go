package route_test

import (
	"errors"

	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/commands/route"
	"code.cloudfoundry.org/cli/cf/commands/route/routefakes"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/flags"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"

	"code.cloudfoundry.org/cli/cf/api/apifakes"

	testconfig "code.cloudfoundry.org/cli/cf/util/testhelpers/configuration"
	testterm "code.cloudfoundry.org/cli/cf/util/testhelpers/terminal"

	"strings"

	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MapRoute", func() {
	var (
		ui         *testterm.FakeUI
		configRepo coreconfig.Repository
		routeRepo  *apifakes.FakeRouteRepository

		cmd         commandregistry.Command
		deps        commandregistry.Dependency
		factory     *requirementsfakes.FakeFactory
		flagContext flags.FlagContext

		loginRequirement       requirements.Requirement
		applicationRequirement *requirementsfakes.FakeApplicationRequirement
		domainRequirement      *requirementsfakes.FakeDomainRequirement

		originalCreateRouteCmd commandregistry.Command
		fakeCreateRouteCmd     commandregistry.Command

		fakeDomain models.DomainFields
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		routeRepo = new(apifakes.FakeRouteRepository)
		repoLocator := deps.RepoLocator.SetRouteRepository(routeRepo)

		deps = commandregistry.Dependency{
			UI:          ui,
			Config:      configRepo,
			RepoLocator: repoLocator,
		}

		originalCreateRouteCmd = commandregistry.Commands.FindCommand("create-route")
		fakeCreateRouteCmd = new(routefakes.OldFakeRouteCreator)
		commandregistry.Register(fakeCreateRouteCmd)

		cmd = &route.MapRoute{}
		cmd.SetDependency(deps, false)

		flagContext = flags.NewFlagContext(cmd.MetaData().Flags)

		factory = new(requirementsfakes.FakeFactory)

		loginRequirement = &passingRequirement{Name: "login-requirement"}
		factory.NewLoginRequirementReturns(loginRequirement)

		applicationRequirement = new(requirementsfakes.FakeApplicationRequirement)
		factory.NewApplicationRequirementReturns(applicationRequirement)

		fakeApplication := models.Application{}
		fakeApplication.GUID = "fake-app-guid"
		applicationRequirement.GetApplicationReturns(fakeApplication)

		domainRequirement = new(requirementsfakes.FakeDomainRequirement)
		factory.NewDomainRequirementReturns(domainRequirement)

		fakeDomain = models.DomainFields{
			GUID: "fake-domain-guid",
			Name: "fake-domain-name",
		}
		domainRequirement.GetDomainReturns(fakeDomain)
	})

	AfterEach(func() {
		commandregistry.Register(originalCreateRouteCmd)
	})

	Describe("Help text", func() {
		var usage []string

		BeforeEach(func() {
			cmd := &route.MapRoute{}
			up := commandregistry.CLICommandUsagePresenter(cmd)

			usage = strings.Split(up.Usage(), "\n")
		})

		It("contains an example", func() {
			Expect(usage).To(ContainElement("   cf map-route my-app example.com --port 50000                 # example.com:50000"))
		})

		It("contains the options", func() {
			Expect(usage).To(ContainElement("   --hostname, -n      Hostname for the HTTP route (required for shared domains)"))
			Expect(usage).To(ContainElement("   --path              Path for the HTTP route"))
			Expect(usage).To(ContainElement("   --port              Port for the TCP route"))
			Expect(usage).To(ContainElement("   --random-port       Create a random port for the TCP route"))
		})

		It("shows the usage", func() {
			Expect(usage).To(ContainElement("   Map an HTTP route:"))
			Expect(usage).To(ContainElement("      cf map-route APP_NAME DOMAIN [--hostname HOSTNAME] [--path PATH]"))

			Expect(usage).To(ContainElement("   Map a TCP route:"))
			Expect(usage).To(ContainElement("      cf map-route APP_NAME DOMAIN (--port PORT | --random-port)"))
		})
	})

	Describe("Requirements", func() {
		Context("when not provided exactly two args", func() {
			BeforeEach(func() {
				flagContext.Parse("app-name")
			})

			It("fails with usage", func() {
				_, err := cmd.Requirements(factory, flagContext)
				Expect(err).To(HaveOccurred())
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Incorrect Usage. Requires APP_NAME and DOMAIN as arguments"},
					[]string{"NAME"},
					[]string{"USAGE"},
				))
			})
		})

		Context("when provided exactly two args", func() {
			BeforeEach(func() {
				flagContext.Parse("app-name", "domain-name")
			})

			It("returns a LoginRequirement", func() {
				actualRequirements, err := cmd.Requirements(factory, flagContext)
				Expect(err).NotTo(HaveOccurred())
				Expect(factory.NewLoginRequirementCallCount()).To(Equal(1))

				Expect(actualRequirements).To(ContainElement(loginRequirement))
			})

			It("returns an ApplicationRequirement", func() {
				actualRequirements, err := cmd.Requirements(factory, flagContext)
				Expect(err).NotTo(HaveOccurred())
				Expect(factory.NewApplicationRequirementCallCount()).To(Equal(1))

				Expect(factory.NewApplicationRequirementArgsForCall(0)).To(Equal("app-name"))
				Expect(actualRequirements).To(ContainElement(applicationRequirement))
			})

			It("returns a DomainRequirement", func() {
				actualRequirements, err := cmd.Requirements(factory, flagContext)
				Expect(err).NotTo(HaveOccurred())
				Expect(factory.NewDomainRequirementCallCount()).To(Equal(1))

				Expect(factory.NewDomainRequirementArgsForCall(0)).To(Equal("domain-name"))
				Expect(actualRequirements).To(ContainElement(domainRequirement))
			})

			Context("when passing port with a hostname", func() {
				BeforeEach(func() {
					flagContext.Parse("app-name", "example.com", "--port", "8080", "--hostname", "something-else")
				})

				It("fails", func() {
					_, err := cmd.Requirements(factory, flagContext)
					Expect(err).To(HaveOccurred())
					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"FAILED"},
						[]string{"Cannot specify port together with hostname and/or path."},
					))
				})
			})

			Context("when passing port with a path", func() {
				BeforeEach(func() {
					flagContext.Parse("app-name", "example.com", "--port", "8080", "--path", "something-else")
				})

				It("fails", func() {
					_, err := cmd.Requirements(factory, flagContext)
					Expect(err).To(HaveOccurred())
					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"FAILED"},
						[]string{"Cannot specify port together with hostname and/or path."},
					))
				})
			})

			Context("when both --port and --random-port are given", func() {
				BeforeEach(func() {
					err := flagContext.Parse("app-name", "domain-name", "--port", "9090", "--random-port")
					Expect(err).NotTo(HaveOccurred())
				})

				It("fails with error", func() {
					_, err := cmd.Requirements(factory, flagContext)
					Expect(err).To(HaveOccurred())
					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"FAILED"},
						[]string{"Cannot specify random-port together with port, hostname and/or path."},
					))
				})
			})

			Context("when both --random-port and --hostname are given", func() {
				BeforeEach(func() {
					err := flagContext.Parse("app-name", "domain-name", "--hostname", "host", "--random-port")
					Expect(err).NotTo(HaveOccurred())
				})

				It("fails with error", func() {
					_, err := cmd.Requirements(factory, flagContext)
					Expect(err).To(HaveOccurred())
					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"FAILED"},
						[]string{"Cannot specify random-port together with port, hostname and/or path."},
					))
				})
			})

			Context("when --random-port and --path are given", func() {
				BeforeEach(func() {
					err := flagContext.Parse("app-name", "domain-name", "--path", "path", "--random-port")
					Expect(err).NotTo(HaveOccurred())
				})

				It("fails with error", func() {
					_, err := cmd.Requirements(factory, flagContext)
					Expect(err).To(HaveOccurred())
					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"FAILED"},
						[]string{"Cannot specify random-port together with port, hostname and/or path."},
					))
				})
			})
		})
	})

	Describe("Execute", func() {
		var err error

		BeforeEach(func() {
			err := flagContext.Parse("app-name", "domain-name")
			Expect(err).NotTo(HaveOccurred())
			cmd.Requirements(factory, flagContext)
		})

		JustBeforeEach(func() {
			err = cmd.Execute(flagContext)
		})

		It("tries to create the route", func() {
			Expect(err).ToNot(HaveOccurred())
			fakeRouteCreator, ok := fakeCreateRouteCmd.(*routefakes.OldFakeRouteCreator)
			Expect(ok).To(BeTrue())

			Expect(fakeRouteCreator.CreateRouteCallCount()).To(Equal(1))
			host, path, port, randomPort, domain, space := fakeRouteCreator.CreateRouteArgsForCall(0)
			Expect(host).To(Equal(""))
			Expect(path).To(Equal(""))
			Expect(port).To(Equal(0))
			Expect(randomPort).To(BeFalse())
			Expect(domain).To(Equal(fakeDomain))
			Expect(space).To(Equal(models.SpaceFields{
				Name: "my-space",
				GUID: "my-space-guid",
			}))
		})

		Context("when a port is passed", func() {
			BeforeEach(func() {
				err := flagContext.Parse("app-name", "domain-name", "--port", "60000")
				Expect(err).NotTo(HaveOccurred())
				cmd.Requirements(factory, flagContext)
			})

			It("tries to create the route with the port", func() {
				fakeRouteCreator, ok := fakeCreateRouteCmd.(*routefakes.OldFakeRouteCreator)
				Expect(ok).To(BeTrue())

				Expect(err).ToNot(HaveOccurred())
				Expect(fakeRouteCreator.CreateRouteCallCount()).To(Equal(1))
				_, _, port, _, _, _ := fakeRouteCreator.CreateRouteArgsForCall(0)
				Expect(port).To(Equal(60000))
			})
		})

		Context("when a random-port is passed", func() {
			BeforeEach(func() {
				err := flagContext.Parse("app-name", "domain-name", "--random-port")
				Expect(err).NotTo(HaveOccurred())
				cmd.Requirements(factory, flagContext)
			})

			It("tries to create the route with a random port", func() {
				fakeRouteCreator, ok := fakeCreateRouteCmd.(*routefakes.OldFakeRouteCreator)
				Expect(ok).To(BeTrue())

				Expect(err).ToNot(HaveOccurred())
				Expect(fakeRouteCreator.CreateRouteCallCount()).To(Equal(1))
				_, _, _, randomPort, _, _ := fakeRouteCreator.CreateRouteArgsForCall(0)
				Expect(randomPort).To(BeTrue())
			})
		})

		Context("when creating the route fails", func() {
			BeforeEach(func() {
				fakeRouteCreator, ok := fakeCreateRouteCmd.(*routefakes.OldFakeRouteCreator)
				Expect(ok).To(BeTrue())
				fakeRouteCreator.CreateRouteReturns(models.Route{}, errors.New("create-route-err"))
			})

			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("create-route-err"))
			})
		})

		Context("when creating the route succeeds", func() {
			BeforeEach(func() {
				fakeRouteCreator, ok := fakeCreateRouteCmd.(*routefakes.OldFakeRouteCreator)
				Expect(ok).To(BeTrue())
				fakeRouteCreator.CreateRouteReturns(models.Route{GUID: "fake-route-guid"}, nil)
			})

			It("tells the user that it is adding the route", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Adding route", "to app", "in org"},
				))
			})

			It("tries to bind the route", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(routeRepo.BindCallCount()).To(Equal(1))
				routeGUID, appGUID := routeRepo.BindArgsForCall(0)
				Expect(routeGUID).To(Equal("fake-route-guid"))
				Expect(appGUID).To(Equal("fake-app-guid"))
			})

			Context("when binding the route succeeds", func() {
				BeforeEach(func() {
					routeRepo.BindReturns(nil)
				})

				It("tells the user that it succeeded", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"OK"},
					))
				})
			})

			Context("when binding the route fails", func() {
				BeforeEach(func() {
					routeRepo.BindReturns(errors.New("bind-error"))
				})

				It("returns an error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("bind-error"))
				})
			})
		})

		Context("when a hostname is passed", func() {
			BeforeEach(func() {
				err := flagContext.Parse("app-name", "domain-name", "-n", "the-hostname")
				Expect(err).NotTo(HaveOccurred())
				cmd.Requirements(factory, flagContext)
			})

			It("tries to create the route with the hostname", func() {
				Expect(err).ToNot(HaveOccurred())
				fakeRouteCreator, ok := fakeCreateRouteCmd.(*routefakes.OldFakeRouteCreator)
				Expect(ok).To(BeTrue())
				Expect(fakeRouteCreator.CreateRouteCallCount()).To(Equal(1))
				hostName, _, _, _, _, _ := fakeRouteCreator.CreateRouteArgsForCall(0)
				Expect(hostName).To(Equal("the-hostname"))
			})
		})

		Context("when a hostname is not passed", func() {
			BeforeEach(func() {
				err := flagContext.Parse("app-name", "domain-name")
				Expect(err).NotTo(HaveOccurred())
				cmd.Requirements(factory, flagContext)
			})

			It("tries to create the route without a hostname", func() {
				Expect(err).ToNot(HaveOccurred())
				fakeRouteCreator, ok := fakeCreateRouteCmd.(*routefakes.OldFakeRouteCreator)
				Expect(ok).To(BeTrue())
				Expect(fakeRouteCreator.CreateRouteCallCount()).To(Equal(1))
				hostName, _, _, _, _, _ := fakeRouteCreator.CreateRouteArgsForCall(0)
				Expect(hostName).To(Equal(""))
			})
		})

		Context("when a path is passed", func() {
			BeforeEach(func() {
				err := flagContext.Parse("app-name", "domain-name", "--path", "the-path")
				Expect(err).NotTo(HaveOccurred())
				cmd.Requirements(factory, flagContext)
			})

			It("tries to create the route with the path", func() {
				Expect(err).ToNot(HaveOccurred())
				fakeRouteCreator, ok := fakeCreateRouteCmd.(*routefakes.OldFakeRouteCreator)
				Expect(ok).To(BeTrue())
				Expect(fakeRouteCreator.CreateRouteCallCount()).To(Equal(1))
				_, path, _, _, _, _ := fakeRouteCreator.CreateRouteArgsForCall(0)
				Expect(path).To(Equal("the-path"))
			})
		})
	})
})
