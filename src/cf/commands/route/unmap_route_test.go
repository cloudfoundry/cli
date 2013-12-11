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

func TestUnmapRouteFailsWithUsage(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	routeRepo := &testapi.FakeRouteRepository{}

	fakeUI := callUnmapRoute(t, []string{}, reqFactory, routeRepo)
	assert.True(t, fakeUI.FailedWithUsage)

	fakeUI = callUnmapRoute(t, []string{"foo"}, reqFactory, routeRepo)
	assert.True(t, fakeUI.FailedWithUsage)

	fakeUI = callUnmapRoute(t, []string{"foo", "bar"}, reqFactory, routeRepo)
	assert.False(t, fakeUI.FailedWithUsage)
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

	assert.Contains(t, ui.Outputs[0], "Removing route")
	assert.Contains(t, ui.Outputs[0], "foo.example.com")
	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[0], "my-org")
	assert.Contains(t, ui.Outputs[0], "my-space")
	assert.Contains(t, ui.Outputs[0], "my-user")

	assert.Equal(t, routeRepo.UnboundRouteGuid, "my-route-guid")
	assert.Equal(t, routeRepo.UnboundAppGuid, "my-app-guid")

	assert.Contains(t, ui.Outputs[1], "OK")
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
