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

func callListServiceAuthTokens(requirementsFactory *testreq.FakeReqFactory, authTokenRepo *testapi.FakeAuthTokenRepo) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}

	config := testconfig.NewRepositoryWithDefaults()

	cmd := NewListServiceAuthTokens(ui, config, authTokenRepo)
	testcmd.RunCommand(cmd, []string{}, requirementsFactory)

	return
}

var _ = Describe("Testing with ginkgo", func() {
	It("TestListServiceAuthTokensRequirements", func() {
		authTokenRepo := &testapi.FakeAuthTokenRepo{}
		requirementsFactory := &testreq.FakeReqFactory{}

		requirementsFactory.LoginSuccess = false
		callListServiceAuthTokens(requirementsFactory, authTokenRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		requirementsFactory.LoginSuccess = true
		callListServiceAuthTokens(requirementsFactory, authTokenRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())
	})
	It("TestListServiceAuthTokens", func() {

		requirementsFactory := &testreq.FakeReqFactory{LoginSuccess: true}
		authTokenRepo := &testapi.FakeAuthTokenRepo{}
		authToken := models.ServiceAuthTokenFields{}
		authToken.Label = "a label"
		authToken.Provider = "a provider"
		authToken2 := models.ServiceAuthTokenFields{}
		authToken2.Label = "a second label"
		authToken2.Provider = "a second provider"
		authTokenRepo.FindAllAuthTokens = []models.ServiceAuthTokenFields{authToken, authToken2}

		ui := callListServiceAuthTokens(requirementsFactory, authTokenRepo)
		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Getting service auth tokens as", "my-user"},
			[]string{"OK"},
			[]string{"label", "provider"},
			[]string{"a label", "a provider"},
			[]string{"a second label", "a second provider"},
		))
	})
})
