package application_test

import (
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/commands/application"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("set-env command", func() {
	var (
		ui                  *testterm.FakeUI
		configRepo          configuration.ReadWriter
		app                 models.Application
		appRepo             *testapi.FakeApplicationRepository
		requirementsFactory *testreq.FakeReqFactory
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		app = models.Application{}
		app.Name = "my-app"
		app.Guid = "my-app-guid"
		appRepo = &testapi.FakeApplicationRepository{}
		requirementsFactory = &testreq.FakeReqFactory{}
		configRepo = testconfig.NewRepositoryWithDefaults()
	})

	runCommand := func(args ...string) {
		testcmd.RunCommand(NewSetEnv(ui, configRepo, appRepo), args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("fails when login is not successful", func() {
			requirementsFactory.Application = app
			requirementsFactory.TargetedSpaceSuccess = true

			runCommand("hey", "gabba", "gabba")
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})

		It("fails when a space is not targeted", func() {
			requirementsFactory.Application = app
			requirementsFactory.LoginSuccess = true

			runCommand("hey", "gabba", "gabba")
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})

		It("fails with usage when not provided with exactly three args", func() {
			requirementsFactory.Application = app
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedSpaceSuccess = true

			runCommand("zomg", "too", "many", "args")
			Expect(ui.FailedWithUsage).To(BeTrue())
		})
	})

	Context("when logged in, a space is targeted and given enough args", func() {
		BeforeEach(func() {
			app.EnvironmentVars = map[string]string{"foo": "bar"}
			requirementsFactory.Application = app
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedSpaceSuccess = true
		})

		Context("when it is new", func() {
			It("is created", func() {
				runCommand("my-app", "DATABASE_URL", "mysql://new-example.com/my-db")

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{
						"Setting env variable",
						"DATABASE_URL",
						"mysql://new-example.com/my-db",
						"my-app",
						"my-org",
						"my-space",
						"my-user",
					},
					[]string{"OK"},
					[]string{"TIP"},
				))

				Expect(requirementsFactory.ApplicationName).To(Equal("my-app"))
				Expect(appRepo.UpdateAppGuid).To(Equal(app.Guid))
				Expect(*appRepo.UpdateParams.EnvironmentVars).To(Equal(map[string]string{
					"DATABASE_URL": "mysql://new-example.com/my-db",
					"foo":          "bar",
				}))
			})
		})

		Context("when it already exists", func() {
			BeforeEach(func() {
				app.EnvironmentVars["DATABASE_URL"] = "mysql://old-example.com/my-db"
			})

			It("is updated", func() {
				runCommand("my-app", "DATABASE_URL", "mysql://new-example.com/my-db")

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{
						"Setting env variable",
						"DATABASE_URL",
						"mysql://new-example.com/my-db",
						"my-app",
						"my-org",
						"my-space",
						"my-user",
					},
					[]string{"OK"},
					[]string{"TIP"},
				))
			})
		})

		It("allows the variable value to begin with a hyphen", func() {
			runCommand("my-app", "MY_VAR", "--has-a-cool-value")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{
					"Setting env variable",
					"MY_VAR",
					"--has-a-cool-value",
				},
				[]string{"OK"},
				[]string{"TIP"},
			))
			Expect(*appRepo.UpdateParams.EnvironmentVars).To(Equal(map[string]string{
				"MY_VAR": "--has-a-cool-value",
				"foo":    "bar",
			}))
		})

		Context("when setting fails", func() {
			BeforeEach(func() {
				appRepo.UpdateErr = true
			})

			It("tells the user", func() {
				runCommand("please", "dont", "fail")

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Setting env variable"},
					[]string{"FAILED"},
					[]string{"Error updating app."},
				))
			})
		})
	})
})
