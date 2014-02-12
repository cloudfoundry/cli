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

func callUnbindService(args []string, reqFactory *testreq.FakeReqFactory, serviceBindingRepo api.ServiceBindingRepository) (fakeUI *testterm.FakeUI) {
	fakeUI = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("unbind-service", args)

	config := testconfig.NewRepositoryWithDefaults()

	cmd := NewUnbindService(fakeUI, config, serviceBindingRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}

var _ = Describe("Testing with ginkgo", func() {
	It("TestUnbindCommand", func() {
		app := models.Application{}
		app.Name = "my-app"
		app.Guid = "my-app-guid"
		serviceInstance := models.ServiceInstance{}
		serviceInstance.Name = "my-service"
		serviceInstance.Guid = "my-service-guid"
		reqFactory := &testreq.FakeReqFactory{
			Application:     app,
			ServiceInstance: serviceInstance,
		}
		serviceBindingRepo := &testapi.FakeServiceBindingRepo{}
		ui := callUnbindService([]string{"my-app", "my-service"}, reqFactory, serviceBindingRepo)

		Expect(reqFactory.ApplicationName).To(Equal("my-app"))
		Expect(reqFactory.ServiceInstanceName).To(Equal("my-service"))

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Unbinding app", "my-service", "my-app", "my-org", "my-space", "my-user"},
			{"OK"},
		})
		Expect(serviceBindingRepo.DeleteServiceInstance).To(Equal(serviceInstance))
		Expect(serviceBindingRepo.DeleteApplicationGuid).To(Equal("my-app-guid"))
	})
	It("TestUnbindCommandWhenBindingIsNonExistent", func() {

		app := models.Application{}
		app.Name = "my-app"
		app.Guid = "my-app-guid"
		serviceInstance := models.ServiceInstance{}
		serviceInstance.Name = "my-service"
		serviceInstance.Guid = "my-service-guid"
		reqFactory := &testreq.FakeReqFactory{
			Application:     app,
			ServiceInstance: serviceInstance,
		}
		serviceBindingRepo := &testapi.FakeServiceBindingRepo{DeleteBindingNotFound: true}
		ui := callUnbindService([]string{"my-app", "my-service"}, reqFactory, serviceBindingRepo)

		Expect(reqFactory.ApplicationName).To(Equal("my-app"))
		Expect(reqFactory.ServiceInstanceName).To(Equal("my-service"))

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Unbinding app", "my-service", "my-app"},
			{"OK"},
			{"my-service", "my-app", "did not exist"},
		})
		Expect(serviceBindingRepo.DeleteServiceInstance).To(Equal(serviceInstance))
		Expect(serviceBindingRepo.DeleteApplicationGuid).To(Equal("my-app-guid"))
	})
	It("TestUnbindCommandFailsWithUsage", func() {

		reqFactory := &testreq.FakeReqFactory{}
		serviceBindingRepo := &testapi.FakeServiceBindingRepo{}

		ui := callUnbindService([]string{"my-service"}, reqFactory, serviceBindingRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callUnbindService([]string{"my-app"}, reqFactory, serviceBindingRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callUnbindService([]string{"my-app", "my-service"}, reqFactory, serviceBindingRepo)
		Expect(ui.FailedWithUsage).To(BeFalse())
	})
})
