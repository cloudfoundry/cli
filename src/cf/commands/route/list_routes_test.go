package route_test

import (
	"cf"
	. "cf/commands/route"
	"cf/configuration"
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

func callListRoutes(t mr.TestingT, args []string, reqFactory *testreq.FakeReqFactory, routeRepo *testapi.FakeRouteRepository) (ui *testterm.FakeUI) {

	ui = &testterm.FakeUI{}

	ctxt := testcmd.NewContext("routes", args)

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

	cmd := NewListRoutes(ui, config, routeRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)

	return
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestListingRoutes", func() {
			domain := cf.DomainFields{}
			domain.Name = "example.com"
			domain2 := cf.DomainFields{}
			domain2.Name = "cfapps.com"
			domain3 := cf.DomainFields{}
			domain3.Name = "another-example.com"

			app1 := cf.ApplicationFields{}
			app1.Name = "dora"
			app2 := cf.ApplicationFields{}
			app2.Name = "dora2"

			app3 := cf.ApplicationFields{}
			app3.Name = "my-app"
			app4 := cf.ApplicationFields{}
			app4.Name = "my-app2"

			app5 := cf.ApplicationFields{}
			app5.Name = "july"

			route := cf.Route{}
			route.Host = "hostname-1"
			route.Domain = domain
			route.Apps = []cf.ApplicationFields{app1, app2}
			route2 := cf.Route{}
			route2.Host = "hostname-2"
			route2.Domain = domain2
			route2.Apps = []cf.ApplicationFields{app3, app4}
			route3 := cf.Route{}
			route3.Host = "hostname-3"
			route3.Domain = domain3
			route3.Apps = []cf.ApplicationFields{app5}
			routes := []cf.Route{route, route2, route3}

			routeRepo := &testapi.FakeRouteRepository{Routes: routes}

			ui := callListRoutes(mr.T(), []string{}, &testreq.FakeReqFactory{}, routeRepo)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Getting routes", "my-user"},
				{"host", "domain", "apps"},
				{"hostname-1", "example.com", "dora", "dora2"},
				{"hostname-2", "cfapps.com", "my-app", "my-app2"},
				{"hostname-3", "another-example.com", "july"},
			})
		})
		It("TestListingRoutesWhenNoneExist", func() {

			routes := []cf.Route{}
			routeRepo := &testapi.FakeRouteRepository{Routes: routes}

			ui := callListRoutes(mr.T(), []string{}, &testreq.FakeReqFactory{}, routeRepo)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Getting routes"},
				{"No routes found"},
			})
		})
		It("TestListingRoutesWhenFindFails", func() {

			routeRepo := &testapi.FakeRouteRepository{ListErr: true}

			ui := callListRoutes(mr.T(), []string{}, &testreq.FakeReqFactory{}, routeRepo)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Getting routes"},
				{"FAILED"},
			})
		})
	})
}
