package route_test

import (
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/commands/route"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/flags"

	fakeapi "github.com/cloudfoundry/cli/cf/api/fakes"
	fakerequirements "github.com/cloudfoundry/cli/cf/requirements/fakes"

	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DeleteRoute", func() {
	var (
		ui         *testterm.FakeUI
		configRepo core_config.Repository
		routeRepo  *fakeapi.FakeRouteRepository

		cmd         command_registry.Command
		deps        command_registry.Dependency
		factory     *fakerequirements.FakeFactory
		flagContext flags.FlagContext

		loginRequirement  requirements.Requirement
		domainRequirement *fakerequirements.FakeDomainRequirement

		fakeDomain models.DomainFields
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		ui.InputsChan = make(chan string)

		configRepo = testconfig.NewRepositoryWithDefaults()
		routeRepo = &fakeapi.FakeRouteRepository{}
		repoLocator := deps.RepoLocator.SetRouteRepository(routeRepo)

		deps = command_registry.Dependency{
			Ui:          ui,
			Config:      configRepo,
			RepoLocator: repoLocator,
		}

		cmd = &route.DeleteRoute{}
		cmd.SetDependency(deps, false)

		flagContext = flags.NewFlagContext(cmd.MetaData().Flags)

		factory = &fakerequirements.FakeFactory{}

		loginRequirement = &passingRequirement{Name: "login-requirement"}
		factory.NewLoginRequirementReturns(loginRequirement)

		domainRequirement = &fakerequirements.FakeDomainRequirement{}
		factory.NewDomainRequirementReturns(domainRequirement)

		fakeDomain = models.DomainFields{
			Guid: "fake-domain-guid",
			Name: "fake-domain-name",
		}
		domainRequirement.GetDomainReturns(fakeDomain)
	})

	Describe("Requirements", func() {
		Context("when not provided exactly one arg", func() {
			BeforeEach(func() {
				flagContext.Parse("app-name", "extra-arg")
			})

			It("fails with usage", func() {
				Expect(func() { cmd.Requirements(factory, flagContext) }).To(Panic())
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"Incorrect Usage. Requires an argument"},
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
		BeforeEach(func() {
			err := flagContext.Parse("domain-name")
			Expect(err).NotTo(HaveOccurred())
			_, err = cmd.Requirements(factory, flagContext)
			Expect(err).NotTo(HaveOccurred())
		})

		It("asks the user if they would like to proceed", func() {
			go cmd.Execute(flagContext)
			Eventually(func() []string { return ui.Prompts }).Should(ContainSubstrings(
				[]string{"Really delete the route"},
			))
		})

		It("tries to delete the route when the response is to proceed", func() {
			go cmd.Execute(flagContext)
			ui.InputsChan <- "y"
			Eventually(routeRepo.DeleteCallCount()).Should(Equal(1))
		})

		It("does not try to delete the route when the response is not to proceed", func() {
			go cmd.Execute(flagContext)
			ui.InputsChan <- "n"
			Consistently(routeRepo.DeleteCallCount()).Should(BeZero())
		})

		Context("when force is set", func() {
			BeforeEach(func() {
				err := flagContext.Parse("domain-name", "-f")
				Expect(err).NotTo(HaveOccurred())
			})

			It("does not ask the user if they would like to proceed", func() {
				cmd.Execute(flagContext)
				Consistently(func() []string { return ui.Prompts }).ShouldNot(ContainSubstrings(
					[]string{"Really delete the route"},
				))
			})

			It("tries to find the route", func() {
				cmd.Execute(flagContext)
				Eventually(routeRepo.FindCallCount()).Should(Equal(1))
				host, domain, path := routeRepo.FindArgsForCall(0)
				Expect(host).To(Equal(""))
				Expect(path).To(Equal(""))
				Expect(domain).To(Equal(fakeDomain))
			})

			Context("when there is an error finding the route	", func() {
				BeforeEach(func() {
					routeRepo.FindReturns(models.Route{}, errors.New("find-err"))
				})

				It("fails with error", func() {
					Expect(func() { cmd.Execute(flagContext) }).To(Panic())
					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"FAILED"},
						[]string{"find-err"},
					))
				})

				It("does not try to delete the route", func() {
					Expect(func() { cmd.Execute(flagContext) }).To(Panic())
					Expect(routeRepo.DeleteCallCount()).To(BeZero())
				})
			})

			Context("when there is a ModelNotFoundError when finding the route	", func() {
				BeforeEach(func() {
					routeRepo.FindReturns(models.Route{}, errors.NewModelNotFoundError("model-type", "model-name"))
				})

				It("tells the user that it could not delete the route", func() {
					cmd.Execute(flagContext)
					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Unable to delete, route", "does not exist"},
					))
				})

				It("does not try to delete the route", func() {
					cmd.Execute(flagContext)
					Expect(routeRepo.DeleteCallCount()).To(BeZero())
				})
			})

			Context("when the route can be found", func() {
				BeforeEach(func() {
					routeRepo.FindReturns(models.Route{
						Guid: "route-guid",
					}, nil)
				})

				It("tries to delete the route", func() {
					cmd.Execute(flagContext)
					Expect(routeRepo.DeleteCallCount()).To(Equal(1))
					Expect(routeRepo.DeleteArgsForCall(0)).To(Equal("route-guid"))
				})

				Context("when deleting the route fails", func() {
					BeforeEach(func() {
						routeRepo.DeleteReturns(errors.New("delete-err"))
					})

					It("fails with error", func() {
						Expect(func() { cmd.Execute(flagContext) }).To(Panic())
						Expect(ui.Outputs).To(ContainSubstrings(
							[]string{"FAILED"},
							[]string{"delete-err"},
						))
					})
				})

				Context("when deleting the route succeeds", func() {
					BeforeEach(func() {
						routeRepo.DeleteReturns(nil)
					})

					It("tells the user that it succeeded", func() {
						cmd.Execute(flagContext)
						Expect(ui.Outputs).To(ContainSubstrings(
							[]string{"OK"},
						))
					})
				})
			})
		})
	})
})
