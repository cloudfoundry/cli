package route_test

import (
	. "github.com/cloudfoundry/cli/cf/commands/route"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/models"
	testapi "github.com/cloudfoundry/cli/testhelpers/api"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

var _ = Describe("unmap-route command", func() {
	var (
		ui                  *testterm.FakeUI
		configRepo          configuration.ReadWriter
		routeRepo           *testapi.FakeRouteRepository
		requirementsFactory *testreq.FakeReqFactory
	)

	BeforeEach(func() {
		ui = new(testterm.FakeUI)
		configRepo = testconfig.NewRepositoryWithDefaults()
		routeRepo = new(testapi.FakeRouteRepository)
		requirementsFactory = new(testreq.FakeReqFactory)
	})

	runCommand := func(args ...string) {
		cmd := NewUnmapRoute(ui, configRepo, routeRepo)
		testcmd.RunCommand(cmd, args, requirementsFactory)
	}

	Context("when the user is not logged in", func() {
		It("fails requirements", func() {
			runCommand("my-app", "some-domain.com")
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})
	})

	Context("when the user is logged in", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
		})

		Context("when the user does not provide two args", func() {
			It("fails with usage", func() {
				runCommand()
				Expect(ui.FailedWithUsage).To(BeTrue())
			})
		})

		Context("when the user provides an app and a domain", func() {
			BeforeEach(func() {
				requirementsFactory.Application = models.Application{ApplicationFields: models.ApplicationFields{
					Guid: "my-app-guid",
					Name: "my-app",
				}}
				requirementsFactory.Domain = models.DomainFields{
					Guid: "my-domain-guid",
					Name: "example.com",
				}
				routeRepo.FindByHostAndDomainReturns.Route = models.Route{
					Domain: requirementsFactory.Domain,
					Guid:   "my-route-guid",
					Host:   "foo",
				}
				runCommand("-n", "my-host", "my-app", "my-domain.com")
			})

			It("passes requirements", func() {
				Expect(testcmd.CommandDidPassRequirements).To(BeTrue())

			})

			It("reads the app and domain from its requirements", func() {
				Expect(requirementsFactory.ApplicationName).To(Equal("my-app"))
				Expect(requirementsFactory.DomainName).To(Equal("my-domain.com"))
			})

			It("unmaps the route", func() {
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Removing route", "foo.example.com", "my-app", "my-org", "my-space", "my-user"},
					[]string{"OK"},
				))

				Expect(routeRepo.UnboundRouteGuid).To(Equal("my-route-guid"))
				Expect(routeRepo.UnboundAppGuid).To(Equal("my-app-guid"))
			})
		})
	})
})
