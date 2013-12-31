package route_test

import (
	"cf"
	. "cf/commands/route"
	"cf/configuration"
	"github.com/codegangsta/cli"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestUnmapRouteFailsWithUsage(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	routeRepo := &testapi.FakeRouteRepository{}

	ui := callUnmapRoute(t, []string{}, reqFactory, routeRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callUnmapRoute(t, []string{"foo"}, reqFactory, routeRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callUnmapRoute(t, []string{"foo", "bar"}, reqFactory, routeRepo)
	assert.False(t, ui.FailedWithUsage)
}

func TestUnmapRouteRequirements(t *testing.T) {
	routeRepo := &testapi.FakeRouteRepository{}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}

	callUnmapRoute(t, []string{"-n", "my-host", "my-app", "my-domain.com"}, reqFactory, routeRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)
	assert.Equal(t, reqFactory.ApplicationName, "my-app")
	assert.Equal(t, reqFactory.DomainName, "my-domain.com")
}

func TestUnmapRouteWhenUnbinding(t *testing.T) {
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

	routeRepo := &testapi.FakeRouteRepository{FindByHostAndDomainRoute: route}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, Application: app, Domain: domain}

	ui := callUnmapRoute(t, []string{"-n", "my-host", "my-app", "my-domain.com"}, reqFactory, routeRepo)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Removing route", "foo.example.com", "my-app", "my-org", "my-space", "my-user"},
		{"OK"},
	})

	assert.Equal(t, routeRepo.UnboundRouteGuid, "my-route-guid")
	assert.Equal(t, routeRepo.UnboundAppGuid, "my-app-guid")
}

func callUnmapRoute(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory, routeRepo *testapi.FakeRouteRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	var ctxt *cli.Context = testcmd.NewContext("unmap-route", args)

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

	cmd := NewUnmapRoute(ui, config, routeRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
