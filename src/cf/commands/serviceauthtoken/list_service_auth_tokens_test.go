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

func callListServiceAuthTokens(t mr.TestingT, reqFactory *testreq.FakeReqFactory, authTokenRepo *testapi.FakeAuthTokenRepo) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}

	config := testconfig.NewRepositoryWithDefaults()

	cmd := NewListServiceAuthTokens(ui, config, authTokenRepo)
	ctxt := testcmd.NewContext("service-auth-tokens", []string{})
	testcmd.RunCommand(cmd, ctxt, reqFactory)

	return
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestListServiceAuthTokensRequirements", func() {
			authTokenRepo := &testapi.FakeAuthTokenRepo{}
			reqFactory := &testreq.FakeReqFactory{}

			reqFactory.LoginSuccess = false
			callListServiceAuthTokens(mr.T(), reqFactory, authTokenRepo)
			assert.False(mr.T(), testcmd.CommandDidPassRequirements)

			reqFactory.LoginSuccess = true
			callListServiceAuthTokens(mr.T(), reqFactory, authTokenRepo)
			assert.True(mr.T(), testcmd.CommandDidPassRequirements)
		})
		It("TestListServiceAuthTokens", func() {

			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
			authTokenRepo := &testapi.FakeAuthTokenRepo{}
			authToken := models.ServiceAuthTokenFields{}
			authToken.Label = "a label"
			authToken.Provider = "a provider"
			authToken2 := models.ServiceAuthTokenFields{}
			authToken2.Label = "a second label"
			authToken2.Provider = "a second provider"
			authTokenRepo.FindAllAuthTokens = []models.ServiceAuthTokenFields{authToken, authToken2}

			ui := callListServiceAuthTokens(mr.T(), reqFactory, authTokenRepo)
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Getting service auth tokens as", "my-user"},
				{"OK"},
				{"label", "provider"},
				{"a label", "a provider"},
				{"a second label", "a second provider"},
			})
		})
	})
}
