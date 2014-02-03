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

func callUnmapRoute(t mr.TestingT, args []string, reqFactory *testreq.FakeReqFactory, routeRepo *testapi.FakeRouteRepository) (ui *testterm.FakeUI) {
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
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestUnmapRouteFailsWithUsage", func() {
			reqFactory := &testreq.FakeReqFactory{}
			routeRepo := &testapi.FakeRouteRepository{}

			ui := callUnmapRoute(mr.T(), []string{}, reqFactory, routeRepo)
			assert.True(mr.T(), ui.FailedWithUsage)

			ui = callUnmapRoute(mr.T(), []string{"foo"}, reqFactory, routeRepo)
			assert.True(mr.T(), ui.FailedWithUsage)

			ui = callUnmapRoute(mr.T(), []string{"foo", "bar"}, reqFactory, routeRepo)
			assert.False(mr.T(), ui.FailedWithUsage)
		})
		It("TestUnmapRouteRequirements", func() {

			routeRepo := &testapi.FakeRouteRepository{}
			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}

			callUnmapRoute(mr.T(), []string{"-n", "my-host", "my-app", "my-domain.com"}, reqFactory, routeRepo)
			assert.True(mr.T(), testcmd.CommandDidPassRequirements)
			assert.Equal(mr.T(), reqFactory.ApplicationName, "my-app")
			assert.Equal(mr.T(), reqFactory.DomainName, "my-domain.com")
		})
		It("TestUnmapRouteWhenUnbinding", func() {

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

			ui := callUnmapRoute(mr.T(), []string{"-n", "my-host", "my-app", "my-domain.com"}, reqFactory, routeRepo)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Removing route", "foo.example.com", "my-app", "my-org", "my-space", "my-user"},
				{"OK"},
			})

			assert.Equal(mr.T(), routeRepo.UnboundRouteGuid, "my-route-guid")
			assert.Equal(mr.T(), routeRepo.UnboundAppGuid, "my-app-guid")
		})
	})
}
