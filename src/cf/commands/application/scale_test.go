package application_test

import (
	"cf"
	"cf/api"
	. "cf/commands/application"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestScaleRequirements(t *testing.T) {
	args := []string{"-d", "1G", "my-app"}
	reqFactory, restarter, appRepo := getScaleDependencies()

	reqFactory.LoginSuccess = false
	reqFactory.TargetedSpaceSuccess = true
	callScale(args, reqFactory, restarter, appRepo)
	assert.False(t, testhelpers.CommandDidPassRequirements)

	reqFactory.LoginSuccess = true
	reqFactory.TargetedSpaceSuccess = false
	callScale(args, reqFactory, restarter, appRepo)
	assert.False(t, testhelpers.CommandDidPassRequirements)

	reqFactory.LoginSuccess = true
	reqFactory.TargetedSpaceSuccess = true
	callScale(args, reqFactory, restarter, appRepo)
	assert.True(t, testhelpers.CommandDidPassRequirements)
	assert.Equal(t, reqFactory.ApplicationName, "my-app")
}

func TestScaleFailsWithUsage(t *testing.T) {
	reqFactory, restarter, appRepo := getScaleDependencies()

	ui := callScale([]string{}, reqFactory, restarter, appRepo)

	assert.True(t, ui.FailedWithUsage)
	assert.False(t, testhelpers.CommandDidPassRequirements)
}

func TestScaleAll(t *testing.T) {
	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}
	reqFactory, restarter, appRepo := getScaleDependencies()
	reqFactory.Application = app

	ui := callScale([]string{"-d", "2G", "-i", "5", "-m", "512M", "my-app"}, reqFactory, restarter, appRepo)

	assert.Contains(t, ui.Outputs[0], "Scaling")
	assert.Contains(t, ui.Outputs[0], "my-app")

	assert.Equal(t, appRepo.ScaledApp.Guid, "my-app-guid")
	assert.Equal(t, appRepo.ScaledApp.DiskQuota, uint64(2048)) // in megabytes
	assert.Equal(t, appRepo.ScaledApp.Memory, uint64(512))
	assert.Equal(t, appRepo.ScaledApp.Instances, 5)

	assert.Equal(t, restarter.AppToRestart.Guid, app.Guid)
	assert.Equal(t, restarter.AppToRestart.Name, app.Name)
}

func TestScaleOnlyDisk(t *testing.T) {
	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}
	reqFactory, restarter, appRepo := getScaleDependencies()
	reqFactory.Application = app

	callScale([]string{"-d", "2G", "my-app"}, reqFactory, restarter, appRepo)

	assert.Equal(t, appRepo.ScaledApp.Guid, "my-app-guid")
	assert.Equal(t, appRepo.ScaledApp.DiskQuota, uint64(2048)) // in megabytes
	assert.Equal(t, appRepo.ScaledApp.Memory, uint64(0))
	assert.Equal(t, appRepo.ScaledApp.Instances, 0)
}

func TestScaleOnlyInstances(t *testing.T) {
	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}
	reqFactory, restarter, appRepo := getScaleDependencies()
	reqFactory.Application = app

	callScale([]string{"-i", "5", "my-app"}, reqFactory, restarter, appRepo)

	assert.Equal(t, appRepo.ScaledApp.Guid, "my-app-guid")
	assert.Equal(t, appRepo.ScaledApp.DiskQuota, uint64(0))
	assert.Equal(t, appRepo.ScaledApp.Memory, uint64(0))
	assert.Equal(t, appRepo.ScaledApp.Instances, 5)
}

func TestScaleOnlyMemory(t *testing.T) {
	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}
	reqFactory, restarter, appRepo := getScaleDependencies()
	reqFactory.Application = app

	callScale([]string{"-m", "512M", "my-app"}, reqFactory, restarter, appRepo)

	assert.Equal(t, appRepo.ScaledApp.Guid, "my-app-guid")
	assert.Equal(t, appRepo.ScaledApp.DiskQuota, uint64(0))
	assert.Equal(t, appRepo.ScaledApp.Memory, uint64(512))
	assert.Equal(t, appRepo.ScaledApp.Instances, 0)
}

func getScaleDependencies() (reqFactory *testhelpers.FakeReqFactory, restarter *testhelpers.FakeAppRestarter, appRepo *testhelpers.FakeApplicationRepository) {
	reqFactory = &testhelpers.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}
	restarter = &testhelpers.FakeAppRestarter{}
	appRepo = &testhelpers.FakeApplicationRepository{}
	return
}

func callScale(args []string, reqFactory *testhelpers.FakeReqFactory, restarter *testhelpers.FakeAppRestarter, appRepo api.ApplicationRepository) (ui *testhelpers.FakeUI) {
	ui = new(testhelpers.FakeUI)
	ctxt := testhelpers.NewContext("scale", args)
	cmd := NewScale(ui, restarter, appRepo)
	testhelpers.RunCommand(cmd, ctxt, reqFactory)
	return
}
