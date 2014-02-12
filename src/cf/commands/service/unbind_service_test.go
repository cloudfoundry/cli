package service_test

import (
	"cf/api"
	. "cf/commands/service"
	"cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

func callUnbindService(t mr.TestingT, args []string, reqFactory *testreq.FakeReqFactory, serviceBindingRepo api.ServiceBindingRepository) (fakeUI *testterm.FakeUI) {
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
		ui := callUnbindService(mr.T(), []string{"my-app", "my-service"}, reqFactory, serviceBindingRepo)

		Expect(reqFactory.ApplicationName).To(Equal("my-app"))
		assert.Equal(mr.T(), reqFactory.ServiceInstanceName, "my-service")

		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Unbinding app", "my-service", "my-app", "my-org", "my-space", "my-user"},
			{"OK"},
		})
		assert.Equal(mr.T(), serviceBindingRepo.DeleteServiceInstance, serviceInstance)
		assert.Equal(mr.T(), serviceBindingRepo.DeleteApplicationGuid, "my-app-guid")
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
		ui := callUnbindService(mr.T(), []string{"my-app", "my-service"}, reqFactory, serviceBindingRepo)

		assert.Equal(mr.T(), reqFactory.ApplicationName, "my-app")
		assert.Equal(mr.T(), reqFactory.ServiceInstanceName, "my-service")

		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Unbinding app", "my-service", "my-app"},
			{"OK"},
			{"my-service", "my-app", "did not exist"},
		})
		assert.Equal(mr.T(), serviceBindingRepo.DeleteServiceInstance, serviceInstance)
		assert.Equal(mr.T(), serviceBindingRepo.DeleteApplicationGuid, "my-app-guid")
	})
	It("TestUnbindCommandFailsWithUsage", func() {

		reqFactory := &testreq.FakeReqFactory{}
		serviceBindingRepo := &testapi.FakeServiceBindingRepo{}

		ui := callUnbindService(mr.T(), []string{"my-service"}, reqFactory, serviceBindingRepo)
		assert.True(mr.T(), ui.FailedWithUsage)

		ui = callUnbindService(mr.T(), []string{"my-app"}, reqFactory, serviceBindingRepo)
		assert.True(mr.T(), ui.FailedWithUsage)

		ui = callUnbindService(mr.T(), []string{"my-app", "my-service"}, reqFactory, serviceBindingRepo)
		assert.False(mr.T(), ui.FailedWithUsage)
	})
})
