package service_test

import (
	"errors"
	"fmt"

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

var _ = Describe("purge-service-offering command", func() {
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
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("purge-service-offering").SetDependency(deps, pluginCall))
	}

	runCommand := func(args []string) bool {
		return testcmd.RunCliCommand("purge-service-offering", args, requirementsFactory, updateCommandDependency, false)
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

	It("works when given -p and a provider name", func() {
		offering := maker.NewServiceOffering("the-service-name")
		serviceRepo.FindServiceOfferingByLabelAndProviderReturns(offering, nil)

		ui.Inputs = []string{"yes"}

		runCommand([]string{"-p", "the-provider", "the-service-name"})

		name, provider := serviceRepo.FindServiceOfferingByLabelAndProviderArgsForCall(0)
		Expect(name).To(Equal("the-service-name"))
		Expect(provider).To(Equal("the-provider"))

		Expect(serviceRepo.PurgeServiceOfferingArgsForCall(0)).To(Equal(offering))
	})

	It("works when not given a provider", func() {
		offering := maker.NewServiceOffering("the-service-name")
		serviceRepo.FindServiceOfferingByLabelAndProviderReturns(offering, nil)

		ui.Inputs = []string{"yes"}

		runCommand([]string{"the-service-name"})

		Expect(ui.Outputs).To(ContainSubstrings([]string{"WARNING"}))
		Expect(ui.Prompts).To(ContainSubstrings([]string{"Really purge service", "the-service-name"}))
		Expect(ui.Outputs).To(ContainSubstrings([]string{"Purging service the-service-name..."}))

		name, provider := serviceRepo.FindServiceOfferingByLabelAndProviderArgsForCall(0)
		Expect(name).To(Equal("the-service-name"))
		Expect(provider).To(BeEmpty())
		Expect(serviceRepo.PurgeServiceOfferingArgsForCall(0)).To(Equal(offering))

		Expect(ui.Outputs).To(ContainSubstrings([]string{"OK"}))
	})

	It("exits when the user does not acknowledge the confirmation", func() {
		ui.Inputs = []string{"no"}

		runCommand([]string{"the-service-name"})

		Expect(serviceRepo.FindServiceOfferingByLabelAndProviderCallCount()).To(Equal(1))
		Expect(serviceRepo.PurgeServiceOfferingCallCount()).To(BeZero())
	})

	It("does not prompt with confirmation when -f is passed", func() {
		offering := maker.NewServiceOffering("the-service-name")
		serviceRepo.FindServiceOfferingByLabelAndProviderReturns(offering, nil)

		runCommand(
			[]string{"-f", "the-service-name"},
		)

		Expect(len(ui.Prompts)).To(Equal(0))
		Expect(serviceRepo.PurgeServiceOfferingCallCount()).To(Equal(1))
	})

	It("fails with an error message when the request fails", func() {
		serviceRepo.FindServiceOfferingByLabelAndProviderReturns(models.ServiceOffering{}, errors.New("oh no!"))

		runCommand(
			[]string{"-f", "-p", "the-provider", "the-service-name"},
		)

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"FAILED"},
			[]string{"oh no!"},
		))

		Expect(serviceRepo.PurgeServiceOfferingCallCount()).To(BeZero())
	})

	It("fails with an error message when the purging request fails", func() {
		serviceRepo.PurgeServiceOfferingReturns(fmt.Errorf("crumpets insufficiently buttered"))

		runCommand(
			[]string{"-f", "-p", "the-provider", "the-service-name"},
		)

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"FAILED"},
			[]string{"crumpets insufficiently buttered"},
		))
	})

	It("indicates when a service doesn't exist", func() {
		serviceRepo.FindServiceOfferingByLabelAndProviderReturns(models.ServiceOffering{}, cferrors.NewModelNotFoundError("Service Offering", ""))

		ui.Inputs = []string{"yes"}

		runCommand(
			[]string{"-p", "the-provider", "the-service-name"},
		)

		Expect(ui.Outputs).To(ContainSubstrings([]string{"Service offering", "does not exist"}))
		Expect(ui.Outputs).ToNot(ContainSubstrings([]string{"WARNING"}))
		Expect(ui.Outputs).ToNot(ContainSubstrings([]string{"Ok"}))

		Expect(serviceRepo.PurgeServiceOfferingCallCount()).To(BeZero())
	})
})
