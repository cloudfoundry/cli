package servicebroker_test

import (
	. "cf/commands/servicebroker"
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

func callUpdateServiceBroker(t mr.TestingT, args []string, reqFactory *testreq.FakeReqFactory, repo *testapi.FakeServiceBrokerRepo) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}

	config := testconfig.NewRepositoryWithDefaults()

	cmd := NewUpdateServiceBroker(ui, config, repo)
	ctxt := testcmd.NewContext("update-service-broker", args)
	testcmd.RunCommand(cmd, ctxt, reqFactory)

	return
}

var _ = Describe("Testing with ginkgo", func() {
	It("TestUpdateServiceBrokerFailsWithUsage", func() {
		reqFactory := &testreq.FakeReqFactory{}
		repo := &testapi.FakeServiceBrokerRepo{}

		ui := callUpdateServiceBroker(mr.T(), []string{}, reqFactory, repo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callUpdateServiceBroker(mr.T(), []string{"arg1"}, reqFactory, repo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callUpdateServiceBroker(mr.T(), []string{"arg1", "arg2"}, reqFactory, repo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callUpdateServiceBroker(mr.T(), []string{"arg1", "arg2", "arg3"}, reqFactory, repo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callUpdateServiceBroker(mr.T(), []string{"arg1", "arg2", "arg3", "arg4"}, reqFactory, repo)
		Expect(ui.FailedWithUsage).To(BeFalse())
	})
	It("TestUpdateServiceBrokerRequirements", func() {

		reqFactory := &testreq.FakeReqFactory{}
		repo := &testapi.FakeServiceBrokerRepo{}
		args := []string{"arg1", "arg2", "arg3", "arg4"}

		reqFactory.LoginSuccess = false
		callUpdateServiceBroker(mr.T(), args, reqFactory, repo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		reqFactory.LoginSuccess = true
		callUpdateServiceBroker(mr.T(), args, reqFactory, repo)
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())
	})
	It("TestUpdateServiceBroker", func() {

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
		broker := models.ServiceBroker{}
		broker.Name = "my-found-broker"
		broker.Guid = "my-found-broker-guid"
		repo := &testapi.FakeServiceBrokerRepo{
			FindByNameServiceBroker: broker,
		}
		args := []string{"my-broker", "new-username", "new-password", "new-url"}

		ui := callUpdateServiceBroker(mr.T(), args, reqFactory, repo)

		Expect(repo.FindByNameName).To(Equal("my-broker"))

		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Updating service broker", "my-found-broker", "my-user"},
			{"OK"},
		})

		expectedServiceBroker := models.ServiceBroker{}
		expectedServiceBroker.Name = "my-found-broker"
		expectedServiceBroker.Username = "new-username"
		expectedServiceBroker.Password = "new-password"
		expectedServiceBroker.Url = "new-url"
		expectedServiceBroker.Guid = "my-found-broker-guid"

		Expect(repo.UpdatedServiceBroker).To(Equal(expectedServiceBroker))
	})
})
