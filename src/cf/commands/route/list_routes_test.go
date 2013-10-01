package route_test

import (
	"cf"
	. "cf/commands/route"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestListingRoutes(t *testing.T) {
	routes := []cf.Route{
		cf.Route{
			Host:   "hostname-1",
			Domain: cf.Domain{Name: "example.com"},
		},
		cf.Route{
			Host:   "hostname-2",
			Domain: cf.Domain{Name: "cfapps.com"},
		},
	}
	routeRepo := &testhelpers.FakeRouteRepository{FindAllRoutes: routes}

	ui := &testhelpers.FakeUI{}

	cmd := NewListRoutes(ui, routeRepo)
	cmd.Run(testhelpers.NewContext("routes", []string{}))

	assert.Contains(t, ui.Outputs[0], "Getting routes")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[2], "hostname-1.example.com")
	assert.Contains(t, ui.Outputs[3], "hostname-2.cfapps.com")
}

func TestListingRoutesWhenNoneExist(t *testing.T) {
	routes := []cf.Route{}
	routeRepo := &testhelpers.FakeRouteRepository{FindAllRoutes: routes}

	ui := &testhelpers.FakeUI{}

	cmd := NewListRoutes(ui, routeRepo)
	cmd.Run(testhelpers.NewContext("routes", []string{}))

	assert.Contains(t, ui.Outputs[0], "Getting routes")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[2], "No routes found")
}

func TestListingRoutesWhenFindFails(t *testing.T) {
	routeRepo := &testhelpers.FakeRouteRepository{FindAllErr: true}

	ui := &testhelpers.FakeUI{}

	cmd := NewListRoutes(ui, routeRepo)
	cmd.Run(testhelpers.NewContext("routes", []string{}))

	assert.Contains(t, ui.Outputs[0], "Getting routes")
	assert.Contains(t, ui.Outputs[1], "FAILED")
}
