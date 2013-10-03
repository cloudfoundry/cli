package route_test

import (
	"cf"
	. "cf/commands/route"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestParkRouteRequirements(t *testing.T) {
	routeRepo := &testhelpers.FakeRouteRepository{}
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true}

	callParkRoute([]string{"my-space", "example.com", "-n", "foo"}, reqFactory, routeRepo)
	assert.True(t, testhelpers.CommandDidPassRequirements)
	assert.Equal(t, reqFactory.SpaceName, "my-space")
	assert.Equal(t, reqFactory.DomainName, "example.com")

	reqFactory = &testhelpers.FakeReqFactory{LoginSuccess: false}

	callParkRoute([]string{"my-space", "example.com", "-n", "foo"}, reqFactory, routeRepo)
	assert.False(t, testhelpers.CommandDidPassRequirements)
}

func TestParkRouteFailsWithUsage(t *testing.T) {
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true}
	routeRepo := &testhelpers.FakeRouteRepository{}
	ui := callParkRoute([]string{""}, reqFactory, routeRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callParkRoute([]string{"my-space"}, reqFactory, routeRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callParkRoute([]string{"my-space", "example.com", "host"}, reqFactory, routeRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callParkRoute([]string{"my-space", "example.com", "-n", "host"}, reqFactory, routeRepo)
	assert.False(t, ui.FailedWithUsage)

	ui = callParkRoute([]string{"my-space", "example.com"}, reqFactory, routeRepo)
	assert.False(t, ui.FailedWithUsage)
}

func TestParkRoute(t *testing.T) {
	space := cf.Space{Guid: "my-space-guid", Name: "my-space"}
	domain := cf.Domain{Guid: "domain-guid", Name: "example.com"}
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true, Space: space, Domain: domain}
	routeRepo := &testhelpers.FakeRouteRepository{}

	ui := callParkRoute([]string{"-n", "host", "my-space", "example.com"}, reqFactory, routeRepo)

	assert.Contains(t, ui.Outputs[0], "Adding url route")
	assert.Contains(t, ui.Outputs[0], "host.example.com")
	assert.Contains(t, ui.Outputs[0], "my-space")

	assert.Equal(t, routeRepo.CreateInSpaceRoute, cf.Route{Host: "host", Domain: domain})
	assert.Equal(t, routeRepo.CreateInSpaceDomain, domain)
	assert.Equal(t, routeRepo.CreateInSpaceSpace, space)

	assert.Contains(t, ui.Outputs[1], "OK")
}

func callParkRoute(args []string, reqFactory *testhelpers.FakeReqFactory, routeRepo *testhelpers.FakeRouteRepository) (fakeUI *testhelpers.FakeUI) {
	fakeUI = new(testhelpers.FakeUI)
	ctxt := testhelpers.NewContext("park-route", args)
	cmd := NewParkRoute(fakeUI, routeRepo)

	testhelpers.RunCommand(cmd, ctxt, reqFactory)
	return
}
