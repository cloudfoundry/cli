/*
                       WARNING WARNING WARNING

                Attention all potential contributors

   This testfile is not in the best state. We've been slowly transitioning
   from the built in "testing" package to using Ginkgo. As you can see, we've
   changed the format, but a lot of the setup, test body, descriptions, etc
   are either hardcoded, completely lacking, or misleading.

   For example:

   Describe("Testing with ginkgo"...)      // This is not a great description
   It("TestDoesSoemthing"...)              // This is a horrible description

   Describe("create-user command"...       // Describe the actual object under test
   It("creates a user when provided ..."   // this is more descriptive

   For good examples of writing Ginkgo tests for the cli, refer to

   src/cf/commands/application/delete_app_test.go
   src/cf/terminal/ui_test.go
   src/github.com/cloudfoundry/loggregator_consumer/consumer_test.go
*/

package application_test

import (
	. "cf/commands/application"
	"cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())
		Expect(reqFactory.ApplicationName).To(Equal("my-app"))

		reqFactory.LoginSuccess = false
		callEnv([]string{"my-app"}, reqFactory)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})
	It("TestEnvFailsWithUsage", func() {

		reqFactory := getEnvDependencies()
		ui := callEnv([]string{}, reqFactory)

		Expect(ui.FailedWithUsage).To(BeTrue())
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})
	It("TestEnvListsEnvironmentVariables", func() {

		reqFactory := getEnvDependencies()
		reqFactory.Application.EnvironmentVars = map[string]string{
			"my-key":  "my-value",
			"my-key2": "my-value2",
		}

		ui := callEnv([]string{"my-app"}, reqFactory)

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Getting env variables for app", "my-app", "my-org", "my-space", "my-user"},
			{"OK"},
			{"my-key", "my-value", "my-key2", "my-value2"},
		})
	})
	It("TestEnvShowsEmptyMessage", func() {

		reqFactory := getEnvDependencies()
		reqFactory.Application.EnvironmentVars = map[string]string{}

		ui := callEnv([]string{"my-app"}, reqFactory)

		testassert.SliceContains(ui.Outputs, testassert.Lines{
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
