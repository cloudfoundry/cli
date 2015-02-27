package route_test

import (
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	. "github.com/cloudfoundry/cli/cf/commands/route"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
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
		configRepo          core_config.ReadWriter
		routeRepo           *testapi.FakeRouteRepository
		requirementsFactory *testreq.FakeReqFactory
	)

	BeforeEach(func() {
		ui = new(testterm.FakeUI)
		configRepo = testconfig.NewRepositoryWithDefaults()
		routeRepo = new(testapi.FakeRouteRepository)
		requirementsFactory = new(testreq.FakeReqFactory)
	})

	runCommand := func(args ...string) bool {
		cmd := NewUnmapRoute(ui, configRepo, routeRepo)
		return testcmd.RunCommand(cmd, args, requirementsFactory)
	}

	Context("when the user is not logged in", func() {
		It("fails requirements", func() {
			Expect(runCommand("my-app", "some-domain.com")).To(BeFalse())
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
				requirementsFactory.Application = models.Application{
					ApplicationFields: models.ApplicationFields{
						Guid: "my-app-guid",
						Name: "my-app",
					},
					Routes: []models.RouteSummary{
						models.RouteSummary{
							Guid: "my-route-guid",
						},
					},
				}

				requirementsFactory.Domain = models.DomainFields{
					Guid: "my-domain-guid",
					Name: "example.com",
				}
				routeRepo.FindByHostAndDomainReturns.Route = models.Route{
					Domain: requirementsFactory.Domain,
					Guid:   "my-route-guid",
					Host:   "foo",
					Apps: []models.ApplicationFields{
						models.ApplicationFields{
							Guid: "my-app-guid",
							Name: "my-app",
						},
					},
				}
			})

			It("passes requirements", func() {
				Expect(runCommand("-n", "my-host", "my-app", "my-domain.com")).To(BeTrue())
			})

			It("reads the app and domain from its requirements", func() {
				runCommand("-n", "my-host", "my-app", "my-domain.com")
				Expect(requirementsFactory.ApplicationName).To(Equal("my-app"))
				Expect(requirementsFactory.DomainName).To(Equal("my-domain.com"))
			})

			It("unmaps the route", func() {
				runCommand("-n", "my-host", "my-app", "my-domain.com")
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Removing route", "foo.example.com", "my-app", "my-org", "my-space", "my-user"},
					[]string{"OK"},
				))

				Expect(ui.WarnOutputs).ToNot(ContainSubstrings(
					[]string{"Route to be unmapped is not currently mapped to the application."},
				))

				Expect(routeRepo.UnboundRouteGuid).To(Equal("my-route-guid"))
				Expect(routeRepo.UnboundAppGuid).To(Equal("my-app-guid"))
			})

			Context("when the route does not exist for the app", func() {
				BeforeEach(func() {
					requirementsFactory.Application = models.Application{
						ApplicationFields: models.ApplicationFields{
							Guid: "not-my-app-guid",
							Name: "my-app",
						},
						Routes: []models.RouteSummary{
							models.RouteSummary{
								Guid: "my-route-guid",
							},
						},
					}
				})

				It("informs the user the route did not exist on the applicaiton", func() {
					runCommand("-n", "my-host", "my-app", "my-domain.com")
					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Removing route", "foo.example.com", "my-app", "my-org", "my-space", "my-user"},
						[]string{"OK"},
					))

					Expect(ui.WarnOutputs).To(ContainSubstrings(
						[]string{"Route to be unmapped is not currently mapped to the application."},
					))
				})
			})
		})
	})
})
