package space_test

import (
	"cf"
	. "cf/commands/space"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestShowSpaceRequirements(t *testing.T) {
	config := &configuration.Configuration{}

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}
	callShowSpace(t, []string{}, reqFactory, config)
	assert.True(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: false, TargetedSpaceSuccess: true}
	callShowSpace(t, []string{}, reqFactory, config)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: false}
	callShowSpace(t, []string{}, reqFactory, config)
	assert.False(t, testcmd.CommandDidPassRequirements)
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

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}
	ui := callShowSpace(t, []string{}, reqFactory, config)
	assert.Contains(t, ui.Outputs[0], "Getting info for space")
	assert.Contains(t, ui.Outputs[0], "space1")
	assert.Contains(t, ui.Outputs[0], "my-org")
	assert.Contains(t, ui.Outputs[0], "my-user")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[2], "space1")
	assert.Contains(t, ui.Outputs[3], "Org")
	assert.Contains(t, ui.Outputs[3], "org1")
	assert.Contains(t, ui.Outputs[4], "Apps")
	assert.Contains(t, ui.Outputs[4], "app1")
	assert.Contains(t, ui.Outputs[5], "Domains")
	assert.Contains(t, ui.Outputs[5], "domain1")
	assert.Contains(t, ui.Outputs[6], "Services")
	assert.Contains(t, ui.Outputs[6], "service1")
}

func callShowSpace(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory, config *configuration.Configuration) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("space", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)

	config.Organization = cf.Organization{Name: "my-org"}
	config.AccessToken = token

	cmd := NewShowSpace(ui, config)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
