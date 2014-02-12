package serviceauthtoken_test

import (
	. "cf/commands/serviceauthtoken"
	"cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

func callCreateServiceAuthToken(t mr.TestingT, args []string, reqFactory *testreq.FakeReqFactory, authTokenRepo *testapi.FakeAuthTokenRepo) (ui *testterm.FakeUI) {
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

		ui := callCreateServiceAuthToken(mr.T(), []string{}, reqFactory, authTokenRepo)
		assert.True(mr.T(), ui.FailedWithUsage)

		ui = callCreateServiceAuthToken(mr.T(), []string{"arg1"}, reqFactory, authTokenRepo)
		assert.True(mr.T(), ui.FailedWithUsage)

		ui = callCreateServiceAuthToken(mr.T(), []string{"arg1", "arg2"}, reqFactory, authTokenRepo)
		assert.True(mr.T(), ui.FailedWithUsage)

		ui = callCreateServiceAuthToken(mr.T(), []string{"arg1", "arg2", "arg3"}, reqFactory, authTokenRepo)
		assert.False(mr.T(), ui.FailedWithUsage)
	})
	It("TestCreateServiceAuthTokenRequirements", func() {

		authTokenRepo := &testapi.FakeAuthTokenRepo{}
		reqFactory := &testreq.FakeReqFactory{}
		args := []string{"arg1", "arg2", "arg3"}

		reqFactory.LoginSuccess = true
		callCreateServiceAuthToken(mr.T(), args, reqFactory, authTokenRepo)
		assert.True(mr.T(), testcmd.CommandDidPassRequirements)

		reqFactory.LoginSuccess = false
		callCreateServiceAuthToken(mr.T(), args, reqFactory, authTokenRepo)
		assert.False(mr.T(), testcmd.CommandDidPassRequirements)
	})
	It("TestCreateServiceAuthToken", func() {

		authTokenRepo := &testapi.FakeAuthTokenRepo{}
		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
		args := []string{"a label", "a provider", "a value"}

		ui := callCreateServiceAuthToken(mr.T(), args, reqFactory, authTokenRepo)
		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
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
