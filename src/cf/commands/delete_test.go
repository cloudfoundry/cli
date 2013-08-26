package commands_test

import (
	"cf"
	. "cf/commands"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestDelete(t *testing.T) {
	appRepo := &testhelpers.FakeApplicationRepository{
		AppByName: cf.Application{Name: "app-to-delete", Guid: "app-to-delete-guid"},
	}
	ui := &testhelpers.FakeUI{}
	config := &configuration.Configuration{}
	cmd := NewDelete(ui, config, appRepo)

	ui.Inputs = []string{"y"}

	cmd.Run(testhelpers.NewContext("delete", []string{"app-to-delete"}))

	assert.Equal(t, appRepo.AppName, "app-to-delete")
	assert.Contains(t, ui.Prompts[0], "Really delete app-to-delete?>")

	assert.Contains(t, ui.Outputs[0], "Deleting app-to-delete")
	assert.Equal(t, appRepo.DeletedApp, cf.Application{Name: "app-to-delete", Guid: "app-to-delete-guid"})
	assert.Contains(t, ui.Outputs[1], "OK")
}
