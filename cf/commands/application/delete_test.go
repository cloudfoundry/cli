package application_test

import (
	testApplication "github.com/cloudfoundry/cli/cf/api/applications/fakes"
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

var _ = Describe("delete app command", func() {
	var (
		ui                  *testterm.FakeUI
		app                 models.Application
		configRepo          core_config.Repository
		appRepo             *testApplication.FakeApplicationRepository
		routeRepo           *testapi.FakeRouteRepository
		requirementsFactory *testreq.FakeReqFactory
		deps                command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.Config = configRepo
		deps.RepoLocator = deps.RepoLocator.SetApplicationRepository(appRepo)
		deps.RepoLocator = deps.RepoLocator.SetRouteRepository(routeRepo)
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("delete").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		app = models.Application{}
		app.Name = "app-to-delete"
		app.Guid = "app-to-delete-guid"

		ui = &testterm.FakeUI{}
		appRepo = &testApplication.FakeApplicationRepository{}
		routeRepo = &testapi.FakeRouteRepository{}
		requirementsFactory = &testreq.FakeReqFactory{}

		configRepo = testconfig.NewRepositoryWithDefaults()
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCliCommand("delete", args, requirementsFactory, updateCommandDependency, false)
	}

	It("fails requirements when not logged in", func() {
		requirementsFactory.LoginSuccess = false
		Expect(runCommand("-f", "delete-this-app-plz")).To(BeFalse())
	})
	It("fails if a space is not targeted", func() {
		requirementsFactory.LoginSuccess = true
		requirementsFactory.TargetedSpaceSuccess = false
		Expect(runCommand("-f", "delete-this-app-plz")).To(BeFalse())
	})

	Context("when logged in", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedSpaceSuccess = true
		})

		It("fails with usage when not provided exactly one arg", func() {
			runCommand()
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires", "argument"},
			))
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

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Deleting", "app-to-delete", "my-org", "my-space", "my-user"},
					[]string{"OK"},
				))
			})

			It("does not prompt when the -f flag is provided", func() {
				runCommand("-f", "app-to-delete")

				Expect(appRepo.ReadArgsForCall(0)).To(Equal("app-to-delete"))
				Expect(appRepo.DeleteArgsForCall(0)).To(Equal("app-to-delete-guid"))
				Expect(ui.Prompts).To(BeEmpty())

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Deleting", "app-to-delete"},
					[]string{"OK"},
				))
			})

			Describe("mapped routes", func() {
				BeforeEach(func() {
					route1 := models.RouteSummary{}
					route1.Guid = "the-first-route-guid"
					route1.Host = "my-app-is-good.com"

					route2 := models.RouteSummary{}
					route2.Guid = "the-second-route-guid"
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

							Expect(ui.Outputs).To(ContainSubstrings(
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

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Deleting", "app-to-delete"},
					[]string{"OK"},
				))

				Expect(ui.WarnOutputs).To(ContainSubstrings([]string{"app-to-delete", "does not exist"}))
			})
		})
	})
})
