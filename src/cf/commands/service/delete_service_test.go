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

func callDeleteService(t mr.TestingT, confirmation string, args []string, reqFactory *testreq.FakeReqFactory, serviceRepo api.ServiceRepository) (fakeUI *testterm.FakeUI) {
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
		ui := callDeleteService(mr.T(), "Y", []string{"my-service"}, reqFactory, serviceRepo)

		testassert.SliceContains(mr.T(), ui.Prompts, testassert.Lines{
			{"Are you sure"},
		})

		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
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
		ui := callDeleteService(mr.T(), "Yes", []string{"my-service"}, reqFactory, serviceRepo)

		testassert.SliceContains(mr.T(), ui.Prompts, testassert.Lines{{"Are you sure"}})

		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Deleting service", "my-service"},
			{"OK"},
		})

		Expect(serviceRepo.DeleteServiceServiceInstance).To(Equal(serviceInstance))
	})
	It("TestDeleteServiceCommandOnNonExistentService", func() {

		reqFactory := &testreq.FakeReqFactory{}
		serviceRepo := &testapi.FakeServiceRepo{FindInstanceByNameNotFound: true}
		ui := callDeleteService(mr.T(), "", []string{"-f", "my-service"}, reqFactory, serviceRepo)

		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Deleting service", "my-service"},
			{"OK"},
			{"my-service", "does not exist"},
		})
	})
	It("TestDeleteServiceCommandFailsWithUsage", func() {

		reqFactory := &testreq.FakeReqFactory{}
		serviceRepo := &testapi.FakeServiceRepo{}

		ui := callDeleteService(mr.T(), "", []string{"-f"}, reqFactory, serviceRepo)
		assert.True(mr.T(), ui.FailedWithUsage)

		ui = callDeleteService(mr.T(), "", []string{"-f", "my-service"}, reqFactory, serviceRepo)
		assert.False(mr.T(), ui.FailedWithUsage)
	})
	It("TestDeleteServiceForceFlagSkipsConfirmation", func() {

		reqFactory := &testreq.FakeReqFactory{}
		serviceRepo := &testapi.FakeServiceRepo{}

		ui := callDeleteService(mr.T(), "", []string{"-f", "foo.com"}, reqFactory, serviceRepo)

		Expect(len(ui.Prompts)).To(Equal(0))
		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Deleting service", "foo.com"},
			{"OK"},
		})
	})
})
