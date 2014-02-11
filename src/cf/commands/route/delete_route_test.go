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

func init() {
	Describe("delete-route command", func() {
		var routeRepo *testapi.FakeRouteRepository
		var reqFactory *testreq.FakeReqFactory

		BeforeEach(func() {
			routeRepo = &testapi.FakeRouteRepository{}
			reqFactory = &testreq.FakeReqFactory{}
		})

		It("fails requirements when not logged in", func() {
			callDeleteRoute("y", []string{"-n", "my-host", "example.com"}, reqFactory, routeRepo)
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})

		Context("when logged in successfully", func() {
			var ui *testterm.FakeUI

			BeforeEach(func() {
				reqFactory.LoginSuccess = true
			})

			It("passes requirements when logged in", func() {
				callDeleteRoute("y", []string{"-n", "my-host", "example.com"}, reqFactory, routeRepo)
				Expect(testcmd.CommandDidPassRequirements).To(BeTrue())
			})

			It("fails with usage when given zero args", func() {
				ui = callDeleteRoute("y", []string{}, reqFactory, routeRepo)
				Expect(ui.FailedWithUsage).To(BeTrue())
			})

			It("does not fail with usage when provided with a domain", func() {
				ui = callDeleteRoute("y", []string{"example.com"}, reqFactory, routeRepo)
				Expect(ui.FailedWithUsage).To(BeFalse())
			})

			It("does not fail with usage when provided a hostname", func() {
				ui = callDeleteRoute("y", []string{"-n", "my-host", "example.com"}, reqFactory, routeRepo)
				Expect(ui.FailedWithUsage).To(BeFalse())
			})

			It("TestDeleteRouteWithConfirmation", func() {
				domain := models.DomainFields{}
				domain.Guid = "domain-guid"
				domain.Name = "example.com"

				route := models.Route{}
				route.Guid = "route-guid"
				route.Host = "my-host"
				route.Domain = domain
				routeRepo.FindByHostAndDomainRoute = route

				ui := callDeleteRoute("y", []string{"-n", "my-host", "example.com"}, reqFactory, routeRepo)

				testassert.SliceContains(GinkgoT(), ui.Prompts, testassert.Lines{
					{"Really delete", "my-host"},
				})

				testassert.SliceContains(GinkgoT(), ui.Outputs, testassert.Lines{
					{"Deleting route", "my-host.example.com"},
					{"OK"},
				})
				Expect(routeRepo.DeleteRouteGuid).To(Equal("route-guid"))
			})

			It("TestDeleteRouteWithForce", func() {
				domain := models.DomainFields{}
				domain.Guid = "domain-guid"
				domain.Name = "example.com"

				route := models.Route{}
				route.Guid = "route-guid"
				route.Host = "my-host"
				route.Domain = domain
				routeRepo.FindByHostAndDomainRoute = route

				ui := callDeleteRoute("", []string{"-f", "-n", "my-host", "example.com"}, reqFactory, routeRepo)

				Expect(len(ui.Prompts)).To(Equal(0))

				testassert.SliceContains(GinkgoT(), ui.Outputs, testassert.Lines{
					{"Deleting", "my-host.example.com"},
					{"OK"},
				})
				Expect(routeRepo.DeleteRouteGuid).To(Equal("route-guid"))
			})

			It("TestDeleteRouteWhenRouteDoesNotExist", func() {
				routeRepo.FindByHostAndDomainNotFound = true

				ui := callDeleteRoute("y", []string{"-n", "my-host", "example.com"}, reqFactory, routeRepo)

				testassert.SliceContains(GinkgoT(), ui.Outputs, testassert.Lines{
					{"Deleting", "my-host.example.com"},
					{"OK"},
					{"my-host", "does not exist"},
				})
			})
		})
	})
}

func callDeleteRoute(confirmation string, args []string, reqFactory *testreq.FakeReqFactory, routeRepo *testapi.FakeRouteRepository) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{Inputs: []string{confirmation}}
	ctxt := testcmd.NewContext("delete-route", args)
	configRepo := testconfig.NewRepositoryWithDefaults()

	cmd := NewDeleteRoute(ui, configRepo, routeRepo)

	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
