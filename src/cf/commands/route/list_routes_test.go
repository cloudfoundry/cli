package route_test

import (
	"cf"
	. "cf/commands/route"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestListingRoutes(t *testing.T) {
	routes := []cf.Route{
		cf.Route{
			Host:     "hostname-1",
			Domain:   cf.Domain{Name: "example.com"},
			AppNames: []string{"dora", "dora2"},
		},
		cf.Route{
			Host:     "hostname-2",
			Domain:   cf.Domain{Name: "cfapps.com"},
			AppNames: []string{"my-app", "my-app2"},
		},
	}
	routeRepo := &testhelpers.FakeRouteRepository{FindAllRoutes: routes}
	config := &configuration.Configuration{
		Space:        cf.Space{Name: "my-space"},
		Organization: cf.Organization{Name: "my-org"},
	}
	ui := &testhelpers.FakeUI{}

	cmd := NewListRoutes(ui, config, routeRepo)
	cmd.Run(testhelpers.NewContext("routes", []string{}))

	assert.Contains(t, ui.Outputs[0], "Getting routes in space")
	assert.Contains(t, ui.Outputs[0], "my-space")
	assert.Contains(t, ui.Outputs[0], "my-org")
	assert.Contains(t, ui.Outputs[1], "OK")

	assert.Contains(t, ui.Outputs[3], "host")
	assert.Contains(t, ui.Outputs[3], "domain")
	assert.Contains(t, ui.Outputs[3], "apps")

	assert.Contains(t, ui.Outputs[4], "hostname-1")
	assert.Contains(t, ui.Outputs[4], "example.com")
	assert.Contains(t, ui.Outputs[4], "dora, dora2")

	assert.Contains(t, ui.Outputs[5], "hostname-2")
	assert.Contains(t, ui.Outputs[5], "cfapps.com")
	assert.Contains(t, ui.Outputs[5], "my-app, my-app2")
}

func TestListingRoutesWhenNoneExist(t *testing.T) {
	routes := []cf.Route{}
	routeRepo := &testhelpers.FakeRouteRepository{FindAllRoutes: routes}
	config := &configuration.Configuration{}
	ui := &testhelpers.FakeUI{}

	cmd := NewListRoutes(ui, config, routeRepo)
	cmd.Run(testhelpers.NewContext("routes", []string{}))

	assert.Contains(t, ui.Outputs[0], "Getting routes")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[3], "No routes found")
}

func TestListingRoutesWhenFindFails(t *testing.T) {
	routeRepo := &testhelpers.FakeRouteRepository{FindAllErr: true}
	config := &configuration.Configuration{}
	ui := &testhelpers.FakeUI{}

	cmd := NewListRoutes(ui, config, routeRepo)
	cmd.Run(testhelpers.NewContext("routes", []string{}))

	assert.Contains(t, ui.Outputs[0], "Getting routes")
	assert.Contains(t, ui.Outputs[1], "FAILED")
}
