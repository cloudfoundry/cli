package route_test

import (
	"errors"

	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/flags"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/cli/cf/commands/route"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

var _ = Describe("routes command", func() {
	var (
		ui                  *testterm.FakeUI
		routeRepo           *testapi.FakeRouteRepository
		domainRepo          *testapi.FakeDomainRepository
		configRepo          core_config.Repository
		requirementsFactory *testreq.FakeReqFactory
		deps                command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.RepoLocator = deps.RepoLocator.SetRouteRepository(routeRepo).SetDomainRepository(domainRepo)
		deps.Config = configRepo
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("routes").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{
			LoginSuccess:         true,
			TargetedSpaceSuccess: true,
		}
		routeRepo = new(testapi.FakeRouteRepository)
		domainRepo = new(testapi.FakeDomainRepository)
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCliCommand("routes", args, requirementsFactory, updateCommandDependency, false)
	}

	Describe("login requirements", func() {
		It("fails if the user is not logged in", func() {
			requirementsFactory.LoginSuccess = false
			Expect(runCommand()).To(BeFalse())
		})

		It("fails when an org and space is not targeted", func() {
			requirementsFactory.TargetedSpaceSuccess = false

			Expect(runCommand()).To(BeFalse())
		})

		Context("when arguments are provided", func() {
			var cmd command_registry.Command
			var flagContext flags.FlagContext

			BeforeEach(func() {
				cmd = &route.ListRoutes{}
				cmd.SetDependency(deps, false)
				flagContext = flags.NewFlagContext(cmd.MetaData().Flags)
			})

			It("should fail with usage", func() {
				flagContext.Parse("blahblah")

				reqs := cmd.Requirements(requirementsFactory, flagContext)

				err := testcmd.RunRequirements(reqs)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Incorrect Usage"))
				Expect(err.Error()).To(ContainSubstring("No argument required"))
			})
		})
	})

	Context("when there are routes", func() {
		BeforeEach(func() {
			cookieClickerGuid := "cookie-clicker-guid"

			domainRepo.ListDomainsForOrgStub = func(_ string, cb func(models.DomainFields) bool) error {
				tcpDomain := models.DomainFields{
					Guid: cookieClickerGuid,
					RouterGroupTypes: []string{
						"tcp",
						"bar",
					},
				}
				cb(tcpDomain)
				return nil
			}

			routeRepo.ListRoutesStub = func(cb func(models.Route) bool) error {
				app1 := models.ApplicationFields{Name: "dora"}
				app2 := models.ApplicationFields{Name: "bora"}

				route := models.Route{
					Space: models.SpaceFields{
						Name: "my-space",
					},
					Host:   "hostname-1",
					Domain: models.DomainFields{Name: "example.com"},
					Apps:   []models.ApplicationFields{app1},
					ServiceInstance: models.ServiceInstanceFields{
						Name: "test-service",
						Guid: "service-guid",
					},
				}

				route2 := models.Route{
					Space: models.SpaceFields{
						Name: "my-space",
					},
					Host:   "hostname-2",
					Path:   "/foo",
					Domain: models.DomainFields{Name: "cookieclicker.co"},
					Apps:   []models.ApplicationFields{app1, app2},
				}

				route3 := models.Route{
					Space: models.SpaceFields{
						Name: "my-space",
					},
					Domain: models.DomainFields{
						Guid: cookieClickerGuid,
						Name: "cookieclicker.co",
					},
					Apps: []models.ApplicationFields{app1, app2},
					Port: 9090,
				}

				cb(route)
				cb(route2)
				cb(route3)

				return nil
			}
		})

		It("lists routes", func() {
			runCommand()

			Expect(ui.Outputs).To(BeInDisplayOrder(
				[]string{"Getting routes for org my-org / space my-space as my-user ..."},
				[]string{"space", "host", "domain", "port", "path", "type", "apps", "service"},
			))

			Expect(ui.Outputs).To(ContainElement(MatchRegexp(`^my-space\s+hostname-1\s+example.com\s+dora\s+test-service\s*$`)))
			Expect(ui.Outputs).To(ContainElement(MatchRegexp(`^my-space\s+hostname-2\s+cookieclicker\.co\s+/foo\s+dora,bora\s*$`)))
			Expect(ui.Outputs).To(ContainElement(MatchRegexp(`^my-space\s+cookieclicker\.co\s+9090\s+tcp,bar\s+dora,bora\s*$`)))
		})
	})

	Context("when there are routes in different spaces", func() {
		BeforeEach(func() {
			routeRepo.ListAllRoutesStub = func(cb func(models.Route) bool) error {
				space1 := models.SpaceFields{Name: "space-1"}
				space2 := models.SpaceFields{Name: "space-2"}

				domain := models.DomainFields{Name: "example.com"}
				domain2 := models.DomainFields{Name: "cookieclicker.co"}

				app1 := models.ApplicationFields{Name: "dora"}
				app2 := models.ApplicationFields{Name: "bora"}

				route := models.Route{}
				route.Host = "hostname-1"
				route.Domain = domain
				route.Apps = []models.ApplicationFields{app1}
				route.Space = space1
				route.ServiceInstance = models.ServiceInstanceFields{
					Name: "test-service",
					Guid: "service-guid",
				}

				route2 := models.Route{}
				route2.Host = "hostname-2"
				route2.Path = "/foo"
				route2.Domain = domain2
				route2.Apps = []models.ApplicationFields{app1, app2}
				route2.Space = space2

				cb(route)
				cb(route2)

				return nil
			}
		})

		It("lists routes at orglevel", func() {
			runCommand("--orglevel")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Getting routes for org", "my-org", "my-user"},
				[]string{"space", "host", "domain", "apps", "service"},
				[]string{"space-1", "hostname-1", "example.com", "dora", "test-service"},
				[]string{"space-2", "hostname-2", "cookieclicker.co", "dora", "bora"},
			))
		})
	})

	Context("when there are not routes", func() {
		It("tells the user when no routes were found", func() {
			runCommand()

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Getting routes"},
				[]string{"No routes found"},
			))
		})
	})

	Context("when there is an error listing routes", func() {
		BeforeEach(func() {
			routeRepo.ListRoutesReturns(errors.New("an-error"))
		})

		It("returns an error to the user", func() {
			runCommand()

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Getting routes"},
				[]string{"FAILED"},
			))
		})
	})
})
