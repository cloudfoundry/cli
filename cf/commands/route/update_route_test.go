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

	"code.cloudfoundry.org/cli/cf/api/apifakes"

	testconfig "code.cloudfoundry.org/cli/cf/util/testhelpers/configuration"
	testterm "code.cloudfoundry.org/cli/cf/util/testhelpers/terminal"

	"strings"

	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("UpdateRoute", func() {
	var (
		ui         *testterm.FakeUI
		configRepo coreconfig.Repository
		routeRepo  *apifakes.FakeRouteRepository

		cmd         commandregistry.Command
		deps        commandregistry.Dependency
		factory     *requirementsfakes.FakeFactory
		flagContext flags.FlagContext

		loginRequirement  requirements.Requirement
		domainRequirement *requirementsfakes.FakeDomainRequirement

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

		cmd = &route.UpdateRoute{}
		cmd.SetDependency(deps, false)

		flagContext = flags.NewFlagContext(cmd.MetaData().Flags)

		factory = new(requirementsfakes.FakeFactory)

		loginRequirement = &passingRequirement{Name: "login-requirement"}
		factory.NewLoginRequirementReturns(loginRequirement)

		domainRequirement = new(requirementsfakes.FakeDomainRequirement)
		factory.NewDomainRequirementReturns(domainRequirement)

		fakeDomain = models.DomainFields{
			GUID: "fake-domain-guid",
			Name: "fake-domain-name",
		}
		domainRequirement.GetDomainReturns(fakeDomain)
	})

	AfterEach(func() {

	})

	Describe("Help text", func() {
		var usage []string

		BeforeEach(func() {
			cmd := &route.UpdateRoute{}
			up := commandregistry.CLICommandUsagePresenter(cmd)

			usage = strings.Split(up.Usage(), "\n")
		})

		It("contains an example", func() {
			Expect(usage).To(ContainElement("   cf update-route example.com -o loadbalancing=round-robin"))
		})

		It("contains the options", func() {
			Expect(usage).To(ContainElement("   --hostname, -n           Hostname for the HTTP route (required for shared domains)"))
			Expect(usage).To(ContainElement("   --path                   Path for the HTTP route"))
			Expect(usage).To(ContainElement("   --option, -o             Set the value of a per-route option"))
			Expect(usage).To(ContainElement("   --remove-option, -r      Remove an option with the given name"))
		})

		It("shows the usage", func() {
			Expect(usage).To(ContainElement("   Update an existing HTTP route:"))
			Expect(usage).To(ContainElement("      cf update-route DOMAIN [--hostname HOSTNAME] [--path PATH] [--option OPTION=VALUE] [--remove-option OPTION]"))
		})
	})

	Describe("Requirements", func() {
		Context("when not provided exactly one arg", func() {
			BeforeEach(func() {
			})

			It("fails with usage", func() {
				_, err := cmd.Requirements(factory, flagContext)
				Expect(err).To(HaveOccurred())
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Incorrect Usage. Requires DOMAIN as an argument"},
					[]string{"NAME"},
					[]string{"USAGE"},
				))
			})
		})

		Context("when provided exactly one arg", func() {
			BeforeEach(func() {
				flagContext.Parse("domain-name")
			})

			It("returns a LoginRequirement", func() {
				actualRequirements, err := cmd.Requirements(factory, flagContext)
				Expect(err).NotTo(HaveOccurred())
				Expect(factory.NewLoginRequirementCallCount()).To(Equal(1))

				Expect(actualRequirements).To(ContainElement(loginRequirement))
			})

			It("returns a DomainRequirement", func() {
				actualRequirements, err := cmd.Requirements(factory, flagContext)
				Expect(err).NotTo(HaveOccurred())
				Expect(factory.NewDomainRequirementCallCount()).To(Equal(1))

				Expect(factory.NewDomainRequirementArgsForCall(0)).To(Equal("domain-name"))
				Expect(actualRequirements).To(ContainElement(domainRequirement))
			})
		})
	})

	Describe("Execute", func() {
		var err error

		BeforeEach(func() {
			err := flagContext.Parse("domain-name")
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

		Context("when a hostname and a path are passed", func() {
			BeforeEach(func() {
				err := flagContext.Parse("domain-name", "--hostname", "the-hostname", "--path", "the-path")
				Expect(err).NotTo(HaveOccurred())
				cmd.Requirements(factory, flagContext)
			})

			It("tries to find the route with the hostname and path", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(routeRepo.FindCallCount()).To(Equal(1))
				hostname, _, path, _ := routeRepo.FindArgsForCall(0)
				Expect(hostname).To(Equal("the-hostname"))
				Expect(path).To(Equal("the-path"))
			})
		})

		Context("when the route can be found and an option is passed", func() {
			BeforeEach(func() {
				routeRepo.FindReturns(models.Route{GUID: "route-guid"}, nil)
				err := flagContext.Parse("domain-name", "--option", "loadbalancing=round-robin")
				Expect(err).NotTo(HaveOccurred())
				cmd.Requirements(factory, flagContext)
			})

			It("tries to set the given route option", func() {
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Setting route option loadbalancing", "for", "to round-robin"},
				))
			})
		})

		Context("when the route can be found and an option is passed, which already exists", func() {
			BeforeEach(func() {
				routeRepo.FindReturns(models.Route{GUID: "route-guid",
					Options: map[string]string{"loadbalancing": "round-robin", "a": "b"}}, nil)
				err := flagContext.Parse("domain-name", "--option", "loadbalancing=least-connection")
				Expect(err).NotTo(HaveOccurred())
				cmd.Requirements(factory, flagContext)
			})

			It("tries to set the given route option", func() {
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Setting route option loadbalancing", "for", "to least-connection"},
				))
			})
		})

		Context("when the route can be found a remove option is passed", func() {
			BeforeEach(func() {
				routeRepo.FindReturns(
					models.Route{GUID: "route-guid",
						Options: map[string]string{"loadbalancing": "round-robin", "a": "b"}},
					nil,
				)
				err := flagContext.Parse("domain-name", "--option", "a=b", "--remove-option", "loadbalancing")
				Expect(err).NotTo(HaveOccurred())
				cmd.Requirements(factory, flagContext)
			})

			It("tries to remove the given route option if it exists and gives an error message", func() {
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Removing route option loadbalancing", "for"},
				))
			})

		})

		Context("when the route can be found a remove option for non existent option is passed", func() {
			BeforeEach(func() {
				routeRepo.FindReturns(
					models.Route{GUID: "route-guid",
						Options: map[string]string{"a": "b"}},
					nil,
				)
				err := flagContext.Parse("domain-name", "--remove-option", "loadbalancing")
				Expect(err).NotTo(HaveOccurred())
				cmd.Requirements(factory, flagContext)
			})

			It("gives an error message that the given route option does not exist", func() {
				routeRepo.FindReturns(models.Route{GUID: "route-guid"}, nil)
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"No route option loadbalancing", "for"},
				))
			})
		})

		Context("when the route can be found and multiple options are passed", func() {
			BeforeEach(func() {
				routeRepo.FindReturns(
					models.Route{GUID: "route-guid",
						Options: map[string]string{"a": "b", "c": "d"}},
					nil,
				)
				err := flagContext.Parse("domain-name", "--option", "a=b", "--option", "c=d")
				Expect(err).NotTo(HaveOccurred())
				cmd.Requirements(factory, flagContext)
			})

			It("tries to set the given route options", func() {
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Setting route option a for domain-name to b"},
					[]string{"Setting route option c for domain-name to d"},
				))
			})
		})

		Context("when the route can be found and multiple remove options are passed", func() {
			BeforeEach(func() {
				routeRepo.FindReturns(
					models.Route{GUID: "route-guid",
						Options: map[string]string{"a": "b", "c": "d"}},
					nil,
				)
				err := flagContext.Parse("domain-name", "--remove-option", "a", "--remove-option", "c")
				Expect(err).NotTo(HaveOccurred())
				cmd.Requirements(factory, flagContext)
			})

			It("tries to set the given route options", func() {
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Removing route option a"},
					[]string{"Removing route option c"},
				))
			})
		})

		Context("when the route cannot be found", func() {
			BeforeEach(func() {
				err := flagContext.Parse("domain-name")
				Expect(err).NotTo(HaveOccurred())
				cmd.Requirements(factory, flagContext)
				routeRepo.FindReturns(models.Route{}, errors.New("find-by-host-and-domain-err"))
			})

			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("find-by-host-and-domain-err"))
			})

			It("tells the user a route with the given domain does not exist", func() {
				Expect(err).To(HaveOccurred())
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Route with domain", "does not exist"},
				))
			})
		})
	})
})
