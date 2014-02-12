package service_test

import (
	"cf/api"
	. "cf/commands/service"
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

func callUpdateUserProvidedService(args []string, reqFactory *testreq.FakeReqFactory, userProvidedServiceInstanceRepo api.UserProvidedServiceInstanceRepository) (fakeUI *testterm.FakeUI) {
	fakeUI = &testterm.FakeUI{}
	ctxt := testcmd.NewContext("update-user-provided-service", args)

	config := testconfig.NewRepositoryWithDefaults()

	cmd := NewUpdateUserProvidedService(fakeUI, config, userProvidedServiceInstanceRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}

var _ = Describe("Testing with ginkgo", func() {
	It("TestUpdateUserProvidedServiceFailsWithUsage", func() {
		reqFactory := &testreq.FakeReqFactory{}
		userProvidedServiceInstanceRepo := &testapi.FakeUserProvidedServiceInstanceRepo{}

		ui := callUpdateUserProvidedService([]string{}, reqFactory, userProvidedServiceInstanceRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callUpdateUserProvidedService([]string{"foo"}, reqFactory, userProvidedServiceInstanceRepo)
		Expect(ui.FailedWithUsage).To(BeFalse())
	})

	It("TestUpdateUserProvidedServiceRequirements", func() {
		args := []string{"service-name"}
		reqFactory := &testreq.FakeReqFactory{}
		userProvidedServiceInstanceRepo := &testapi.FakeUserProvidedServiceInstanceRepo{}

		reqFactory.LoginSuccess = false
		callUpdateUserProvidedService(args, reqFactory, userProvidedServiceInstanceRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		reqFactory.LoginSuccess = true
		callUpdateUserProvidedService(args, reqFactory, userProvidedServiceInstanceRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())

		Expect(reqFactory.ServiceInstanceName).To(Equal("service-name"))
	})
	It("TestUpdateUserProvidedServiceWhenNoFlagsArePresent", func() {

		args := []string{"service-name"}
		serviceInstance := models.ServiceInstance{}
		serviceInstance.Name = "found-service-name"
		reqFactory := &testreq.FakeReqFactory{
			LoginSuccess:    true,
			ServiceInstance: serviceInstance,
		}
		repo := &testapi.FakeUserProvidedServiceInstanceRepo{}
		ui := callUpdateUserProvidedService(args, reqFactory, repo)

		testassert.SliceContains(GinkgoT(), ui.Outputs, testassert.Lines{
			{"Updating user provided service", "found-service-name", "my-org", "my-space", "my-user"},
			{"OK"},
			{"No changes"},
		})
	})
	It("TestUpdateUserProvidedServiceWithJson", func() {

		args := []string{"-p", `{"foo":"bar"}`, "-l", "syslog://example.com", "service-name"}
		serviceInstance := models.ServiceInstance{}
		serviceInstance.Name = "found-service-name"
		reqFactory := &testreq.FakeReqFactory{
			LoginSuccess:    true,
			ServiceInstance: serviceInstance,
		}
		repo := &testapi.FakeUserProvidedServiceInstanceRepo{}
		ui := callUpdateUserProvidedService(args, reqFactory, repo)

		testassert.SliceContains(GinkgoT(), ui.Outputs, testassert.Lines{
			{"Updating user provided service", "found-service-name", "my-org", "my-space", "my-user"},
			{"OK"},
			{"TIP"},
		})
		Expect(repo.UpdateServiceInstance.Name).To(Equal(serviceInstance.Name))
		Expect(repo.UpdateServiceInstance.Params).To(Equal(map[string]string{"foo": "bar"}))
		Expect(repo.UpdateServiceInstance.SysLogDrainUrl).To(Equal("syslog://example.com"))
	})
	It("TestUpdateUserProvidedServiceWithoutJson", func() {

		args := []string{"-l", "syslog://example.com", "service-name"}
		serviceInstance := models.ServiceInstance{}
		serviceInstance.Name = "found-service-name"
		reqFactory := &testreq.FakeReqFactory{
			LoginSuccess:    true,
			ServiceInstance: serviceInstance,
		}
		repo := &testapi.FakeUserProvidedServiceInstanceRepo{}
		ui := callUpdateUserProvidedService(args, reqFactory, repo)

		testassert.SliceContains(GinkgoT(), ui.Outputs, testassert.Lines{
			{"Updating user provided service"},
			{"OK"},
		})
	})
	It("TestUpdateUserProvidedServiceWithInvalidJson", func() {

		args := []string{"-p", `{"foo":"ba`, "service-name"}
		serviceInstance := models.ServiceInstance{}
		serviceInstance.Name = "found-service-name"
		reqFactory := &testreq.FakeReqFactory{
			LoginSuccess:    true,
			ServiceInstance: serviceInstance,
		}
		userProvidedServiceInstanceRepo := &testapi.FakeUserProvidedServiceInstanceRepo{}

		ui := callUpdateUserProvidedService(args, reqFactory, userProvidedServiceInstanceRepo)

		Expect(userProvidedServiceInstanceRepo.UpdateServiceInstance).NotTo(Equal(serviceInstance))

		testassert.SliceContains(GinkgoT(), ui.Outputs, testassert.Lines{
			{"FAILED"},
			{"JSON is invalid"},
		})
	})
	It("TestUpdateUserProvidedServiceWithAServiceInstanceThatIsNotUserProvided", func() {

		args := []string{"-p", `{"foo":"bar"}`, "service-name"}
		plan := models.ServicePlanFields{}
		plan.Guid = "my-plan-guid"
		serviceInstance := models.ServiceInstance{}
		serviceInstance.Name = "found-service-name"
		serviceInstance.ServicePlan = plan

		reqFactory := &testreq.FakeReqFactory{
			LoginSuccess:    true,
			ServiceInstance: serviceInstance,
		}
		userProvidedServiceInstanceRepo := &testapi.FakeUserProvidedServiceInstanceRepo{}

		ui := callUpdateUserProvidedService(args, reqFactory, userProvidedServiceInstanceRepo)

		Expect(userProvidedServiceInstanceRepo.UpdateServiceInstance).NotTo(Equal(serviceInstance))

		testassert.SliceContains(GinkgoT(), ui.Outputs, testassert.Lines{
			{"FAILED"},
			{"Service Instance is not user provided"},
		})
	})
})
