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

func TestCreateRouteRequirements(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	routeRepo := &testapi.FakeRouteRepository{}

	callCreateRoute(t, []string{"my-space", "example.com", "-n", "foo"}, reqFactory, routeRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: false}

	callCreateRoute(t, []string{"my-space", "example.com", "-n", "foo"}, reqFactory, routeRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)
}

func TestCreateRouteFailsWithUsage(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	routeRepo := &testapi.FakeRouteRepository{}

	ui := callCreateRoute(t, []string{""}, reqFactory, routeRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callCreateRoute(t, []string{"my-space"}, reqFactory, routeRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callCreateRoute(t, []string{"my-space", "example.com", "host"}, reqFactory, routeRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callCreateRoute(t, []string{"my-space", "example.com", "-n", "host"}, reqFactory, routeRepo)
	assert.False(t, ui.FailedWithUsage)

	ui = callCreateRoute(t, []string{"my-space", "example.com"}, reqFactory, routeRepo)
	assert.False(t, ui.FailedWithUsage)
}

func TestCreateRoute(t *testing.T) {
	space := cf.Space{Guid: "my-space-guid", Name: "my-space"}
	domain := cf.Domain{Guid: "domain-guid", Name: "example.com"}
	reqFactory := &testreq.FakeReqFactory{
		LoginSuccess: true,
		Domain:       domain,
		Space:        space,
	}
	routeRepo := &testapi.FakeRouteRepository{}

	ui := callCreateRoute(t, []string{"-n", "host", "my-space", "example.com"}, reqFactory, routeRepo)

	assert.Contains(t, ui.Outputs[0], "Reserving route")
	assert.Contains(t, ui.Outputs[0], "host.example.com")
	assert.Contains(t, ui.Outputs[0], "my-org")
	assert.Contains(t, ui.Outputs[0], "my-space")
	assert.Contains(t, ui.Outputs[0], "my-user")

	assert.Equal(t, routeRepo.CreateInSpaceRoute, cf.Route{Host: "host", Domain: domain})
	assert.Equal(t, routeRepo.CreateInSpaceDomain, domain)
	assert.Equal(t, routeRepo.CreateInSpaceSpace, space)

}

func TestRouteCreator(t *testing.T) {
	space := cf.Space{Guid: "my-space-guid", Name: "my-space"}
	domain := cf.Domain{Guid: "domain-guid", Name: "example.com"}
	routeRepo := &testapi.FakeRouteRepository{
		CreateInSpaceCreatedRoute: cf.Route{
			Host: "my-host",
			Guid: "my-route-guid",
			Domain: cf.Domain{
				Name: "example.com",
			},
		},
	}

	ui := new(testterm.FakeUI)
	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)
	config := &configuration.Configuration{
		Organization: cf.Organization{Name: "my-org"},
		AccessToken:  token,
	}

	cmd := NewCreateRoute(ui, config, routeRepo)
	route, apiResponse := cmd.CreateRoute("my-host", domain, space)

	assert.True(t, apiResponse.IsSuccessful())
	assert.Contains(t, ui.Outputs[0], "Reserving route")
	assert.Contains(t, ui.Outputs[0], "my-host.example.com")
	assert.Contains(t, ui.Outputs[0], "my-org")
	assert.Contains(t, ui.Outputs[0], "my-space")
	assert.Contains(t, ui.Outputs[0], "my-user")

	assert.Equal(t, routeRepo.CreateInSpaceRoute.Host, "my-host")
	assert.Equal(t, routeRepo.CreateInSpaceDomain, domain)
	assert.Equal(t, routeRepo.CreateInSpaceSpace, space)
	assert.Equal(t, routeRepo.CreateInSpaceCreatedRoute, route)
	assert.Contains(t, ui.Outputs[1], "OK")
}

func callCreateRoute(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory, routeRepo *testapi.FakeRouteRepository) (fakeUI *testterm.FakeUI) {
	fakeUI = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("create-route", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)

	config := &configuration.Configuration{
		Space:        cf.Space{Name: "my-space"},
		Organization: cf.Organization{Name: "my-org"},
		AccessToken:  token,
	}

	cmd := NewCreateRoute(fakeUI, config, routeRepo)

	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
