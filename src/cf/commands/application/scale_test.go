package application_test

import (
	"cf"
	. "cf/commands/application"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	"testhelpers/maker"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestScaleRequirements(t *testing.T) {
	args := []string{"-m", "1G", "my-app"}
	deps := getScaleDependencies()

	deps.reqFactory.LoginSuccess = false
	deps.reqFactory.TargetedSpaceSuccess = true
	callScale(t, args, deps)
	assert.False(t, testcmd.CommandDidPassRequirements)

	deps.reqFactory.LoginSuccess = true
	deps.reqFactory.TargetedSpaceSuccess = false
	callScale(t, args, deps)
	assert.False(t, testcmd.CommandDidPassRequirements)

	deps.reqFactory.LoginSuccess = true
	deps.reqFactory.TargetedSpaceSuccess = true
	callScale(t, args, deps)
	assert.True(t, testcmd.CommandDidPassRequirements)
	assert.Equal(t, deps.reqFactory.ApplicationName, "my-app")
}

func TestScaleFailsWithUsage(t *testing.T) {
	deps := getScaleDependencies()

	ui := callScale(t, []string{}, deps)

	assert.True(t, ui.FailedWithUsage)
	assert.False(t, testcmd.CommandDidPassRequirements)
}

func TestScaleFailsWithoutFlags(t *testing.T) {
	args := []string{"my-app"}
	deps := getScaleDependencies()
	deps.reqFactory.LoginSuccess = true
	deps.reqFactory.TargetedSpaceSuccess = true

	callScale(t, args, deps)
	assert.False(t, testcmd.CommandDidPassRequirements)
}

func TestScaleAll(t *testing.T) {
	app := maker.NewApp(maker.Overrides{"name": "my-app", "guid": "my-app-guid"})
	deps := getScaleDependencies()
	deps.reqFactory.Application = app
	deps.appRepo.UpdateAppResult = app

	ui := callScale(t, []string{"-i", "5", "-m", "512M", "my-app"}, deps)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Scaling", "my-app", "my-org", "my-space", "my-user"},
		{"OK"},
	})

	assert.Equal(t, deps.restarter.AppToRestart.Guid, "my-app-guid")
	assert.Equal(t, deps.appRepo.UpdateAppGuid, "my-app-guid")
	assert.Equal(t, deps.appRepo.UpdateParams.Get("memory"), uint64(512))
	assert.Equal(t, deps.appRepo.UpdateParams.Get("instances"), 5)
}

func TestScaleOnlyInstances(t *testing.T) {
	app := maker.NewApp(maker.Overrides{"name": "my-app", "guid": "my-app-guid"})
	deps := getScaleDependencies()
	deps.reqFactory.Application = app
	deps.appRepo.UpdateAppResult = app

	callScale(t, []string{"-i", "5", "my-app"}, deps)

	assert.Equal(t, deps.restarter.AppToRestart.Guid, "")
	assert.Equal(t, deps.appRepo.UpdateAppGuid, "my-app-guid")
	assert.Equal(t, deps.appRepo.UpdateParams.Get("instances"), 5)
	assert.False(t, deps.appRepo.UpdateParams.Has("disk_quota"))
	assert.False(t, deps.appRepo.UpdateParams.Has("memory"))
}

func TestScaleOnlyMemory(t *testing.T) {
	app := maker.NewApp(maker.Overrides{"name": "my-app", "guid": "my-app-guid"})
	deps := getScaleDependencies()
	deps.reqFactory.Application = app
	deps.appRepo.UpdateAppResult = app

	callScale(t, []string{"-m", "512M", "my-app"}, deps)

	assert.Equal(t, deps.restarter.AppToRestart.Guid, "my-app-guid")
	assert.Equal(t, deps.appRepo.UpdateAppGuid, "my-app-guid")
	assert.Equal(t, deps.appRepo.UpdateParams.Get("memory").(uint64), uint64(512))
	assert.False(t, deps.appRepo.UpdateParams.Has("disk_quota"))
	assert.False(t, deps.appRepo.UpdateParams.Has("instances"))
}

type scaleDependencies struct {
	reqFactory *testreq.FakeReqFactory
	restarter  *testcmd.FakeAppRestarter
	appRepo    *testapi.FakeApplicationRepository
}

func getScaleDependencies() (deps scaleDependencies) {
	deps = scaleDependencies{
		reqFactory: &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true},
		restarter:  &testcmd.FakeAppRestarter{},
		appRepo:    &testapi.FakeApplicationRepository{},
	}
	return
}

func callScale(t *testing.T, args []string, deps scaleDependencies) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("scale", args)

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

	cmd := NewScale(ui, config, deps.restarter, deps.appRepo)
	testcmd.RunCommand(cmd, ctxt, deps.reqFactory)
	return
}
