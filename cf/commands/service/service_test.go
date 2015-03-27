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
			createServiceInstanceWithState := func(state string) {
				offering := models.ServiceOfferingFields{Label: "mysql", DocumentationUrl: "http://documentation.url", Description: "the-description"}
				plan := models.ServicePlanFields{Guid: "plan-guid", Name: "plan-name"}

				serviceInstance := models.ServiceInstance{}
				serviceInstance.Name = "service1"
				serviceInstance.Guid = "service1-guid"
				serviceInstance.LastOperation.Type = "create"
				serviceInstance.LastOperation.State = "in progress"
				serviceInstance.LastOperation.Description = "creating resource - step 1"
				serviceInstance.ServicePlan = plan
				serviceInstance.ServiceOffering = offering
				serviceInstance.DashboardUrl = "some-url"
				serviceInstance.LastOperation.State = state
				requirementsFactory.ServiceInstance = serviceInstance
			}

			createServiceInstance := func() {
				createServiceInstanceWithState("")
			}

			It("shows the service", func() {
				createServiceInstanceWithState("in progress")
				runCommand("service1")

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Service instance:", "service1"},
					[]string{"Service: ", "mysql"},
					[]string{"Plan: ", "plan-name"},
					[]string{"Description: ", "the-description"},
					[]string{"Documentation url: ", "http://documentation.url"},
					[]string{"Dashboard: ", "some-url"},
					[]string{"Last Operation"},
					[]string{"Status: ", "create in progress"},
					[]string{"Message: ", "creating resource - step 1"},
				))
				Expect(requirementsFactory.ServiceInstanceName).To(Equal("service1"))
			})

			Context("shows correct status information based on service instance state", func() {
				It("shows status: `create in progress` when state is `in progress`", func() {
					createServiceInstanceWithState("in progress")
					runCommand("service1")

					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Status: ", "create in progress"},
					))
					Expect(requirementsFactory.ServiceInstanceName).To(Equal("service1"))
				})

				It("shows status: `create succeeded` when state is `succeeded`", func() {
					createServiceInstanceWithState("succeeded")
					runCommand("service1")

					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Status: ", "create succeeded"},
					))
					Expect(requirementsFactory.ServiceInstanceName).To(Equal("service1"))
				})

				It("shows status: `create failed` when state is `failed`", func() {
					createServiceInstanceWithState("failed")
					runCommand("service1")

					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Status: ", "create failed"},
					))
					Expect(requirementsFactory.ServiceInstanceName).To(Equal("service1"))
				})

				It("shows status: `` when state is ``", func() {
					createServiceInstanceWithState("")
					runCommand("service1")

					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Status: ", ""},
					))
					Expect(requirementsFactory.ServiceInstanceName).To(Equal("service1"))
				})
			})

			Context("when the guid flag is provided", func() {
				It("shows only the service guid", func() {
					createServiceInstance()
					runCommand("--guid", "service1")

					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"service1-guid"},
					))

					Expect(ui.Outputs).ToNot(ContainSubstrings(
						[]string{"Service instance:", "service1"},
					))
				})
			})
		})

		Context("when the service is user provided", func() {
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

var _ = Describe("ServiceInstanceStateToStatus", func() {
	var operationType string
	Context("when the service is not user provided", func() {
		isUserProvided := false

		Context("when operationType is `create`", func() {
			BeforeEach(func() { operationType = "create" })

			It("returns status: `create in progress` when state: `in progress`", func() {
				status := ServiceInstanceStateToStatus(operationType, "in progress", isUserProvided)
				Expect(status).To(Equal("create in progress"))
			})

			It("returns status: `create succeeded` when state: `succeeded`", func() {
				status := ServiceInstanceStateToStatus(operationType, "succeeded", isUserProvided)
				Expect(status).To(Equal("create succeeded"))
			})

			It("returns status: `create failed` when state: `failed`", func() {
				status := ServiceInstanceStateToStatus(operationType, "failed", isUserProvided)
				Expect(status).To(Equal("create failed"))
			})

			It("returns status: `` when state: ``", func() {
				status := ServiceInstanceStateToStatus(operationType, "", isUserProvided)
				Expect(status).To(Equal(""))
			})
		})
	})

	Context("when the service is user provided", func() {
		isUserProvided := true

		It("returns status: `` when state: ``", func() {
			status := ServiceInstanceStateToStatus(operationType, "", isUserProvided)
			Expect(status).To(Equal(""))
		})
	})
})
