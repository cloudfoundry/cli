package service_test

import (
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

func callRenameService(t mr.TestingT, args []string, reqFactory *testreq.FakeReqFactory) (ui *testterm.FakeUI, serviceRepo *testapi.FakeServiceRepo) {
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

		fakeUI, _ := callRenameService(mr.T(), []string{}, reqFactory)
		assert.True(mr.T(), fakeUI.FailedWithUsage)

		fakeUI, _ = callRenameService(mr.T(), []string{"my-service"}, reqFactory)
		assert.True(mr.T(), fakeUI.FailedWithUsage)

		fakeUI, _ = callRenameService(mr.T(), []string{"my-service", "new-name", "extra"}, reqFactory)
		assert.True(mr.T(), fakeUI.FailedWithUsage)
	})
	It("TestRenameServiceRequirements", func() {

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: false, TargetedSpaceSuccess: true}
		callRenameService(mr.T(), []string{"my-service", "new-name"}, reqFactory)
		assert.False(mr.T(), testcmd.CommandDidPassRequirements)

		reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: false}
		callRenameService(mr.T(), []string{"my-service", "new-name"}, reqFactory)
		assert.False(mr.T(), testcmd.CommandDidPassRequirements)

		Expect(reqFactory.ServiceInstanceName).To(Equal("my-service"))
	})
	It("TestRenameService", func() {

		serviceInstance := models.ServiceInstance{}
		serviceInstance.Name = "different-name"
		serviceInstance.Guid = "different-name-guid"
		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, ServiceInstance: serviceInstance}
		ui, fakeServiceRepo := callRenameService(mr.T(), []string{"my-service", "new-name"}, reqFactory)

		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Renaming service", "different-name", "new-name", "my-org", "my-space", "my-user"},
			{"OK"},
		})

		assert.Equal(mr.T(), fakeServiceRepo.RenameServiceServiceInstance, serviceInstance)
		assert.Equal(mr.T(), fakeServiceRepo.RenameServiceNewName, "new-name")
	})
})
