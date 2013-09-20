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
	args := []string{"my-app", "-d 1G -i 12"}
	starter := &testhelpers.FakeAppStarter{}
	stopper := &testhelpers.FakeAppStopper{}
	appRepo := &testhelpers.FakeApplicationRepository{}

	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: false, TargetedSpaceSuccess: true, Application: cf.Application{}}
	callScale(args, reqFactory, starter, stopper, appRepo)
	assert.False(t, testhelpers.CommandDidPassRequirements)

	reqFactory = &testhelpers.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: false, Application: cf.Application{}}
	callScale(args, reqFactory, starter, stopper, appRepo)
	assert.False(t, testhelpers.CommandDidPassRequirements)

	reqFactory = &testhelpers.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, Application: cf.Application{}}
	callScale(args, reqFactory, starter, stopper, appRepo)
	assert.True(t, testhelpers.CommandDidPassRequirements)
	assert.Equal(t, reqFactory.ApplicationName, "my-app")
}

func TestScaleFailsWithUsage(t *testing.T) {
	starter := &testhelpers.FakeAppStarter{}
	stopper := &testhelpers.FakeAppStopper{}
	appRepo := &testhelpers.FakeApplicationRepository{}
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, Application: cf.Application{}}

	ui := callScale([]string{}, reqFactory, starter, stopper, appRepo)

	assert.True(t, ui.FailedWithUsage)
	assert.False(t, testhelpers.CommandDidPassRequirements)
}

func TestScaleDisk(t *testing.T) {
	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}

	starter := &testhelpers.FakeAppStarter{}
	stopper := &testhelpers.FakeAppStopper{}
	appRepo := &testhelpers.FakeApplicationRepository{}
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, Application: app}

	ui := callScale([]string{"-d", "2G", "my-app"}, reqFactory, starter, stopper, appRepo)

	assert.Contains(t, ui.Outputs[0], "Scaling")
	assert.Contains(t, ui.Outputs[0], "my-app")

	assert.Equal(t, appRepo.ScaledApp.Guid, "my-app-guid")
	assert.Equal(t, appRepo.ScaledApp.DiskQuota, int64(2*GIGABYTE))

	assert.Equal(t, stopper.StoppedApp, app)
	assert.Equal(t, starter.StartedApp, app)

}

func callScale(args []string, reqFactory *testhelpers.FakeReqFactory, starter ApplicationStarter, stopper ApplicationStopper, appRepo api.ApplicationRepository) (ui *testhelpers.FakeUI) {
	ui = new(testhelpers.FakeUI)
	ctxt := testhelpers.NewContext("scale", args)
	cmd := NewScale(ui, starter, stopper, appRepo)
	testhelpers.RunCommand(cmd, ctxt, reqFactory)

	return
}
