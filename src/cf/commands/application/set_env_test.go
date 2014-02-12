package application_test

import (
	"cf/api"
	. "cf/commands/application"
	"cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	mr "github.com/tjarratt/mr_t"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

var _ = Describe("set-env command", func() {
	It("TestSetEnvRequirements", func() {
		app := models.Application{}
		app.Name = "my-app"
		app.Guid = "my-app-guid"
		appRepo := &testapi.FakeApplicationRepository{}
		args := []string{"my-app", "DATABASE_URL", "mysql://example.com/my-db"}

		reqFactory := &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: true}
		callSetEnv(args, reqFactory, appRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())

		reqFactory = &testreq.FakeReqFactory{Application: app, LoginSuccess: false, TargetedSpaceSuccess: true}
		callSetEnv(args, reqFactory, appRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		testcmd.CommandDidPassRequirements = true

		reqFactory = &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: false}
		callSetEnv(args, reqFactory, appRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})

	It("TestRunWhenApplicationExists", func() {

		app := models.Application{}
		app.Name = "my-app"
		app.Guid = "my-app-guid"
		app.EnvironmentVars = map[string]string{"foo": "bar"}
		reqFactory := &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: true}
		appRepo := &testapi.FakeApplicationRepository{}

		args := []string{"my-app", "DATABASE_URL", "mysql://example.com/my-db"}
		ui := callSetEnv(args, reqFactory, appRepo)

		Expect(len(ui.Outputs)).To(Equal(3))
		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{
				"Setting env variable",
				"DATABASE_URL",
				"mysql://example.com/my-db",
				"my-app",
				"my-org",
				"my-space",
				"my-user",
			},
			{"OK"},
			{"TIP"},
		})

		Expect(reqFactory.ApplicationName).To(Equal("my-app"))
		Expect(appRepo.UpdateAppGuid).To(Equal(app.Guid))
		Expect(*appRepo.UpdateParams.EnvironmentVars).To(Equal(map[string]string{
			"DATABASE_URL": "mysql://example.com/my-db",
			"foo":          "bar",
		}))
	})

	It("TestSetEnvWhenItAlreadyExists", func() {

		app := models.Application{}
		app.Name = "my-app"
		app.Guid = "my-app-guid"
		app.EnvironmentVars = map[string]string{"DATABASE_URL": "mysql://example.com/my-db"}
		reqFactory := &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: true}
		appRepo := &testapi.FakeApplicationRepository{}

		args := []string{"my-app", "DATABASE_URL", "mysql://example2.com/my-db"}
		ui := callSetEnv(args, reqFactory, appRepo)

		Expect(len(ui.Outputs)).To(Equal(3))
		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{
				"Setting env variable",
				"DATABASE_URL",
				"mysql://example2.com/my-db",
				"my-app",
				"my-org",
				"my-space",
				"my-user",
			},
			{"OK"},
			{"TIP"},
		})

		Expect(reqFactory.ApplicationName).To(Equal("my-app"))
		Expect(appRepo.UpdateAppGuid).To(Equal(app.Guid))
		Expect(*appRepo.UpdateParams.EnvironmentVars).To(Equal(map[string]string{
			"DATABASE_URL": "mysql://example2.com/my-db",
		}))
	})

	It("TestRunWhenSettingTheEnvFails", func() {

		app := models.Application{}
		app.Name = "my-app"
		app.Guid = "my-app-guid"
		reqFactory := &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: true}
		appRepo := &testapi.FakeApplicationRepository{
			ReadApp:   app,
			UpdateErr: true,
		}

		args := []string{"does-not-exist", "DATABASE_URL", "mysql://example.com/my-db"}
		ui := callSetEnv(args, reqFactory, appRepo)

		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Setting env variable"},
			{"FAILED"},
			{"Error updating app."},
		})
	})

	It("TestSetEnvFailsWithUsage", func() {

		app := models.Application{}
		app.Name = "my-app"
		app.Guid = "my-app-guid"
		reqFactory := &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: true}
		appRepo := &testapi.FakeApplicationRepository{ReadApp: app}

		args := []string{"my-app", "DATABASE_URL", "..."}
		ui := callSetEnv(args, reqFactory, appRepo)
		Expect(ui.FailedWithUsage).To(BeFalse())

		args = []string{"my-app", "DATABASE_URL"}
		ui = callSetEnv(args, reqFactory, appRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		args = []string{"my-app"}
		ui = callSetEnv(args, reqFactory, appRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		args = []string{}
		ui = callSetEnv(args, reqFactory, appRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())
	})
})

func callSetEnv(args []string, reqFactory *testreq.FakeReqFactory, appRepo api.ApplicationRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("set-env", args)
	configRepo := testconfig.NewRepositoryWithDefaults()
	cmd := NewSetEnv(ui, configRepo, appRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
