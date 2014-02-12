package route_test

import (
	. "cf/commands/route"
	"cf/configuration"
	"cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	mr "github.com/tjarratt/mr_t"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

var _ = Describe("Testing with ginkgo", func() {
	It("TestCreateRouteRequirements", func() {
		routeRepo := &testapi.FakeRouteRepository{}

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: false, TargetedOrgSuccess: true}
		callCreateRoute([]string{"my-space", "example.com", "-n", "foo"}, reqFactory, routeRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: false}
		callCreateRoute([]string{"my-space", "example.com", "-n", "foo"}, reqFactory, routeRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}
		callCreateRoute([]string{"my-space", "example.com", "-n", "foo"}, reqFactory, routeRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())
	})

	It("TestCreateRouteFailsWithUsage", func() {
		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}
		routeRepo := &testapi.FakeRouteRepository{}

		ui := callCreateRoute([]string{""}, reqFactory, routeRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callCreateRoute([]string{"my-space"}, reqFactory, routeRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callCreateRoute([]string{"my-space", "example.com", "host"}, reqFactory, routeRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callCreateRoute([]string{"my-space", "example.com", "-n", "host"}, reqFactory, routeRepo)
		Expect(ui.FailedWithUsage).To(BeFalse())

		ui = callCreateRoute([]string{"my-space", "example.com"}, reqFactory, routeRepo)
		Expect(ui.FailedWithUsage).To(BeFalse())
	})

	It("creates routes", func() {
		reqFactory := &testreq.FakeReqFactory{
			LoginSuccess:       true,
			TargetedOrgSuccess: true,
			Domain: models.DomainFields{
				Guid: "domain-guid",
				Name: "example.com",
			},
			Space: models.Space{SpaceFields: models.SpaceFields{
				Guid: "my-space-guid",
				Name: "my-space",
			}},
		}

		routeRepo := &testapi.FakeRouteRepository{}

		ui := callCreateRoute([]string{"-n", "host", "my-space", "example.com"}, reqFactory, routeRepo)

		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Creating route", "host.example.com", "my-org", "my-space", "my-user"},
			{"OK"},
		})

		Expect(routeRepo.CreateInSpaceHost).To(Equal("host"))
		Expect(routeRepo.CreateInSpaceDomainGuid).To(Equal("domain-guid"))
		Expect(routeRepo.CreateInSpaceSpaceGuid).To(Equal("my-space-guid"))
	})

	It("is idempotent", func() {
		domain := models.DomainFields{
			Guid: "domain-guid",
			Name: "example.com",
		}

		space := models.Space{SpaceFields: models.SpaceFields{
			Guid: "my-space-guid",
			Name: "my-space",
		}}

		reqFactory := &testreq.FakeReqFactory{
			LoginSuccess:       true,
			TargetedOrgSuccess: true,
			Domain:             domain,
			Space:              space,
		}

		routeRepo := &testapi.FakeRouteRepository{
			CreateInSpaceErr: true,
			FindByHostAndDomainRoute: models.Route{
				Space: space.SpaceFields,
				RouteSummary: models.RouteSummary{
					RouteFields: models.RouteFields{
						Guid: "my-route-guid",
						Host: "host",
					},
					Domain: domain,
				},
			},
		}

		ui := callCreateRoute([]string{"-n", "host", "my-space", "example.com"}, reqFactory, routeRepo)

		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Creating route"},
			{"OK"},
			{"host.example.com", "already exists"},
		})

		Expect(routeRepo.CreateInSpaceHost).To(Equal("host"))
		Expect(routeRepo.CreateInSpaceDomainGuid).To(Equal("domain-guid"))
		Expect(routeRepo.CreateInSpaceSpaceGuid).To(Equal("my-space-guid"))
	})

	It("TestRouteCreator", func() {
		space := models.SpaceFields{}
		space.Guid = "my-space-guid"
		space.Name = "my-space"
		domain := models.DomainFields{}
		domain.Guid = "domain-guid"
		domain.Name = "example.com"

		createdRoute := models.Route{}
		createdRoute.Host = "my-host"
		createdRoute.Guid = "my-route-guid"
		routeRepo := &testapi.FakeRouteRepository{
			CreateInSpaceCreatedRoute: createdRoute,
		}

		ui := new(testterm.FakeUI)
		configRepo := testconfig.NewRepositoryWithAccessToken(configuration.TokenInfo{Username: "my-user"})
		orgFields := models.OrganizationFields{}
		orgFields.Name = "my-org"
		configRepo.SetOrganizationFields(orgFields)

		cmd := NewCreateRoute(ui, configRepo, routeRepo)
		route, apiResponse := cmd.CreateRoute("my-host", domain, space)

		Expect(route.Guid).To(Equal(createdRoute.Guid))

		Expect(apiResponse.IsSuccessful()).To(BeTrue())

		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Creating route", "my-host.example.com", "my-org", "my-space", "my-user"},
			{"OK"},
		})

		Expect(routeRepo.CreateInSpaceHost).To(Equal("my-host"))
		Expect(routeRepo.CreateInSpaceDomainGuid).To(Equal("domain-guid"))
		Expect(routeRepo.CreateInSpaceSpaceGuid).To(Equal("my-space-guid"))
	})
})

func callCreateRoute(args []string, reqFactory *testreq.FakeReqFactory, routeRepo *testapi.FakeRouteRepository) (fakeUI *testterm.FakeUI) {
	fakeUI = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("create-route", args)

	configRepo := testconfig.NewRepositoryWithAccessToken(configuration.TokenInfo{Username: "my-user"})

	spaceFields := models.SpaceFields{}
	spaceFields.Name = "my-space"

	orgFields := models.OrganizationFields{}
	orgFields.Name = "my-org"
	configRepo.SetSpaceFields(spaceFields)
	configRepo.SetOrganizationFields(orgFields)

	cmd := NewCreateRoute(fakeUI, configRepo, routeRepo)

	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
