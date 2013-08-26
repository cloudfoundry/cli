package commands_test

import (
	"cf"
	. "cf/commands"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestApps(t *testing.T) {
	app1Urls := []string{"app1.cfapps.io", "app1.example.com"}
	app2Urls := []string{"app2.cfapps.io"}

	apps := []cf.Application{
		cf.Application{Name: "Application-1", State: "started", Instances: 1, Memory: 512, Urls: app1Urls},
		cf.Application{Name: "Application-2", State: "started", Instances: 2, Memory: 256, Urls: app2Urls},
	}
	appRepo := &testhelpers.FakeApplicationRepository{FindAllApps: apps}
	ui := &testhelpers.FakeUI{}
	config := &configuration.Configuration{
		Space: cf.Space{Name: "development", Guid: "development-guid"},
	}

	cmd := NewApps(ui, config, appRepo)
	cmd.Run(testhelpers.NewContext(5, []string{}))

	assert.Contains(t, ui.Outputs[0], "Getting applications in development")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[3], "Application-1")
	assert.Contains(t, ui.Outputs[3], "running")
	assert.Contains(t, ui.Outputs[3], "1 x 512M")
	assert.Contains(t, ui.Outputs[3], "app1.cfapps.io, app1.example.com")

	assert.Contains(t, ui.Outputs[4], "Application-2")
	assert.Contains(t, ui.Outputs[4], "running")
	assert.Contains(t, ui.Outputs[4], "2 x 256M")
	assert.Contains(t, ui.Outputs[4], "app2.cfapps.io")
}
