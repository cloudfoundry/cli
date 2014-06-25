package route_test

import (
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/commands/route"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("create-route command", func() {
	var (
		ui                  *testterm.FakeUI
		routeRepo           *testapi.FakeRouteRepository
		requirementsFactory *testreq.FakeReqFactory
		config              configuration.ReadWriter
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		routeRepo = &testapi.FakeRouteRepository{}
		requirementsFactory = &testreq.FakeReqFactory{}
		config = testconfig.NewRepositoryWithDefaults()
	})

	runCommand := func(args ...string) {
		testcmd.RunCommand(NewCreateRoute(ui, config, routeRepo), args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("fails when not logged in", func() {
			requirementsFactory.TargetedOrgSuccess = true
			runCommand("my-space", "example.com", "-n", "foo")
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})

		It("fails when an org is not targeted", func() {
			requirementsFactory.LoginSuccess = true
			runCommand("my-space", "example.com", "-n", "foo")
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})

		It("fails with usage when not provided two args", func() {
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedOrgSuccess = true

			runCommand("my-space")
			Expect(ui.FailedWithUsage).To(BeTrue())
		})
	})

	Context("when logged in, targeted a space and given a domain that exists", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedOrgSuccess = true
			requirementsFactory.Domain = models.DomainFields{
				Guid: "domain-guid",
				Name: "example.com",
			}
			requirementsFactory.Space = models.Space{SpaceFields: models.SpaceFields{
				Guid: "my-space-guid",
				Name: "my-space",
			}}
		})

		It("creates routes, obviously", func() {
			runCommand("-n", "host", "my-space", "example.com")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Creating route", "host.example.com", "my-org", "my-space", "my-user"},
				[]string{"OK"},
			))

			Expect(routeRepo.CreateInSpaceHost).To(Equal("host"))
			Expect(routeRepo.CreateInSpaceDomainGuid).To(Equal("domain-guid"))
			Expect(routeRepo.CreateInSpaceSpaceGuid).To(Equal("my-space-guid"))
		})

		It("is idempotent", func() {
			routeRepo.CreateInSpaceErr = true
			routeRepo.FindByHostAndDomainReturns.Route = models.Route{
				Space:  requirementsFactory.Space.SpaceFields,
				Guid:   "my-route-guid",
				Host:   "host",
				Domain: requirementsFactory.Domain,
			}

			runCommand("-n", "host", "my-space", "example.com")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Creating route"},
				[]string{"OK"},
				[]string{"host.example.com", "already exists"},
			))

			Expect(routeRepo.CreateInSpaceHost).To(Equal("host"))
			Expect(routeRepo.CreateInSpaceDomainGuid).To(Equal("domain-guid"))
			Expect(routeRepo.CreateInSpaceSpaceGuid).To(Equal("my-space-guid"))
		})

		Describe("RouteCreator interface", func() {
			It("creates a route, given a domain and space", func() {
				createdRoute := models.Route{}
				createdRoute.Host = "my-host"
				createdRoute.Guid = "my-route-guid"
				routeRepo := &testapi.FakeRouteRepository{
					CreateInSpaceCreatedRoute: createdRoute,
				}

				cmd := NewCreateRoute(ui, config, routeRepo)
				route, apiErr := cmd.CreateRoute("my-host", requirementsFactory.Domain, requirementsFactory.Space.SpaceFields)

				Expect(apiErr).NotTo(HaveOccurred())
				Expect(route.Guid).To(Equal(createdRoute.Guid))
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Creating route", "my-host.example.com", "my-org", "my-space", "my-user"},
					[]string{"OK"},
				))

				Expect(routeRepo.CreateInSpaceHost).To(Equal("my-host"))
				Expect(routeRepo.CreateInSpaceDomainGuid).To(Equal("domain-guid"))
				Expect(routeRepo.CreateInSpaceSpaceGuid).To(Equal("my-space-guid"))
			})
		})
	})
})
