package application_test

import (
	"cf"
	. "cf/commands/application"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
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
	appSummaryRepo := &testapi.FakeAppSummaryRepo{
		GetSummariesInCurrentSpaceApps: apps,
	}

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}

	ui := callApps(t, appSummaryRepo, reqFactory)

	assert.True(t, testcmd.CommandDidPassRequirements)

	assert.Contains(t, ui.Outputs[0], "Getting apps in")
	assert.Contains(t, ui.Outputs[0], "my-org")
	assert.Contains(t, ui.Outputs[0], "development")
	assert.Contains(t, ui.Outputs[0], "my-user")
	assert.Contains(t, ui.Outputs[1], "OK")

	assert.Contains(t, ui.Outputs[4], "Application-1")
	assert.Contains(t, ui.Outputs[4], "started")
	assert.Contains(t, ui.Outputs[4], "1/1")
	assert.Contains(t, ui.Outputs[4], "512M")
	assert.Contains(t, ui.Outputs[4], "1G")
	assert.Contains(t, ui.Outputs[4], "app1.cfapps.io, app1.example.com")

	assert.Contains(t, ui.Outputs[5], "Application-2")
	assert.Contains(t, ui.Outputs[5], "started")
	assert.Contains(t, ui.Outputs[5], "1/2")
	assert.Contains(t, ui.Outputs[5], "256M")
	assert.Contains(t, ui.Outputs[5], "1G")
	assert.Contains(t, ui.Outputs[5], "app2.cfapps.io")
}

func TestAppsRequiresLogin(t *testing.T) {
	appSummaryRepo := &testapi.FakeAppSummaryRepo{}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: false, TargetedSpaceSuccess: true}

	callApps(t, appSummaryRepo, reqFactory)

	assert.False(t, testcmd.CommandDidPassRequirements)
}

func TestAppsRequiresASelectedSpaceAndOrg(t *testing.T) {
	appSummaryRepo := &testapi.FakeAppSummaryRepo{}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: false}

	callApps(t, appSummaryRepo, reqFactory)

	assert.False(t, testcmd.CommandDidPassRequirements)
}

func callApps(t *testing.T, appSummaryRepo *testapi.FakeAppSummaryRepo, reqFactory *testreq.FakeReqFactory) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}
	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)

	config := &configuration.Configuration{
		Space:        cf.Space{Name: "development"},
		Organization: cf.Organization{Name: "my-org"},
		AccessToken:  token,
	}

	ctxt := testcmd.NewContext("apps", []string{})
	cmd := NewListApps(ui, config, appSummaryRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)

	return
}
