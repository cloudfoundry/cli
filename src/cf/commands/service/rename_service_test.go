package service_test

import (
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

func callRenameService(args []string, reqFactory *testreq.FakeReqFactory) (ui *testterm.FakeUI, serviceRepo *testapi.FakeServiceRepo) {
	ui = &testterm.FakeUI{}
	serviceRepo = &testapi.FakeServiceRepo{}

	config := testconfig.NewRepositoryWithDefaults()

	cmd := NewRenameService(ui, config, serviceRepo)
	ctxt := testcmd.NewContext("rename-service", args)

	testcmd.RunCommand(cmd, ctxt, reqFactory)

	return
}

var _ = Describe("Testing with ginkgo", func() {
	It("TestRenameServiceFailsWithUsage", func() {
		reqFactory := &testreq.FakeReqFactory{}

		fakeUI, _ := callRenameService([]string{}, reqFactory)
		Expect(fakeUI.FailedWithUsage).To(BeTrue())

		fakeUI, _ = callRenameService([]string{"my-service"}, reqFactory)
		Expect(fakeUI.FailedWithUsage).To(BeTrue())

		fakeUI, _ = callRenameService([]string{"my-service", "new-name", "extra"}, reqFactory)
		Expect(fakeUI.FailedWithUsage).To(BeTrue())
	})
	It("TestRenameServiceRequirements", func() {

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: false, TargetedSpaceSuccess: true}
		callRenameService([]string{"my-service", "new-name"}, reqFactory)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: false}
		callRenameService([]string{"my-service", "new-name"}, reqFactory)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		Expect(reqFactory.ServiceInstanceName).To(Equal("my-service"))
	})
	It("TestRenameService", func() {

		serviceInstance := models.ServiceInstance{}
		serviceInstance.Name = "different-name"
		serviceInstance.Guid = "different-name-guid"
		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, ServiceInstance: serviceInstance}
		ui, fakeServiceRepo := callRenameService([]string{"my-service", "new-name"}, reqFactory)

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Renaming service", "different-name", "new-name", "my-org", "my-space", "my-user"},
			{"OK"},
		})

		Expect(fakeServiceRepo.RenameServiceServiceInstance).To(Equal(serviceInstance))
		Expect(fakeServiceRepo.RenameServiceNewName).To(Equal("new-name"))
	})
})
