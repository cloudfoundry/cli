package application_test

import (
	. "cf/commands/application"
	"cf/configuration"
	"cf/models"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

func callEnv(t mr.TestingT, args []string, reqFactory *testreq.FakeReqFactory) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}
	ctxt := testcmd.NewContext("env", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)
	org := models.OrganizationFields{}
	org.Name = "my-org"
	space := models.SpaceFields{}
	space.Name = "my-space"
	config := &configuration.Configuration{
		SpaceFields:        space,
		OrganizationFields: org,
		AccessToken:        token,
	}

	cmd := NewEnv(ui, config)
	testcmd.RunCommand(cmd, ctxt, reqFactory)

	return
}

func getEnvDependencies() (reqFactory *testreq.FakeReqFactory) {
	app := models.Application{}
	app.Name = "my-app"
	reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, Application: app}
	return
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestEnvRequirements", func() {
			reqFactory := getEnvDependencies()

			reqFactory.LoginSuccess = true
			callEnv(mr.T(), []string{"my-app"}, reqFactory)
			assert.True(mr.T(), testcmd.CommandDidPassRequirements)
			assert.Equal(mr.T(), reqFactory.ApplicationName, "my-app")

			reqFactory.LoginSuccess = false
			callEnv(mr.T(), []string{"my-app"}, reqFactory)
			assert.False(mr.T(), testcmd.CommandDidPassRequirements)
		})
		It("TestEnvFailsWithUsage", func() {

			reqFactory := getEnvDependencies()
			ui := callEnv(mr.T(), []string{}, reqFactory)

			assert.True(mr.T(), ui.FailedWithUsage)
			assert.False(mr.T(), testcmd.CommandDidPassRequirements)
		})
		It("TestEnvListsEnvironmentVariables", func() {

			reqFactory := getEnvDependencies()
			reqFactory.Application.EnvironmentVars = map[string]string{
				"my-key":  "my-value",
				"my-key2": "my-value2",
			}

			ui := callEnv(mr.T(), []string{"my-app"}, reqFactory)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Getting env variables for app", "my-app", "my-org", "my-space", "my-user"},
				{"OK"},
				{"my-key", "my-value", "my-key2", "my-value2"},
			})
		})
		It("TestEnvShowsEmptyMessage", func() {

			reqFactory := getEnvDependencies()
			reqFactory.Application.EnvironmentVars = map[string]string{}

			ui := callEnv(mr.T(), []string{"my-app"}, reqFactory)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Getting env variables for app", "my-app"},
				{"OK"},
				{"No env variables exist"},
			})
		})
	})
}
