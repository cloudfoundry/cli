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
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func callUpdateServiceAuthToken(args []string, requirementsFactory *testreq.FakeReqFactory, authTokenRepo *testapi.FakeAuthTokenRepo) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)

	config := testconfig.NewRepositoryWithDefaults()

	cmd := NewUpdateServiceAuthToken(ui, config, authTokenRepo)
	ctxt := testcmd.NewContext("update-service-auth-token", args)

	testcmd.RunCommand(cmd, ctxt, requirementsFactory)
	return
}

var _ = Describe("Testing with ginkgo", func() {
	It("TestUpdateServiceAuthTokenFailsWithUsage", func() {
		authTokenRepo := &testapi.FakeAuthTokenRepo{}
		requirementsFactory := &testreq.FakeReqFactory{}

		ui := callUpdateServiceAuthToken([]string{}, requirementsFactory, authTokenRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callUpdateServiceAuthToken([]string{"MY-TOKEN-LABEL"}, requirementsFactory, authTokenRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callUpdateServiceAuthToken([]string{"MY-TOKEN-LABEL", "my-token-abc123"}, requirementsFactory, authTokenRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callUpdateServiceAuthToken([]string{"MY-TOKEN-LABEL", "my-provider", "my-token-abc123"}, requirementsFactory, authTokenRepo)
		Expect(ui.FailedWithUsage).To(BeFalse())
	})
	It("TestUpdateServiceAuthTokenRequirements", func() {

		authTokenRepo := &testapi.FakeAuthTokenRepo{}
		requirementsFactory := &testreq.FakeReqFactory{}
		args := []string{"MY-TOKEN-LABLE", "my-provider", "my-token-abc123"}

		requirementsFactory.LoginSuccess = true
		callUpdateServiceAuthToken(args, requirementsFactory, authTokenRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())

		requirementsFactory.LoginSuccess = false
		callUpdateServiceAuthToken(args, requirementsFactory, authTokenRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})
	It("TestUpdateServiceAuthToken", func() {

		foundAuthToken := models.ServiceAuthTokenFields{}
		foundAuthToken.Guid = "found-auth-token-guid"
		foundAuthToken.Label = "found label"
		foundAuthToken.Provider = "found provider"

		authTokenRepo := &testapi.FakeAuthTokenRepo{FindByLabelAndProviderServiceAuthTokenFields: foundAuthToken}
		requirementsFactory := &testreq.FakeReqFactory{LoginSuccess: true}
		args := []string{"a label", "a provider", "a value"}

		ui := callUpdateServiceAuthToken(args, requirementsFactory, authTokenRepo)
		expectedAuthToken := models.ServiceAuthTokenFields{}
		expectedAuthToken.Guid = "found-auth-token-guid"
		expectedAuthToken.Label = "found label"
		expectedAuthToken.Provider = "found provider"
		expectedAuthToken.Token = "a value"

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Updating service auth token as", "my-user"},
			[]string{"OK"},
		))

		Expect(authTokenRepo.FindByLabelAndProviderLabel).To(Equal("a label"))
		Expect(authTokenRepo.FindByLabelAndProviderProvider).To(Equal("a provider"))
		Expect(authTokenRepo.UpdatedServiceAuthTokenFields).To(Equal(expectedAuthToken))
		Expect(authTokenRepo.UpdatedServiceAuthTokenFields).To(Equal(expectedAuthToken))
	})
})
