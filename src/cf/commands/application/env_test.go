package application_test

import (
	. "cf/commands/application"
	"cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

var _ = Describe("Testing with ginkgo", func() {
	It("TestEnvRequirements", func() {
		reqFactory := getEnvDependencies()

		reqFactory.LoginSuccess = true
		callEnv([]string{"my-app"}, reqFactory)
		assert.True(mr.T(), testcmd.CommandDidPassRequirements)
		Expect(reqFactory.ApplicationName).To(Equal("my-app"))

		reqFactory.LoginSuccess = false
		callEnv([]string{"my-app"}, reqFactory)
		assert.False(mr.T(), testcmd.CommandDidPassRequirements)
	})
	It("TestEnvFailsWithUsage", func() {

		reqFactory := getEnvDependencies()
		ui := callEnv([]string{}, reqFactory)

		assert.True(mr.T(), ui.FailedWithUsage)
		assert.False(mr.T(), testcmd.CommandDidPassRequirements)
	})
	It("TestEnvListsEnvironmentVariables", func() {

		reqFactory := getEnvDependencies()
		reqFactory.Application.EnvironmentVars = map[string]string{
			"my-key":  "my-value",
			"my-key2": "my-value2",
		}

		ui := callEnv([]string{"my-app"}, reqFactory)

		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Getting env variables for app", "my-app", "my-org", "my-space", "my-user"},
			{"OK"},
			{"my-key", "my-value", "my-key2", "my-value2"},
		})
	})
	It("TestEnvShowsEmptyMessage", func() {

		reqFactory := getEnvDependencies()
		reqFactory.Application.EnvironmentVars = map[string]string{}

		ui := callEnv([]string{"my-app"}, reqFactory)

		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Getting env variables for app", "my-app"},
			{"OK"},
			{"No env variables exist"},
		})
	})
})

func callEnv(args []string, reqFactory *testreq.FakeReqFactory) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}
	ctxt := testcmd.NewContext("env", args)

	configRepo := testconfig.NewRepositoryWithDefaults()
	cmd := NewEnv(ui, configRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)

	return
}

func getEnvDependencies() (reqFactory *testreq.FakeReqFactory) {
	app := models.Application{}
	app.Name = "my-app"
	reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, Application: app}
	return
}
