package route_test

import (
	"errors"

	"github.com/simonleung8/flags"

	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/commands/route"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"

	fakeapi "github.com/cloudfoundry/cli/cf/api/fakes"
	fakeroute "github.com/cloudfoundry/cli/cf/commands/route/fakes"
	fakerequirements "github.com/cloudfoundry/cli/cf/requirements/fakes"

	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MapRoute", func() {
	var (
		ui         *testterm.FakeUI
		configRepo core_config.Repository
		routeRepo  *fakeapi.FakeRouteRepository

		cmd         command_registry.Command
		deps        command_registry.Dependency
		factory     *fakerequirements.FakeFactory
		flagContext flags.FlagContext

		loginRequirement       requirements.Requirement
		applicationRequirement *fakerequirements.FakeApplicationRequirement
		domainRequirement      *fakerequirements.FakeDomainRequirement

		originalCreateRouteCmd command_registry.Command
		fakeCreateRouteCmd     command_registry.Command

		fakeDomain models.DomainFields
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		routeRepo = &fakeapi.FakeRouteRepository{}
		repoLocator := deps.RepoLocator.SetRouteRepository(routeRepo)

		deps = command_registry.Dependency{
			Ui:          ui,
			Config:      configRepo,
			RepoLocator: repoLocator,
		}

		originalCreateRouteCmd = command_registry.Commands.FindCommand("create-route")
		fakeCreateRouteCmd = &fakeroute.FakeRouteCreator{}
		command_registry.Register(fakeCreateRouteCmd)

		cmd = &route.MapRoute{}
		cmd.SetDependency(deps, false)

		flagContext = flags.NewFlagContext(cmd.MetaData().Flags)

		factory = &fakerequirements.FakeFactory{}

		loginRequirement = &passingRequirement{}
		factory.NewLoginRequirementReturns(loginRequirement)

		applicationRequirement = &fakerequirements.FakeApplicationRequirement{}
		factory.NewApplicationRequirementReturns(applicationRequirement)

		fakeApplication := models.Application{}
		fakeApplication.Guid = "fake-app-guid"
		applicationRequirement.GetApplicationReturns(fakeApplication)

		domainRequirement = &fakerequirements.FakeDomainRequirement{}
		factory.NewDomainRequirementReturns(domainRequirement)

		fakeDomain = models.DomainFields{
			Guid: "fake-domain-guid",
			Name: "fake-domain-name",
		}
		domainRequirement.GetDomainReturns(fakeDomain)
	})

	AfterEach(func() {
		command_registry.Register(originalCreateRouteCmd)
	})

	Describe("Requirements", func() {
		Context("when not provided exactly two args", func() {
			BeforeEach(func() {
				flagContext.Parse("app-name")
			})

			It("fails with usage", func() {
				Expect(func() { cmd.Requirements(factory, flagContext) }).To(Panic())
				Expect(ui.Outputs).To(ContainSubstrings(
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
		})
	})

	Describe("Execute", func() {
		BeforeEach(func() {
			err := flagContext.Parse("app-name", "domain-name")
			Expect(err).NotTo(HaveOccurred())
			_, err = cmd.Requirements(factory, flagContext)
			Expect(err).NotTo(HaveOccurred())
		})

		It("tries to create the route", func() {
			cmd.Execute(flagContext)
			fakeRouteCreator, ok := fakeCreateRouteCmd.(*fakeroute.FakeRouteCreator)
			Expect(ok).To(BeTrue())

			Expect(fakeRouteCreator.CreateRouteCallCount()).To(Equal(1))
			_, path, domain, space := fakeRouteCreator.CreateRouteArgsForCall(0)
			Expect(path).To(Equal(""))
			Expect(domain).To(Equal(fakeDomain))
			Expect(space).To(Equal(models.SpaceFields{
				Name: "my-space",
				Guid: "my-space-guid",
			}))
		})

		Context("when creating the route fails", func() {
			BeforeEach(func() {
				fakeRouteCreator, ok := fakeCreateRouteCmd.(*fakeroute.FakeRouteCreator)
				Expect(ok).To(BeTrue())
				fakeRouteCreator.CreateRouteReturns(models.Route{}, errors.New("create-route-err"))
			})

			It("panics and prints a failure message", func() {
				Expect(func() { cmd.Execute(flagContext) }).To(Panic())
				Expect(ui.Outputs).To(BeInDisplayOrder(
					[]string{"FAILED"},
					[]string{"create-route-err"},
				))
			})
		})

		Context("when creating the route succeeds", func() {
			BeforeEach(func() {
				fakeRouteCreator, ok := fakeCreateRouteCmd.(*fakeroute.FakeRouteCreator)
				Expect(ok).To(BeTrue())
				fakeRouteCreator.CreateRouteReturns(models.Route{Guid: "fake-route-guid"}, nil)
			})

			It("tells the user that it is adding the route", func() {
				cmd.Execute(flagContext)
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Adding route", "to app", "in org"},
				))
			})

			It("tries to bind the route", func() {
				cmd.Execute(flagContext)
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
					cmd.Execute(flagContext)
					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"OK"},
					))
				})
			})

			Context("when binding the route fails", func() {
				BeforeEach(func() {
					routeRepo.BindReturns(errors.New("bind-error"))
				})

				It("panics and prints a failure message", func() {
					Expect(func() { cmd.Execute(flagContext) }).To(Panic())
					Expect(ui.Outputs).To(BeInDisplayOrder(
						[]string{"FAILED"},
						[]string{"bind-error"},
					))
				})
			})
		})

		Context("when a hostname is passed", func() {
			BeforeEach(func() {
				err := flagContext.Parse("app-name", "domain-name", "-n", "the-hostname")
				Expect(err).NotTo(HaveOccurred())
				_, err = cmd.Requirements(factory, flagContext)
				Expect(err).NotTo(HaveOccurred())
			})

			It("tries to create the route with the hostname", func() {
				cmd.Execute(flagContext)
				fakeRouteCreator, ok := fakeCreateRouteCmd.(*fakeroute.FakeRouteCreator)
				Expect(ok).To(BeTrue())
				Expect(fakeRouteCreator.CreateRouteCallCount()).To(Equal(1))
				hostName, _, _, _ := fakeRouteCreator.CreateRouteArgsForCall(0)
				Expect(hostName).To(Equal("the-hostname"))
			})
		})

		Context("when a hostname is not passed", func() {
			BeforeEach(func() {
				err := flagContext.Parse("app-name", "domain-name")
				Expect(err).NotTo(HaveOccurred())
				_, err = cmd.Requirements(factory, flagContext)
				Expect(err).NotTo(HaveOccurred())
			})

			It("tries to create the route without a hostname", func() {
				cmd.Execute(flagContext)
				fakeRouteCreator, ok := fakeCreateRouteCmd.(*fakeroute.FakeRouteCreator)
				Expect(ok).To(BeTrue())
				Expect(fakeRouteCreator.CreateRouteCallCount()).To(Equal(1))
				hostName, _, _, _ := fakeRouteCreator.CreateRouteArgsForCall(0)
				Expect(hostName).To(Equal(""))
			})
		})
	})
})
