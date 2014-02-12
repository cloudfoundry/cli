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
