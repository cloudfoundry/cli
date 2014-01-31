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

func callRenameServiceBroker(t mr.TestingT, args []string, reqFactory *testreq.FakeReqFactory, repo *testapi.FakeServiceBrokerRepo) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)
	space := cf.SpaceFields{}
	space.Name = "my-space"
	org := cf.OrganizationFields{}
	org.Name = "my-org"
	config := &configuration.Configuration{
		SpaceFields:        space,
		OrganizationFields: org,
		AccessToken:        token,
	}

	cmd := NewRenameServiceBroker(ui, config, repo)
	ctxt := testcmd.NewContext("rename-service-broker", args)
	testcmd.RunCommand(cmd, ctxt, reqFactory)

	return
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestRenameServiceBrokerFailsWithUsage", func() {
			reqFactory := &testreq.FakeReqFactory{}
			repo := &testapi.FakeServiceBrokerRepo{}

			ui := callRenameServiceBroker(mr.T(), []string{}, reqFactory, repo)
			assert.True(mr.T(), ui.FailedWithUsage)

			ui = callRenameServiceBroker(mr.T(), []string{"arg1"}, reqFactory, repo)
			assert.True(mr.T(), ui.FailedWithUsage)

			ui = callRenameServiceBroker(mr.T(), []string{"arg1", "arg2"}, reqFactory, repo)
			assert.False(mr.T(), ui.FailedWithUsage)
		})
		It("TestRenameServiceBrokerRequirements", func() {

			reqFactory := &testreq.FakeReqFactory{}
			repo := &testapi.FakeServiceBrokerRepo{}
			args := []string{"arg1", "arg2"}

			reqFactory.LoginSuccess = false
			callRenameServiceBroker(mr.T(), args, reqFactory, repo)
			assert.False(mr.T(), testcmd.CommandDidPassRequirements)

			reqFactory.LoginSuccess = true
			callRenameServiceBroker(mr.T(), args, reqFactory, repo)
			assert.True(mr.T(), testcmd.CommandDidPassRequirements)
		})
		It("TestRenameServiceBroker", func() {

			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
			broker := cf.ServiceBroker{}
			broker.Name = "my-found-broker"
			broker.Guid = "my-found-broker-guid"
			repo := &testapi.FakeServiceBrokerRepo{
				FindByNameServiceBroker: broker,
			}
			args := []string{"my-broker", "my-new-broker"}

			ui := callRenameServiceBroker(mr.T(), args, reqFactory, repo)

			assert.Equal(mr.T(), repo.FindByNameName, "my-broker")

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Renaming service broker", "my-found-broker", "my-new-broker", "my-user"},
				{"OK"},
			})

			assert.Equal(mr.T(), repo.RenamedServiceBrokerGuid, "my-found-broker-guid")
			assert.Equal(mr.T(), repo.RenamedServiceBrokerName, "my-new-broker")
		})
	})
}
