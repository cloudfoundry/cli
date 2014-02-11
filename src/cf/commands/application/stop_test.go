package application_test

import (
	"cf/api"
	. "cf/commands/application"
	"cf/models"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

func init() {
	Describe("Testing with ginkgo", func() {
		It("TestStopCommandFailsWithUsage", func() {
			app := models.Application{}
			app.Name = "my-app"
			app.Guid = "my-app-guid"
			appRepo := &testapi.FakeApplicationRepository{ReadApp: app}
			reqFactory := &testreq.FakeReqFactory{Application: app}

			ui := callStop([]string{}, reqFactory, appRepo)
			assert.True(mr.T(), ui.FailedWithUsage)

			ui = callStop([]string{"my-app"}, reqFactory, appRepo)
			assert.False(mr.T(), ui.FailedWithUsage)
		})
		It("TestStopApplication", func() {

			app := models.Application{}
			app.Name = "my-app"
			app.Guid = "my-app-guid"
			appRepo := &testapi.FakeApplicationRepository{ReadApp: app}
			args := []string{"my-app"}
			reqFactory := &testreq.FakeReqFactory{Application: app}
			ui := callStop(args, reqFactory, appRepo)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Stopping app", "my-app", "my-org", "my-space", "my-user"},
				{"OK"},
			})

			assert.Equal(mr.T(), reqFactory.ApplicationName, "my-app")
			assert.Equal(mr.T(), appRepo.UpdateAppGuid, "my-app-guid")
		})
		It("TestStopApplicationWhenStopFails", func() {

			app := models.Application{}
			app.Name = "my-app"
			app.Guid = "my-app-guid"
			appRepo := &testapi.FakeApplicationRepository{ReadApp: app, UpdateErr: true}
			args := []string{"my-app"}
			reqFactory := &testreq.FakeReqFactory{Application: app}
			ui := callStop(args, reqFactory, appRepo)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Stopping", "my-app"},
				{"FAILED"},
				{"Error updating app."},
			})
			assert.Equal(mr.T(), appRepo.UpdateAppGuid, "my-app-guid")
		})
		It("TestStopApplicationIsAlreadyStopped", func() {

			app := models.Application{}
			app.Name = "my-app"
			app.Guid = "my-app-guid"
			app.State = "stopped"
			appRepo := &testapi.FakeApplicationRepository{ReadApp: app}
			args := []string{"my-app"}
			reqFactory := &testreq.FakeReqFactory{Application: app}
			ui := callStop(args, reqFactory, appRepo)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"my-app", "is already stopped"},
			})
			assert.Equal(mr.T(), appRepo.UpdateAppGuid, "")
		})
		It("TestApplicationStopReturnsUpdatedApp", func() {

			appToStop := models.Application{}
			appToStop.Name = "my-app"
			appToStop.Guid = "my-app-guid"
			appToStop.State = "started"
			expectedStoppedApp := models.Application{}
			expectedStoppedApp.Name = "my-stopped-app"
			expectedStoppedApp.Guid = "my-stopped-app-guid"
			expectedStoppedApp.State = "stopped"

			appRepo := &testapi.FakeApplicationRepository{UpdateAppResult: expectedStoppedApp}
			config := testconfig.NewRepository()
			stopper := NewStop(new(testterm.FakeUI), config, appRepo)
			actualStoppedApp, err := stopper.ApplicationStop(appToStop)

			assert.NoError(mr.T(), err)
			assert.Equal(mr.T(), expectedStoppedApp, actualStoppedApp)
		})
		It("TestApplicationStopReturnsUpdatedAppWhenAppIsAlreadyStopped", func() {

			appToStop := models.Application{}
			appToStop.Name = "my-app"
			appToStop.Guid = "my-app-guid"
			appToStop.State = "stopped"
			appRepo := &testapi.FakeApplicationRepository{}
			config := testconfig.NewRepository()
			stopper := NewStop(new(testterm.FakeUI), config, appRepo)
			updatedApp, err := stopper.ApplicationStop(appToStop)

			assert.NoError(mr.T(), err)
			assert.Equal(mr.T(), appToStop, updatedApp)
		})
	})
}

func callStop(args []string, reqFactory *testreq.FakeReqFactory, appRepo api.ApplicationRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("stop", args)

	configRepo := testconfig.NewRepositoryWithDefaults()
	cmd := NewStop(ui, configRepo, appRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
