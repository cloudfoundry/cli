package route_test

import (
	"errors"

	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/commands/route"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/flags"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"
	"github.com/blang/semver"

	"code.cloudfoundry.org/cli/cf/api/apifakes"

	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"

	"strings"

	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UnmapRoute", func() {
	var (
		ui         *testterm.FakeUI
		configRepo coreconfig.Repository
		routeRepo  *apifakes.FakeRouteRepository

		cmd         commandregistry.Command
		deps        commandregistry.Dependency
		factory     *requirementsfakes.FakeFactory
		flagContext flags.FlagContext

		loginRequirement         requirements.Requirement
		applicationRequirement   *requirementsfakes.FakeApplicationRequirement
		domainRequirement        *requirementsfakes.FakeDomainRequirement
		minAPIVersionRequirement requirements.Requirement

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

		cmd = &route.UnmapRoute{}
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

		minAPIVersionRequirement = &passingRequirement{Name: "min-api-version-requirement"}
		factory.NewMinAPIVersionRequirementReturns(minAPIVersionRequirement)
	})

	Describe("Help text", func() {
		var usage []string

		BeforeEach(func() {
			cmd := &route.UnmapRoute{}
			up := commandregistry.CLICommandUsagePresenter(cmd)

			usage = strings.Split(up.Usage(), "\n")
		})

		It("contains an example", func() {
			Expect(usage).To(ContainElement("   cf unmap-route my-app example.com --port 5000                  # example.com:5000"))
		})

		It("contains the options", func() {
			Expect(usage).To(ContainElement("   --hostname, -n      Hostname used to identify the HTTP route"))
			Expect(usage).To(ContainElement("   --path              Path used to identify the HTTP route"))
			Expect(usage).To(ContainElement("   --port              Port used to identify the TCP route"))
		})

		It("shows the usage", func() {
			Expect(usage).To(ContainElement("   Unmap an HTTP route:"))
			Expect(usage).To(ContainElement("      cf unmap-route APP_NAME DOMAIN [--hostname HOSTNAME] [--path PATH]"))

			Expect(usage).To(ContainElement("   Unmap a TCP route:"))
			Expect(usage).To(ContainElement("      cf unmap-route APP_NAME DOMAIN --port PORT"))
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
					[]string{"Incorrect Usage. Requires app_name, domain_name as arguments"},
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

			Context("when a path is passed", func() {
				BeforeEach(func() {
					flagContext.Parse("app-name", "domain-name", "--path", "the-path")
				})

				It("returns a MinAPIVersionRequirement as the first requirement", func() {
					actualRequirements, err := cmd.Requirements(factory, flagContext)
					Expect(err).NotTo(HaveOccurred())

					expectedVersion, err := semver.Make("2.36.0")
					Expect(err).NotTo(HaveOccurred())

					Expect(factory.NewMinAPIVersionRequirementCallCount()).To(Equal(1))
					feature, requiredVersion := factory.NewMinAPIVersionRequirementArgsForCall(0)
					Expect(feature).To(Equal("Option '--path'"))
					Expect(requiredVersion).To(Equal(expectedVersion))
					Expect(actualRequirements[0]).To(Equal(minAPIVersionRequirement))
				})
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

			Context("when no options are passed", func() {
				BeforeEach(func() {
					flagContext.Parse("app-name", "domain-name")
				})

				It("does not return a MinAPIVersionRequirement", func() {
					actualRequirements, err := cmd.Requirements(factory, flagContext)
					Expect(err).NotTo(HaveOccurred())
					Expect(factory.NewMinAPIVersionRequirementCallCount()).To(Equal(0))
					Expect(actualRequirements).NotTo(ContainElement(minAPIVersionRequirement))
				})
			})

			Context("when a port is passed", func() {
				BeforeEach(func() {
					flagContext.Parse("app-name", "domain-name", "--port", "5000")
				})

				It("returns a MinAPIVersionRequirement as the first requirement", func() {
					actualRequirements, err := cmd.Requirements(factory, flagContext)
					Expect(err).NotTo(HaveOccurred())

					expectedVersion, err := semver.Make("2.53.0")
					Expect(err).NotTo(HaveOccurred())

					Expect(factory.NewMinAPIVersionRequirementCallCount()).To(Equal(1))
					feature, requiredVersion := factory.NewMinAPIVersionRequirementArgsForCall(0)
					Expect(feature).To(Equal("Option '--port'"))
					Expect(requiredVersion).To(Equal(expectedVersion))
					Expect(actualRequirements[0]).To(Equal(minAPIVersionRequirement))
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

		It("tries to find the route", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(routeRepo.FindCallCount()).To(Equal(1))
			hostname, domain, path, port := routeRepo.FindArgsForCall(0)
			Expect(hostname).To(Equal(""))
			Expect(domain).To(Equal(fakeDomain))
			Expect(path).To(Equal(""))
			Expect(port).To(Equal(0))
		})

		Context("when a path is passed", func() {
			BeforeEach(func() {
				err := flagContext.Parse("app-name", "domain-name", "--path", "the-path")
				Expect(err).NotTo(HaveOccurred())
				cmd.Requirements(factory, flagContext)
			})

			It("tries to find the route with the path", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(routeRepo.FindCallCount()).To(Equal(1))
				_, _, path, _ := routeRepo.FindArgsForCall(0)
				Expect(path).To(Equal("the-path"))
			})
		})

		Context("when a port is passed", func() {
			BeforeEach(func() {
				err := flagContext.Parse("app-name", "domain-name", "--port", "60000")
				Expect(err).NotTo(HaveOccurred())
				cmd.Requirements(factory, flagContext)
			})

			It("tries to find the route with the port", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(routeRepo.FindCallCount()).To(Equal(1))
				_, _, _, port := routeRepo.FindArgsForCall(0)
				Expect(port).To(Equal(60000))
			})
		})

		Context("when the route can be found", func() {
			BeforeEach(func() {
				routeRepo.FindReturns(models.Route{GUID: "route-guid"}, nil)
			})

			It("tells the user that it is removing the route", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Removing route", "from app", "in org"},
				))
			})

			Context("when the returned route has an app with the requested app's guid", func() {
				BeforeEach(func() {
					route := models.Route{
						GUID: "route-guid",
						Apps: []models.ApplicationFields{
							{GUID: "fake-app-guid"},
						},
					}
					routeRepo.FindReturns(route, nil)
				})

				It("tries to unbind the route from the app", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(routeRepo.UnbindCallCount()).To(Equal(1))
					routeGUID, appGUID := routeRepo.UnbindArgsForCall(0)
					Expect(routeGUID).To(Equal("route-guid"))
					Expect(appGUID).To(Equal("fake-app-guid"))
				})

				Context("when unbinding the route from the app fails", func() {
					BeforeEach(func() {
						routeRepo.UnbindReturns(errors.New("unbind-err"))
					})

					It("returns an error", func() {
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(Equal("unbind-err"))
					})
				})

				Context("when unbinding the route from the app succeeds", func() {
					BeforeEach(func() {
						routeRepo.UnbindReturns(nil)
					})

					It("tells the user it succeeded", func() {
						Expect(err).NotTo(HaveOccurred())
						Expect(ui.Outputs()).To(BeInDisplayOrder(
							[]string{"OK"},
						))
					})
				})
			})

			Context("when the returned route does not have an app with the requested app's guid", func() {
				BeforeEach(func() {
					route := models.Route{
						GUID: "route-guid",
						Apps: []models.ApplicationFields{
							{GUID: "other-fake-app-guid"},
						},
					}
					routeRepo.FindReturns(route, nil)
				})

				It("does not unbind the route from the app", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(routeRepo.UnbindCallCount()).To(Equal(0))
				})

				It("tells the user 'OK'", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"OK"},
					))
				})

				It("warns the user the route was not mapped to the application", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"Route to be unmapped is not currently mapped to the application."},
					))
				})
			})
		})

		Context("when the route cannot be found", func() {
			BeforeEach(func() {
				routeRepo.FindReturns(models.Route{}, errors.New("find-by-host-and-domain-err"))
			})

			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("find-by-host-and-domain-err"))
			})
		})

	})
})
