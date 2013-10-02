package route_test

import (
	"cf"
	. "cf/commands/route"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestMapRouteFailsWithUsage(t *testing.T) {
	reqFactory := &testhelpers.FakeReqFactory{}
	routeRepo := &testhelpers.FakeRouteRepository{}

	fakeUI := callMapRoute([]string{}, reqFactory, routeRepo)
	assert.True(t, fakeUI.FailedWithUsage)

	fakeUI = callMapRoute([]string{"foo"}, reqFactory, routeRepo)
	assert.True(t, fakeUI.FailedWithUsage)

	fakeUI = callMapRoute([]string{"foo", "bar"}, reqFactory, routeRepo)
	assert.False(t, fakeUI.FailedWithUsage)
}

func TestMapRouteRequirements(t *testing.T) {
	routeRepo := &testhelpers.FakeRouteRepository{}
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true}

	callMapRoute([]string{"-n", "my-host", "my-app", "my-domain.com"}, reqFactory, routeRepo)
	assert.True(t, testhelpers.CommandDidPassRequirements)
	assert.Equal(t, reqFactory.ApplicationName, "my-app")
	assert.Equal(t, reqFactory.RouteHost, "my-host")
	assert.Equal(t, reqFactory.RouteDomain, "my-domain.com")
}

func TestMapRoute(t *testing.T) {
	route := cf.Route{
		Guid:   "my-route-guid",
		Host:   "foo",
		Domain: cf.Domain{Guid: "my-domain-guid", Name: "example.com"},
	}
	app := cf.Application{Guid: "my-app-guid", Name: "my-app"}

	routeRepo := &testhelpers.FakeRouteRepository{}
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true, Route: route, Application: app}

	ui := callMapRoute([]string{"-n", "my-host", "my-app", "my-domain.com"}, reqFactory, routeRepo)

	assert.Contains(t, ui.Outputs[0], "Adding url route")
	assert.Contains(t, ui.Outputs[0], "foo.example.com")
	assert.Contains(t, ui.Outputs[0], "my-app")

	assert.Equal(t, route, routeRepo.BoundRoute)
	assert.Equal(t, app, routeRepo.BoundApp)

	assert.Contains(t, ui.Outputs[1], "OK")
}

func callMapRoute(args []string, reqFactory *testhelpers.FakeReqFactory, routeRepo *testhelpers.FakeRouteRepository) (ui *testhelpers.FakeUI) {
	ui = new(testhelpers.FakeUI)
	ctxt := testhelpers.NewContext("map-route", args)
	cmd := NewMapRoute(ui, routeRepo)
	testhelpers.RunCommand(cmd, ctxt, reqFactory)
	return
}
