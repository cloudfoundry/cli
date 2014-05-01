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

   src/github.com/cloudfoundry/cli/cf/commands/application/delete_app_test.go
   src/github.com/cloudfoundry/cli/cf/terminal/ui_test.go
   src/github.com/cloudfoundry/loggregator_consumer/consumer_test.go
*/

package servicebroker_test

import (
	. "github.com/cloudfoundry/cli/cf/commands/servicebroker"
	"github.com/cloudfoundry/cli/cf/models"
	testapi "github.com/cloudfoundry/cli/testhelpers/api"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func callRenameServiceBroker(args []string, requirementsFactory *testreq.FakeReqFactory, repo *testapi.FakeServiceBrokerRepo) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}
	config := testconfig.NewRepositoryWithDefaults()
	cmd := NewRenameServiceBroker(ui, config, repo)
	ctxt := testcmd.NewContext("rename-service-broker", args)
	testcmd.RunCommand(cmd, ctxt, requirementsFactory)

	return
}

var _ = Describe("Testing with ginkgo", func() {
	It("TestRenameServiceBrokerFailsWithUsage", func() {
		requirementsFactory := &testreq.FakeReqFactory{}
		repo := &testapi.FakeServiceBrokerRepo{}

		ui := callRenameServiceBroker([]string{}, requirementsFactory, repo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callRenameServiceBroker([]string{"arg1"}, requirementsFactory, repo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callRenameServiceBroker([]string{"arg1", "arg2"}, requirementsFactory, repo)
		Expect(ui.FailedWithUsage).To(BeFalse())
	})
	It("TestRenameServiceBrokerRequirements", func() {

		requirementsFactory := &testreq.FakeReqFactory{}
		repo := &testapi.FakeServiceBrokerRepo{}
		args := []string{"arg1", "arg2"}

		requirementsFactory.LoginSuccess = false
		callRenameServiceBroker(args, requirementsFactory, repo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		requirementsFactory.LoginSuccess = true
		callRenameServiceBroker(args, requirementsFactory, repo)
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())
	})
	It("TestRenameServiceBroker", func() {

		requirementsFactory := &testreq.FakeReqFactory{LoginSuccess: true}
		broker := models.ServiceBroker{}
		broker.Name = "my-found-broker"
		broker.Guid = "my-found-broker-guid"
		repo := &testapi.FakeServiceBrokerRepo{
			FindByNameServiceBroker: broker,
		}
		args := []string{"my-broker", "my-new-broker"}

		ui := callRenameServiceBroker(args, requirementsFactory, repo)

		Expect(repo.FindByNameName).To(Equal("my-broker"))

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Renaming service broker", "my-found-broker", "my-new-broker", "my-user"},
			[]string{"OK"},
		))

		Expect(repo.RenamedServiceBrokerGuid).To(Equal("my-found-broker-guid"))
		Expect(repo.RenamedServiceBrokerName).To(Equal("my-new-broker"))
	})
})
