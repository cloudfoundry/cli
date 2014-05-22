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

   src/github.com/cloudfoundry/cli/cf/commands/application/delete_app_test.go
   src/github.com/cloudfoundry/cli/cf/terminal/ui_test.go
   src/github.com/cloudfoundry/loggregator_consumer/consumer_test.go
*/

package serviceauthtoken_test

import (
	. "github.com/cloudfoundry/cli/cf/commands/serviceauthtoken"
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

func callCreateServiceAuthToken(args []string, requirementsFactory *testreq.FakeReqFactory, authTokenRepo *testapi.FakeAuthTokenRepo) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)

	config := testconfig.NewRepositoryWithDefaults()

	cmd := NewCreateServiceAuthToken(ui, config, authTokenRepo)
	testcmd.RunCommand(cmd, args, requirementsFactory)
	return
}

var _ = Describe("Testing with ginkgo", func() {
	It("TestCreateServiceAuthTokenFailsWithUsage", func() {
		authTokenRepo := &testapi.FakeAuthTokenRepo{}
		requirementsFactory := &testreq.FakeReqFactory{}

		ui := callCreateServiceAuthToken([]string{}, requirementsFactory, authTokenRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callCreateServiceAuthToken([]string{"arg1"}, requirementsFactory, authTokenRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callCreateServiceAuthToken([]string{"arg1", "arg2"}, requirementsFactory, authTokenRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callCreateServiceAuthToken([]string{"arg1", "arg2", "arg3"}, requirementsFactory, authTokenRepo)
		Expect(ui.FailedWithUsage).To(BeFalse())
	})

	It("TestCreateServiceAuthTokenRequirements", func() {
		authTokenRepo := &testapi.FakeAuthTokenRepo{}
		requirementsFactory := &testreq.FakeReqFactory{}
		args := []string{"arg1", "arg2", "arg3"}

		requirementsFactory.LoginSuccess = true
		callCreateServiceAuthToken(args, requirementsFactory, authTokenRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())

		requirementsFactory.LoginSuccess = false
		callCreateServiceAuthToken(args, requirementsFactory, authTokenRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})

	It("TestCreateServiceAuthToken", func() {
		authTokenRepo := &testapi.FakeAuthTokenRepo{}
		requirementsFactory := &testreq.FakeReqFactory{LoginSuccess: true}
		args := []string{"a label", "a provider", "a value"}

		ui := callCreateServiceAuthToken(args, requirementsFactory, authTokenRepo)
		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Creating service auth token as", "my-user"},
			[]string{"OK"},
		))

		authToken := models.ServiceAuthTokenFields{}
		authToken.Label = "a label"
		authToken.Provider = "a provider"
		authToken.Token = "a value"
		Expect(authTokenRepo.CreatedServiceAuthTokenFields).To(Equal(authToken))
	})
})
