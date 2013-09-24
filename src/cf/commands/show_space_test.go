package commands_test

import (
	"cf"
	. "cf/commands"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestShowSpaceRequirements(t *testing.T) {
	config := &configuration.Configuration{
		Space:        cf.Space{},
		Organization: cf.Organization{},
	}

	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}
	callShowSpace([]string{}, reqFactory, config)
	assert.True(t, testhelpers.CommandDidPassRequirements)

	reqFactory = &testhelpers.FakeReqFactory{LoginSuccess: false, TargetedSpaceSuccess: true}
	callShowSpace([]string{}, reqFactory, config)
	assert.False(t, testhelpers.CommandDidPassRequirements)

	reqFactory = &testhelpers.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: false}
	callShowSpace([]string{}, reqFactory, config)
	assert.False(t, testhelpers.CommandDidPassRequirements)
}

func TestShowSpaceInfoSuccess(t *testing.T) {
	org := cf.Organization{Name: "org1"}
	apps := []cf.Application{
		cf.Application{Name: "app1", Guid: "app1-guid"},
	}
	domains := []cf.Domain{
		cf.Domain{Name: "domain1", Guid: "domain1-guid"},
	}
	services := []cf.ServiceInstance{
		cf.ServiceInstance{Name: "service1", Guid: "service1-guid"},
	}
	space := cf.Space{Name: "space1", Organization: org, Applications: apps, Domains: domains, ServiceInstances: services}
	config := &configuration.Configuration{Space: space}

	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}
	ui := callShowSpace([]string{}, reqFactory, config)
	assert.Contains(t, ui.Outputs[0], "Getting info for space")
	assert.Contains(t, ui.Outputs[0], "space1")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[2], "space1")
	assert.Contains(t, ui.Outputs[3], "organization")
	assert.Contains(t, ui.Outputs[3], "org1")
	assert.Contains(t, ui.Outputs[4], "apps")
	assert.Contains(t, ui.Outputs[4], "app1")
	assert.Contains(t, ui.Outputs[5], "domains")
	assert.Contains(t, ui.Outputs[5], "domain1")
	assert.Contains(t, ui.Outputs[6], "services")
	assert.Contains(t, ui.Outputs[6], "service1")
}

func callShowSpace(args []string, reqFactory *testhelpers.FakeReqFactory, config *configuration.Configuration) (ui *testhelpers.FakeUI) {
	ui = new(testhelpers.FakeUI)
	ctxt := testhelpers.NewContext("space", args)

	cmd := NewShowSpace(ui, config)
	testhelpers.RunCommand(cmd, ctxt, reqFactory)
	return
}
