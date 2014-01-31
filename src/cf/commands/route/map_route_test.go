package route_test

import (
	"cf"
	. "cf/commands/route"
	"cf/configuration"
	"github.com/codegangsta/cli"
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

func callMapRoute(t mr.TestingT, args []string, reqFactory *testreq.FakeReqFactory, routeRepo *testapi.FakeRouteRepository, createRoute *testcmd.FakeRouteCreator) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	var ctxt *cli.Context = testcmd.NewContext("map-route", args)

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

	cmd := NewMapRoute(ui, config, routeRepo, createRoute)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestMapRouteFailsWithUsage", func() {
			reqFactory := &testreq.FakeReqFactory{}
			routeRepo := &testapi.FakeRouteRepository{}

			ui := callMapRoute(mr.T(), []string{}, reqFactory, routeRepo, &testcmd.FakeRouteCreator{})
			assert.True(mr.T(), ui.FailedWithUsage)

			ui = callMapRoute(mr.T(), []string{"foo"}, reqFactory, routeRepo, &testcmd.FakeRouteCreator{})
			assert.True(mr.T(), ui.FailedWithUsage)

			ui = callMapRoute(mr.T(), []string{"foo", "bar"}, reqFactory, routeRepo, &testcmd.FakeRouteCreator{})
			assert.False(mr.T(), ui.FailedWithUsage)
		})
		It("TestMapRouteRequirements", func() {

			routeRepo := &testapi.FakeRouteRepository{}
			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}

			callMapRoute(mr.T(), []string{"-n", "my-host", "my-app", "my-domain.com"}, reqFactory, routeRepo, &testcmd.FakeRouteCreator{})
			assert.True(mr.T(), testcmd.CommandDidPassRequirements)
			assert.Equal(mr.T(), reqFactory.ApplicationName, "my-app")
			assert.Equal(mr.T(), reqFactory.DomainName, "my-domain.com")
		})
		It("TestMapRouteWhenBinding", func() {

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

			routeRepo := &testapi.FakeRouteRepository{}
			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, Application: app, Domain: domain}
			routeCreator := &testcmd.FakeRouteCreator{ReservedRoute: route}

			ui := callMapRoute(mr.T(), []string{"-n", "my-host", "my-app", "my-domain.com"}, reqFactory, routeRepo, routeCreator)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Adding route", "foo.example.com", "my-app", "my-org", "my-space", "my-user"},
				{"OK"},
			})

			assert.Equal(mr.T(), routeRepo.BoundRouteGuid, "my-route-guid")
			assert.Equal(mr.T(), routeRepo.BoundAppGuid, "my-app-guid")
		})
		It("TestMapRouteWhenRouteNotReserved", func() {

			domain := cf.DomainFields{}
			domain.Name = "my-domain.com"
			route := cf.Route{}
			route.Guid = "my-app-guid"
			route.Host = "my-host"
			route.Domain = domain
			app := cf.Application{}
			app.Guid = "my-app-guid"
			app.Name = "my-app"

			routeRepo := &testapi.FakeRouteRepository{}
			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, Application: app}
			routeCreator := &testcmd.FakeRouteCreator{ReservedRoute: route}

			callMapRoute(mr.T(), []string{"-n", "my-host", "my-app", "my-domain.com"}, reqFactory, routeRepo, routeCreator)

			assert.Equal(mr.T(), routeCreator.ReservedRoute, route)
		})
	})
}
