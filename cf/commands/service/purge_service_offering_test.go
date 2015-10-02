package service_test

import (
	"errors"

	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	cferrors "github.com/cloudfoundry/cli/cf/errors"
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
		serviceRepo         *testapi.FakeServiceRepo
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
		serviceRepo = &testapi.FakeServiceRepo{}
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
		serviceRepo.FindServiceOfferingByLabelAndProviderServiceOffering = offering

		ui.Inputs = []string{"yes"}

		runCommand([]string{"-p", "the-provider", "the-service-name"})

		Expect(serviceRepo.FindServiceOfferingByLabelAndProviderName).To(Equal("the-service-name"))
		Expect(serviceRepo.FindServiceOfferingByLabelAndProviderProvider).To(Equal("the-provider"))
		Expect(serviceRepo.PurgedServiceOffering).To(Equal(offering))
	})

	It("works when not given a provider", func() {
		offering := maker.NewServiceOffering("the-service-name")
		serviceRepo.FindServiceOfferingByLabelAndProviderServiceOffering = offering

		ui.Inputs = []string{"yes"}

		runCommand([]string{"the-service-name"})

		Expect(ui.Outputs).To(ContainSubstrings([]string{"WARNING"}))
		Expect(ui.Prompts).To(ContainSubstrings([]string{"Really purge service", "the-service-name"}))
		Expect(ui.Outputs).To(ContainSubstrings([]string{"Purging service the-service-name..."}))

		Expect(serviceRepo.FindServiceOfferingByLabelAndProviderName).To(Equal("the-service-name"))
		Expect(serviceRepo.FindServiceOfferingByLabelAndProviderProvider).To(Equal(""))
		Expect(serviceRepo.PurgedServiceOffering).To(Equal(offering))

		Expect(ui.Outputs).To(ContainSubstrings([]string{"OK"}))
	})

	It("exits when the user does not acknowledge the confirmation", func() {
		ui.Inputs = []string{"no"}

		runCommand([]string{"the-service-name"})

		Expect(serviceRepo.FindServiceOfferingByLabelAndProviderCalled).To(Equal(true))
		Expect(serviceRepo.PurgeServiceOfferingCalled).To(Equal(false))
	})

	It("does not prompt with confirmation when -f is passed", func() {
		offering := maker.NewServiceOffering("the-service-name")
		serviceRepo.FindServiceOfferingByLabelAndProviderServiceOffering = offering

		runCommand(
			[]string{"-f", "the-service-name"},
		)

		Expect(len(ui.Prompts)).To(Equal(0))
		Expect(serviceRepo.PurgeServiceOfferingCalled).To(Equal(true))
	})

	It("fails with an error message when the request fails", func() {
		serviceRepo.FindServiceOfferingByLabelAndProviderApiResponse = cferrors.NewWithError("oh no!", errors.New("!"))

		runCommand(
			[]string{"-f", "-p", "the-provider", "the-service-name"},
		)

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"FAILED"},
			[]string{"oh no!"},
		))

		Expect(serviceRepo.PurgeServiceOfferingCalled).To(Equal(false))
	})

	It("fails with an error message when the purging request fails", func() {
		serviceRepo.PurgeServiceOfferingApiResponse = cferrors.New("crumpets insufficiently buttered")

		runCommand(
			[]string{"-f", "-p", "the-provider", "the-service-name"},
		)

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"FAILED"},
			[]string{"crumpets insufficiently buttered"},
		))
	})

	It("indicates when a service doesn't exist", func() {
		serviceRepo.FindServiceOfferingByLabelAndProviderApiResponse = cferrors.NewModelNotFoundError("Service Offering", "")

		ui.Inputs = []string{"yes"}

		runCommand(
			[]string{"-p", "the-provider", "the-service-name"},
		)

		Expect(ui.Outputs).To(ContainSubstrings([]string{"Service offering", "does not exist"}))
		Expect(ui.Outputs).ToNot(ContainSubstrings([]string{"WARNING"}))
		Expect(ui.Outputs).ToNot(ContainSubstrings([]string{"Ok"}))

		Expect(serviceRepo.PurgeServiceOfferingCalled).To(Equal(false))
	})
})
