package commands_test

import (
	"cf"
	. "cf/commands"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func deleteApp(confirmation string) (ui *testhelpers.FakeUI, reqFactory *testhelpers.FakeReqFactory, appRepo *testhelpers.FakeApplicationRepository) {
	app := cf.Application{Name: "app-to-delete", Guid: "app-to-delete-guid"}
	reqFactory = &testhelpers.FakeReqFactory{Application: app}
	appRepo = &testhelpers.FakeApplicationRepository{}
	config := &configuration.Configuration{}

	ui = &testhelpers.FakeUI{
		Inputs: []string{confirmation},
	}
	ctxt := testhelpers.NewContext("delete", []string{"app-to-delete"})

	cmd := NewDelete(ui, config, appRepo)
	testhelpers.RunCommand(cmd, ctxt, reqFactory)
	return
}

func TestDeleteConfirmingWithY(t *testing.T) {
	ui, reqFactory, appRepo := deleteApp("y")

	assert.Equal(t, reqFactory.ApplicationName, "app-to-delete")
	assert.Contains(t, ui.Prompts[0], "Really delete app-to-delete?>")

	assert.Contains(t, ui.Outputs[0], "Deleting app-to-delete")
	assert.Equal(t, appRepo.DeletedApp, reqFactory.Application)
	assert.Contains(t, ui.Outputs[1], "OK")
}

func TestDeleteConfirmingWithYes(t *testing.T) {
	ui, reqFactory, appRepo := deleteApp("Yes")

	assert.Equal(t, reqFactory.ApplicationName, "app-to-delete")
	assert.Contains(t, ui.Prompts[0], "Really delete app-to-delete?>")

	assert.Contains(t, ui.Outputs[0], "Deleting app-to-delete")
	assert.Equal(t, appRepo.DeletedApp, reqFactory.Application)
	assert.Contains(t, ui.Outputs[1], "OK")
}
