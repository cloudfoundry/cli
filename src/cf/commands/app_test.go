package commands_test

import (
	"cf"
	. "cf/commands"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
	"time"
)

func TestAppRequirements(t *testing.T) {
	args := []string{"my-app", "/foo"}
	appSummaryRepo := &testhelpers.FakeAppSummaryRepo{}

	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: false, TargetedSpaceSuccess: true, Application: cf.Application{}}
	callApp(args, reqFactory, appSummaryRepo)
	assert.False(t, testhelpers.CommandDidPassRequirements)

	reqFactory = &testhelpers.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: false, Application: cf.Application{}}
	callApp(args, reqFactory, appSummaryRepo)
	assert.False(t, testhelpers.CommandDidPassRequirements)

	reqFactory = &testhelpers.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, Application: cf.Application{}}
	callApp(args, reqFactory, appSummaryRepo)
	assert.True(t, testhelpers.CommandDidPassRequirements)
	assert.Equal(t, reqFactory.ApplicationName, "my-app")
}

func TestAppFailsWithUsage(t *testing.T) {
	appSummaryRepo := &testhelpers.FakeAppSummaryRepo{}
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, Application: cf.Application{}}
	ui := callApp([]string{}, reqFactory, appSummaryRepo)

	assert.True(t, ui.FailedWithUsage)
	assert.False(t, testhelpers.CommandDidPassRequirements)
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
			DiskQuota: 1024 * 1024 * 1024, // 1GB
			DiskUsage: 32 * 1024 * 1024,   //32MB
			MemQuota:  64 * 1024 * 1024,   // 64MB
			MemUsage:  13,                 // 13 B
		},
		cf.ApplicationInstance{
			State:     cf.InstanceDown,
			Since:     time2,
			CpuUsage:  1.5,
			DiskQuota: 1024 * 1024 * 1024 * 1024, // 1TB
			DiskUsage: 16 * 1024 * 1024,          //16MB
			MemQuota:  64 * 1024 * 1024,          // 64MB
			MemUsage:  13 * 1024,                 // 13 KB
		},
	}
	appSummary := cf.AppSummary{App: app, Instances: instances}

	appSummaryRepo := &testhelpers.FakeAppSummaryRepo{GetSummarySummary: appSummary}
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, Application: reqApp}
	ui := callApp([]string{"my-app"}, reqFactory, appSummaryRepo)

	assert.Equal(t, appSummaryRepo.GetSummaryApp.Name, "my-app")

	assert.Contains(t, ui.Outputs[0], "Showing health and status")
	assert.Contains(t, ui.Outputs[0], "my-app")

	assert.Contains(t, ui.Outputs[2], "health")
	assert.Contains(t, ui.Outputs[2], "running")

	assert.Contains(t, ui.Outputs[3], "usage")
	assert.Contains(t, ui.Outputs[3], "256M x 2 instances")

	assert.Contains(t, ui.Outputs[4], "urls")
	assert.Contains(t, ui.Outputs[4], "my-app.example.com, foo.example.com")

	assert.Contains(t, ui.Outputs[6], "#0")
	assert.Contains(t, ui.Outputs[6], "running")
	assert.Contains(t, ui.Outputs[6], "2012-01-02 03:04:05 PM")
	assert.Contains(t, ui.Outputs[6], "1.0%")
	assert.Contains(t, ui.Outputs[6], "13 of 64M")
	assert.Contains(t, ui.Outputs[6], "32M of 1G")

	assert.Contains(t, ui.Outputs[7], "#1")
	assert.Contains(t, ui.Outputs[7], "down")
	assert.Contains(t, ui.Outputs[7], "2012-04-01 03:04:05 PM")
	assert.Contains(t, ui.Outputs[7], "1.5%")
	assert.Contains(t, ui.Outputs[7], "13K of 64M")
	assert.Contains(t, ui.Outputs[7], "16M of 1T")
}

func callApp(args []string, reqFactory *testhelpers.FakeReqFactory, appSummaryRepo *testhelpers.FakeAppSummaryRepo) (ui *testhelpers.FakeUI) {
	ui = &testhelpers.FakeUI{}
	ctxt := testhelpers.NewContext("app", args)
	cmd := NewApp(ui, appSummaryRepo)
	testhelpers.RunCommand(cmd, ctxt, reqFactory)

	return
}
