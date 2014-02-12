package servicebroker_test

import (
	. "cf/commands/servicebroker"
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

func callRenameServiceBroker(t mr.TestingT, args []string, reqFactory *testreq.FakeReqFactory, repo *testapi.FakeServiceBrokerRepo) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}
	config := testconfig.NewRepositoryWithDefaults()
	cmd := NewRenameServiceBroker(ui, config, repo)
	ctxt := testcmd.NewContext("rename-service-broker", args)
	testcmd.RunCommand(cmd, ctxt, reqFactory)

	return
}

var _ = Describe("Testing with ginkgo", func() {
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
		broker := models.ServiceBroker{}
		broker.Name = "my-found-broker"
		broker.Guid = "my-found-broker-guid"
		repo := &testapi.FakeServiceBrokerRepo{
			FindByNameServiceBroker: broker,
		}
		args := []string{"my-broker", "my-new-broker"}

		ui := callRenameServiceBroker(mr.T(), args, reqFactory, repo)

		Expect(repo.FindByNameName).To(Equal("my-broker"))

		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Renaming service broker", "my-found-broker", "my-new-broker", "my-user"},
			{"OK"},
		})

		Expect(repo.RenamedServiceBrokerGuid).To(Equal("my-found-broker-guid"))
		Expect(repo.RenamedServiceBrokerName).To(Equal("my-new-broker"))
	})
})
