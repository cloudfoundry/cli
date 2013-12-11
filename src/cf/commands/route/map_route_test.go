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

func TestMapRouteFailsWithUsage(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	routeRepo := &testapi.FakeRouteRepository{}

	fakeUI := callMapRoute(t, []string{}, reqFactory, routeRepo, &testcmd.FakeRouteCreator{})
	assert.True(t, fakeUI.FailedWithUsage)

	fakeUI = callMapRoute(t, []string{"foo"}, reqFactory, routeRepo, &testcmd.FakeRouteCreator{})
	assert.True(t, fakeUI.FailedWithUsage)

	fakeUI = callMapRoute(t, []string{"foo", "bar"}, reqFactory, routeRepo, &testcmd.FakeRouteCreator{})
	assert.False(t, fakeUI.FailedWithUsage)
}

func TestMapRouteRequirements(t *testing.T) {
	routeRepo := &testapi.FakeRouteRepository{}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}

	callMapRoute(t, []string{"-n", "my-host", "my-app", "my-domain.com"}, reqFactory, routeRepo, &testcmd.FakeRouteCreator{})
	assert.True(t, testcmd.CommandDidPassRequirements)
	assert.Equal(t, reqFactory.ApplicationName, "my-app")
	assert.Equal(t, reqFactory.DomainName, "my-domain.com")
}

func TestMapRouteWhenBinding(t *testing.T) {

	domain := cf.Domain{}
	domain.Guid = "my-domain-guid"
	domain.Name = "example.com"
	route := cf.Route{}
	route.Guid = "my-route-guid"
	route.Host = "foo"
	route.Domain = domain.DomainFields

	app := cf.Application{}
	app.Guid = "my-app-guid"
	app.Name = "my-app"

	routeRepo := &testapi.FakeRouteRepository{}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, Application: app, Domain: domain}
	routeCreator := &testcmd.FakeRouteCreator{ReservedRoute: route}

	ui := callMapRoute(t, []string{"-n", "my-host", "my-app", "my-domain.com"}, reqFactory, routeRepo, routeCreator)

	assert.Contains(t, ui.Outputs[0], "Adding route")
	assert.Contains(t, ui.Outputs[0], "foo.example.com")
	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[0], "my-org")
	assert.Contains(t, ui.Outputs[0], "my-space")
	assert.Contains(t, ui.Outputs[0], "my-user")

	assert.Equal(t, routeRepo.BoundRouteGuid, "my-route-guid")
	assert.Equal(t, routeRepo.BoundAppGuid, "my-app-guid")

	assert.Contains(t, ui.Outputs[1], "OK")
}

func TestMapRouteWhenRouteNotReserved(t *testing.T) {
	domain := cf.DomainFields{}
	domain.Name = "my-domain.com"
	route := cf.Route{}
	route.Guid = "my-app-guid"
	route.Host = "my-host"
	route.Domain = domain
	app := cf.Application{}
	app.Guid = "my-app-guid"
	app.Name = "my-app"

	routeRepo := &testapi.FakeRouteRepository{}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, Application: app}
	routeCreator := &testcmd.FakeRouteCreator{ReservedRoute: route}

	callMapRoute(t, []string{"-n", "my-host", "my-app", "my-domain.com"}, reqFactory, routeRepo, routeCreator)

	assert.Equal(t, routeCreator.ReservedRoute, route)
}

func callMapRoute(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory, routeRepo *testapi.FakeRouteRepository, createRoute *testcmd.FakeRouteCreator) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	var ctxt *cli.Context = testcmd.NewContext("map-route", args)

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

	cmd := NewMapRoute(ui, config, routeRepo, createRoute)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
