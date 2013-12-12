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
	routeRepo := &testapi.FakeRouteRepository{}

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: false, TargetedOrgSuccess: true}
	callCreateRoute(t, []string{"my-space", "example.com", "-n", "foo"}, reqFactory, routeRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: false}
	callCreateRoute(t, []string{"my-space", "example.com", "-n", "foo"}, reqFactory, routeRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}
	callCreateRoute(t, []string{"my-space", "example.com", "-n", "foo"}, reqFactory, routeRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)
}

func TestCreateRouteFailsWithUsage(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}
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
	space := cf.SpaceFields{}
	space.Guid = "my-space-guid"
	space.Name = "my-space"
	domain := cf.DomainFields{}
	domain.Guid = "domain-guid"
	domain.Name = "example.com"
	reqFactory := &testreq.FakeReqFactory{
		LoginSuccess:       true,
		TargetedOrgSuccess: true,
		Domain:             cf.Domain{DomainFields: domain},
		Space:              cf.Space{SpaceFields: space},
	}
	routeRepo := &testapi.FakeRouteRepository{}

	ui := callCreateRoute(t, []string{"-n", "host", "my-space", "example.com"}, reqFactory, routeRepo)

	assert.Contains(t, ui.Outputs[0], "Creating route")
	assert.Contains(t, ui.Outputs[0], "host.example.com")
	assert.Contains(t, ui.Outputs[0], "my-org")
	assert.Contains(t, ui.Outputs[0], "my-space")
	assert.Contains(t, ui.Outputs[0], "my-user")
	assert.Contains(t, ui.Outputs[1], "OK")

	assert.Equal(t, routeRepo.CreateInSpaceHost, "host")
	assert.Equal(t, routeRepo.CreateInSpaceDomainGuid, "domain-guid")
	assert.Equal(t, routeRepo.CreateInSpaceSpaceGuid, "my-space-guid")

}

func TestCreateRouteIsIdempotent(t *testing.T) {
	space := cf.SpaceFields{}
	space.Guid = "my-space-guid"
	space.Name = "my-space"
	domain := cf.DomainFields{}
	domain.Guid = "domain-guid"
	domain.Name = "example.com"
	reqFactory := &testreq.FakeReqFactory{
		LoginSuccess:       true,
		TargetedOrgSuccess: true,
		Domain:             cf.Domain{DomainFields: domain},
		Space:              cf.Space{SpaceFields: space},
	}

	route := cf.Route{}
	route.Guid = "my-route-guid"
	route.Host = "host"
	route.Domain = domain
	route.Space = space
	routeRepo := &testapi.FakeRouteRepository{
		CreateInSpaceErr:         true,
		FindByHostAndDomainRoute: route,
	}

	ui := callCreateRoute(t, []string{"-n", "host", "my-space", "example.com"}, reqFactory, routeRepo)

	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[2], "host.example.com")
	assert.Contains(t, ui.Outputs[2], "already exists")
	assert.Equal(t, routeRepo.CreateInSpaceHost, "host")
	assert.Equal(t, routeRepo.CreateInSpaceDomainGuid, "domain-guid")
	assert.Equal(t, routeRepo.CreateInSpaceSpaceGuid, "my-space-guid")

}

func TestRouteCreator(t *testing.T) {
	space := cf.SpaceFields{}
	space.Guid = "my-space-guid"
	space.Name = "my-space"
	domain := cf.DomainFields{}
	domain.Guid = "domain-guid"
	domain.Name = "example.com"

	createdRoute := cf.Route{}
	createdRoute.Host = "my-host"
	createdRoute.Guid = "my-route-guid"
	routeRepo := &testapi.FakeRouteRepository{
		CreateInSpaceCreatedRoute: createdRoute,
	}

	ui := new(testterm.FakeUI)
	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)
	org := cf.OrganizationFields{}
	org.Name = "my-org"
	config := &configuration.Configuration{
		OrganizationFields: org,
		AccessToken:        token,
	}

	cmd := NewCreateRoute(ui, config, routeRepo)
	route, apiResponse := cmd.CreateRoute("my-host", domain, space)

	assert.Equal(t, route.Guid, createdRoute.Guid)

	assert.True(t, apiResponse.IsSuccessful())
	assert.Contains(t, ui.Outputs[0], "Creating route")
	assert.Contains(t, ui.Outputs[0], "my-host.example.com")
	assert.Contains(t, ui.Outputs[0], "my-org")
	assert.Contains(t, ui.Outputs[0], "my-space")
	assert.Contains(t, ui.Outputs[0], "my-user")

	assert.Equal(t, routeRepo.CreateInSpaceHost, "my-host")
	assert.Equal(t, routeRepo.CreateInSpaceDomainGuid, "domain-guid")
	assert.Equal(t, routeRepo.CreateInSpaceSpaceGuid, "my-space-guid")
	assert.Contains(t, ui.Outputs[1], "OK")
}

func callCreateRoute(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory, routeRepo *testapi.FakeRouteRepository) (fakeUI *testterm.FakeUI) {
	fakeUI = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("create-route", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)
	org := cf.OrganizationFields{}
	org.Name = "my-org"
	space := cf.SpaceFields{}
	space.Name = "my-space"
	config := &configuration.Configuration{
		SpaceFields:        space,
		OrganizationFields: org,
		AccessToken:        token,
	}

	cmd := NewCreateRoute(fakeUI, config, routeRepo)

	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
