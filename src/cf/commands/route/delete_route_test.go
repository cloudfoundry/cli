package route_test

import (
	. "cf/commands/route"
	"cf/configuration"
	"cf/models"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

func callDeleteRoute(t mr.TestingT, confirmation string, args []string, reqFactory *testreq.FakeReqFactory, routeRepo *testapi.FakeRouteRepository) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{
		Inputs: []string{confirmation},
	}
	ctxt := testcmd.NewContext("delete-route", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)
	org := models.OrganizationFields{}
	org.Name = "my-org"
	space := models.SpaceFields{}
	space.Name = "my-space"
	config := &configuration.Configuration{
		SpaceFields:        space,
		OrganizationFields: org,
		AccessToken:        token,
	}

	cmd := NewDeleteRoute(ui, config, routeRepo)

	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestDeleteRouteRequirements", func() {
			routeRepo := &testapi.FakeRouteRepository{}
			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}

			callDeleteRoute(mr.T(), "y", []string{"-n", "my-host", "example.com"}, reqFactory, routeRepo)
			assert.True(mr.T(), testcmd.CommandDidPassRequirements)

			reqFactory = &testreq.FakeReqFactory{LoginSuccess: false}
			callDeleteRoute(mr.T(), "y", []string{"-n", "my-host", "example.com"}, reqFactory, routeRepo)
			assert.False(mr.T(), testcmd.CommandDidPassRequirements)
		})
		It("TestDeleteRouteFailsWithUsage", func() {

			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
			routeRepo := &testapi.FakeRouteRepository{}
			ui := callDeleteRoute(mr.T(), "y", []string{}, reqFactory, routeRepo)
			assert.True(mr.T(), ui.FailedWithUsage)

			ui = callDeleteRoute(mr.T(), "y", []string{"example.com"}, reqFactory, routeRepo)
			assert.False(mr.T(), ui.FailedWithUsage)

			ui = callDeleteRoute(mr.T(), "y", []string{"-n", "my-host", "example.com"}, reqFactory, routeRepo)
			assert.False(mr.T(), ui.FailedWithUsage)
		})
		It("TestDeleteRouteWithConfirmation", func() {

			domain := models.DomainFields{}
			domain.Guid = "domain-guid"
			domain.Name = "example.com"
			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
			route := models.Route{}
			route.Guid = "route-guid"
			route.Host = "my-host"
			route.Domain = domain
			routeRepo := &testapi.FakeRouteRepository{
				FindByHostAndDomainRoute: route,
			}

			ui := callDeleteRoute(mr.T(), "y", []string{"-n", "my-host", "example.com"}, reqFactory, routeRepo)

			testassert.SliceContains(mr.T(), ui.Prompts, testassert.Lines{
				{"Really delete", "my-host"},
			})

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Deleting route", "my-host.example.com"},
				{"OK"},
			})
			assert.Equal(mr.T(), routeRepo.DeleteRouteGuid, "route-guid")
		})
		It("TestDeleteRouteWithForce", func() {

			domain := models.DomainFields{}
			domain.Guid = "domain-guid"
			domain.Name = "example.com"
			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
			route := models.Route{}
			route.Guid = "route-guid"
			route.Host = "my-host"
			route.Domain = domain
			routeRepo := &testapi.FakeRouteRepository{
				FindByHostAndDomainRoute: route,
			}

			ui := callDeleteRoute(mr.T(), "", []string{"-f", "-n", "my-host", "example.com"}, reqFactory, routeRepo)

			assert.Equal(mr.T(), len(ui.Prompts), 0)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Deleting", "my-host.example.com"},
				{"OK"},
			})
			assert.Equal(mr.T(), routeRepo.DeleteRouteGuid, "route-guid")
		})
		It("TestDeleteRouteWhenRouteDoesNotExist", func() {

			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
			routeRepo := &testapi.FakeRouteRepository{
				FindByHostAndDomainNotFound: true,
			}

			ui := callDeleteRoute(mr.T(), "y", []string{"-n", "my-host", "example.com"}, reqFactory, routeRepo)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Deleting", "my-host.example.com"},
				{"OK"},
				{"my-host", "does not exist"},
			})
		})
	})
}
