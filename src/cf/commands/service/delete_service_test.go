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

func callDeleteService(confirmation string, args []string, reqFactory *testreq.FakeReqFactory, serviceRepo api.ServiceRepository) (fakeUI *testterm.FakeUI) {
	fakeUI = &testterm.FakeUI{
		Inputs: []string{confirmation},
	}
	ctxt := testcmd.NewContext("delete-service", args)

	config := testconfig.NewRepositoryWithDefaults()

	cmd := NewDeleteService(fakeUI, config, serviceRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}

var _ = Describe("Testing with ginkgo", func() {
	It("TestDeleteServiceCommandWithY", func() {
		serviceInstance := models.ServiceInstance{}
		serviceInstance.Name = "my-service"
		serviceInstance.Guid = "my-service-guid"
		reqFactory := &testreq.FakeReqFactory{}
		serviceRepo := &testapi.FakeServiceRepo{FindInstanceByNameServiceInstance: serviceInstance}
		ui := callDeleteService("Y", []string{"my-service"}, reqFactory, serviceRepo)

		testassert.SliceContains(ui.Prompts, testassert.Lines{
			{"Are you sure"},
		})

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Deleting service", "my-service", "my-org", "my-space", "my-user"},
			{"OK"},
		})

		Expect(serviceRepo.DeleteServiceServiceInstance).To(Equal(serviceInstance))
	})
	It("TestDeleteServiceCommandWithYes", func() {

		serviceInstance := models.ServiceInstance{}
		serviceInstance.Name = "my-service"
		serviceInstance.Guid = "my-service-guid"
		reqFactory := &testreq.FakeReqFactory{}
		serviceRepo := &testapi.FakeServiceRepo{FindInstanceByNameServiceInstance: serviceInstance}
		ui := callDeleteService("Yes", []string{"my-service"}, reqFactory, serviceRepo)

		testassert.SliceContains(ui.Prompts, testassert.Lines{{"Are you sure"}})

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Deleting service", "my-service"},
			{"OK"},
		})

		Expect(serviceRepo.DeleteServiceServiceInstance).To(Equal(serviceInstance))
	})
	It("TestDeleteServiceCommandOnNonExistentService", func() {

		reqFactory := &testreq.FakeReqFactory{}
		serviceRepo := &testapi.FakeServiceRepo{FindInstanceByNameNotFound: true}
		ui := callDeleteService("", []string{"-f", "my-service"}, reqFactory, serviceRepo)

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Deleting service", "my-service"},
			{"OK"},
			{"my-service", "does not exist"},
		})
	})
	It("TestDeleteServiceCommandFailsWithUsage", func() {

		reqFactory := &testreq.FakeReqFactory{}
		serviceRepo := &testapi.FakeServiceRepo{}

		ui := callDeleteService("", []string{"-f"}, reqFactory, serviceRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callDeleteService("", []string{"-f", "my-service"}, reqFactory, serviceRepo)
		Expect(ui.FailedWithUsage).To(BeFalse())
	})
	It("TestDeleteServiceForceFlagSkipsConfirmation", func() {

		reqFactory := &testreq.FakeReqFactory{}
		serviceRepo := &testapi.FakeServiceRepo{}

		ui := callDeleteService("", []string{"-f", "foo.com"}, reqFactory, serviceRepo)

		Expect(len(ui.Prompts)).To(Equal(0))
		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Deleting service", "foo.com"},
			{"OK"},
		})
	})
})
