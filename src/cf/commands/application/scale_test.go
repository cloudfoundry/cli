package application_test

import (
	"cf"
	"cf/api"
	. "cf/commands/application"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestScaleRequirements(t *testing.T) {
	args := []string{"-m", "1G", "my-app"}
	reqFactory, restarter, appRepo := getScaleDependencies()

	reqFactory.LoginSuccess = false
	reqFactory.TargetedSpaceSuccess = true
	callScale(t, args, reqFactory, restarter, appRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory.LoginSuccess = true
	reqFactory.TargetedSpaceSuccess = false
	callScale(t, args, reqFactory, restarter, appRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory.LoginSuccess = true
	reqFactory.TargetedSpaceSuccess = true
	callScale(t, args, reqFactory, restarter, appRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)
	assert.Equal(t, reqFactory.ApplicationName, "my-app")
}

func TestScaleFailsWithUsage(t *testing.T) {
	reqFactory, restarter, appRepo := getScaleDependencies()

	ui := callScale(t, []string{}, reqFactory, restarter, appRepo)

	assert.True(t, ui.FailedWithUsage)
	assert.False(t, testcmd.CommandDidPassRequirements)
}

func TestScaleFailsWithoutFlags(t *testing.T) {
	args := []string{"my-app"}
	reqFactory, restarter, appRepo := getScaleDependencies()
	reqFactory.LoginSuccess = true
	reqFactory.TargetedSpaceSuccess = true

	callScale(t, args, reqFactory, restarter, appRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)
}

func TestScaleAll(t *testing.T) {
	app := cf.Application{}
	app.Name = "my-app"
	app.Guid = "my-app-guid"
	reqFactory, restarter, appRepo := getScaleDependencies()
	reqFactory.Application = app

	ui := callScale(t, []string{"-i", "5", "-m", "512M", "my-app"}, reqFactory, restarter, appRepo)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Scaling", "my-app", "my-org", "my-space", "my-user"},
		{"OK"},
	})

	assert.Equal(t, appRepo.UpdateAppGuid, "my-app-guid")
	assert.Equal(t, appRepo.UpdateParams.Get("memory"), uint64(512))
	assert.Equal(t, appRepo.UpdateParams.Get("instances"), 5)
}

func TestScaleOnlyInstances(t *testing.T) {
	app := cf.Application{}
	app.Name = "my-app"
	app.Guid = "my-app-guid"
	reqFactory, restarter, appRepo := getScaleDependencies()
	reqFactory.Application = app

	callScale(t, []string{"-i", "5", "my-app"}, reqFactory, restarter, appRepo)

	assert.Equal(t, appRepo.UpdateAppGuid, "my-app-guid")
	assert.Equal(t, appRepo.UpdateParams.Get("instances"), 5)

	assert.False(t, appRepo.UpdateParams.Has("disk_quota"))
	assert.False(t, appRepo.UpdateParams.Has("memory"))
}

func TestScaleOnlyMemory(t *testing.T) {
	app := cf.Application{}
	app.Name = "my-app"
	app.Guid = "my-app-guid"
	reqFactory, restarter, appRepo := getScaleDependencies()
	reqFactory.Application = app

	callScale(t, []string{"-m", "512M", "my-app"}, reqFactory, restarter, appRepo)

	assert.Equal(t, appRepo.UpdateAppGuid, "my-app-guid")
	assert.Equal(t, appRepo.UpdateParams.Get("memory").(uint64), uint64(512))

	assert.False(t, appRepo.UpdateParams.Has("disk_quota"))
	assert.False(t, appRepo.UpdateParams.Has("instances"))
}

func TestScaleMemoryWithIntAssumesMegabytes(t *testing.T) {
	app := cf.Application{}
	app.Name = "my-app"
	app.Guid = "my-app-guid"
	reqFactory, restarter, appRepo := getScaleDependencies()
	reqFactory.Application = app

	callScale(t, []string{"-m", "1024", "my-app"}, reqFactory, restarter, appRepo)

	assert.Equal(t, appRepo.UpdateAppGuid, "my-app-guid")
	assert.Equal(t, appRepo.UpdateParams.Get("memory").(uint64), uint64(1024))

	assert.False(t, appRepo.UpdateParams.Has("disk_quota"))
	assert.False(t, appRepo.UpdateParams.Has("instances"))
}

func getScaleDependencies() (reqFactory *testreq.FakeReqFactory, restarter *testcmd.FakeAppRestarter, appRepo *testapi.FakeApplicationRepository) {
	reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}
	restarter = &testcmd.FakeAppRestarter{}
	appRepo = &testapi.FakeApplicationRepository{}
	return
}

func callScale(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory, restarter *testcmd.FakeAppRestarter, appRepo api.ApplicationRepository) (ui *testterm.FakeUI) {
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

	cmd := NewScale(ui, config, restarter, appRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
