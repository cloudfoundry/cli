package service_test

import (
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/commands/service"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("service command", func() {
	var (
		ui                  *testterm.FakeUI
		requirementsFactory *testreq.FakeReqFactory
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		requirementsFactory = &testreq.FakeReqFactory{}
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCommand(NewShowService(ui), args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("fails when not provided the name of the service to show", func() {
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedSpaceSuccess = true
			runCommand()

			Expect(ui.FailedWithUsage).To(BeTrue())
		})

		It("fails when not logged in", func() {
			requirementsFactory.TargetedSpaceSuccess = true

			Expect(runCommand("come-ON")).To(BeFalse())
		})

		It("fails when a space is not targeted", func() {
			requirementsFactory.LoginSuccess = true

			Expect(runCommand("okay-this-time-please??")).To(BeFalse())
		})
	})

	Context("when logged in, a space is targeted, and provided the name of a service that exists", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedSpaceSuccess = true
		})

		Context("when the service is externally provided", func() {
			BeforeEach(func() {
				offering := models.ServiceOfferingFields{Label: "mysql", DocumentationUrl: "http://documentation.url", Description: "the-description"}
				plan := models.ServicePlanFields{Guid: "plan-guid", Name: "plan-name"}

				serviceInstance := models.ServiceInstance{}
				serviceInstance.Name = "service1"
				serviceInstance.Guid = "service1-guid"
				serviceInstance.ServicePlan = plan
				serviceInstance.ServiceOffering = offering
				requirementsFactory.ServiceInstance = serviceInstance
			})

			It("shows the service", func() {
				runCommand("service1")

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Service instance:", "service1"},
					[]string{"Service: ", "mysql"},
					[]string{"Plan: ", "plan-name"},
					[]string{"Description: ", "the-description"},
					[]string{"Documentation url: ", "http://documentation.url"},
				))
				Expect(requirementsFactory.ServiceInstanceName).To(Equal("service1"))
			})
		})

		Context("when th e service is user provided", func() {
			BeforeEach(func() {
				serviceInstance := models.ServiceInstance{}
				serviceInstance.Name = "service1"
				serviceInstance.Guid = "service1-guid"
				requirementsFactory.ServiceInstance = serviceInstance
			})

			It("shows user provided services", func() {
				runCommand("service1")

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Service instance: ", "service1"},
					[]string{"Service: ", "user-provided"},
				))
			})
		})
	})
})
