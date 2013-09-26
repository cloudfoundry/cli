package commands_test

import (
	"cf"
	"cf/api"
	. "cf/commands"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestShowServiceRequirements(t *testing.T) {
	serviceRepo := &testhelpers.FakeServiceRepo{}
	args := []string{"service1"}
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}
	callShowService(args, reqFactory, serviceRepo)
	assert.True(t, testhelpers.CommandDidPassRequirements)

	reqFactory = &testhelpers.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: false}
	callShowService(args, reqFactory, serviceRepo)
	assert.False(t, testhelpers.CommandDidPassRequirements)

	reqFactory = &testhelpers.FakeReqFactory{LoginSuccess: false, TargetedSpaceSuccess: true}
	callShowService(args, reqFactory, serviceRepo)
	assert.False(t, testhelpers.CommandDidPassRequirements)
}

func TestShowServiceFailsWithUsage(t *testing.T) {
	serviceRepo := &testhelpers.FakeServiceRepo{}
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}

	ui := callShowService([]string{}, reqFactory, serviceRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callShowService([]string{"my-service"}, reqFactory, serviceRepo)
	assert.False(t, ui.FailedWithUsage)
}

func TestShowServiceOutput(t *testing.T) {
	serviceRepo := &testhelpers.FakeServiceRepo{
		FindInstanceByNameServiceInstance: cf.ServiceInstance{
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
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}
	ui := callShowService([]string{"service1"}, reqFactory, serviceRepo)

	assert.Contains(t, ui.Outputs[0], "Getting service instance")
	assert.Contains(t, ui.Outputs[0], "service1")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[2], "")
	assert.Contains(t, ui.Outputs[3], "service instance: ")
	assert.Contains(t, ui.Outputs[3], "service1")
	assert.Contains(t, ui.Outputs[4], "service: ")
	assert.Contains(t, ui.Outputs[4], "mysql")
	assert.Contains(t, ui.Outputs[5], "plan: ")
	assert.Contains(t, ui.Outputs[5], "plan-name")
	assert.Contains(t, ui.Outputs[6], "description: ")
	assert.Contains(t, ui.Outputs[6], "the-description")
	assert.Contains(t, ui.Outputs[7], "documentation url: ")
	assert.Contains(t, ui.Outputs[7], "http://documentation.url")
}

func callShowService(args []string, reqFactory *testhelpers.FakeReqFactory, serviceRepo api.ServiceRepository) (ui *testhelpers.FakeUI) {
	ui = new(testhelpers.FakeUI)
	ctxt := testhelpers.NewContext("service", args)
	cmd := NewShowService(ui, serviceRepo)
	testhelpers.RunCommand(cmd, ctxt, reqFactory)
	return
}
