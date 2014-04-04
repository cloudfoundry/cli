/*
                       WARNING WARNING WARNING

                Attention all potential contributors

   This testfile is not in the best state. We've been slowly transitioning
   from the built in "testing" package to using Ginkgo. As you can see, we've
   changed the format, but a lot of the setup, test body, descriptions, etc
   are either hardcoded, completely lacking, or misleading.

   For example:

   Describe("Testing with ginkgo"...)      // This is not a great description
   It("TestDoesSoemthing"...)              // This is a horrible description

   Describe("create-user command"...       // Describe the actual object under test
   It("creates a user when provided ..."   // this is more descriptive

   For good examples of writing Ginkgo tests for the cli, refer to

   src/cf/commands/application/delete_app_test.go
   src/cf/terminal/ui_test.go
   src/github.com/cloudfoundry/loggregator_consumer/consumer_test.go
*/

package servicebroker_test

import (
	. "cf/commands/servicebroker"
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

func callDeleteServiceBroker(args []string, reqFactory *testreq.FakeReqFactory, repo *testapi.FakeServiceBrokerRepo) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}
	ctxt := testcmd.NewContext("delete-service-broker", args)
	config := testconfig.NewRepositoryWithDefaults()

	cmd := NewDeleteServiceBroker(ui, config, repo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}

func deleteServiceBroker(confirmation string, args []string) (ui *testterm.FakeUI, reqFactory *testreq.FakeReqFactory, repo *testapi.FakeServiceBrokerRepo) {
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
		ui, _, _ := deleteServiceBroker("y", []string{})
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui, _, _ = deleteServiceBroker("y", []string{"my-broker"})
		Expect(ui.FailedWithUsage).To(BeFalse())
	})
	It("TestDeleteServiceBrokerRequirements", func() {

		reqFactory := &testreq.FakeReqFactory{}
		repo := &testapi.FakeServiceBrokerRepo{}

		reqFactory.LoginSuccess = false
		callDeleteServiceBroker([]string{"-f", "my-broker"}, reqFactory, repo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		reqFactory.LoginSuccess = true
		callDeleteServiceBroker([]string{"-f", "my-broker"}, reqFactory, repo)
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())
	})
	It("TestDeleteConfirmingWithY", func() {

		ui, _, repo := deleteServiceBroker("y", []string{"service-broker-to-delete"})

		Expect(repo.FindByNameName).To(Equal("service-broker-to-delete"))
		Expect(repo.DeletedServiceBrokerGuid).To(Equal("service-broker-to-delete-guid"))
		Expect(len(ui.Outputs)).To(Equal(2))
		testassert.SliceContains(ui.Prompts, testassert.Lines{
			{"Really delete", "service-broker-to-delete"},
		})
		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Deleting service broker", "service-broker-to-delete", "my-user"},
			{"OK"},
		})
	})
	It("TestDeleteConfirmingWithYes", func() {

		ui, _, repo := deleteServiceBroker("Yes", []string{"service-broker-to-delete"})

		Expect(repo.FindByNameName).To(Equal("service-broker-to-delete"))
		Expect(repo.DeletedServiceBrokerGuid).To(Equal("service-broker-to-delete-guid"))
		Expect(len(ui.Outputs)).To(Equal(2))
		testassert.SliceContains(ui.Prompts, testassert.Lines{
			{"Really delete", "service-broker-to-delete"},
		})

		testassert.SliceContains(ui.Outputs, testassert.Lines{
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
		ui := callDeleteServiceBroker([]string{"-f", "service-broker-to-delete"}, reqFactory, repo)

		Expect(repo.FindByNameName).To(Equal("service-broker-to-delete"))
		Expect(repo.DeletedServiceBrokerGuid).To(Equal("service-broker-to-delete-guid"))
		Expect(len(ui.Prompts)).To(Equal(0))
		Expect(len(ui.Outputs)).To(Equal(2))
		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Deleting service broker", "service-broker-to-delete", "my-user"},
			{"OK"},
		})
	})
	It("TestDeleteAppThatDoesNotExist", func() {

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
		repo := &testapi.FakeServiceBrokerRepo{FindByNameNotFound: true}
		ui := callDeleteServiceBroker([]string{"-f", "service-broker-to-delete"}, reqFactory, repo)

		Expect(repo.FindByNameName).To(Equal("service-broker-to-delete"))
		Expect(repo.DeletedServiceBrokerGuid).To(Equal(""))
		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Deleting service broker", "service-broker-to-delete"},
			{"OK"},
			{"service-broker-to-delete", "does not exist"},
		})
	})
})
