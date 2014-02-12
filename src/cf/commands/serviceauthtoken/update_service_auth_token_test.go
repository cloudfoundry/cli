package serviceauthtoken_test

import (
	. "cf/commands/serviceauthtoken"
	"cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	mr "github.com/tjarratt/mr_t"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

func callUpdateServiceAuthToken(t mr.TestingT, args []string, reqFactory *testreq.FakeReqFactory, authTokenRepo *testapi.FakeAuthTokenRepo) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)

	config := testconfig.NewRepositoryWithDefaults()

	cmd := NewUpdateServiceAuthToken(ui, config, authTokenRepo)
	ctxt := testcmd.NewContext("update-service-auth-token", args)

	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}

var _ = Describe("Testing with ginkgo", func() {
	It("TestUpdateServiceAuthTokenFailsWithUsage", func() {
		authTokenRepo := &testapi.FakeAuthTokenRepo{}
		reqFactory := &testreq.FakeReqFactory{}

		ui := callUpdateServiceAuthToken(mr.T(), []string{}, reqFactory, authTokenRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callUpdateServiceAuthToken(mr.T(), []string{"MY-TOKEN-LABEL"}, reqFactory, authTokenRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callUpdateServiceAuthToken(mr.T(), []string{"MY-TOKEN-LABEL", "my-token-abc123"}, reqFactory, authTokenRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callUpdateServiceAuthToken(mr.T(), []string{"MY-TOKEN-LABEL", "my-provider", "my-token-abc123"}, reqFactory, authTokenRepo)
		Expect(ui.FailedWithUsage).To(BeFalse())
	})
	It("TestUpdateServiceAuthTokenRequirements", func() {

		authTokenRepo := &testapi.FakeAuthTokenRepo{}
		reqFactory := &testreq.FakeReqFactory{}
		args := []string{"MY-TOKEN-LABLE", "my-provider", "my-token-abc123"}

		reqFactory.LoginSuccess = true
		callUpdateServiceAuthToken(mr.T(), args, reqFactory, authTokenRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())

		reqFactory.LoginSuccess = false
		callUpdateServiceAuthToken(mr.T(), args, reqFactory, authTokenRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})
	It("TestUpdateServiceAuthToken", func() {

		foundAuthToken := models.ServiceAuthTokenFields{}
		foundAuthToken.Guid = "found-auth-token-guid"
		foundAuthToken.Label = "found label"
		foundAuthToken.Provider = "found provider"

		authTokenRepo := &testapi.FakeAuthTokenRepo{FindByLabelAndProviderServiceAuthTokenFields: foundAuthToken}
		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
		args := []string{"a label", "a provider", "a value"}

		ui := callUpdateServiceAuthToken(mr.T(), args, reqFactory, authTokenRepo)
		expectedAuthToken := models.ServiceAuthTokenFields{}
		expectedAuthToken.Guid = "found-auth-token-guid"
		expectedAuthToken.Label = "found label"
		expectedAuthToken.Provider = "found provider"
		expectedAuthToken.Token = "a value"

		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Updating service auth token as", "my-user"},
			{"OK"},
		})

		Expect(authTokenRepo.FindByLabelAndProviderLabel).To(Equal("a label"))
		Expect(authTokenRepo.FindByLabelAndProviderProvider).To(Equal("a provider"))
		Expect(authTokenRepo.UpdatedServiceAuthTokenFields).To(Equal(expectedAuthToken))
		Expect(authTokenRepo.UpdatedServiceAuthTokenFields).To(Equal(expectedAuthToken))
	})
})
