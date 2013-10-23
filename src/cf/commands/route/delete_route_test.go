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

func TestDeleteRouteRequirements(t *testing.T) {
	routeRepo := &testapi.FakeRouteRepository{}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}

	callDeleteRoute(t, "y", []string{"-n", "my-host", "example.com"}, reqFactory, routeRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: false}
	callDeleteRoute(t, "y", []string{"-n", "my-host", "example.com"}, reqFactory, routeRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)
}

func TestDeleteRouteFailsWithUsage(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	routeRepo := &testapi.FakeRouteRepository{}
	ui := callDeleteRoute(t, "y", []string{}, reqFactory, routeRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callDeleteRoute(t, "y", []string{"example.com"}, reqFactory, routeRepo)
	assert.False(t, ui.FailedWithUsage)

	ui = callDeleteRoute(t, "y", []string{"-n", "my-host", "example.com"}, reqFactory, routeRepo)
	assert.False(t, ui.FailedWithUsage)
}

func TestDeleteRouteWithConfirmation(t *testing.T) {
	domain := cf.Domain{Guid: "domain-guid", Name: "example.com"}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	routeRepo := &testapi.FakeRouteRepository{
		FindByHostAndDomainRoute: cf.Route{Host: "my-host", Domain: domain},
	}

	ui := callDeleteRoute(t, "y", []string{"-n", "my-host", "example.com"}, reqFactory, routeRepo)

	assert.Contains(t, ui.Prompts[0], "Really delete")

	assert.Contains(t, ui.Outputs[0], "Deleting route")
	assert.Contains(t, ui.Outputs[0], "my-host.example.com")
	assert.Contains(t, ui.Outputs[0], "my-org")
	assert.Contains(t, ui.Outputs[0], "my-space")

	assert.Equal(t, routeRepo.DeleteRoute, cf.Route{Host: "my-host", Domain: domain})

	assert.Contains(t, ui.Outputs[1], "OK")
}

func TestDeleteRouteWithForce(t *testing.T) {
	domain := cf.Domain{Guid: "domain-guid", Name: "example.com"}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	routeRepo := &testapi.FakeRouteRepository{
		FindByHostAndDomainRoute: cf.Route{Host: "my-host", Domain: domain},
	}

	ui := callDeleteRoute(t, "", []string{"-f", "-n", "my-host", "example.com"}, reqFactory, routeRepo)

	assert.Equal(t, len(ui.Prompts), 0)

	assert.Contains(t, ui.Outputs[0], "Deleting")
	assert.Contains(t, ui.Outputs[0], "my-host.example.com")

	assert.Equal(t, routeRepo.DeleteRoute, cf.Route{Host: "my-host", Domain: domain})

	assert.Contains(t, ui.Outputs[1], "OK")
}

func TestDeleteRouteWhenRouteDoesNotExist(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	routeRepo := &testapi.FakeRouteRepository{
		FindByHostAndDomainNotFound: true,
	}

	ui := callDeleteRoute(t, "y", []string{"-n", "my-host", "example.com"}, reqFactory, routeRepo)

	assert.Contains(t, ui.Outputs[0], "Deleting")
	assert.Contains(t, ui.Outputs[0], "my-host.example.com")

	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[2], "my-host")
	assert.Contains(t, ui.Outputs[2], "does not exist")
}

func callDeleteRoute(t *testing.T, confirmation string, args []string, reqFactory *testreq.FakeReqFactory, routeRepo *testapi.FakeRouteRepository) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{
		Inputs: []string{confirmation},
	}
	ctxt := testcmd.NewContext("delete-route", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)

	config := &configuration.Configuration{
		Space:        cf.Space{Name: "my-space"},
		Organization: cf.Organization{Name: "my-org"},
		AccessToken:  token,
	}

	cmd := NewDeleteRoute(ui, config, routeRepo)

	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
