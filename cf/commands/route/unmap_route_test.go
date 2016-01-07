package route_test

import (
	"errors"

	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/commands/route"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/simonleung8/flags"

	fakeapi "github.com/cloudfoundry/cli/cf/api/fakes"
	fakerequirements "github.com/cloudfoundry/cli/cf/requirements/fakes"

	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UnmapRoute", func() {
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

		cmd = &route.UnmapRoute{}
		cmd.SetDependency(deps, false)

		flagContext = flags.NewFlagContext(cmd.MetaData().Flags)

		factory = &fakerequirements.FakeFactory{}

		loginRequirement = &passingRequirement{Name: "login-requirement"}
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

	Describe("Requirements", func() {
		Context("when not provided exactly two args", func() {
			BeforeEach(func() {
				flagContext.Parse("app-name")
			})

			It("fails with usage", func() {
				Expect(func() { cmd.Requirements(factory, flagContext) }).To(Panic())
				Expect(ui.Outputs).To(ContainSubstrings(
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
		})
	})

	Describe("Execute", func() {
		BeforeEach(func() {
			err := flagContext.Parse("app-name", "domain-name")
			Expect(err).NotTo(HaveOccurred())
			_, err = cmd.Requirements(factory, flagContext)
			Expect(err).NotTo(HaveOccurred())
		})

		It("tries to find the host and domain in the route repo", func() {
			cmd.Execute(flagContext)
			Expect(routeRepo.FindByHostAndDomainCallCount()).To(Equal(1))
			hostname, domain := routeRepo.FindByHostAndDomainArgsForCall(0)
			Expect(hostname).To(Equal(""))
			Expect(domain).To(Equal(fakeDomain))
		})

		Context("when the route can be found", func() {
			BeforeEach(func() {
				routeRepo.FindByHostAndDomainReturns(models.Route{Guid: "route-guid"}, nil)
			})

			It("tells the user that it is removing the route", func() {
				cmd.Execute(flagContext)
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Removing route", "from app", "in org"},
				))
			})

			Context("when the returned route has an app with the requested app's guid", func() {
				BeforeEach(func() {
					route := models.Route{
						Guid: "route-guid",
						Apps: []models.ApplicationFields{
							{Guid: "fake-app-guid"},
						},
					}
					routeRepo.FindByHostAndDomainReturns(route, nil)
				})

				It("tries to unbind the route from the app", func() {
					cmd.Execute(flagContext)
					Expect(routeRepo.UnbindCallCount()).To(Equal(1))
					routeGUID, appGUID := routeRepo.UnbindArgsForCall(0)
					Expect(routeGUID).To(Equal("route-guid"))
					Expect(appGUID).To(Equal("fake-app-guid"))
				})

				Context("when unbinding the route from the app fails", func() {
					BeforeEach(func() {
						routeRepo.UnbindReturns(errors.New("unbind-err"))
					})

					It("panics and prints a failure message", func() {
						Expect(func() { cmd.Execute(flagContext) }).To(Panic())
						Expect(ui.Outputs).To(BeInDisplayOrder(
							[]string{"FAILED"},
							[]string{"unbind-err"},
						))
					})
				})

				Context("when unbinding the route from the app succeeds", func() {
					BeforeEach(func() {
						routeRepo.UnbindReturns(nil)
					})

					It("tells the user it succeeded", func() {
						cmd.Execute(flagContext)
						Expect(ui.Outputs).To(BeInDisplayOrder(
							[]string{"OK"},
						))
					})
				})
			})

			Context("when the returned route does not have an app with the requested app's guid", func() {
				BeforeEach(func() {
					route := models.Route{
						Guid: "route-guid",
						Apps: []models.ApplicationFields{
							{Guid: "other-fake-app-guid"},
						},
					}
					routeRepo.FindByHostAndDomainReturns(route, nil)
				})

				It("does not unbind the route from the app", func() {
					cmd.Execute(flagContext)
					Expect(routeRepo.UnbindCallCount()).To(Equal(0))
				})

				It("tells the user 'OK'", func() {
					cmd.Execute(flagContext)
					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"OK"},
					))
				})

				It("warns the user the route was not mapped to the application", func() {
					cmd.Execute(flagContext)
					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Route to be unmapped is not currently mapped to the application."},
					))
				})
			})
		})

		Context("when the route cannot be found", func() {
			BeforeEach(func() {
				routeRepo.FindByHostAndDomainReturns(models.Route{}, errors.New("find-by-host-and-domain-err"))
			})

			It("panics and prints a failure message", func() {
				Expect(func() { cmd.Execute(flagContext) }).To(Panic())
				Expect(ui.Outputs).To(BeInDisplayOrder(
					[]string{"FAILED"},
					[]string{"find-by-host-and-domain-err"},
				))
			})
		})
	})
})
