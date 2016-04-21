package servicebroker_test

import (
	"github.com/cloudfoundry/cli/cf/api/apifakes"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
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
		configRepo          coreconfig.Repository
		serviceBrokerRepo   *apifakes.FakeServiceBrokerRepository
		deps                commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.RepoLocator = deps.RepoLocator.SetServiceBrokerRepository(serviceBrokerRepo)
		deps.Config = configRepo
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("update-service-broker").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		configRepo = testconfig.NewRepositoryWithDefaults()
		ui = &testterm.FakeUI{}
		requirementsFactory = &testreq.FakeReqFactory{}
		serviceBrokerRepo = new(apifakes.FakeServiceBrokerRepository)
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCLICommand("update-service-broker", args, requirementsFactory, updateCommandDependency, false)
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
			broker.GUID = "my-found-broker-guid"
			serviceBrokerRepo.FindByNameReturns(broker, nil)
		})

		It("updates the service broker with the provided properties", func() {
			runCommand("my-broker", "new-username", "new-password", "new-url")

			Expect(serviceBrokerRepo.FindByNameArgsForCall(0)).To(Equal("my-broker"))

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Updating service broker", "my-found-broker", "my-user"},
				[]string{"OK"},
			))

			expectedServiceBroker := models.ServiceBroker{}
			expectedServiceBroker.Name = "my-found-broker"
			expectedServiceBroker.Username = "new-username"
			expectedServiceBroker.Password = "new-password"
			expectedServiceBroker.URL = "new-url"
			expectedServiceBroker.GUID = "my-found-broker-guid"

			Expect(serviceBrokerRepo.UpdateArgsForCall(0)).To(Equal(expectedServiceBroker))
		})
	})
})
