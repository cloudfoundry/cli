package service_test

import (
	"cf/api"
	. "cf/commands/service"
	"cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	mr "github.com/tjarratt/mr_t"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

func callBindService(t mr.TestingT, args []string, reqFactory *testreq.FakeReqFactory, serviceBindingRepo api.ServiceBindingRepository) (fakeUI *testterm.FakeUI) {
	fakeUI = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("bind-service", args)

	config := testconfig.NewRepositoryWithDefaults()

	cmd := NewBindService(fakeUI, config, serviceBindingRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}

var _ = Describe("Testing with ginkgo", func() {
	It("TestBindCommand", func() {
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
		ui := callBindService(mr.T(), []string{"my-app", "my-service"}, reqFactory, serviceBindingRepo)

		Expect(reqFactory.ApplicationName).To(Equal("my-app"))
		Expect(reqFactory.ServiceInstanceName).To(Equal("my-service"))

		Expect(len(ui.Outputs)).To(Equal(3))
		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Binding service", "my-service", "my-app", "my-org", "my-space", "my-user"},
			{"OK"},
			{"TIP"},
		})
		Expect(serviceBindingRepo.CreateServiceInstanceGuid).To(Equal("my-service-guid"))
		Expect(serviceBindingRepo.CreateApplicationGuid).To(Equal("my-app-guid"))
	})
	It("TestBindCommandIfServiceIsAlreadyBound", func() {

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
		serviceBindingRepo := &testapi.FakeServiceBindingRepo{CreateErrorCode: "90003"}
		ui := callBindService(mr.T(), []string{"my-app", "my-service"}, reqFactory, serviceBindingRepo)

		Expect(len(ui.Outputs)).To(Equal(3))
		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Binding service"},
			{"OK"},
			{"my-app", "is already bound", "my-service"},
		})
	})
	It("TestBindCommandFailsWithUsage", func() {

		reqFactory := &testreq.FakeReqFactory{}
		serviceBindingRepo := &testapi.FakeServiceBindingRepo{}

		ui := callBindService(mr.T(), []string{"my-service"}, reqFactory, serviceBindingRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callBindService(mr.T(), []string{"my-app"}, reqFactory, serviceBindingRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callBindService(mr.T(), []string{"my-app", "my-service"}, reqFactory, serviceBindingRepo)
		Expect(ui.FailedWithUsage).To(BeFalse())
	})
})
