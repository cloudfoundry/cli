package application_test

import (
	"cf/api"
	. "cf/commands/application"
	"cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

var _ = Describe("Testing with ginkgo", func() {
	It("TestUnsetEnvRequirements", func() {
		app := models.Application{}
		app.Name = "my-app"
		app.Guid = "my-app-guid"
		appRepo := &testapi.FakeApplicationRepository{}
		args := []string{"my-app", "DATABASE_URL"}

		reqFactory := &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: true}
		callUnsetEnv(args, reqFactory, appRepo)
		assert.True(mr.T(), testcmd.CommandDidPassRequirements)

		reqFactory = &testreq.FakeReqFactory{Application: app, LoginSuccess: false, TargetedSpaceSuccess: true}
		callUnsetEnv(args, reqFactory, appRepo)
		assert.False(mr.T(), testcmd.CommandDidPassRequirements)

		reqFactory = &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: false}
		callUnsetEnv(args, reqFactory, appRepo)
		assert.False(mr.T(), testcmd.CommandDidPassRequirements)
	})
	It("TestUnsetEnvWhenApplicationExists", func() {

		app := models.Application{}
		app.Name = "my-app"
		app.Guid = "my-app-guid"
		app.EnvironmentVars = map[string]string{"foo": "bar", "DATABASE_URL": "mysql://example.com/my-db"}
		reqFactory := &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: true}
		appRepo := &testapi.FakeApplicationRepository{}

		args := []string{"my-app", "DATABASE_URL"}
		ui := callUnsetEnv(args, reqFactory, appRepo)

		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Removing env variable", "DATABASE_URL", "my-app", "my-org", "my-space", "my-user"},
			{"OK"},
		})

		Expect(reqFactory.ApplicationName).To(Equal("my-app"))
		Expect(appRepo.UpdateAppGuid).To(Equal("my-app-guid"))
		Expect(*appRepo.UpdateParams.EnvironmentVars).To(Equal(map[string]string{
			"foo": "bar",
		}))
	})

	It("TestUnsetEnvWhenUnsettingTheEnvFails", func() {

		app := models.Application{}
		app.Name = "my-app"
		app.Guid = "my-app-guid"
		app.EnvironmentVars = map[string]string{"DATABASE_URL": "mysql://example.com/my-db"}
		reqFactory := &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: true}
		appRepo := &testapi.FakeApplicationRepository{
			ReadApp:   app,
			UpdateErr: true,
		}

		args := []string{"does-not-exist", "DATABASE_URL"}
		ui := callUnsetEnv(args, reqFactory, appRepo)

		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Removing env variable"},
			{"FAILED"},
			{"Error updating app."},
		})
	})
	It("TestUnsetEnvWhenEnvVarDoesNotExist", func() {

		app := models.Application{}
		app.Name = "my-app"
		app.Guid = "my-app-guid"
		reqFactory := &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: true}
		appRepo := &testapi.FakeApplicationRepository{}

		args := []string{"my-app", "DATABASE_URL"}
		ui := callUnsetEnv(args, reqFactory, appRepo)

		Expect(len(ui.Outputs)).To(Equal(3))
		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Removing env variable"},
			{"OK"},
			{"DATABASE_URL", "was not set."},
		})
	})
	It("TestUnsetEnvFailsWithUsage", func() {

		app := models.Application{}
		app.Name = "my-app"
		app.Guid = "my-app-guid"
		reqFactory := &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: true}
		appRepo := &testapi.FakeApplicationRepository{ReadApp: app}

		args := []string{"my-app", "DATABASE_URL"}
		ui := callUnsetEnv(args, reqFactory, appRepo)
		assert.False(mr.T(), ui.FailedWithUsage)

		args = []string{"my-app"}
		ui = callUnsetEnv(args, reqFactory, appRepo)
		assert.True(mr.T(), ui.FailedWithUsage)

		args = []string{}
		ui = callUnsetEnv(args, reqFactory, appRepo)
		assert.True(mr.T(), ui.FailedWithUsage)
	})
})

func callUnsetEnv(args []string, reqFactory *testreq.FakeReqFactory, appRepo api.ApplicationRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("unset-env", args)
	configRepo := testconfig.NewRepositoryWithDefaults()
	cmd := NewUnsetEnv(ui, configRepo, appRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
