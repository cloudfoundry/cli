package service_test

import (
	"cf"
	. "cf/commands/service"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestShowServiceRequirements(t *testing.T) {
	args := []string{"service1"}
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}
	callShowService(args, reqFactory)
	assert.True(t, testhelpers.CommandDidPassRequirements)

	reqFactory = &testhelpers.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: false}
	callShowService(args, reqFactory)
	assert.False(t, testhelpers.CommandDidPassRequirements)

	reqFactory = &testhelpers.FakeReqFactory{LoginSuccess: false, TargetedSpaceSuccess: true}
	callShowService(args, reqFactory)
	assert.False(t, testhelpers.CommandDidPassRequirements)

	assert.Equal(t, reqFactory.ServiceInstanceName, "service1")
}

func TestShowServiceFailsWithUsage(t *testing.T) {
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}

	ui := callShowService([]string{}, reqFactory)
	assert.True(t, ui.FailedWithUsage)

	ui = callShowService([]string{"my-service"}, reqFactory)
	assert.False(t, ui.FailedWithUsage)
}

func TestShowServiceOutput(t *testing.T) {
	reqFactory := &testhelpers.FakeReqFactory{
		LoginSuccess:         true,
		TargetedSpaceSuccess: true,
		ServiceInstance: cf.ServiceInstance{
			Name:        "service1",
			Guid:        "service1-guid",
			ServicePlan: cf.ServicePlan{Name: "plan-name"},
			ServiceOffering: cf.ServiceOffering{
				Label:            "mysql",
				DocumentationUrl: "http://documentation.url",
				Description:      "the-description",
			},
		},
	}
	ui := callShowService([]string{"service1"}, reqFactory)

	assert.Contains(t, ui.Outputs[0], "")
	assert.Contains(t, ui.Outputs[1], "Service instance: ")
	assert.Contains(t, ui.Outputs[1], "service1")
	assert.Contains(t, ui.Outputs[2], "Service: ")
	assert.Contains(t, ui.Outputs[2], "mysql")
	assert.Contains(t, ui.Outputs[3], "Plan: ")
	assert.Contains(t, ui.Outputs[3], "plan-name")
	assert.Contains(t, ui.Outputs[4], "Description: ")
	assert.Contains(t, ui.Outputs[4], "the-description")
	assert.Contains(t, ui.Outputs[5], "Documentation url: ")
	assert.Contains(t, ui.Outputs[5], "http://documentation.url")
}

func callShowService(args []string, reqFactory *testhelpers.FakeReqFactory) (ui *testhelpers.FakeUI) {
	ui = new(testhelpers.FakeUI)
	ctxt := testhelpers.NewContext("service", args)
	cmd := NewShowService(ui)
	testhelpers.RunCommand(cmd, ctxt, reqFactory)
	return
}
