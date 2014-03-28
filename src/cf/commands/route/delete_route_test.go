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
		routeRepo  *testapi.FakeRouteRepository
		reqFactory *testreq.FakeReqFactory
		ui         *testterm.FakeUI
		cmd        *DeleteRoute
	)

	BeforeEach(func() {
		configRepo := testconfig.NewRepositoryWithDefaults()
		ui = &testterm.FakeUI{}
		routeRepo = &testapi.FakeRouteRepository{}
		reqFactory = &testreq.FakeReqFactory{}
		cmd = NewDeleteRoute(ui, configRepo, routeRepo)
	})

	var callDeleteRoute = func(confirmation string, args []string) {
		ui.Inputs = []string{confirmation}
		testcmd.RunCommand(cmd, testcmd.NewContext("delete-route", args), reqFactory)
	}

	It("fails requirements when not logged in", func() {
		callDeleteRoute("y", []string{"-n", "my-host", "example.com"})
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})

	Context("when logged in successfully", func() {
		BeforeEach(func() {
			reqFactory.LoginSuccess = true
			route := models.Route{}
			route.Guid = "route-guid"
			route.Host = "my-host"
			route.Domain = models.DomainFields{
				Guid: "domain-guid",
				Name: "example.com",
			}
			routeRepo.FindByHostAndDomainRoute = route

		})

		It("passes requirements when logged in", func() {
			callDeleteRoute("y", []string{"-n", "my-host", "example.com"})
			Expect(testcmd.CommandDidPassRequirements).To(BeTrue())
		})

		It("fails with usage when given zero args", func() {
			callDeleteRoute("y", []string{})
			Expect(ui.FailedWithUsage).To(BeTrue())
		})

		It("does not fail with usage when provided with a domain", func() {
			callDeleteRoute("y", []string{"example.com"})
			Expect(ui.FailedWithUsage).To(BeFalse())
		})

		It("does not fail with usage when provided a hostname", func() {
			callDeleteRoute("y", []string{"-n", "my-host", "example.com"})
			Expect(ui.FailedWithUsage).To(BeFalse())
		})

		It("deletes routes when the user confirms", func() {
			callDeleteRoute("y", []string{"-n", "my-host", "example.com"})

			testassert.SliceContains(ui.Prompts, testassert.Lines{
				{"Really delete", "my-host"},
			})

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"Deleting route", "my-host.example.com"},
				{"OK"},
			})
			Expect(routeRepo.DeleteRouteGuid).To(Equal("route-guid"))
		})

		It("does not prompt the user to confirm when they pass the '-f' flag", func() {
			callDeleteRoute("", []string{"-f", "-n", "my-host", "example.com"})

			Expect(len(ui.Prompts)).To(Equal(0))

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"Deleting", "my-host.example.com"},
				{"OK"},
			})
			Expect(routeRepo.DeleteRouteGuid).To(Equal("route-guid"))
		})

		It("succeeds with a warning when the route does not exist", func() {
			routeRepo.FindByHostAndDomainNotFound = true

			callDeleteRoute("y", []string{"-n", "my-host", "example.com"})

			testassert.SliceContains(ui.WarnOutputs, testassert.Lines{
				{"my-host", "does not exist"},
			})

			testassert.SliceDoesNotContain(ui.Outputs, testassert.Lines{
				{"OK"},
			})

		})
	})
})
