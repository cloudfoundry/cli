package cf

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestApplicationHealth(t *testing.T) {
	app := Application{State: "down"}
	assert.Equal(t, app.Health(), "down")

	app.State = "started"
	app.Instances = 0
	assert.Equal(t, app.Health(), "N/A")

	app.Instances = 3
	app.RunningInstances = 1
	assert.Equal(t, app.Health(), "33%")

	app.RunningInstances = 3
	assert.Equal(t, app.Health(), "running")
}
