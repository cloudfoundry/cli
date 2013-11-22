package application_test

import (
	"cf"
	. "cf/commands/application"
	"cf/configuration"
	"cf/formatters"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
	"time"
)

func TestAppRequirements(t *testing.T) {
	args := []string{"my-app", "/foo"}
	appSummaryRepo := &testapi.FakeAppSummaryRepo{}
	appInstancesRepo := &testapi.FakeAppInstancesRepo{}

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: false, TargetedSpaceSuccess: true, Application: cf.Application{}}
	callApp(t, args, reqFactory, appSummaryRepo, appInstancesRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: false, Application: cf.Application{}}
	callApp(t, args, reqFactory, appSummaryRepo, appInstancesRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, Application: cf.Application{}}
	callApp(t, args, reqFactory, appSummaryRepo, appInstancesRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)
	assert.Equal(t, reqFactory.ApplicationName, "my-app")
}

func TestAppFailsWithUsage(t *testing.T) {
	appSummaryRepo := &testapi.FakeAppSummaryRepo{}
	appInstancesRepo := &testapi.FakeAppInstancesRepo{}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, Application: cf.Application{}}
	ui := callApp(t, []string{}, reqFactory, appSummaryRepo, appInstancesRepo)

	assert.True(t, ui.FailedWithUsage)
	assert.False(t, testcmd.CommandDidPassRequirements)
}

func TestDisplayingAppSummary(t *testing.T) {
	reqApp := cf.Application{}
	reqApp.Name = "my-app"
	reqApp.Guid = "my-app-guid"

	route1 := cf.RouteSummary{}
	route1.Host = "my-app"

	domain := cf.DomainFields{}
	domain.Name = "example.com"
	route1.Domain = domain

	route2 := cf.RouteSummary{}
	route2.Host = "foo"
	domain2 := cf.DomainFields{}
	domain2.Name = "example.com"
	route2.Domain = domain2

	appSummary := cf.AppSummary{}
	appSummary.State = "started"
	appSummary.InstanceCount = 2
	appSummary.RunningInstances = 2
	appSummary.Memory = 256
	appSummary.RouteSummaries = []cf.RouteSummary{route1, route2}

	time1, err := time.Parse("Mon Jan 2 15:04:05 -0700 MST 2006", "Mon Jan 2 15:04:05 -0700 MST 2012")
	assert.NoError(t, err)

	time2, err := time.Parse("Mon Jan 2 15:04:05 -0700 MST 2006", "Mon Apr 1 15:04:05 -0700 MST 2012")
	assert.NoError(t, err)

	appInstance := cf.AppInstanceFields{}
	appInstance.State = cf.InstanceRunning
	appInstance.Since = time1
	appInstance.CpuUsage = 1.0
	appInstance.DiskQuota = 1 * formatters.GIGABYTE
	appInstance.DiskUsage = 32 * formatters.MEGABYTE
	appInstance.MemQuota = 64 * formatters.MEGABYTE
	appInstance.MemUsage = 13 * formatters.BYTE

	appInstance2 := cf.AppInstanceFields{}
	appInstance2.State = cf.InstanceDown
	appInstance2.Since = time2

	instances := []cf.AppInstanceFields{appInstance, appInstance2}

	appSummaryRepo := &testapi.FakeAppSummaryRepo{GetSummarySummary: appSummary}
	appInstancesRepo := &testapi.FakeAppInstancesRepo{GetInstancesResponses: [][]cf.AppInstanceFields{instances}}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, Application: reqApp}
	ui := callApp(t, []string{"my-app"}, reqFactory, appSummaryRepo, appInstancesRepo)

	assert.Equal(t, appSummaryRepo.GetSummaryAppGuid, "my-app-guid")

	assert.Contains(t, ui.Outputs[0], "Showing health and status")
	assert.Contains(t, ui.Outputs[0], "my-app")

	assert.Contains(t, ui.Outputs[2], "state")
	assert.Contains(t, ui.Outputs[2], "started")

	assert.Contains(t, ui.Outputs[3], "instances")
	assert.Contains(t, ui.Outputs[3], "2/2")

	assert.Contains(t, ui.Outputs[4], "usage")
	assert.Contains(t, ui.Outputs[4], "256M x 2 instances")

	assert.Contains(t, ui.Outputs[5], "urls")
	assert.Contains(t, ui.Outputs[5], "my-app.example.com, foo.example.com")

	assert.Contains(t, ui.Outputs[7], "#0")
	assert.Contains(t, ui.Outputs[7], "running")
	assert.Contains(t, ui.Outputs[7], "2012-01-02 03:04:05 PM")
	assert.Contains(t, ui.Outputs[7], "1.0%")
	assert.Contains(t, ui.Outputs[7], "13 of 64M")
	assert.Contains(t, ui.Outputs[7], "32M of 1G")

	assert.Contains(t, ui.Outputs[8], "#1")
	assert.Contains(t, ui.Outputs[8], "down")
	assert.Contains(t, ui.Outputs[8], "2012-04-01 03:04:05 PM")
	assert.Contains(t, ui.Outputs[8], "0%")
	assert.Contains(t, ui.Outputs[8], "0 of 0")
	assert.Contains(t, ui.Outputs[8], "0 of 0")
}

func TestDisplayingStoppedAppSummary(t *testing.T) {
	testDisplayingAppSummaryWithErrorCode(t, cf.APP_STOPPED)
}

func TestDisplayingNotStagedAppSummary(t *testing.T) {
	testDisplayingAppSummaryWithErrorCode(t, cf.APP_NOT_STAGED)
}

func testDisplayingAppSummaryWithErrorCode(t *testing.T, errorCode string) {
	reqApp := cf.Application{}
	reqApp.Name = "my-app"
	reqApp.Guid = "my-app-guid"

	domain3 := cf.DomainFields{}
	domain3.Name = "example.com"
	domain4 := cf.DomainFields{}
	domain4.Name = "example.com"

	route1 := cf.RouteSummary{}
	route1.Host = "my-app"
	route1.Domain = domain3

	route2 := cf.RouteSummary{}
	route2.Host = "foo"
	route2.Domain = domain4

	routes := []cf.RouteSummary{
		route1,
		route2,
	}

	app := cf.ApplicationFields{}
	app.State = "stopped"
	app.InstanceCount = 2
	app.RunningInstances = 0
	app.Memory = 256

	appSummary := cf.AppSummary{}
	appSummary.ApplicationFields = app
	appSummary.RouteSummaries = routes

	appSummaryRepo := &testapi.FakeAppSummaryRepo{GetSummarySummary: appSummary, GetSummaryErrorCode: errorCode}
	appInstancesRepo := &testapi.FakeAppInstancesRepo{}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, Application: reqApp}
	ui := callApp(t, []string{"my-app"}, reqFactory, appSummaryRepo, appInstancesRepo)

	assert.Equal(t, appSummaryRepo.GetSummaryAppGuid, "my-app-guid")
	assert.Equal(t, appInstancesRepo.GetInstancesAppGuid, "my-app-guid")
	assert.Equal(t, len(ui.Outputs), 6)

	assert.Contains(t, ui.Outputs[0], "Showing health and status")
	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[0], "my-org")
	assert.Contains(t, ui.Outputs[0], "my-space")
	assert.Contains(t, ui.Outputs[0], "my-user")

	assert.Contains(t, ui.Outputs[2], "state")
	assert.Contains(t, ui.Outputs[2], "stopped")

	assert.Contains(t, ui.Outputs[3], "instances")
	assert.Contains(t, ui.Outputs[3], "0/2")

	assert.Contains(t, ui.Outputs[4], "usage")
	assert.Contains(t, ui.Outputs[4], "256M x 2 instances")

	assert.Contains(t, ui.Outputs[5], "urls")
	assert.Contains(t, ui.Outputs[5], "my-app.example.com, foo.example.com")
}

func callApp(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory, appSummaryRepo *testapi.FakeAppSummaryRepo, appInstancesRepo *testapi.FakeAppInstancesRepo) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}
	ctxt := testcmd.NewContext("app", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)
	space := cf.SpaceFields{}
	space.Name = "my-space"
	org := cf.OrganizationFields{}
	org.Name = "my-org"
	config := &configuration.Configuration{
		SpaceFields:        space,
		OrganizationFields: org,
		AccessToken:        token,
	}

	cmd := NewShowApp(ui, config, appSummaryRepo, appInstancesRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)

	return
}
