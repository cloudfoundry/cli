package route_test

import (
	. "github.com/cloudfoundry/cli/cf/commands/route"
	"github.com/cloudfoundry/cli/cf/errors"
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
		testcmd.RunCommand(cmd, args, requirementsFactory)
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
			routeRepo.FindByHostAndDomainReturns.Route = route
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

			Expect(ui.Prompts).To(ContainSubstrings([]string{"Really delete the route my-host"}))

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Deleting route", "my-host.example.com"},
				[]string{"OK"},
			))
			Expect(routeRepo.DeletedRouteGuids).To(Equal([]string{"route-guid"}))
		})

		It("does not prompt the user to confirm when they pass the '-f' flag", func() {
			ui.Inputs = []string{}
			runCommand("-f", "-n", "my-host", "example.com")

			Expect(ui.Prompts).To(BeEmpty())

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Deleting", "my-host.example.com"},
				[]string{"OK"},
			))
			Expect(routeRepo.DeletedRouteGuids).To(Equal([]string{"route-guid"}))
		})

		It("succeeds with a warning when the route does not exist", func() {
			routeRepo.FindByHostAndDomainReturns.Error = errors.NewModelNotFoundError("Org", "not found")

			runCommand("-n", "my-host", "example.com")

			Expect(ui.WarnOutputs).To(ContainSubstrings([]string{"my-host", "does not exist"}))

			Expect(ui.Outputs).ToNot(ContainSubstrings([]string{"OK"}))
		})
	})
})
