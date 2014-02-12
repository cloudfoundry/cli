package service_test

import (
	. "cf/commands/service"
	"cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

func callShowService(args []string, reqFactory *testreq.FakeReqFactory) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("service", args)
	cmd := NewShowService(ui)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}

var _ = Describe("Testing with ginkgo", func() {
	It("TestShowServiceRequirements", func() {
		args := []string{"service1"}
		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}
		callShowService(args, reqFactory)
		assert.True(mr.T(), testcmd.CommandDidPassRequirements)

		reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: false}
		callShowService(args, reqFactory)
		assert.False(mr.T(), testcmd.CommandDidPassRequirements)

		reqFactory = &testreq.FakeReqFactory{LoginSuccess: false, TargetedSpaceSuccess: true}
		callShowService(args, reqFactory)
		assert.False(mr.T(), testcmd.CommandDidPassRequirements)

		Expect(reqFactory.ServiceInstanceName).To(Equal("service1"))
	})
	It("TestShowServiceFailsWithUsage", func() {

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}

		ui := callShowService([]string{}, reqFactory)
		assert.True(mr.T(), ui.FailedWithUsage)

		ui = callShowService([]string{"my-service"}, reqFactory)
		assert.False(mr.T(), ui.FailedWithUsage)
	})
	It("TestShowServiceOutput", func() {

		offering := models.ServiceOfferingFields{}
		offering.Label = "mysql"
		offering.DocumentationUrl = "http://documentation.url"
		offering.Description = "the-description"

		plan := models.ServicePlanFields{}
		plan.Guid = "plan-guid"
		plan.Name = "plan-name"

		serviceInstance := models.ServiceInstance{}
		serviceInstance.Name = "service1"
		serviceInstance.Guid = "service1-guid"
		serviceInstance.ServicePlan = plan
		serviceInstance.ServiceOffering = offering
		reqFactory := &testreq.FakeReqFactory{
			LoginSuccess:         true,
			TargetedSpaceSuccess: true,
			ServiceInstance:      serviceInstance,
		}
		ui := callShowService([]string{"service1"}, reqFactory)

		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Service instance:", "service1"},
			{"Service: ", "mysql"},
			{"Plan: ", "plan-name"},
			{"Description: ", "the-description"},
			{"Documentation url: ", "http://documentation.url"},
		})
	})
	It("TestShowUserProvidedServiceOutput", func() {

		serviceInstance2 := models.ServiceInstance{}
		serviceInstance2.Name = "service1"
		serviceInstance2.Guid = "service1-guid"
		reqFactory := &testreq.FakeReqFactory{
			LoginSuccess:         true,
			TargetedSpaceSuccess: true,
			ServiceInstance:      serviceInstance2,
		}
		ui := callShowService([]string{"service1"}, reqFactory)

		Expect(len(ui.Outputs)).To(Equal(3))
		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Service instance: ", "service1"},
			{"Service: ", "user-provided"},
		})
	})
})
