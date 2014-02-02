package application_test

import (
	"cf"
	. "cf/commands/application"
	"cf/configuration"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	"testhelpers/maker"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

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

func callScale(t mr.TestingT, args []string, deps scaleDependencies) (ui *testterm.FakeUI) {
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
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestScaleRequirements", func() {
			args := []string{"-m", "1G", "my-app"}
			deps := getScaleDependencies()

			deps.reqFactory.LoginSuccess = false
			deps.reqFactory.TargetedSpaceSuccess = true
			callScale(mr.T(), args, deps)
			assert.False(mr.T(), testcmd.CommandDidPassRequirements)

			deps.reqFactory.LoginSuccess = true
			deps.reqFactory.TargetedSpaceSuccess = false
			callScale(mr.T(), args, deps)
			assert.False(mr.T(), testcmd.CommandDidPassRequirements)

			deps.reqFactory.LoginSuccess = true
			deps.reqFactory.TargetedSpaceSuccess = true
			callScale(mr.T(), args, deps)
			assert.True(mr.T(), testcmd.CommandDidPassRequirements)
			assert.Equal(mr.T(), deps.reqFactory.ApplicationName, "my-app")
		})
		It("TestScaleFailsWithUsage", func() {

			deps := getScaleDependencies()

			ui := callScale(mr.T(), []string{}, deps)

			assert.True(mr.T(), ui.FailedWithUsage)
			assert.False(mr.T(), testcmd.CommandDidPassRequirements)
		})
		It("TestScaleFailsWithoutFlags", func() {

			args := []string{"my-app"}
			deps := getScaleDependencies()
			deps.reqFactory.LoginSuccess = true
			deps.reqFactory.TargetedSpaceSuccess = true

			callScale(mr.T(), args, deps)
			assert.False(mr.T(), testcmd.CommandDidPassRequirements)
		})
		It("TestScaleAll", func() {

			app := maker.NewApp(maker.Overrides{"name": "my-app", "guid": "my-app-guid"})
			deps := getScaleDependencies()
			deps.reqFactory.Application = app
			deps.appRepo.UpdateAppResult = app

			ui := callScale(mr.T(), []string{"-i", "5", "-m", "512M", "my-app"}, deps)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Scaling", "my-app", "my-org", "my-space", "my-user"},
				{"OK"},
			})

			assert.Equal(mr.T(), deps.restarter.AppToRestart.Guid, "my-app-guid")
			assert.Equal(mr.T(), deps.appRepo.UpdateAppGuid, "my-app-guid")
			assert.Equal(mr.T(), deps.appRepo.UpdateParams.Get("memory"), uint64(512))
			assert.Equal(mr.T(), deps.appRepo.UpdateParams.Get("instances"), 5)
		})
		It("TestScaleOnlyInstances", func() {

			app := maker.NewApp(maker.Overrides{"name": "my-app", "guid": "my-app-guid"})
			deps := getScaleDependencies()
			deps.reqFactory.Application = app
			deps.appRepo.UpdateAppResult = app

			callScale(mr.T(), []string{"-i", "5", "my-app"}, deps)

			assert.Equal(mr.T(), deps.restarter.AppToRestart.Guid, "")
			assert.Equal(mr.T(), deps.appRepo.UpdateAppGuid, "my-app-guid")
			assert.Equal(mr.T(), deps.appRepo.UpdateParams.Get("instances"), 5)
			assert.False(mr.T(), deps.appRepo.UpdateParams.Has("disk_quota"))
			assert.False(mr.T(), deps.appRepo.UpdateParams.Has("memory"))
		})
		It("TestScaleOnlyMemory", func() {

			app := maker.NewApp(maker.Overrides{"name": "my-app", "guid": "my-app-guid"})
			deps := getScaleDependencies()
			deps.reqFactory.Application = app
			deps.appRepo.UpdateAppResult = app

			callScale(mr.T(), []string{"-m", "512M", "my-app"}, deps)

			assert.Equal(mr.T(), deps.restarter.AppToRestart.Guid, "my-app-guid")
			assert.Equal(mr.T(), deps.appRepo.UpdateAppGuid, "my-app-guid")
			assert.Equal(mr.T(), deps.appRepo.UpdateParams.Get("memory").(uint64), uint64(512))
			assert.False(mr.T(), deps.appRepo.UpdateParams.Has("disk_quota"))
			assert.False(mr.T(), deps.appRepo.UpdateParams.Has("instances"))
		})
	})
}
