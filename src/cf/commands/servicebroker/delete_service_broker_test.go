package servicebroker_test

import (
	. "cf/commands/servicebroker"
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

func callDeleteServiceBroker(t mr.TestingT, args []string, reqFactory *testreq.FakeReqFactory, repo *testapi.FakeServiceBrokerRepo) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}
	ctxt := testcmd.NewContext("delete-service-broker", args)
	config := testconfig.NewRepositoryWithDefaults()

	cmd := NewDeleteServiceBroker(ui, config, repo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}

func deleteServiceBroker(t mr.TestingT, confirmation string, args []string) (ui *testterm.FakeUI, reqFactory *testreq.FakeReqFactory, repo *testapi.FakeServiceBrokerRepo) {
	serviceBroker := models.ServiceBroker{}
	serviceBroker.Name = "service-broker-to-delete"
	serviceBroker.Guid = "service-broker-to-delete-guid"

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: true}
	repo = &testapi.FakeServiceBrokerRepo{FindByNameServiceBroker: serviceBroker}
	ui = &testterm.FakeUI{
		Inputs: []string{confirmation},
	}
	config := testconfig.NewRepositoryWithDefaults()

	ctxt := testcmd.NewContext("delete-service-broker", args)
	cmd := NewDeleteServiceBroker(ui, config, repo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}

var _ = Describe("Testing with ginkgo", func() {
	It("TestDeleteServiceBrokerFailsWithUsage", func() {
		ui, _, _ := deleteServiceBroker(mr.T(), "y", []string{})
		assert.True(mr.T(), ui.FailedWithUsage)

		ui, _, _ = deleteServiceBroker(mr.T(), "y", []string{"my-broker"})
		assert.False(mr.T(), ui.FailedWithUsage)
	})
	It("TestDeleteServiceBrokerRequirements", func() {

		reqFactory := &testreq.FakeReqFactory{}
		repo := &testapi.FakeServiceBrokerRepo{}

		reqFactory.LoginSuccess = false
		callDeleteServiceBroker(mr.T(), []string{"-f", "my-broker"}, reqFactory, repo)
		assert.False(mr.T(), testcmd.CommandDidPassRequirements)

		reqFactory.LoginSuccess = true
		callDeleteServiceBroker(mr.T(), []string{"-f", "my-broker"}, reqFactory, repo)
		assert.True(mr.T(), testcmd.CommandDidPassRequirements)
	})
	It("TestDeleteConfirmingWithY", func() {

		ui, _, repo := deleteServiceBroker(mr.T(), "y", []string{"service-broker-to-delete"})

		assert.Equal(mr.T(), repo.FindByNameName, "service-broker-to-delete")
		assert.Equal(mr.T(), repo.DeletedServiceBrokerGuid, "service-broker-to-delete-guid")
		assert.Equal(mr.T(), len(ui.Outputs), 2)
		testassert.SliceContains(mr.T(), ui.Prompts, testassert.Lines{
			{"Really delete", "service-broker-to-delete"},
		})
		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Deleting service broker", "service-broker-to-delete", "my-user"},
			{"OK"},
		})
	})
	It("TestDeleteConfirmingWithYes", func() {

		ui, _, repo := deleteServiceBroker(mr.T(), "Yes", []string{"service-broker-to-delete"})

		assert.Equal(mr.T(), repo.FindByNameName, "service-broker-to-delete")
		assert.Equal(mr.T(), repo.DeletedServiceBrokerGuid, "service-broker-to-delete-guid")
		assert.Equal(mr.T(), len(ui.Outputs), 2)
		testassert.SliceContains(mr.T(), ui.Prompts, testassert.Lines{
			{"Really delete", "service-broker-to-delete"},
		})

		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Deleting service broker", "service-broker-to-delete", "my-user"},
			{"OK"},
		})
	})
	It("TestDeleteWithForceOption", func() {

		serviceBroker := models.ServiceBroker{}
		serviceBroker.Name = "service-broker-to-delete"
		serviceBroker.Guid = "service-broker-to-delete-guid"

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
		repo := &testapi.FakeServiceBrokerRepo{FindByNameServiceBroker: serviceBroker}
		ui := callDeleteServiceBroker(mr.T(), []string{"-f", "service-broker-to-delete"}, reqFactory, repo)

		assert.Equal(mr.T(), repo.FindByNameName, "service-broker-to-delete")
		assert.Equal(mr.T(), repo.DeletedServiceBrokerGuid, "service-broker-to-delete-guid")
		assert.Equal(mr.T(), len(ui.Prompts), 0)
		assert.Equal(mr.T(), len(ui.Outputs), 2)
		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Deleting service broker", "service-broker-to-delete", "my-user"},
			{"OK"},
		})
	})
	It("TestDeleteAppThatDoesNotExist", func() {

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
		repo := &testapi.FakeServiceBrokerRepo{FindByNameNotFound: true}
		ui := callDeleteServiceBroker(mr.T(), []string{"-f", "service-broker-to-delete"}, reqFactory, repo)

		assert.Equal(mr.T(), repo.FindByNameName, "service-broker-to-delete")
		assert.Equal(mr.T(), repo.DeletedServiceBrokerGuid, "")
		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Deleting service broker", "service-broker-to-delete"},
			{"OK"},
			{"service-broker-to-delete", "does not exist"},
		})
	})
})
