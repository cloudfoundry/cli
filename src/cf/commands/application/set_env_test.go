package application_test

import (
	"cf/api"
	. "cf/commands/application"
	"cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"

	. "testhelpers/matchers"
)

var _ = Describe("set-env command", func() {
	var (
		app                 models.Application
		appRepo             *testapi.FakeApplicationRepository
		args                []string
		requirementsFactory *testreq.FakeReqFactory
	)

	BeforeEach(func() {
		app = models.Application{}
		app.Name = "my-app"
		app.Guid = "my-app-guid"
		appRepo = &testapi.FakeApplicationRepository{}
	})

	JustBeforeEach(func() {
		requirementsFactory = &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: true}
	})

	Describe("requirements", func() {
		BeforeEach(func() {
			args = []string{"my-app", "DATABASE_URL", "mysql://example.com/my-db"}
		})

		It("passes when all requirements are present", func() {
			requirementsFactory = &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: true}
			callSetEnv(args, requirementsFactory, appRepo)
			Expect(testcmd.CommandDidPassRequirements).To(BeTrue())
		})

		It("fails when login is not successful", func() {
			requirementsFactory = &testreq.FakeReqFactory{Application: app, LoginSuccess: false, TargetedSpaceSuccess: true}
			callSetEnv(args, requirementsFactory, appRepo)
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})

		It("fails when a space is not targeted", func() {
			requirementsFactory = &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: false}
			callSetEnv(args, requirementsFactory, appRepo)
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})
	})

	Context("setting an environment variable", func() {
		BeforeEach(func() {
			app.EnvironmentVars = map[string]string{"foo": "bar"}
			args = []string{"my-app", "DATABASE_URL", "mysql://new-example.com/my-db"}
		})

		Context("when it is new", func() {
			It("is created", func() {
				ui := callSetEnv(args, requirementsFactory, appRepo)

				Expect(len(ui.Outputs)).To(Equal(3))
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
				ui := callSetEnv(args, requirementsFactory, appRepo)

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

		XIt("allows the variable value to begin with a hyphen", func() {
			args = []string{"my-app", "MY_VAR", "--has-a-cool-value"}
			ui := callSetEnv(args, requirementsFactory, appRepo)

			Expect(len(ui.Outputs)).To(Equal(3))
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{
					"Setting env variable",
					"MY_VAR",
					"--has-a-cool-value",
				},
				[]string{"OK"},
				[]string{"TIP"},
			))

			Expect(appRepo.UpdateAppGuid).To(Equal(app.Guid))
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
				ui := callSetEnv(args, requirementsFactory, appRepo)

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Setting env variable"},
					[]string{"FAILED"},
					[]string{"Error updating app."},
				))
			})
		})
	})

	Describe("usage requirements", func() {
		Context("when an app, key, and value are all present", func() {
			BeforeEach(func() {
				args = []string{"my-app", "DATABASE_URL", "..."}
			})

			It("does not fail with usage", func() {
				ui := callSetEnv(args, requirementsFactory, appRepo)
				Expect(ui.FailedWithUsage).To(BeFalse())
			})
		})

		Context("when the value is missing", func() {
			BeforeEach(func() {
				args = []string{"my-app", "DATABASE_URL"}
			})

			It("fails with usage", func() {
				ui := callSetEnv(args, requirementsFactory, appRepo)
				Expect(ui.FailedWithUsage).To(BeTrue())
			})
		})

		Context("when the key and value are missing", func() {
			BeforeEach(func() {
				args = []string{"my-app"}
			})

			It("fails with usage", func() {
				ui := callSetEnv(args, requirementsFactory, appRepo)
				Expect(ui.FailedWithUsage).To(BeTrue())
			})
		})

		Context("when all parameters are missing", func() {
			BeforeEach(func() {
				args = []string{}
			})

			It("fails with usage", func() {
				ui := callSetEnv(args, requirementsFactory, appRepo)
				Expect(ui.FailedWithUsage).To(BeTrue())
			})
		})
	})
})

func callSetEnv(args []string, requirementsFactory *testreq.FakeReqFactory, appRepo api.ApplicationRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("set-env", args)
	configRepo := testconfig.NewRepositoryWithDefaults()
	cmd := NewSetEnv(ui, configRepo, appRepo)
	testcmd.RunCommand(cmd, ctxt, requirementsFactory)
	return
}
