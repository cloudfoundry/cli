package servicebroker_test

import (
	"github.com/cloudfoundry/cli/cf/api/apifakes"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/requirements/requirementsfakes"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("rename-service-broker command", func() {
	var (
		ui                  *testterm.FakeUI
		requirementsFactory *requirementsfakes.FakeFactory
		configRepo          coreconfig.Repository
		serviceBrokerRepo   *apifakes.FakeServiceBrokerRepository
		deps                commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.RepoLocator = deps.RepoLocator.SetServiceBrokerRepository(serviceBrokerRepo)
		deps.Config = configRepo
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("rename-service-broker").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		configRepo = testconfig.NewRepositoryWithDefaults()

		ui = &testterm.FakeUI{}
		requirementsFactory = new(requirementsfakes.FakeFactory)
		serviceBrokerRepo = new(apifakes.FakeServiceBrokerRepository)
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCLICommand("rename-service-broker", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	Describe("requirements", func() {
		It("fails with usage when not invoked with exactly two args", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
			runCommand("welp")
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires", "arguments"},
			))
		})

		It("fails when not logged in", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Failing{Message: "not logged in"})
			Expect(runCommand("okay", "DO---IIIIT")).To(BeFalse())
		})
	})

	Context("when logged in", func() {
		BeforeEach(func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
			broker := models.ServiceBroker{}
			broker.Name = "my-found-broker"
			broker.GUID = "my-found-broker-guid"
			serviceBrokerRepo.FindByNameReturns(broker, nil)
		})

		It("renames the given service broker", func() {
			runCommand("my-broker", "my-new-broker")
			Expect(serviceBrokerRepo.FindByNameCallCount()).To(Equal(1))
			Expect(serviceBrokerRepo.FindByNameArgsForCall(0)).To(Equal("my-broker"))

			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Renaming service broker", "my-found-broker", "my-new-broker", "my-user"},
				[]string{"OK"},
			))

			Expect(serviceBrokerRepo.RenameCallCount()).To(Equal(1))
			guid, name := serviceBrokerRepo.RenameArgsForCall(0)
			Expect(guid).To(Equal("my-found-broker-guid"))
			Expect(name).To(Equal("my-new-broker"))
		})
	})
})
