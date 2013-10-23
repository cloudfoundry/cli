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

func TestReserveRouteRequirements(t *testing.T) {
	routeRepo := &testapi.FakeRouteRepository{}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}

	callReserveRoute(t, []string{"my-space", "example.com", "-n", "foo"}, reqFactory, routeRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)
	assert.Equal(t, reqFactory.SpaceName, "my-space")
	assert.Equal(t, reqFactory.DomainName, "example.com")

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: false}

	callReserveRoute(t, []string{"my-space", "example.com", "-n", "foo"}, reqFactory, routeRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)
}

func TestReserveRouteFailsWithUsage(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	routeRepo := &testapi.FakeRouteRepository{}
	ui := callReserveRoute(t, []string{""}, reqFactory, routeRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callReserveRoute(t, []string{"my-space"}, reqFactory, routeRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callReserveRoute(t, []string{"my-space", "example.com", "host"}, reqFactory, routeRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callReserveRoute(t, []string{"my-space", "example.com", "-n", "host"}, reqFactory, routeRepo)
	assert.False(t, ui.FailedWithUsage)

	ui = callReserveRoute(t, []string{"my-space", "example.com"}, reqFactory, routeRepo)
	assert.False(t, ui.FailedWithUsage)
}

func TestReserveRoute(t *testing.T) {
	space := cf.Space{Guid: "my-space-guid", Name: "my-space"}
	domain := cf.Domain{Guid: "domain-guid", Name: "example.com"}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, Space: space, Domain: domain}
	routeRepo := &testapi.FakeRouteRepository{}

	ui := callReserveRoute(t, []string{"-n", "host", "my-space", "example.com"}, reqFactory, routeRepo)

	assert.Contains(t, ui.Outputs[0], "Reserving route")
	assert.Contains(t, ui.Outputs[0], "host.example.com")
	assert.Contains(t, ui.Outputs[0], "my-org")
	assert.Contains(t, ui.Outputs[0], "my-space")
	assert.Contains(t, ui.Outputs[0], "my-user")

	assert.Equal(t, routeRepo.CreateInSpaceRoute, cf.Route{Host: "host", Domain: domain})
	assert.Equal(t, routeRepo.CreateInSpaceDomain, domain)
	assert.Equal(t, routeRepo.CreateInSpaceSpace, space)

	assert.Contains(t, ui.Outputs[1], "OK")
}

func callReserveRoute(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory, routeRepo *testapi.FakeRouteRepository) (fakeUI *testterm.FakeUI) {
	fakeUI = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("reserve-route", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)

	config := &configuration.Configuration{
		Space:        cf.Space{Name: "my-space"},
		Organization: cf.Organization{Name: "my-org"},
		AccessToken:  token,
	}

	cmd := NewReserveRoute(fakeUI, config, routeRepo)

	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
