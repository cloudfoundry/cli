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

func callCreateRoute(t mr.TestingT, args []string, reqFactory *testreq.FakeReqFactory, routeRepo *testapi.FakeRouteRepository) (fakeUI *testterm.FakeUI) {
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
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestCreateRouteRequirements", func() {
			routeRepo := &testapi.FakeRouteRepository{}

			reqFactory := &testreq.FakeReqFactory{LoginSuccess: false, TargetedOrgSuccess: true}
			callCreateRoute(mr.T(), []string{"my-space", "example.com", "-n", "foo"}, reqFactory, routeRepo)
			assert.False(mr.T(), testcmd.CommandDidPassRequirements)

			reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: false}
			callCreateRoute(mr.T(), []string{"my-space", "example.com", "-n", "foo"}, reqFactory, routeRepo)
			assert.False(mr.T(), testcmd.CommandDidPassRequirements)

			reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}
			callCreateRoute(mr.T(), []string{"my-space", "example.com", "-n", "foo"}, reqFactory, routeRepo)
			assert.True(mr.T(), testcmd.CommandDidPassRequirements)
		})
		It("TestCreateRouteFailsWithUsage", func() {

			reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}
			routeRepo := &testapi.FakeRouteRepository{}

			ui := callCreateRoute(mr.T(), []string{""}, reqFactory, routeRepo)
			assert.True(mr.T(), ui.FailedWithUsage)

			ui = callCreateRoute(mr.T(), []string{"my-space"}, reqFactory, routeRepo)
			assert.True(mr.T(), ui.FailedWithUsage)

			ui = callCreateRoute(mr.T(), []string{"my-space", "example.com", "host"}, reqFactory, routeRepo)
			assert.True(mr.T(), ui.FailedWithUsage)

			ui = callCreateRoute(mr.T(), []string{"my-space", "example.com", "-n", "host"}, reqFactory, routeRepo)
			assert.False(mr.T(), ui.FailedWithUsage)

			ui = callCreateRoute(mr.T(), []string{"my-space", "example.com"}, reqFactory, routeRepo)
			assert.False(mr.T(), ui.FailedWithUsage)
		})
		It("TestCreateRoute", func() {

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

			ui := callCreateRoute(mr.T(), []string{"-n", "host", "my-space", "example.com"}, reqFactory, routeRepo)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Creating route", "host.example.com", "my-org", "my-space", "my-user"},
				{"OK"},
			})

			assert.Equal(mr.T(), routeRepo.CreateInSpaceHost, "host")
			assert.Equal(mr.T(), routeRepo.CreateInSpaceDomainGuid, "domain-guid")
			assert.Equal(mr.T(), routeRepo.CreateInSpaceSpaceGuid, "my-space-guid")
		})
		It("TestCreateRouteIsIdempotent", func() {

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

			ui := callCreateRoute(mr.T(), []string{"-n", "host", "my-space", "example.com"}, reqFactory, routeRepo)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Creating route"},
				{"OK"},
				{"host.example.com", "already exists"},
			})

			assert.Equal(mr.T(), routeRepo.CreateInSpaceHost, "host")
			assert.Equal(mr.T(), routeRepo.CreateInSpaceDomainGuid, "domain-guid")
			assert.Equal(mr.T(), routeRepo.CreateInSpaceSpaceGuid, "my-space-guid")
		})
		It("TestRouteCreator", func() {

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
			assert.NoError(mr.T(), err)
			org := cf.OrganizationFields{}
			org.Name = "my-org"
			config := &configuration.Configuration{
				OrganizationFields: org,
				AccessToken:        token,
			}

			cmd := NewCreateRoute(ui, config, routeRepo)
			route, apiResponse := cmd.CreateRoute("my-host", domain, space)

			assert.Equal(mr.T(), route.Guid, createdRoute.Guid)

			assert.True(mr.T(), apiResponse.IsSuccessful())

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Creating route", "my-host.example.com", "my-org", "my-space", "my-user"},
				{"OK"},
			})

			assert.Equal(mr.T(), routeRepo.CreateInSpaceHost, "my-host")
			assert.Equal(mr.T(), routeRepo.CreateInSpaceDomainGuid, "domain-guid")
			assert.Equal(mr.T(), routeRepo.CreateInSpaceSpaceGuid, "my-space-guid")
		})
	})
}
