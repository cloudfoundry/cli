package servicebroker_test

import (
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	"github.com/cloudfoundry/cli/cf/command_registry"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("create-service-broker command", func() {
	var (
		ui                  *testterm.FakeUI
		requirementsFactory *testreq.FakeReqFactory
		configRepo          core_config.Repository
		serviceBrokerRepo   *testapi.FakeServiceBrokerRepo
		deps                command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.RepoLocator = deps.RepoLocator.SetServiceBrokerRepository(serviceBrokerRepo)
		deps.Config = configRepo
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("create-service-broker").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		configRepo = testconfig.NewRepositoryWithDefaults()

		ui = &testterm.FakeUI{}
		requirementsFactory = &testreq.FakeReqFactory{}
		serviceBrokerRepo = &testapi.FakeServiceBrokerRepo{}
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCliCommand("create-service-broker", args, requirementsFactory, updateCommandDependency, false)
	}

	Describe("requirements", func() {
		It("fails with usage when called without exactly four args", func() {
			requirementsFactory.LoginSuccess = true
			runCommand("whoops", "not-enough", "args")
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires", "arguments"},
			))
		})

		It("fails when not logged in", func() {
			Expect(runCommand("Just", "Enough", "Args", "Provided")).To(BeFalse())
		})
	})

	Context("when logged in", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
		})

		It("creates a service broker, obviously", func() {
			runCommand("my-broker", "my-username", "my-password", "http://example.com")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Creating service broker", "my-broker", "my-user"},
				[]string{"OK"},
			))

			Expect(serviceBrokerRepo.CreateName).To(Equal("my-broker"))
			Expect(serviceBrokerRepo.CreateUrl).To(Equal("http://example.com"))
			Expect(serviceBrokerRepo.CreateUsername).To(Equal("my-username"))
			Expect(serviceBrokerRepo.CreatePassword).To(Equal("my-password"))
		})
	})
})
