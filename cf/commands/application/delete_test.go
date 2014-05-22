package application_test

import (
	. "github.com/cloudfoundry/cli/cf/commands/application"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	testapi "github.com/cloudfoundry/cli/testhelpers/api"
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

		configRepo = testconfig.NewRepositoryWithDefaults()
		cmd = NewDeleteApp(ui, configRepo, appRepo, routeRepo)
	})

	runCommand := func(args ...string) {
		testcmd.RunCommand(cmd, args, requirementsFactory)
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

				Expect(ui.Prompts).To(ContainSubstrings([]string{"Really delete the app app-to-delete"}))

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Deleting", "app-to-delete", "my-org", "my-space", "my-user"},
					[]string{"OK"},
				))
			})

			It("does not prompt when the -f flag is provided", func() {
				runCommand("-f", "app-to-delete")

				Expect(appRepo.ReadArgs.Name).To(Equal("app-to-delete"))
				Expect(appRepo.DeletedAppGuid).To(Equal("app-to-delete-guid"))
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
						Expect(routeRepo.DeletedRouteGuids).To(BeEmpty())
					})
				})
			})
		})

		Context("when the app provided is not found", func() {
			BeforeEach(func() {
				appRepo.ReadReturns.Error = errors.NewModelNotFoundError("App", "the-app")
			})

			It("warns the user when the provided app does not exist", func() {
				runCommand("-f", "app-to-delete")

				Expect(appRepo.ReadArgs.Name).To(Equal("app-to-delete"))
				Expect(appRepo.DeletedAppGuid).To(Equal(""))

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Deleting", "app-to-delete"},
					[]string{"OK"},
				))

				Expect(ui.WarnOutputs).To(ContainSubstrings([]string{"app-to-delete", "does not exist"}))
			})
		})
	})
})
