package service_test

import (
	. "cf/commands/service"
	"cf/configuration"
	cferrors "cf/errors"
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	"testhelpers/maker"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"

	. "testhelpers/matchers"
)

var _ = Describe("purge-service command", func() {
	Describe("requirements", func() {
		It("fails when not logged in", func() {
			deps := setupDependencies()
			deps.requirementsFactory.LoginSuccess = false

			cmd := NewPurgeServiceOffering(deps.ui, deps.config, deps.serviceRepo)
			testcmd.RunCommand(
				cmd,
				testcmd.NewContext("purge-service-offering", []string{"-f", "whatever"}),
				deps.requirementsFactory,
			)

			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})

		It("fails when called without exactly one arg", func() {
			deps := setupDependencies()
			deps.requirementsFactory.LoginSuccess = true

			testcmd.RunCommand(
				NewPurgeServiceOffering(deps.ui, deps.config, deps.serviceRepo),
				testcmd.NewContext("purge-service-offering", []string{}),
				deps.requirementsFactory,
			)

			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
			Expect(deps.ui.FailedWithUsage).To(BeTrue())
			Expect(deps.ui.FailedWithUsageCommandName).To(Equal("purge-service-offering"))
		})
	})

	It("works when given -p and a provider name", func() {
		deps := setupDependencies()

		offering := maker.NewServiceOffering("the-service-name")
		deps.serviceRepo.FindServiceOfferingByLabelAndProviderServiceOffering = offering

		deps.ui.Inputs = []string{"yes"}

		testcmd.RunCommand(
			NewPurgeServiceOffering(deps.ui, deps.config, deps.serviceRepo),
			testcmd.NewContext("purge-service-offering", []string{"-p", "the-provider", "the-service-name"}),
			deps.requirementsFactory,
		)

		Expect(deps.serviceRepo.FindServiceOfferingByLabelAndProviderName).To(Equal("the-service-name"))
		Expect(deps.serviceRepo.FindServiceOfferingByLabelAndProviderProvider).To(Equal("the-provider"))
		Expect(deps.serviceRepo.PurgedServiceOffering).To(Equal(offering))
	})

	It("works when not given a provider", func() {
		deps := setupDependencies()

		offering := maker.NewServiceOffering("the-service-name")
		deps.serviceRepo.FindServiceOfferingByLabelAndProviderServiceOffering = offering

		deps.ui.Inputs = []string{"yes"}

		testcmd.RunCommand(
			NewPurgeServiceOffering(deps.ui, deps.config, deps.serviceRepo),
			testcmd.NewContext("purge-service-offering", []string{"the-service-name"}),
			deps.requirementsFactory,
		)

		Expect(deps.ui.Outputs).To(ContainSubstrings([]string{"Warning"}))
		Expect(deps.ui.Prompts).To(ContainSubstrings([]string{"Really purge service", "the-service-name"}))
		Expect(deps.ui.Outputs).To(ContainSubstrings([]string{"Purging service the-service-name..."}))

		Expect(deps.serviceRepo.FindServiceOfferingByLabelAndProviderName).To(Equal("the-service-name"))
		Expect(deps.serviceRepo.FindServiceOfferingByLabelAndProviderProvider).To(Equal(""))
		Expect(deps.serviceRepo.PurgedServiceOffering).To(Equal(offering))

		Expect(deps.ui.Outputs).To(ContainSubstrings([]string{"OK"}))
	})

	It("exits when the user does not acknowledge the confirmation", func() {
		deps := setupDependencies()

		deps.ui.Inputs = []string{"no"}

		testcmd.RunCommand(
			NewPurgeServiceOffering(deps.ui, deps.config, deps.serviceRepo),
			testcmd.NewContext("purge-service-offering", []string{"the-service-name"}),
			deps.requirementsFactory,
		)

		Expect(deps.serviceRepo.FindServiceOfferingByLabelAndProviderCalled).To(Equal(true))
		Expect(deps.serviceRepo.PurgeServiceOfferingCalled).To(Equal(false))
	})

	It("does not prompt with confirmation when -f is passed", func() {
		deps := setupDependencies()

		offering := maker.NewServiceOffering("the-service-name")
		deps.serviceRepo.FindServiceOfferingByLabelAndProviderServiceOffering = offering

		testcmd.RunCommand(
			NewPurgeServiceOffering(deps.ui, deps.config, deps.serviceRepo),
			testcmd.NewContext("purge-service-offering", []string{"-f", "the-service-name"}),
			deps.requirementsFactory,
		)

		Expect(len(deps.ui.Prompts)).To(Equal(0))
		Expect(deps.serviceRepo.PurgeServiceOfferingCalled).To(Equal(true))
	})

	It("fails with an error message when the request fails", func() {
		deps := setupDependencies()

		deps.serviceRepo.FindServiceOfferingByLabelAndProviderApiResponse = cferrors.NewWithError("oh no!", errors.New("!"))

		testcmd.RunCommand(
			NewPurgeServiceOffering(deps.ui, deps.config, deps.serviceRepo),
			testcmd.NewContext("purge-service-offering", []string{"-f", "-p", "the-provider", "the-service-name"}),
			deps.requirementsFactory,
		)

		Expect(deps.ui.Outputs).To(ContainSubstrings(
			[]string{"FAILED"},
			[]string{"oh no!"},
		))

		Expect(deps.serviceRepo.PurgeServiceOfferingCalled).To(Equal(false))
	})

	It("indicates when a service doesn't exist", func() {
		deps := setupDependencies()

		deps.serviceRepo.FindServiceOfferingByLabelAndProviderApiResponse = cferrors.NewModelNotFoundError("Service Offering", "")

		deps.ui.Inputs = []string{"yes"}

		testcmd.RunCommand(
			NewPurgeServiceOffering(deps.ui, deps.config, deps.serviceRepo),
			testcmd.NewContext("purge-service-offering", []string{"-p", "the-provider", "the-service-name"}),
			deps.requirementsFactory,
		)

		Expect(deps.ui.Outputs).To(ContainSubstrings([]string{"Service offering", "does not exist"}))
		Expect(deps.ui.Outputs).ToNot(ContainSubstrings([]string{"Warning"}))
		Expect(deps.ui.Outputs).ToNot(ContainSubstrings([]string{"Ok"}))

		Expect(deps.serviceRepo.PurgeServiceOfferingCalled).To(Equal(false))
	})
})

type commandDependencies struct {
	ui                  *testterm.FakeUI
	config              configuration.ReadWriter
	serviceRepo         *testapi.FakeServiceRepo
	requirementsFactory *testreq.FakeReqFactory
}

func setupDependencies() (obj commandDependencies) {
	obj.ui = &testterm.FakeUI{}

	obj.config = testconfig.NewRepository()
	obj.requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}
	obj.serviceRepo = new(testapi.FakeServiceRepo)
	return
}
