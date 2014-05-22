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

package service_test

import (
	. "github.com/cloudfoundry/cli/cf/commands/service"
	"github.com/cloudfoundry/cli/cf/models"
	testapi "github.com/cloudfoundry/cli/testhelpers/api"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

func callRenameService(args []string, requirementsFactory *testreq.FakeReqFactory) (ui *testterm.FakeUI, serviceRepo *testapi.FakeServiceRepo) {
	ui = &testterm.FakeUI{}
	serviceRepo = &testapi.FakeServiceRepo{}

	config := testconfig.NewRepositoryWithDefaults()

	cmd := NewRenameService(ui, config, serviceRepo)
	testcmd.RunCommand(cmd, args, requirementsFactory)
	return
}

var _ = Describe("Testing with ginkgo", func() {
	It("TestRenameServiceFailsWithUsage", func() {
		requirementsFactory := &testreq.FakeReqFactory{}

		fakeUI, _ := callRenameService([]string{}, requirementsFactory)
		Expect(fakeUI.FailedWithUsage).To(BeTrue())

		fakeUI, _ = callRenameService([]string{"my-service"}, requirementsFactory)
		Expect(fakeUI.FailedWithUsage).To(BeTrue())

		fakeUI, _ = callRenameService([]string{"my-service", "new-name", "extra"}, requirementsFactory)
		Expect(fakeUI.FailedWithUsage).To(BeTrue())
	})
	It("TestRenameServiceRequirements", func() {

		requirementsFactory := &testreq.FakeReqFactory{LoginSuccess: false, TargetedSpaceSuccess: true}
		callRenameService([]string{"my-service", "new-name"}, requirementsFactory)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: false}
		callRenameService([]string{"my-service", "new-name"}, requirementsFactory)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		Expect(requirementsFactory.ServiceInstanceName).To(Equal("my-service"))
	})
	It("TestRenameService", func() {

		serviceInstance := models.ServiceInstance{}
		serviceInstance.Name = "different-name"
		serviceInstance.Guid = "different-name-guid"
		requirementsFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, ServiceInstance: serviceInstance}
		ui, fakeServiceRepo := callRenameService([]string{"my-service", "new-name"}, requirementsFactory)

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Renaming service", "different-name", "new-name", "my-org", "my-space", "my-user"},
			[]string{"OK"},
		))

		Expect(fakeServiceRepo.RenameServiceServiceInstance).To(Equal(serviceInstance))
		Expect(fakeServiceRepo.RenameServiceNewName).To(Equal("new-name"))
	})
})
