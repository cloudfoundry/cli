package serviceauthtoken_test

import (
	. "cf/commands/serviceauthtoken"
	"cf/models"
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

func callUpdateServiceAuthToken(t mr.TestingT, args []string, reqFactory *testreq.FakeReqFactory, authTokenRepo *testapi.FakeAuthTokenRepo) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)

	config := testconfig.NewRepositoryWithDefaults()

	cmd := NewUpdateServiceAuthToken(ui, config, authTokenRepo)
	ctxt := testcmd.NewContext("update-service-auth-token", args)

	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestUpdateServiceAuthTokenFailsWithUsage", func() {
			authTokenRepo := &testapi.FakeAuthTokenRepo{}
			reqFactory := &testreq.FakeReqFactory{}

			ui := callUpdateServiceAuthToken(mr.T(), []string{}, reqFactory, authTokenRepo)
			assert.True(mr.T(), ui.FailedWithUsage)

			ui = callUpdateServiceAuthToken(mr.T(), []string{"MY-TOKEN-LABEL"}, reqFactory, authTokenRepo)
			assert.True(mr.T(), ui.FailedWithUsage)

			ui = callUpdateServiceAuthToken(mr.T(), []string{"MY-TOKEN-LABEL", "my-token-abc123"}, reqFactory, authTokenRepo)
			assert.True(mr.T(), ui.FailedWithUsage)

			ui = callUpdateServiceAuthToken(mr.T(), []string{"MY-TOKEN-LABEL", "my-provider", "my-token-abc123"}, reqFactory, authTokenRepo)
			assert.False(mr.T(), ui.FailedWithUsage)
		})
		It("TestUpdateServiceAuthTokenRequirements", func() {

			authTokenRepo := &testapi.FakeAuthTokenRepo{}
			reqFactory := &testreq.FakeReqFactory{}
			args := []string{"MY-TOKEN-LABLE", "my-provider", "my-token-abc123"}

			reqFactory.LoginSuccess = true
			callUpdateServiceAuthToken(mr.T(), args, reqFactory, authTokenRepo)
			assert.True(mr.T(), testcmd.CommandDidPassRequirements)

			reqFactory.LoginSuccess = false
			callUpdateServiceAuthToken(mr.T(), args, reqFactory, authTokenRepo)
			assert.False(mr.T(), testcmd.CommandDidPassRequirements)
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

			assert.Equal(mr.T(), authTokenRepo.FindByLabelAndProviderLabel, "a label")
			assert.Equal(mr.T(), authTokenRepo.FindByLabelAndProviderProvider, "a provider")
			assert.Equal(mr.T(), authTokenRepo.UpdatedServiceAuthTokenFields, expectedAuthToken)
			assert.Equal(mr.T(), authTokenRepo.UpdatedServiceAuthTokenFields, expectedAuthToken)
		})
	})
}
