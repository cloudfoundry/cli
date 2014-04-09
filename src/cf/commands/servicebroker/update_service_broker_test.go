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

func callUpdateServiceBroker(args []string, requirementsFactory *testreq.FakeReqFactory, repo *testapi.FakeServiceBrokerRepo) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}

	config := testconfig.NewRepositoryWithDefaults()

	cmd := NewUpdateServiceBroker(ui, config, repo)
	ctxt := testcmd.NewContext("update-service-broker", args)
	testcmd.RunCommand(cmd, ctxt, requirementsFactory)

	return
}

var _ = Describe("Testing with ginkgo", func() {
	It("TestUpdateServiceBrokerFailsWithUsage", func() {
		requirementsFactory := &testreq.FakeReqFactory{}
		repo := &testapi.FakeServiceBrokerRepo{}

		ui := callUpdateServiceBroker([]string{}, requirementsFactory, repo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callUpdateServiceBroker([]string{"arg1"}, requirementsFactory, repo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callUpdateServiceBroker([]string{"arg1", "arg2"}, requirementsFactory, repo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callUpdateServiceBroker([]string{"arg1", "arg2", "arg3"}, requirementsFactory, repo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callUpdateServiceBroker([]string{"arg1", "arg2", "arg3", "arg4"}, requirementsFactory, repo)
		Expect(ui.FailedWithUsage).To(BeFalse())
	})
	It("TestUpdateServiceBrokerRequirements", func() {

		requirementsFactory := &testreq.FakeReqFactory{}
		repo := &testapi.FakeServiceBrokerRepo{}
		args := []string{"arg1", "arg2", "arg3", "arg4"}

		requirementsFactory.LoginSuccess = false
		callUpdateServiceBroker(args, requirementsFactory, repo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		requirementsFactory.LoginSuccess = true
		callUpdateServiceBroker(args, requirementsFactory, repo)
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())
	})
	It("TestUpdateServiceBroker", func() {

		requirementsFactory := &testreq.FakeReqFactory{LoginSuccess: true}
		broker := models.ServiceBroker{}
		broker.Name = "my-found-broker"
		broker.Guid = "my-found-broker-guid"
		repo := &testapi.FakeServiceBrokerRepo{
			FindByNameServiceBroker: broker,
		}
		args := []string{"my-broker", "new-username", "new-password", "new-url"}

		ui := callUpdateServiceBroker(args, requirementsFactory, repo)

		Expect(repo.FindByNameName).To(Equal("my-broker"))

		testassert.SliceContains(ui.Outputs, testassert.Lines{
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
