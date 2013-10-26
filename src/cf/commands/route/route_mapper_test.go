package route_test

import (
	"cf"
	. "cf/commands/route"
	"cf/configuration"
	"github.com/codegangsta/cli"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestRouteMapperFailsWithUsage(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	routeRepo := &testapi.FakeRouteRepository{}

	fakeUI := callRouteMapper(t, []string{}, reqFactory, routeRepo, &testcmd.FakeRouteCreator{}, true)
	assert.True(t, fakeUI.FailedWithUsage)

	fakeUI = callRouteMapper(t, []string{"foo"}, reqFactory, routeRepo, &testcmd.FakeRouteCreator{}, true)
	assert.True(t, fakeUI.FailedWithUsage)

	fakeUI = callRouteMapper(t, []string{"foo", "bar"}, reqFactory, routeRepo, &testcmd.FakeRouteCreator{}, true)
	assert.False(t, fakeUI.FailedWithUsage)
}

func TestRouteMapperRequirements(t *testing.T) {
	routeRepo := &testapi.FakeRouteRepository{}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}

	callRouteMapper(t, []string{"-n", "my-host", "my-app", "my-domain.com"}, reqFactory, routeRepo, &testcmd.FakeRouteCreator{}, true)
	assert.True(t, testcmd.CommandDidPassRequirements)
	assert.Equal(t, reqFactory.ApplicationName, "my-app")
	assert.Equal(t, reqFactory.DomainName, "my-domain.com")
}

func TestRouteMapperWhenBinding(t *testing.T) {
	route := cf.Route{
		Guid:   "my-route-guid",
		Host:   "foo",
		Domain: cf.Domain{Guid: "my-domain-guid", Name: "example.com"},
	}
	app := cf.Application{Guid: "my-app-guid", Name: "my-app"}

	routeRepo := &testapi.FakeRouteRepository{}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, Application: app}
	routeCreator := &testcmd.FakeRouteCreator{ReservedRoute: route}

	ui := callRouteMapper(t, []string{"-n", "my-host", "my-app", "my-domain.com"}, reqFactory, routeRepo, routeCreator, true)

	assert.Contains(t, ui.Outputs[0], "Adding route")
	assert.Contains(t, ui.Outputs[0], "foo.example.com")
	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[0], "my-org")
	assert.Contains(t, ui.Outputs[0], "my-space")
	assert.Contains(t, ui.Outputs[0], "my-user")

	assert.Equal(t, route, routeRepo.BoundRoute)
	assert.Equal(t, app, routeRepo.BoundApp)

	assert.Contains(t, ui.Outputs[1], "OK")
}

func TestRouteMapperWhenUnbinding(t *testing.T) {
	route := cf.Route{
		Guid:   "my-route-guid",
		Host:   "foo",
		Domain: cf.Domain{Guid: "my-domain-guid", Name: "example.com"},
	}
	app := cf.Application{Guid: "my-app-guid", Name: "my-app"}

	routeRepo := &testapi.FakeRouteRepository{}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, Application: app}
	routeCreator := &testcmd.FakeRouteCreator{ReservedRoute: route}

	ui := callRouteMapper(t, []string{"-n", "my-host", "my-app", "my-domain.com"}, reqFactory, routeRepo, routeCreator, false)

	assert.Contains(t, ui.Outputs[0], "Removing route")
	assert.Contains(t, ui.Outputs[0], "foo.example.com")
	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[0], "my-org")
	assert.Contains(t, ui.Outputs[0], "my-space")
	assert.Contains(t, ui.Outputs[0], "my-user")

	assert.Equal(t, route, routeRepo.UnboundRoute)
	assert.Equal(t, app, routeRepo.UnboundApp)

	assert.Contains(t, ui.Outputs[1], "OK")
}

func TestRouteMapperWhenRouteNotReserved(t *testing.T) {
	domain := cf.Domain{Name: "my-domain.com"}
	route := cf.Route{Guid: "my-app-guid", Host: "my-host", Domain: domain}
	app := cf.Application{Guid: "my-app-guid", Name: "my-app"}

	routeRepo := &testapi.FakeRouteRepository{}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, Application: app}
	routeCreator := &testcmd.FakeRouteCreator{ReservedRoute: route}

	callRouteMapper(t, []string{"-n", "my-host", "my-app", "my-domain.com"}, reqFactory, routeRepo, routeCreator, true)

	assert.Equal(t, routeCreator.ReservedRoute, route)
}

func callRouteMapper(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory, routeRepo *testapi.FakeRouteRepository, createRoute *testcmd.FakeRouteCreator, bind bool) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	var ctxt *cli.Context
	if bind {
		ctxt = testcmd.NewContext("map-route", args)
	} else {
		ctxt = testcmd.NewContext("unmap-route", args)
	}

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)

	config := &configuration.Configuration{
		Space:        cf.Space{Name: "my-space"},
		Organization: cf.Organization{Name: "my-org"},
		AccessToken:  token,
	}

	cmd := NewRouteMapper(ui, config, routeRepo, createRoute, bind)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
