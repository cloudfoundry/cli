package application_test

import (
	"code.cloudfoundry.org/cli/cf/api/apifakes"
	"code.cloudfoundry.org/cli/cf/api/applications/applicationsfakes"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"
	testcmd "code.cloudfoundry.org/cli/util/testhelpers/commands"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
)

var _ = Describe("delete app command", func() {
	var (
		ui                  *testterm.FakeUI
		app                 models.Application
		configRepo          coreconfig.Repository
		appRepo             *applicationsfakes.FakeRepository
		routeRepo           *apifakes.FakeRouteRepository
		requirementsFactory *requirementsfakes.FakeFactory
		deps                commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.Config = configRepo
		deps.RepoLocator = deps.RepoLocator.SetApplicationRepository(appRepo)
		deps.RepoLocator = deps.RepoLocator.SetRouteRepository(routeRepo)
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("delete").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		app = models.Application{}
		app.Name = "app-to-delete"
		app.GUID = "app-to-delete-guid"

		ui = &testterm.FakeUI{}
		appRepo = new(applicationsfakes.FakeRepository)
		routeRepo = new(apifakes.FakeRouteRepository)
		requirementsFactory = new(requirementsfakes.FakeFactory)

		configRepo = testconfig.NewRepositoryWithDefaults()
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCLICommand("delete", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	It("fails requirements when not logged in", func() {
		requirementsFactory.NewUsageRequirementReturns(requirements.Passing{})
		requirementsFactory.NewLoginRequirementReturns(requirements.Failing{Message: "not logged in"})
		Expect(runCommand("-f", "delete-this-app-plz")).To(BeFalse())
	})

	It("fails if a space is not targeted", func() {
		requirementsFactory.NewUsageRequirementReturns(requirements.Passing{})
		requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
		requirementsFactory.NewTargetedSpaceRequirementReturns(requirements.Failing{Message: "not targeting space"})
		Expect(runCommand("-f", "delete-this-app-plz")).To(BeFalse())
	})

	It("fails with usage when not provided exactly one arg", func() {
		requirementsFactory.NewUsageRequirementReturns(requirements.Failing{})
		Expect(runCommand()).To(BeFalse())
	})

	Context("when passing requirements", func() {
		BeforeEach(func() {
			requirementsFactory.NewUsageRequirementReturns(requirements.Passing{})
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
			requirementsFactory.NewTargetedSpaceRequirementReturns(requirements.Passing{})
		})

		Context("When provided an app that exists", func() {
			BeforeEach(func() {
				appRepo.ReadReturns(app, nil)
			})

			It("deletes an app when the user confirms", func() {
				ui.Inputs = []string{"y"}

				runCommand("app-to-delete")

				Expect(appRepo.ReadArgsForCall(0)).To(Equal("app-to-delete"))
				Expect(appRepo.DeleteArgsForCall(0)).To(Equal("app-to-delete-guid"))

				Expect(ui.Prompts).To(ContainSubstrings([]string{"Really delete the app app-to-delete"}))

				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Deleting", "app-to-delete", "my-org", "my-space", "my-user"},
					[]string{"OK"},
				))
			})

			It("does not prompt when the -f flag is provided", func() {
				runCommand("-f", "app-to-delete")

				Expect(appRepo.ReadArgsForCall(0)).To(Equal("app-to-delete"))
				Expect(appRepo.DeleteArgsForCall(0)).To(Equal("app-to-delete-guid"))
				Expect(ui.Prompts).To(BeEmpty())

				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Deleting", "app-to-delete"},
					[]string{"OK"},
				))
			})

			Describe("mapped routes", func() {
				BeforeEach(func() {
					route1 := models.RouteSummary{}
					route1.GUID = "the-first-route-guid"
					route1.Host = "my-app-is-good.com"

					route2 := models.RouteSummary{}
					route2.GUID = "the-second-route-guid"
					route2.Host = "my-app-is-bad.com"

					appRepo.ReadReturns(models.Application{
						Routes: []models.RouteSummary{route1, route2},
					}, nil)
				})

				Context("when the -r flag is provided", func() {
					Context("when deleting routes succeeds", func() {
						It("deletes the app's routes", func() {
							runCommand("-f", "-r", "app-to-delete")

							Expect(routeRepo.DeleteCallCount()).To(Equal(2))
							Expect(routeRepo.DeleteArgsForCall(0)).To(Equal("the-first-route-guid"))
							Expect(routeRepo.DeleteArgsForCall(1)).To(Equal("the-second-route-guid"))
						})
					})

					Context("when deleting routes fails", func() {
						BeforeEach(func() {
							routeRepo.DeleteReturns(errors.New("an-error"))
						})

						It("fails with the api error message", func() {
							runCommand("-f", "-r", "app-to-delete")

							Expect(ui.Outputs()).To(ContainSubstrings(
								[]string{"Deleting", "app-to-delete"},
								[]string{"FAILED"},
							))
						})
					})
				})

				Context("when the -r flag is not provided", func() {
					It("does not delete mapped routes", func() {
						runCommand("-f", "app-to-delete")
						Expect(routeRepo.DeleteCallCount()).To(BeZero())
					})
				})
			})
		})

		Context("when the app provided is not found", func() {
			BeforeEach(func() {
				appRepo.ReadReturns(models.Application{}, errors.NewModelNotFoundError("App", "the-app"))
			})

			It("warns the user when the provided app does not exist", func() {
				runCommand("-f", "app-to-delete")

				Expect(appRepo.ReadArgsForCall(0)).To(Equal("app-to-delete"))
				Expect(appRepo.DeleteCallCount()).To(BeZero())

				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Deleting", "app-to-delete"},
					[]string{"OK"},
				))

				Expect(ui.WarnOutputs).To(ContainSubstrings([]string{"app-to-delete", "does not exist"}))
			})
		})
	})
})
