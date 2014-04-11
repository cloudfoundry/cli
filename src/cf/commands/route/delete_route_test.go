package route_test

import (
	. "cf/commands/route"
	"cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

var _ = Describe("delete-route command", func() {
	var (
		ui                  *testterm.FakeUI
		requirementsFactory *testreq.FakeReqFactory
		routeRepo           *testapi.FakeRouteRepository
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{Inputs: []string{"yes"}}

		routeRepo = &testapi.FakeRouteRepository{}
		requirementsFactory = &testreq.FakeReqFactory{
			LoginSuccess: true,
		}
	})

	runCommand := func(args ...string) {
		configRepo := testconfig.NewRepositoryWithDefaults()
		cmd := NewDeleteRoute(ui, configRepo, routeRepo)
		testcmd.RunCommand(cmd, testcmd.NewContext("delete-route", args), requirementsFactory)
	}

	Context("when not logged in", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = false
		})

		It("does not pass requirements", func() {
			runCommand("-n", "my-host", "example.com")
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})
	})

	Context("when logged in successfully", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
			route := models.Route{Guid: "route-guid"}
			route.Domain = models.DomainFields{
				Guid: "domain-guid",
				Name: "example.com",
			}
			routeRepo.FindByHostAndDomainRoute = route
		})

		It("fails with usage when given zero args", func() {
			runCommand()
			Expect(ui.FailedWithUsage).To(BeTrue())
		})

		It("does not fail with usage when provided with a domain", func() {
			runCommand("example.com")
			Expect(ui.FailedWithUsage).To(BeFalse())
		})

		It("does not fail with usage when provided a hostname", func() {
			runCommand("-n", "my-host", "example.com")
			Expect(ui.FailedWithUsage).To(BeFalse())
		})

		It("deletes routes when the user confirms", func() {
			runCommand("-n", "my-host", "example.com")

			testassert.SliceContains(ui.Prompts, testassert.Lines{
				{"Really delete the route my-host"},
			})

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"Deleting route", "my-host.example.com"},
				{"OK"},
			})
			Expect(routeRepo.DeletedRouteGuids).To(Equal([]string{"route-guid"}))
		})

		It("does not prompt the user to confirm when they pass the '-f' flag", func() {
			ui.Inputs = []string{}
			runCommand("-f", "-n", "my-host", "example.com")

			Expect(ui.Prompts).To(BeEmpty())

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"Deleting", "my-host.example.com"},
				{"OK"},
			})
			Expect(routeRepo.DeletedRouteGuids).To(Equal([]string{"route-guid"}))
		})

		It("succeeds with a warning when the route does not exist", func() {
			routeRepo.FindByHostAndDomainNotFound = true

			runCommand("-n", "my-host", "example.com")

			testassert.SliceContains(ui.WarnOutputs, testassert.Lines{
				{"my-host", "does not exist"},
			})

			testassert.SliceDoesNotContain(ui.Outputs, testassert.Lines{
				{"OK"},
			})
		})
	})
})
