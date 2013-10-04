package route_test

import (
	"cf"
	. "cf/commands/route"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestReserveRouteRequirements(t *testing.T) {
	routeRepo := &testhelpers.FakeRouteRepository{}
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true}

	callReserveRoute([]string{"my-space", "example.com", "-n", "foo"}, reqFactory, routeRepo)
	assert.True(t, testhelpers.CommandDidPassRequirements)
	assert.Equal(t, reqFactory.SpaceName, "my-space")
	assert.Equal(t, reqFactory.DomainName, "example.com")

	reqFactory = &testhelpers.FakeReqFactory{LoginSuccess: false}

	callReserveRoute([]string{"my-space", "example.com", "-n", "foo"}, reqFactory, routeRepo)
	assert.False(t, testhelpers.CommandDidPassRequirements)
}

func TestReserveRouteFailsWithUsage(t *testing.T) {
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true}
	routeRepo := &testhelpers.FakeRouteRepository{}
	ui := callReserveRoute([]string{""}, reqFactory, routeRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callReserveRoute([]string{"my-space"}, reqFactory, routeRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callReserveRoute([]string{"my-space", "example.com", "host"}, reqFactory, routeRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callReserveRoute([]string{"my-space", "example.com", "-n", "host"}, reqFactory, routeRepo)
	assert.False(t, ui.FailedWithUsage)

	ui = callReserveRoute([]string{"my-space", "example.com"}, reqFactory, routeRepo)
	assert.False(t, ui.FailedWithUsage)
}

func TestReserveRoute(t *testing.T) {
	space := cf.Space{Guid: "my-space-guid", Name: "my-space"}
	domain := cf.Domain{Guid: "domain-guid", Name: "example.com"}
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true, Space: space, Domain: domain}
	routeRepo := &testhelpers.FakeRouteRepository{}

	ui := callReserveRoute([]string{"-n", "host", "my-space", "example.com"}, reqFactory, routeRepo)

	assert.Contains(t, ui.Outputs[0], "Reserving url route")
	assert.Contains(t, ui.Outputs[0], "host.example.com")
	assert.Contains(t, ui.Outputs[0], "my-space")

	assert.Equal(t, routeRepo.CreateInSpaceRoute, cf.Route{Host: "host", Domain: domain})
	assert.Equal(t, routeRepo.CreateInSpaceDomain, domain)
	assert.Equal(t, routeRepo.CreateInSpaceSpace, space)

	assert.Contains(t, ui.Outputs[1], "OK")
}

func callReserveRoute(args []string, reqFactory *testhelpers.FakeReqFactory, routeRepo *testhelpers.FakeRouteRepository) (fakeUI *testhelpers.FakeUI) {
	fakeUI = new(testhelpers.FakeUI)
	ctxt := testhelpers.NewContext("reserve-route", args)
	cmd := NewReserveRoute(fakeUI, routeRepo)

	testhelpers.RunCommand(cmd, ctxt, reqFactory)
	return
}
