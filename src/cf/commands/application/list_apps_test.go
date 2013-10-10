package application_test

import (
	"cf"
	. "cf/commands/application"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestApps(t *testing.T) {
	app1Urls := []string{"app1.cfapps.io", "app1.example.com"}
	app2Urls := []string{"app2.cfapps.io"}

	apps := []cf.Application{
		cf.Application{Name: "Application-1", State: "started", RunningInstances: 1, Instances: 1, Memory: 512, DiskQuota: 1024, Urls: app1Urls},
		cf.Application{Name: "Application-2", State: "started", RunningInstances: 1, Instances: 2, Memory: 256, DiskQuota: 1024, Urls: app2Urls},
	}
	spaceRepo := &testapi.FakeSpaceRepository{
		CurrentSpace: cf.Space{Name: "development", Guid: "development-guid"},
		SummarySpace: cf.Space{Applications: apps},
	}

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}

	ui := callApps(spaceRepo, reqFactory)

	assert.True(t, testcmd.CommandDidPassRequirements)

	assert.Contains(t, ui.Outputs[0], "Getting apps in development")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[3], "Application-1")
	assert.Contains(t, ui.Outputs[3], "started")
	assert.Contains(t, ui.Outputs[3], "1/1")
	assert.Contains(t, ui.Outputs[3], "512M")
	assert.Contains(t, ui.Outputs[3], "1G")
	assert.Contains(t, ui.Outputs[3], "app1.cfapps.io, app1.example.com")

	assert.Contains(t, ui.Outputs[4], "Application-2")
	assert.Contains(t, ui.Outputs[4], "started")
	assert.Contains(t, ui.Outputs[4], "1/2")
	assert.Contains(t, ui.Outputs[4], "256M")
	assert.Contains(t, ui.Outputs[4], "1G")
	assert.Contains(t, ui.Outputs[4], "app2.cfapps.io")
}

func TestAppsRequiresLogin(t *testing.T) {
	spaceRepo := &testapi.FakeSpaceRepository{}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: false, TargetedSpaceSuccess: true}

	callApps(spaceRepo, reqFactory)

	assert.False(t, testcmd.CommandDidPassRequirements)
}

func TestAppsRequiresASelectedSpaceAndOrg(t *testing.T) {
	spaceRepo := &testapi.FakeSpaceRepository{}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: false}

	callApps(spaceRepo, reqFactory)

	assert.False(t, testcmd.CommandDidPassRequirements)
}

func callApps(spaceRepo *testapi.FakeSpaceRepository, reqFactory *testreq.FakeReqFactory) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}

	ctxt := testcmd.NewContext("apps", []string{})
	cmd := NewListApps(ui, spaceRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)

	return
}
