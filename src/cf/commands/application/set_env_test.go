package application_test

import (
	"cf"
	"cf/api"
	. "cf/commands/application"
	"cf/configuration"
	"generic"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

func callSetEnv(t mr.TestingT, args []string, reqFactory *testreq.FakeReqFactory, appRepo api.ApplicationRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("set-env", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)
	space := cf.SpaceFields{}
	space.Name = "my-space"
	org := cf.OrganizationFields{}
	org.Name = "my-org"
	config := &configuration.Configuration{
		SpaceFields:        space,
		OrganizationFields: org,
		AccessToken:        token,
	}

	cmd := NewSetEnv(ui, config, appRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestSetEnvRequirements", func() {
			app := cf.Application{}
			app.Name = "my-app"
			app.Guid = "my-app-guid"
			appRepo := &testapi.FakeApplicationRepository{}
			args := []string{"my-app", "DATABASE_URL", "mysql://example.com/my-db"}

			reqFactory := &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: true}
			callSetEnv(mr.T(), args, reqFactory, appRepo)
			assert.True(mr.T(), testcmd.CommandDidPassRequirements)

			reqFactory = &testreq.FakeReqFactory{Application: app, LoginSuccess: false, TargetedSpaceSuccess: true}
			callSetEnv(mr.T(), args, reqFactory, appRepo)
			assert.False(mr.T(), testcmd.CommandDidPassRequirements)

			testcmd.CommandDidPassRequirements = true

			reqFactory = &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: false}
			callSetEnv(mr.T(), args, reqFactory, appRepo)
			assert.False(mr.T(), testcmd.CommandDidPassRequirements)
		})
		It("TestRunWhenApplicationExists", func() {

			app := cf.Application{}
			app.Name = "my-app"
			app.Guid = "my-app-guid"
			app.EnvironmentVars = map[string]string{"foo": "bar"}
			reqFactory := &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: true}
			appRepo := &testapi.FakeApplicationRepository{}

			args := []string{"my-app", "DATABASE_URL", "mysql://example.com/my-db"}
			ui := callSetEnv(mr.T(), args, reqFactory, appRepo)

			assert.Equal(mr.T(), len(ui.Outputs), 3)
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

			assert.Equal(mr.T(), reqFactory.ApplicationName, "my-app")
			assert.Equal(mr.T(), appRepo.UpdateAppGuid, app.Guid)

			envParams := appRepo.UpdateParams.Get("env").(generic.Map)
			assert.Equal(mr.T(), envParams.Get("DATABASE_URL").(string), "mysql://example.com/my-db")
			assert.Equal(mr.T(), envParams.Get("foo").(string), "bar")
		})
		It("TestSetEnvWhenItAlreadyExists", func() {

			app := cf.Application{}
			app.Name = "my-app"
			app.Guid = "my-app-guid"
			app.EnvironmentVars = map[string]string{"DATABASE_URL": "mysql://example.com/my-db"}
			reqFactory := &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: true}
			appRepo := &testapi.FakeApplicationRepository{}

			args := []string{"my-app", "DATABASE_URL", "mysql://example2.com/my-db"}
			ui := callSetEnv(mr.T(), args, reqFactory, appRepo)

			assert.Equal(mr.T(), len(ui.Outputs), 3)
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

			assert.Equal(mr.T(), reqFactory.ApplicationName, "my-app")
			assert.Equal(mr.T(), appRepo.UpdateAppGuid, app.Guid)

			envParams := appRepo.UpdateParams.Get("env").(generic.Map)
			assert.Equal(mr.T(), envParams.Get("DATABASE_URL").(string), "mysql://example2.com/my-db")
		})
		It("TestRunWhenSettingTheEnvFails", func() {

			app := cf.Application{}
			app.Name = "my-app"
			app.Guid = "my-app-guid"
			reqFactory := &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: true}
			appRepo := &testapi.FakeApplicationRepository{
				ReadApp:   app,
				UpdateErr: true,
			}

			args := []string{"does-not-exist", "DATABASE_URL", "mysql://example.com/my-db"}
			ui := callSetEnv(mr.T(), args, reqFactory, appRepo)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Setting env variable"},
				{"FAILED"},
				{"Error updating app."},
			})
		})
		It("TestSetEnvFailsWithUsage", func() {

			app := cf.Application{}
			app.Name = "my-app"
			app.Guid = "my-app-guid"
			reqFactory := &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: true}
			appRepo := &testapi.FakeApplicationRepository{ReadApp: app}

			args := []string{"my-app", "DATABASE_URL", "..."}
			ui := callSetEnv(mr.T(), args, reqFactory, appRepo)
			assert.False(mr.T(), ui.FailedWithUsage)

			args = []string{"my-app", "DATABASE_URL"}
			ui = callSetEnv(mr.T(), args, reqFactory, appRepo)
			assert.True(mr.T(), ui.FailedWithUsage)

			args = []string{"my-app"}
			ui = callSetEnv(mr.T(), args, reqFactory, appRepo)
			assert.True(mr.T(), ui.FailedWithUsage)

			args = []string{}
			ui = callSetEnv(mr.T(), args, reqFactory, appRepo)
			assert.True(mr.T(), ui.FailedWithUsage)
		})
	})
}
