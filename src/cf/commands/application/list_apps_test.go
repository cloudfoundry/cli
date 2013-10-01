package application_test

import (
	"cf"
	. "cf/commands/application"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestApps(t *testing.T) {
	app1Urls := []string{"app1.cfapps.io", "app1.example.com"}
	app2Urls := []string{"app2.cfapps.io"}

	apps := []cf.Application{
		cf.Application{Name: "Application-1", State: "started", Instances: 1, Memory: 512, Urls: app1Urls},
		cf.Application{Name: "Application-2", State: "started", Instances: 2, Memory: 256, Urls: app2Urls},
	}
	spaceRepo := &testhelpers.FakeSpaceRepository{
		CurrentSpace: cf.Space{Name: "development", Guid: "development-guid"},
		SummarySpace: cf.Space{Applications: apps},
	}

	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}

	ui := callApps(spaceRepo, reqFactory)

	assert.True(t, testhelpers.CommandDidPassRequirements)

	assert.Contains(t, ui.Outputs[0], "Getting applications in development")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[3], "Application-1")
	assert.Contains(t, ui.Outputs[3], "running")
	assert.Contains(t, ui.Outputs[3], "1 x 512M")
	assert.Contains(t, ui.Outputs[3], "app1.cfapps.io, app1.example.com")

	assert.Contains(t, ui.Outputs[4], "Application-2")
	assert.Contains(t, ui.Outputs[4], "running")
	assert.Contains(t, ui.Outputs[4], "2 x 256M")
	assert.Contains(t, ui.Outputs[4], "app2.cfapps.io")
}

func TestAppsRequiresLogin(t *testing.T) {
	spaceRepo := &testhelpers.FakeSpaceRepository{}
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: false, TargetedSpaceSuccess: true}

	callApps(spaceRepo, reqFactory)

	assert.False(t, testhelpers.CommandDidPassRequirements)
}

func TestAppsRequiresASelectedSpaceAndOrg(t *testing.T) {
	spaceRepo := &testhelpers.FakeSpaceRepository{}
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: false}

	callApps(spaceRepo, reqFactory)

	assert.False(t, testhelpers.CommandDidPassRequirements)
}

func callApps(spaceRepo *testhelpers.FakeSpaceRepository, reqFactory *testhelpers.FakeReqFactory) (ui *testhelpers.FakeUI) {
	ui = &testhelpers.FakeUI{}

	ctxt := testhelpers.NewContext("apps", []string{})
	cmd := NewListApps(ui, spaceRepo)
	testhelpers.RunCommand(cmd, ctxt, reqFactory)

	return
}
