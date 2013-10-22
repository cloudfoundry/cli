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

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: false, TargetedSpaceSuccess: true, Application: cf.Application{}}
	callApp(t, args, reqFactory, appSummaryRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: false, Application: cf.Application{}}
	callApp(t, args, reqFactory, appSummaryRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, Application: cf.Application{}}
	callApp(t, args, reqFactory, appSummaryRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)
	assert.Equal(t, reqFactory.ApplicationName, "my-app")
}

func TestAppFailsWithUsage(t *testing.T) {
	appSummaryRepo := &testapi.FakeAppSummaryRepo{}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, Application: cf.Application{}}
	ui := callApp(t, []string{}, reqFactory, appSummaryRepo)

	assert.True(t, ui.FailedWithUsage)
	assert.False(t, testcmd.CommandDidPassRequirements)
}

func TestDisplayingAppSummary(t *testing.T) {
	reqApp := cf.Application{Name: "my-app"}

	app := cf.Application{
		State:            "started",
		Instances:        2,
		RunningInstances: 2,
		Memory:           256,
		Urls:             []string{"my-app.example.com", "foo.example.com"},
	}

	time1, err := time.Parse("Mon Jan 2 15:04:05 -0700 MST 2006", "Mon Jan 2 15:04:05 -0700 MST 2012")
	assert.NoError(t, err)

	time2, err := time.Parse("Mon Jan 2 15:04:05 -0700 MST 2006", "Mon Apr 1 15:04:05 -0700 MST 2012")
	assert.NoError(t, err)

	instances := []cf.ApplicationInstance{
		cf.ApplicationInstance{
			State:     cf.InstanceRunning,
			Since:     time1,
			CpuUsage:  1.0,
			DiskQuota: 1 * formatters.GIGABYTE,
			DiskUsage: 32 * formatters.MEGABYTE,
			MemQuota:  64 * formatters.MEGABYTE,
			MemUsage:  13 * formatters.BYTE,
		},
		cf.ApplicationInstance{
			State: cf.InstanceDown,
			Since: time2,
		},
	}
	appSummary := cf.AppSummary{App: app, Instances: instances}

	appSummaryRepo := &testapi.FakeAppSummaryRepo{GetSummarySummary: appSummary}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, Application: reqApp}
	ui := callApp(t, []string{"my-app"}, reqFactory, appSummaryRepo)

	assert.Equal(t, appSummaryRepo.GetSummaryApp.Name, "my-app")

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
	reqApp := cf.Application{Name: "my-app"}

	app := cf.Application{
		State:            "stopped",
		Instances:        2,
		RunningInstances: 0,
		Memory:           256,
		Urls:             []string{"my-app.example.com", "foo.example.com"},
	}

	appSummary := cf.AppSummary{App: app}

	appSummaryRepo := &testapi.FakeAppSummaryRepo{GetSummarySummary: appSummary, GetSummaryErrorCode: errorCode}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, Application: reqApp}
	ui := callApp(t, []string{"my-app"}, reqFactory, appSummaryRepo)

	assert.Equal(t, appSummaryRepo.GetSummaryApp.Name, "my-app")
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

func callApp(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory, appSummaryRepo *testapi.FakeAppSummaryRepo) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}
	ctxt := testcmd.NewContext("app", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)

	config := &configuration.Configuration{
		Space:        cf.Space{Name: "my-space"},
		Organization: cf.Organization{Name: "my-org"},
		AccessToken:  token,
	}

	cmd := NewShowApp(ui, config, appSummaryRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)

	return
}
