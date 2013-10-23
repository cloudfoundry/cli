package route_test

import (
	"cf"
	. "cf/commands/route"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
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
	routeRepo := &testapi.FakeRouteRepository{FindAllRoutes: routes}

	ui := callListRoutes(t, []string{}, &testreq.FakeReqFactory{}, routeRepo)

	assert.Contains(t, ui.Outputs[0], "Getting routes in space")
	assert.Contains(t, ui.Outputs[0], "my-space")
	assert.Contains(t, ui.Outputs[0], "my-org")
	assert.Contains(t, ui.Outputs[0], "my-user")
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
	routeRepo := &testapi.FakeRouteRepository{FindAllRoutes: routes}

	ui := callListRoutes(t, []string{}, &testreq.FakeReqFactory{}, routeRepo)

	assert.Contains(t, ui.Outputs[0], "Getting routes")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[3], "No routes found")
}

func TestListingRoutesWhenFindFails(t *testing.T) {
	routeRepo := &testapi.FakeRouteRepository{FindAllErr: true}

	ui := callListRoutes(t, []string{}, &testreq.FakeReqFactory{}, routeRepo)

	assert.Contains(t, ui.Outputs[0], "Getting routes")
	assert.Contains(t, ui.Outputs[1], "FAILED")
}

func callListRoutes(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory, routeRepo *testapi.FakeRouteRepository) (ui *testterm.FakeUI) {

	ui = &testterm.FakeUI{}

	ctxt := testcmd.NewContext("list-routes", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)

	config := &configuration.Configuration{
		Space:        cf.Space{Name: "my-space"},
		Organization: cf.Organization{Name: "my-org"},
		AccessToken:  token,
	}

	cmd := NewListRoutes(ui, config, routeRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)

	return
}
