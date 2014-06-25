package service_test

import (
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/commands/service"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("update-user-provided-service test", func() {
	var (
		ui                  *testterm.FakeUI
		configRepo          configuration.ReadWriter
		serviceRepo         *testapi.FakeUserProvidedServiceInstanceRepo
		requirementsFactory *testreq.FakeReqFactory
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		serviceRepo = &testapi.FakeUserProvidedServiceInstanceRepo{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{}
	})

	runCommand := func(args ...string) {
		cmd := NewUpdateUserProvidedService(ui, configRepo, serviceRepo)
		testcmd.RunCommand(cmd, args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("fails with usage when not provided the name of the service to update", func() {
			requirementsFactory.LoginSuccess = true
			runCommand()
			Expect(ui.FailedWithUsage).To(BeTrue())
		})

		It("fails when not logged in", func() {
			runCommand("whoops")
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})
	})

	Context("when logged in", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true

			serviceInstance := models.ServiceInstance{}
			serviceInstance.Name = "service-name"
			requirementsFactory.ServiceInstance = serviceInstance
		})

		Context("when no flags are provided", func() {
			It("tells the user that no changes occurred", func() {
				runCommand("service-name")

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Updating user provided service", "service-name", "my-org", "my-space", "my-user"},
					[]string{"OK"},
					[]string{"No changes"},
				))
			})
		})

		Context("when the user provides valid JSON with the -p flag", func() {
			It("updates the user provided service specified", func() {
				runCommand("-p", `{"foo":"bar"}`, "-l", "syslog://example.com", "service-name")

				Expect(requirementsFactory.ServiceInstanceName).To(Equal("service-name"))
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Updating user provided service", "service-name", "my-org", "my-space", "my-user"},
					[]string{"OK"},
					[]string{"TIP"},
				))
				Expect(serviceRepo.UpdateServiceInstance.Name).To(Equal("service-name"))
				Expect(serviceRepo.UpdateServiceInstance.Params).To(Equal(map[string]string{"foo": "bar"}))
				Expect(serviceRepo.UpdateServiceInstance.SysLogDrainUrl).To(Equal("syslog://example.com"))
			})
		})

		Context("when the user provides invalid JSON with the -p flag", func() {
			It("tells the user the JSON is invalid", func() {
				runCommand("-p", `{"foo":"ba WHOOPS OH MY HOW DID THIS GET HERE???`, "service-name")

				Expect(serviceRepo.UpdateServiceInstance).To(Equal(models.ServiceInstanceFields{}))
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"JSON is invalid"},
				))
			})
		})

		Context("when the service with the given name is not user provided", func() {
			BeforeEach(func() {
				plan := models.ServicePlanFields{Guid: "my-plan-guid"}
				serviceInstance := models.ServiceInstance{}
				serviceInstance.Name = "found-service-name"
				serviceInstance.ServicePlan = plan

				requirementsFactory.ServiceInstance = serviceInstance
			})

			It("fails and tells the user what went wrong", func() {
				runCommand("-p", `{"foo":"bar"}`, "service-name")

				Expect(serviceRepo.UpdateServiceInstance).To(Equal(models.ServiceInstanceFields{}))
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"Service Instance is not user provided"},
				))
			})
		})
	})
})
