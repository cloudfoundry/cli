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

package serviceauthtoken_test

import (
	. "cf/commands/serviceauthtoken"
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

func callCreateServiceAuthToken(args []string, reqFactory *testreq.FakeReqFactory, authTokenRepo *testapi.FakeAuthTokenRepo) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)

	config := testconfig.NewRepositoryWithDefaults()

	cmd := NewCreateServiceAuthToken(ui, config, authTokenRepo)
	ctxt := testcmd.NewContext("create-service-auth-token", args)

	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}

var _ = Describe("Testing with ginkgo", func() {
	It("TestCreateServiceAuthTokenFailsWithUsage", func() {
		authTokenRepo := &testapi.FakeAuthTokenRepo{}
		reqFactory := &testreq.FakeReqFactory{}

		ui := callCreateServiceAuthToken([]string{}, reqFactory, authTokenRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callCreateServiceAuthToken([]string{"arg1"}, reqFactory, authTokenRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callCreateServiceAuthToken([]string{"arg1", "arg2"}, reqFactory, authTokenRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callCreateServiceAuthToken([]string{"arg1", "arg2", "arg3"}, reqFactory, authTokenRepo)
		Expect(ui.FailedWithUsage).To(BeFalse())
	})
	It("TestCreateServiceAuthTokenRequirements", func() {

		authTokenRepo := &testapi.FakeAuthTokenRepo{}
		reqFactory := &testreq.FakeReqFactory{}
		args := []string{"arg1", "arg2", "arg3"}

		reqFactory.LoginSuccess = true
		callCreateServiceAuthToken(args, reqFactory, authTokenRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())

		reqFactory.LoginSuccess = false
		callCreateServiceAuthToken(args, reqFactory, authTokenRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})
	It("TestCreateServiceAuthToken", func() {

		authTokenRepo := &testapi.FakeAuthTokenRepo{}
		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
		args := []string{"a label", "a provider", "a value"}

		ui := callCreateServiceAuthToken(args, reqFactory, authTokenRepo)
		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Creating service auth token as", "my-user"},
			{"OK"},
		})

		authToken := models.ServiceAuthTokenFields{}
		authToken.Label = "a label"
		authToken.Provider = "a provider"
		authToken.Token = "a value"
		Expect(authTokenRepo.CreatedServiceAuthTokenFields).To(Equal(authToken))
	})
})
