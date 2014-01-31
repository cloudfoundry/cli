package servicebroker_test

import (
	"cf"
	. "cf/commands/servicebroker"
	"cf/configuration"
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

func callCreateServiceBroker(t mr.TestingT, args []string, reqFactory *testreq.FakeReqFactory, serviceBrokerRepo *testapi.FakeServiceBrokerRepo) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}
	ctxt := testcmd.NewContext("create-service-broker", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)
	org := cf.OrganizationFields{}
	org.Name = "my-org"
	space := cf.SpaceFields{}
	space.Name = "my-space"
	config := &configuration.Configuration{
		SpaceFields:        space,
		OrganizationFields: org,
		AccessToken:        token,
	}

	cmd := NewCreateServiceBroker(ui, config, serviceBrokerRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestCreateServiceBrokerFailsWithUsage", func() {
			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
			serviceBrokerRepo := &testapi.FakeServiceBrokerRepo{}

			ui := callCreateServiceBroker(mr.T(), []string{}, reqFactory, serviceBrokerRepo)
			assert.True(mr.T(), ui.FailedWithUsage)

			ui = callCreateServiceBroker(mr.T(), []string{"1arg"}, reqFactory, serviceBrokerRepo)
			assert.True(mr.T(), ui.FailedWithUsage)

			ui = callCreateServiceBroker(mr.T(), []string{"1arg", "2arg"}, reqFactory, serviceBrokerRepo)
			assert.True(mr.T(), ui.FailedWithUsage)

			ui = callCreateServiceBroker(mr.T(), []string{"1arg", "2arg", "3arg"}, reqFactory, serviceBrokerRepo)
			assert.True(mr.T(), ui.FailedWithUsage)

			ui = callCreateServiceBroker(mr.T(), []string{"1arg", "2arg", "3arg", "4arg"}, reqFactory, serviceBrokerRepo)
			assert.False(mr.T(), ui.FailedWithUsage)
		})
		It("TestCreateServiceBrokerRequirements", func() {

			reqFactory := &testreq.FakeReqFactory{}
			serviceBrokerRepo := &testapi.FakeServiceBrokerRepo{}
			args := []string{"1arg", "2arg", "3arg", "4arg"}

			reqFactory.LoginSuccess = false
			callCreateServiceBroker(mr.T(), args, reqFactory, serviceBrokerRepo)
			assert.False(mr.T(), testcmd.CommandDidPassRequirements)

			reqFactory.LoginSuccess = true
			callCreateServiceBroker(mr.T(), args, reqFactory, serviceBrokerRepo)
			assert.True(mr.T(), testcmd.CommandDidPassRequirements)
		})
		It("TestCreateServiceBroker", func() {

			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
			serviceBrokerRepo := &testapi.FakeServiceBrokerRepo{}
			args := []string{"my-broker", "my username", "my password", "http://example.com"}
			ui := callCreateServiceBroker(mr.T(), args, reqFactory, serviceBrokerRepo)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Creating service broker", "my-broker", "my-user"},
				{"OK"},
			})

			assert.Equal(mr.T(), serviceBrokerRepo.CreateName, "my-broker")
			assert.Equal(mr.T(), serviceBrokerRepo.CreateUrl, "http://example.com")
			assert.Equal(mr.T(), serviceBrokerRepo.CreateUsername, "my username")
			assert.Equal(mr.T(), serviceBrokerRepo.CreatePassword, "my password")
		})
	})
}
