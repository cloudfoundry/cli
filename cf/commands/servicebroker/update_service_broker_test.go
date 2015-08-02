package servicebroker_test

import (
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/command_registry"
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

var _ = Describe("update-service-broker command", func() {
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
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("update-service-broker").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		configRepo = testconfig.NewRepositoryWithDefaults()
		ui = &testterm.FakeUI{}
		requirementsFactory = &testreq.FakeReqFactory{}
		serviceBrokerRepo = &testapi.FakeServiceBrokerRepo{}
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCliCommand("update-service-broker", args, requirementsFactory, updateCommandDependency, false)
	}

	Describe("requirements", func() {
		It("fails with usage when invoked without exactly four args", func() {
			requirementsFactory.LoginSuccess = true

			runCommand("arg1", "arg2", "arg3")
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires", "arguments"},
			))
		})

		It("fails when not logged in", func() {
			Expect(runCommand("heeeeeeey", "yooouuuuuuu", "guuuuuuuuys", "ヾ(＠*ー⌒ー*@)ノ")).To(BeFalse())
		})
	})

	Context("when logged in", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
			broker := models.ServiceBroker{}
			broker.Name = "my-found-broker"
			broker.Guid = "my-found-broker-guid"
			serviceBrokerRepo.FindByNameServiceBroker = broker
		})

		It("updates the service broker with the provided properties", func() {
			runCommand("my-broker", "new-username", "new-password", "new-url")

			Expect(serviceBrokerRepo.FindByNameName).To(Equal("my-broker"))

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Updating service broker", "my-found-broker", "my-user"},
				[]string{"OK"},
			))

			expectedServiceBroker := models.ServiceBroker{}
			expectedServiceBroker.Name = "my-found-broker"
			expectedServiceBroker.Username = "new-username"
			expectedServiceBroker.Password = "new-password"
			expectedServiceBroker.Url = "new-url"
			expectedServiceBroker.Guid = "my-found-broker-guid"

			Expect(serviceBrokerRepo.UpdatedServiceBroker).To(Equal(expectedServiceBroker))
		})
	})
})
