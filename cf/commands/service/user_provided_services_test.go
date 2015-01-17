package service_test

import (
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	. "github.com/cloudfoundry/cli/cf/commands/service"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
)

var _ = Describe("user-provided-services", func() {
	var (
		ui                  *testterm.FakeUI
		configRepo          core_config.Repository
		requirementsFactory *testreq.FakeReqFactory
		cmd                 ListUserProvidedServices
		serviceRepo         *testapi.FakeUserProvidedServiceInstanceRepository
	)

	Describe("services requirements", func() {
		BeforeEach(func() {
			ui = &testterm.FakeUI{}
			configRepo = testconfig.NewRepositoryWithDefaults()
			requirementsFactory = &testreq.FakeReqFactory{
				LoginSuccess:         true,
				TargetedSpaceSuccess: true,
				TargetedOrgSuccess:   true,
			}
			serviceRepo = &testapi.FakeUserProvidedServiceInstanceRepository{}
			cmd = NewListUserProvidedServices(ui, configRepo, serviceRepo)
		})

		Context("when not logged in", func() {
			BeforeEach(func() {
				requirementsFactory.LoginSuccess = false
			})

			It("fails requirements", func() {
				Expect(testcmd.RunCommand(cmd, []string{}, requirementsFactory)).To(BeFalse())
			})
		})

		Context("when no space is targeted", func() {
			BeforeEach(func() {
				requirementsFactory.TargetedSpaceSuccess = false
			})

			It("fails requirements", func() {
				Expect(testcmd.RunCommand(cmd, []string{}, requirementsFactory)).To(BeFalse())
			})
		})

		It("should fail with usage when provided any arguments", func() {
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedSpaceSuccess = true
			Expect(testcmd.RunCommand(cmd, []string{"blahblah"}, requirementsFactory)).To(BeFalse())
			Expect(ui.FailedWithUsage).To(BeTrue())
		})
	})

	Describe("fail to find any user provided service", func() {
		BeforeEach(func() {
			ui = &testterm.FakeUI{}
			configRepo = testconfig.NewRepositoryWithDefaults()
			requirementsFactory = &testreq.FakeReqFactory{
				LoginSuccess:         true,
				TargetedSpaceSuccess: true,
				TargetedOrgSuccess:   true,
			}
			serviceRepo = &testapi.FakeUserProvidedServiceInstanceRepository{}
			cmd = NewListUserProvidedServices(ui, configRepo, serviceRepo)
		})

		It("lists no services when none are found", func() {
			testcmd.RunCommand(cmd, []string{}, requirementsFactory)

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Getting user provided services in org", "my-org", "my-space", "my-user"},
				[]string{"OK"},
				[]string{"No user provided services found"},
			))
		})
	})

	Describe("successfully found user provided service", func() {
		BeforeEach(func() {
			ui = &testterm.FakeUI{}
			configRepo = testconfig.NewRepositoryWithDefaults()
			requirementsFactory = &testreq.FakeReqFactory{
				LoginSuccess:         true,
				TargetedSpaceSuccess: true,
				TargetedOrgSuccess:   true,
			}

			serviceRepo = &testapi.FakeUserProvidedServiceInstanceRepository{}

			service1 := models.UserProvidedServiceEntity{}
			service1.Name = "my-service-1"
			service1.Credentials = map[string]interface{}{
				"username": "admin1",
				"password": "pass1",
			}
			service1.SysLogDrainUrl = "syslog://sample.com"

			service2 := models.UserProvidedServiceEntity{}
			service2.Name = "my-service-2"

			GetSummariesResult := models.UserProvidedServiceSummary{
				Total:     2,
				Resources: []models.UserProvidedServiceEntity{service1, service2},
			}

			serviceRepo.GetSummariesReturns(GetSummariesResult, nil)

			cmd = NewListUserProvidedServices(ui, configRepo, serviceRepo)
		})

		It("lists available user provided services", func() {
			testcmd.RunCommand(cmd, []string{}, requirementsFactory)

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"OK"},
				[]string{"my-service-1", "username", "admin1", "syslog://sample.com"},
				[]string{"password", "pass1"},
				[]string{"my-service-2"},
			))
		})
	})

})
