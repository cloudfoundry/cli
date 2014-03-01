package service_test

import (
	. "cf/commands/service"
	"cf/configuration"
	cferrors "cf/errors"
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	"testhelpers/maker"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

var _ = Describe("Purging services", func() {
	It("does not run if the requirements are not met", func() {
		deps := setupDependencies()

		testcmd.RunCommand(
			NewPurgeServiceOffering(deps.ui, deps.config, deps.serviceRepo),
			testcmd.NewContext("purge-service-offering", []string{}),
			deps.reqFactory,
		)

		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		Expect(deps.ui.FailedWithUsage).To(BeTrue())
		Expect(deps.ui.FailedWithUsageCommandName).To(Equal("purge-service-offering"))
	})

	It("works when given -p and a provider name", func() {
		deps := setupDependencies()

		offering := maker.NewServiceOffering("the-service-name")
		deps.serviceRepo.FindServiceOfferingByLabelAndProviderServiceOffering = offering

		deps.ui.Inputs = []string{"yes"}

		testcmd.RunCommand(
			NewPurgeServiceOffering(deps.ui, deps.config, deps.serviceRepo),
			testcmd.NewContext("purge-service-offering", []string{"-p", "the-provider", "the-service-name"}),
			deps.reqFactory,
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
			deps.reqFactory,
		)

		testassert.SliceContains(deps.ui.Outputs, testassert.Lines{
			{"Warning"},
		})
		testassert.SliceContains(deps.ui.Prompts, testassert.Lines{
			{"Really purge service", "the-service-name"},
		})
		testassert.SliceContains(deps.ui.Outputs, testassert.Lines{
			{"Purging service the-service-name..."},
		})

		Expect(deps.serviceRepo.FindServiceOfferingByLabelAndProviderName).To(Equal("the-service-name"))
		Expect(deps.serviceRepo.FindServiceOfferingByLabelAndProviderProvider).To(Equal(""))
		Expect(deps.serviceRepo.PurgedServiceOffering).To(Equal(offering))

		testassert.SliceContains(deps.ui.Outputs, testassert.Lines{
			{"OK"},
		})
	})

	It("exits when the user does not acknowledge the confirmation", func() {
		deps := setupDependencies()

		deps.ui.Inputs = []string{"no"}

		testcmd.RunCommand(
			NewPurgeServiceOffering(deps.ui, deps.config, deps.serviceRepo),
			testcmd.NewContext("purge-service-offering", []string{"the-service-name"}),
			deps.reqFactory,
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
			deps.reqFactory,
		)

		Expect(len(deps.ui.Prompts)).To(Equal(0))
		Expect(deps.serviceRepo.PurgeServiceOfferingCalled).To(Equal(true))
	})

	It("fails with an error message when the request fails", func() {
		deps := setupDependencies()

		deps.serviceRepo.FindServiceOfferingByLabelAndProviderApiResponse = cferrors.NewErrorWithError("oh no!", errors.New("!"))

		testcmd.RunCommand(
			NewPurgeServiceOffering(deps.ui, deps.config, deps.serviceRepo),
			testcmd.NewContext("purge-service-offering", []string{"-f", "-p", "the-provider", "the-service-name"}),
			deps.reqFactory,
		)

		testassert.SliceContains(deps.ui.Outputs, testassert.Lines{
			{"FAILED"},
			{"oh no!"},
		})

		Expect(deps.serviceRepo.PurgeServiceOfferingCalled).To(Equal(false))
	})

	It("indicates when a service doesn't exist", func() {
		deps := setupDependencies()

		deps.serviceRepo.FindServiceOfferingByLabelAndProviderApiResponse = cferrors.NewNotFoundError("uh oh cant find it")

		deps.ui.Inputs = []string{"yes"}

		deps.ui.Inputs = []string{"yes"}

		testcmd.RunCommand(
			NewPurgeServiceOffering(deps.ui, deps.config, deps.serviceRepo),
			testcmd.NewContext("purge-service-offering", []string{"-p", "the-provider", "the-service-name"}),
			deps.reqFactory,
		)

		testassert.SliceContains(deps.ui.Outputs, testassert.Lines{{"Service offering", "does not exist"}})
		testassert.SliceDoesNotContain(deps.ui.Outputs, testassert.Lines{{"Warning"}})
		testassert.SliceDoesNotContain(deps.ui.Outputs, testassert.Lines{{"Ok"}})

		Expect(deps.serviceRepo.PurgeServiceOfferingCalled).To(Equal(false))
	})
})

type commandDependencies struct {
	ui          *testterm.FakeUI
	config      configuration.ReadWriter
	serviceRepo *testapi.FakeServiceRepo
	reqFactory  *testreq.FakeReqFactory
}

func setupDependencies() (obj commandDependencies) {
	obj.ui = &testterm.FakeUI{}

	obj.config = testconfig.NewRepository()
	obj.reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}
	obj.serviceRepo = new(testapi.FakeServiceRepo)
	return
}
