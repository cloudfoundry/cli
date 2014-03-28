package application_test

import (
	. "cf/commands/application"
	"cf/configuration"
	"cf/errors"
	"cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

var _ = Describe("delete app command", func() {
	var (
		cmd                 *DeleteApp
		ui                  *testterm.FakeUI
		app                 models.Application
		configRepo          configuration.ReadWriter
		appRepo             *testapi.FakeApplicationRepository
		routeRepo           *testapi.FakeRouteRepository
		requirementsFactory *testreq.FakeReqFactory
	)

	BeforeEach(func() {
		app = models.Application{}
		app.Name = "app-to-delete"
		app.Guid = "app-to-delete-guid"

		ui = &testterm.FakeUI{}
		appRepo = &testapi.FakeApplicationRepository{}
		routeRepo = &testapi.FakeRouteRepository{}
		requirementsFactory = &testreq.FakeReqFactory{}

		ui = &testterm.FakeUI{}

		configRepo = testconfig.NewRepositoryWithDefaults()
		cmd = NewDeleteApp(ui, configRepo, appRepo, routeRepo)
	})

	var runCommand = func(args ...string) {
		testcmd.RunCommand(cmd, testcmd.NewContext("delete", args), requirementsFactory)
	}

	It("fails requirements when not logged in", func() {
		requirementsFactory.LoginSuccess = false
		runCommand("-f", "delete-this-app-plz")
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})

	Context("when logged in", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
		})

		It("provides the user usage text when no app name is given", func() {
			runCommand()
			Expect(ui.FailedWithUsage).To(BeTrue())
		})

		Context("When provided an app that exists", func() {
			BeforeEach(func() {
				appRepo.ReadReturns.App = app
			})

			It("deletes an app when the user confirms", func() {
				ui.Inputs = []string{"y"}

				runCommand("app-to-delete")

				Expect(appRepo.ReadArgs.Name).To(Equal("app-to-delete"))
				Expect(appRepo.DeletedAppGuid).To(Equal("app-to-delete-guid"))

				testassert.SliceContains(ui.Prompts, testassert.Lines{
					{"Really delete"},
				})
				testassert.SliceContains(ui.Outputs, testassert.Lines{
					{"Deleting", "app-to-delete", "my-org", "my-space", "my-user"},
					{"OK"},
				})
			})

			It("does not prompt when the -f flag is provided", func() {
				runCommand("-f", "app-to-delete")

				Expect(appRepo.ReadArgs.Name).To(Equal("app-to-delete"))
				Expect(appRepo.DeletedAppGuid).To(Equal("app-to-delete-guid"))
				Expect(ui.Prompts).To(BeEmpty())

				testassert.SliceContains(ui.Outputs, testassert.Lines{
					{"Deleting", "app-to-delete"},
					{"OK"},
				})
			})

			Describe("mapped routes", func() {
				BeforeEach(func() {
					route1 := models.RouteSummary{}
					route1.Guid = "the-first-route-guid"
					route1.Host = "my-app-is-good.com"

					route2 := models.RouteSummary{}
					route2.Guid = "the-second-route-guid"
					route2.Host = "my-app-is-bad.com"

					appRepo.ReadReturns.App = models.Application{
						Routes: []models.RouteSummary{route1, route2},
					}
				})

				Context("when the -r flag is provided", func() {
					Context("when deleting routes succeeds", func() {
						It("deletes the app's routes", func() {
							runCommand("-f", "-r", "app-to-delete")

							Expect(routeRepo.DeletedRouteGuids).To(ContainElement("the-first-route-guid"))
							Expect(routeRepo.DeletedRouteGuids).To(ContainElement("the-second-route-guid"))
						})
					})

					Context("when deleting routes fails", func() {
						BeforeEach(func() {
							routeRepo.DeleteErr = errors.New("badness")
						})

						It("fails with the api error message", func() {
							runCommand("-f", "-r", "app-to-delete")

							testassert.SliceContains(ui.Outputs, testassert.Lines{
								{"Deleting", "app-to-delete"},
								{"FAILED"},
							})
						})
					})
				})

				Context("when the -r flag is not provided", func() {
					It("does not delete mapped routes", func() {
						runCommand("-f", "app-to-delete")
						Expect(routeRepo.DeletedRouteGuids).To(BeEmpty())
					})
				})
			})
		})

		Context("when the app provided is not found", func() {
			BeforeEach(func() {
				appRepo.ReadReturns.Error = errors.NewModelNotFoundError("App", "the-app")
			})

			It("tells the user when the provided app does not exist", func() {
				runCommand("-f", "app-to-delete")

				Expect(appRepo.ReadArgs.Name).To(Equal("app-to-delete"))
				Expect(appRepo.DeletedAppGuid).To(Equal(""))

				testassert.SliceContains(ui.Outputs, testassert.Lines{
					{"Deleting", "app-to-delete"},
					{"OK"},
					{"app-to-delete", "does not exist"},
				})
			})
		})
	})
})
