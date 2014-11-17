package route_test

import (
	"errors"

	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/commands/route"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("check-route command", func() {
	var (
		ui                  *testterm.FakeUI
		routeRepo           *testapi.FakeRouteRepository
		domainRepo          *testapi.FakeDomainRepository
		requirementsFactory *testreq.FakeReqFactory
		config              core_config.ReadWriter
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		routeRepo = &testapi.FakeRouteRepository{}
		domainRepo = &testapi.FakeDomainRepository{}
		requirementsFactory = &testreq.FakeReqFactory{}
		config = testconfig.NewRepositoryWithDefaults()
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCommand(NewCheckRoute(ui, config, routeRepo, domainRepo), args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("fails when not logged in", func() {
			requirementsFactory.TargetedOrgSuccess = true

			Expect(runCommand("foobar.example.com", "bar.example.com")).To(BeFalse())
		})

		It("fails when no org is targeted", func() {
			requirementsFactory.LoginSuccess = true
			Expect(runCommand("foobar.example.com", "bar.example.com")).To(BeFalse())
		})

		It("fails when the number of arguments is greater than two", func() {
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedOrgSuccess = true
			passed := runCommand("foobar.example.com", "hello", "world")

			Expect(passed).To(BeFalse())
			Expect(ui.FailedWithUsage).To(BeTrue())
		})

		It("fails when the number of arguments is less than two", func() {
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedOrgSuccess = true
			passed := runCommand("foobar.example.com")

			Expect(passed).To(BeFalse())
			Expect(ui.FailedWithUsage).To(BeTrue())
		})
	})

	Context("when the route already exists", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedOrgSuccess = true
			routeRepo.CheckIfExistsFound = true
		})

		It("prints out route does exist", func() {
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedOrgSuccess = true

			runCommand("some-existing-route", "example.com")
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Checking for route..."},
				[]string{"OK"},
				[]string{"Route some-existing-route.example.com does exist"},
			))
		})
	})

	Context("when the route does not exist", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedOrgSuccess = true
			routeRepo.CheckIfExistsFound = false
		})

		It("prints out route does not exist", func() {

			runCommand("non-existent-route", "example.com")
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Checking for route..."},
				[]string{"OK"},
				[]string{"Route non-existent-route.example.com does not exist"},
			))
		})
	})

	Context("when finding the domain returns an error", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedOrgSuccess = true
			domainRepo.FindByNameInOrgApiResponse = errors.New("Domain not found")
		})

		It("prints out route does not exist", func() {

			runCommand("some-silly-route", "some-non-real-domain")
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Checking for route..."},
				[]string{"FAILED"},
				[]string{"Domain not found"},
			))
		})
	})
	Context("when checking if the route exists returns an error", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedOrgSuccess = true
			routeRepo.CheckIfExistsError = errors.New("Some stupid error")
		})

		It("prints out route does not exist", func() {

			runCommand("some-silly-route", "some-non-real-domain")
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Checking for route..."},
				[]string{"FAILED"},
				[]string{"Some stupid error"},
			))
		})
	})

})
