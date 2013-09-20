package commands_test

import (
	"cf"
	"cf/api"
	. "cf/commands"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestScaleRequirements(t *testing.T) {
	args := []string{"-d 1G", "my-app"}
	reqFactory, starter, stopper, appRepo := getDefaultDependencies()

	reqFactory.LoginSuccess = false
	reqFactory.TargetedSpaceSuccess = true
	callScale(args, reqFactory, starter, stopper, appRepo)
	assert.False(t, testhelpers.CommandDidPassRequirements)

	reqFactory.LoginSuccess = true
	reqFactory.TargetedSpaceSuccess = false
	callScale(args, reqFactory, starter, stopper, appRepo)
	assert.False(t, testhelpers.CommandDidPassRequirements)

	reqFactory.LoginSuccess = true
	reqFactory.TargetedSpaceSuccess = true
	callScale(args, reqFactory, starter, stopper, appRepo)
	assert.True(t, testhelpers.CommandDidPassRequirements)
	assert.Equal(t, reqFactory.ApplicationName, "my-app")
}

func TestScaleFailsWithUsage(t *testing.T) {
	reqFactory, starter, stopper, appRepo := getDefaultDependencies()

	ui := callScale([]string{}, reqFactory, starter, stopper, appRepo)

	assert.True(t, ui.FailedWithUsage)
	assert.False(t, testhelpers.CommandDidPassRequirements)
}

func TestScaleAll(t *testing.T) {
	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}
	reqFactory, starter, stopper, appRepo := getDefaultDependencies()
	reqFactory.Application = app

	ui := callScale([]string{"-d", "2G", "-i", "5", "my-app"}, reqFactory, starter, stopper, appRepo)

	assert.Contains(t, ui.Outputs[0], "Scaling")
	assert.Contains(t, ui.Outputs[0], "my-app")

	assert.Equal(t, appRepo.ScaledApp.Guid, "my-app-guid")
	assert.Equal(t, appRepo.ScaledApp.DiskQuota, int64(2*GIGABYTE))
	assert.Equal(t, appRepo.ScaledApp.Instances, 5)

	assert.Equal(t, stopper.StoppedApp, app)
	assert.Equal(t, starter.StartedApp, app)
}

func TestScaleOnlyDisk(t *testing.T) {
	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}
	reqFactory, starter, stopper, appRepo := getDefaultDependencies()
	reqFactory.Application = app

	callScale([]string{"-d", "2G", "my-app"}, reqFactory, starter, stopper, appRepo)

	assert.Equal(t, appRepo.ScaledApp.Guid, "my-app-guid")
	assert.Equal(t, appRepo.ScaledApp.DiskQuota, int64(2*GIGABYTE))
	assert.Equal(t, appRepo.ScaledApp.Instances, 0)
}

func TestScaleOnlyInstances(t *testing.T) {
	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}
	reqFactory, starter, stopper, appRepo := getDefaultDependencies()
	reqFactory.Application = app

	callScale([]string{"-i", "5", "my-app"}, reqFactory, starter, stopper, appRepo)

	assert.Equal(t, appRepo.ScaledApp.Guid, "my-app-guid")
	assert.Equal(t, appRepo.ScaledApp.DiskQuota, 0)
	assert.Equal(t, appRepo.ScaledApp.Instances, 5)
}

func getDefaultDependencies() (reqFactory *testhelpers.FakeReqFactory, starter *testhelpers.FakeAppStarter, stopper *testhelpers.FakeAppStopper, appRepo *testhelpers.FakeApplicationRepository) {
	reqFactory = &testhelpers.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}
	starter = &testhelpers.FakeAppStarter{}
	stopper = &testhelpers.FakeAppStopper{}
	appRepo = &testhelpers.FakeApplicationRepository{}
	return
}

func callScale(args []string, reqFactory *testhelpers.FakeReqFactory, starter ApplicationStarter, stopper ApplicationStopper, appRepo api.ApplicationRepository) (ui *testhelpers.FakeUI) {
	ui = new(testhelpers.FakeUI)
	ctxt := testhelpers.NewContext("scale", args)
	cmd := NewScale(ui, starter, stopper, appRepo)
	testhelpers.RunCommand(cmd, ctxt, reqFactory)
	return
}
