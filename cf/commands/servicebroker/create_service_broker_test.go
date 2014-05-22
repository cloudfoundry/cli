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
	testapi "github.com/cloudfoundry/cli/testhelpers/api"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing with ginkgo", func() {
	It("TestCreateServiceBrokerFailsWithUsage", func() {
		requirementsFactory := &testreq.FakeReqFactory{LoginSuccess: true}
		serviceBrokerRepo := &testapi.FakeServiceBrokerRepo{}

		ui := callCreateServiceBroker([]string{}, requirementsFactory, serviceBrokerRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callCreateServiceBroker([]string{"1arg"}, requirementsFactory, serviceBrokerRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callCreateServiceBroker([]string{"1arg", "2arg"}, requirementsFactory, serviceBrokerRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callCreateServiceBroker([]string{"1arg", "2arg", "3arg"}, requirementsFactory, serviceBrokerRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callCreateServiceBroker([]string{"1arg", "2arg", "3arg", "4arg"}, requirementsFactory, serviceBrokerRepo)
		Expect(ui.FailedWithUsage).To(BeFalse())
	})
	It("TestCreateServiceBrokerRequirements", func() {

		requirementsFactory := &testreq.FakeReqFactory{}
		serviceBrokerRepo := &testapi.FakeServiceBrokerRepo{}
		args := []string{"1arg", "2arg", "3arg", "4arg"}

		requirementsFactory.LoginSuccess = false
		callCreateServiceBroker(args, requirementsFactory, serviceBrokerRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		requirementsFactory.LoginSuccess = true
		callCreateServiceBroker(args, requirementsFactory, serviceBrokerRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())
	})
	It("TestCreateServiceBroker", func() {

		requirementsFactory := &testreq.FakeReqFactory{LoginSuccess: true}
		serviceBrokerRepo := &testapi.FakeServiceBrokerRepo{}
		args := []string{"my-broker", "my username", "my password", "http://example.com"}
		ui := callCreateServiceBroker(args, requirementsFactory, serviceBrokerRepo)

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Creating service broker", "my-broker", "my-user"},
			[]string{"OK"},
		))

		Expect(serviceBrokerRepo.CreateName).To(Equal("my-broker"))
		Expect(serviceBrokerRepo.CreateUrl).To(Equal("http://example.com"))
		Expect(serviceBrokerRepo.CreateUsername).To(Equal("my username"))
		Expect(serviceBrokerRepo.CreatePassword).To(Equal("my password"))
	})
})

func callCreateServiceBroker(args []string, requirementsFactory *testreq.FakeReqFactory, serviceBrokerRepo *testapi.FakeServiceBrokerRepo) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}
	config := testconfig.NewRepositoryWithDefaults()
	cmd := NewCreateServiceBroker(ui, config, serviceBrokerRepo)
	testcmd.RunCommand(cmd, args, requirementsFactory)
	return
}
