package service_test

import (
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	cferrors "github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	"github.com/cloudfoundry/cli/testhelpers/maker"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

var _ = Describe("purge-service-instance command", func() {
	var (
		requirementsFactory *testreq.FakeReqFactory
		config              core_config.Repository
		ui                  *testterm.FakeUI
		serviceRepo         *testapi.FakeServiceRepository
		deps                command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.RepoLocator = deps.RepoLocator.SetServiceRepository(serviceRepo)
		deps.Config = config
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("purge-service-instance").SetDependency(deps, pluginCall))
	}

	runCommand := func(args []string) bool {
		return testcmd.RunCliCommand("purge-service-instance", args, requirementsFactory, updateCommandDependency, false)
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		config = testconfig.NewRepositoryWithDefaults()
		serviceRepo = &testapi.FakeServiceRepository{}
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}
	})

	Describe("requirements", func() {
		It("fails when not logged in", func() {
			requirementsFactory.LoginSuccess = false

			passed := runCommand([]string{"-f", "whatever"})

			Expect(passed).To(BeFalse())
		})

		It("fails when called without exactly one arg", func() {
			requirementsFactory.LoginSuccess = true

			passed := runCommand([]string{})

			Expect(passed).To(BeFalse())
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires", "argument"},
			))
		})
	})

	It("exits when the user does not acknowledge the confirmation", func() {
		ui.Inputs = []string{"no"}

		runCommand([]string{"the-service-name"})

		Expect(serviceRepo.FindInstanceByNameCallCount()).To(Equal(1))
		Expect(serviceRepo.PurgeServiceInstanceCallCount()).To(BeZero())
	})

	It("does not prompt with confirmation when -f is passed", func() {
		instance := maker.NewServiceInstance("the-service-name")
		serviceRepo.FindInstanceByNameReturns(instance, nil)

		runCommand(
			[]string{"-f", "the-service-name"},
		)

		Expect(len(ui.Prompts)).To(Equal(0))
		Expect(serviceRepo.PurgeServiceInstanceCallCount()).To(Equal(1))
		Expect(serviceRepo.PurgeServiceInstanceArgsForCall(0)).To(Equal(instance))
	})

	It("fails with an error message when finding the instance fails", func() {
		serviceRepo.FindInstanceByNameReturns(models.ServiceInstance{}, cferrors.New("Error finding instance"))

		runCommand(
			[]string{"-f", "the-service-name"},
		)

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"FAILED"},
			[]string{"Error finding instance"},
		))
	})

	It("fails with an error message when the purging request fails", func() {
		serviceRepo.PurgeServiceInstanceReturns(cferrors.New("crumpets insufficiently buttered"))

		runCommand(
			[]string{"-f", "the-service-name"},
		)

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"FAILED"},
			[]string{"crumpets insufficiently buttered"},
		))
	})

	It("indicates when a service doesn't exist", func() {
		serviceRepo.FindInstanceByNameReturns(models.ServiceInstance{}, cferrors.NewModelNotFoundError("Service instance", "nonexistent-service"))

		ui.Inputs = []string{"yes"}

		runCommand(
			[]string{"nonexistent-service"},
		)

		Expect(ui.Outputs).To(ContainSubstrings([]string{"Service instance", "nonexistent-service", "not found"}))
		Expect(ui.Outputs).ToNot(ContainSubstrings([]string{"WARNING"}))
		Expect(ui.Outputs).ToNot(ContainSubstrings([]string{"Ok"}))

		Expect(serviceRepo.PurgeServiceInstanceCallCount()).To(BeZero())
	})
})
