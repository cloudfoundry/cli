package service_test

import (
	"cf"
	. "cf/commands/service"
	"github.com/stretchr/testify/assert"
	testcmd "testhelpers/commands"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestShowServiceRequirements(t *testing.T) {
	args := []string{"service1"}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}
	callShowService(args, reqFactory)
	assert.True(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: false}
	callShowService(args, reqFactory)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: false, TargetedSpaceSuccess: true}
	callShowService(args, reqFactory)
	assert.False(t, testcmd.CommandDidPassRequirements)

	assert.Equal(t, reqFactory.ServiceInstanceName, "service1")
}

func TestShowServiceFailsWithUsage(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}

	ui := callShowService([]string{}, reqFactory)
	assert.True(t, ui.FailedWithUsage)

	ui = callShowService([]string{"my-service"}, reqFactory)
	assert.False(t, ui.FailedWithUsage)
}

func TestShowServiceOutput(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{
		LoginSuccess:         true,
		TargetedSpaceSuccess: true,
		ServiceInstance: cf.ServiceInstance{
			Name: "service1",
			Guid: "service1-guid",
			ServicePlan: cf.ServicePlan{
				Guid: "plan-guid",
				Name: "plan-name",
				ServiceOffering: cf.ServiceOffering{
					Label:            "mysql",
					DocumentationUrl: "http://documentation.url",
					Description:      "the-description",
				},
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

func TestShowUserProvidedServiceOutput(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{
		LoginSuccess:         true,
		TargetedSpaceSuccess: true,
		ServiceInstance: cf.ServiceInstance{
			Name: "service1",
			Guid: "service1-guid",
		},
	}
	ui := callShowService([]string{"service1"}, reqFactory)

	assert.Equal(t, len(ui.Outputs), 3)
	assert.Contains(t, ui.Outputs[0], "")
	assert.Contains(t, ui.Outputs[1], "Service instance: ")
	assert.Contains(t, ui.Outputs[1], "service1")
	assert.Contains(t, ui.Outputs[2], "Service: ")
	assert.Contains(t, ui.Outputs[2], "user-provided")
}

func callShowService(args []string, reqFactory *testreq.FakeReqFactory) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("service", args)
	cmd := NewShowService(ui)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
