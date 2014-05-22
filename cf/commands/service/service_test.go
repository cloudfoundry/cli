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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

func callShowService(args []string, requirementsFactory *testreq.FakeReqFactory) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	cmd := NewShowService(ui)
	testcmd.RunCommand(cmd, args, requirementsFactory)
	return
}

var _ = Describe("Testing with ginkgo", func() {
	It("TestShowServiceRequirements", func() {
		args := []string{"service1"}
		requirementsFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}
		callShowService(args, requirementsFactory)
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())

		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: false}
		callShowService(args, requirementsFactory)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: false, TargetedSpaceSuccess: true}
		callShowService(args, requirementsFactory)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		Expect(requirementsFactory.ServiceInstanceName).To(Equal("service1"))
	})

	It("TestShowServiceFailsWithUsage", func() {
		requirementsFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}

		ui := callShowService([]string{}, requirementsFactory)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callShowService([]string{"my-service"}, requirementsFactory)
		Expect(ui.FailedWithUsage).To(BeFalse())
	})

	It("TestShowServiceOutput", func() {
		offering := models.ServiceOfferingFields{}
		offering.Label = "mysql"
		offering.DocumentationUrl = "http://documentation.url"
		offering.Description = "the-description"

		plan := models.ServicePlanFields{}
		plan.Guid = "plan-guid"
		plan.Name = "plan-name"

		serviceInstance := models.ServiceInstance{}
		serviceInstance.Name = "service1"
		serviceInstance.Guid = "service1-guid"
		serviceInstance.ServicePlan = plan
		serviceInstance.ServiceOffering = offering
		requirementsFactory := &testreq.FakeReqFactory{
			LoginSuccess:         true,
			TargetedSpaceSuccess: true,
			ServiceInstance:      serviceInstance,
		}
		ui := callShowService([]string{"service1"}, requirementsFactory)

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Service instance:", "service1"},
			[]string{"Service: ", "mysql"},
			[]string{"Plan: ", "plan-name"},
			[]string{"Description: ", "the-description"},
			[]string{"Documentation url: ", "http://documentation.url"},
		))
	})

	It("TestShowUserProvidedServiceOutput", func() {
		serviceInstance2 := models.ServiceInstance{}
		serviceInstance2.Name = "service1"
		serviceInstance2.Guid = "service1-guid"
		requirementsFactory := &testreq.FakeReqFactory{
			LoginSuccess:         true,
			TargetedSpaceSuccess: true,
			ServiceInstance:      serviceInstance2,
		}
		ui := callShowService([]string{"service1"}, requirementsFactory)

		Expect(len(ui.Outputs)).To(Equal(3))
		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Service instance: ", "service1"},
			[]string{"Service: ", "user-provided"},
		))
	})
})
