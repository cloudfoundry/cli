package service_test

import (
	"cf"
	. "cf/commands/service"
	"cf/configuration"
	"cf/net"
	"errors"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	"testhelpers/maker"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

func init() {
	Describe("Purging services", func() {
		It("does not run if the requirements are not met", func() {
			t := mr.T()
			deps := setupDependencies()

			testcmd.RunCommand(
				NewPurgeServiceOffering(deps.ui, deps.config, deps.serviceRepo),
				testcmd.NewContext("purge-service-offering", []string{}),
				deps.reqFactory,
			)

			assert.False(t, testcmd.CommandDidPassRequirements)
			assert.True(t, deps.ui.FailedWithUsage)
			assert.Equal(t, deps.ui.FailedWithUsageCommandName, "purge-service-offering")
		})

		It("works when given -p and a provider name", func() {
			t := mr.T()
			deps := setupDependencies()

			offering := maker.NewServiceOffering("the-service-name")
			deps.serviceRepo.FindServiceOfferingByLabelAndProviderServiceOffering = offering

			deps.ui.Inputs = []string{"yes"}

			testcmd.RunCommand(
				NewPurgeServiceOffering(deps.ui, deps.config, deps.serviceRepo),
				testcmd.NewContext("purge-service-offering", []string{"-p", "the-provider", "the-service-name"}),
				deps.reqFactory,
			)

			assert.Equal(t, deps.serviceRepo.FindServiceOfferingByLabelAndProviderName, "the-service-name")
			assert.Equal(t, deps.serviceRepo.FindServiceOfferingByLabelAndProviderProvider, "the-provider")
			assert.Equal(t, deps.serviceRepo.PurgedServiceOffering, offering)
		})

		It("works when not given a provider", func() {
			t := mr.T()
			deps := setupDependencies()

			offering := maker.NewServiceOffering("the-service-name")
			deps.serviceRepo.FindServiceOfferingByLabelAndProviderServiceOffering = offering

			deps.ui.Inputs = []string{"yes"}

			testcmd.RunCommand(
				NewPurgeServiceOffering(deps.ui, deps.config, deps.serviceRepo),
				testcmd.NewContext("purge-service-offering", []string{"the-service-name"}),
				deps.reqFactory,
			)

			testassert.SliceContains(t, deps.ui.Outputs, testassert.Lines{
				{"Warning"},
			})
			testassert.SliceContains(t, deps.ui.Prompts, testassert.Lines{
				{"Really purge service", "the-service-name"},
			})

			assert.Equal(t, deps.serviceRepo.FindServiceOfferingByLabelAndProviderName, "the-service-name")
			assert.Equal(t, deps.serviceRepo.FindServiceOfferingByLabelAndProviderProvider, "")
			assert.Equal(t, deps.serviceRepo.PurgedServiceOffering, offering)

			testassert.SliceContains(t, deps.ui.Outputs, testassert.Lines{
				{"OK"},
			})
		})

		It("exits when the user does not acknowledge the confirmation", func() {
			t := mr.T()
			deps := setupDependencies()

			deps.ui.Inputs = []string{"no"}

			testcmd.RunCommand(
				NewPurgeServiceOffering(deps.ui, deps.config, deps.serviceRepo),
				testcmd.NewContext("purge-service-offering", []string{"the-service-name"}),
				deps.reqFactory,
			)

			assert.Equal(t, deps.serviceRepo.FindServiceOfferingByLabelAndProviderCalled, false)
			assert.Equal(t, deps.serviceRepo.PurgeServiceOfferingCalled, false)
		})

		It("does not prompt with confirmation when -f is passed", func() {
			t := mr.T()
			deps := setupDependencies()

			offering := maker.NewServiceOffering("the-service-name")
			deps.serviceRepo.FindServiceOfferingByLabelAndProviderServiceOffering = offering

			testcmd.RunCommand(
				NewPurgeServiceOffering(deps.ui, deps.config, deps.serviceRepo),
				testcmd.NewContext("purge-service-offering", []string{"-f", "the-service-name"}),
				deps.reqFactory,
			)

			assert.Equal(t, len(deps.ui.Prompts), 0)
			assert.Equal(t, deps.serviceRepo.PurgeServiceOfferingCalled, true)
		})

		It("fails with an error message when the request fails", func() {
			t := mr.T()
			deps := setupDependencies()

			deps.serviceRepo.FindServiceOfferingByLabelAndProviderApiResponse = net.NewApiResponseWithError("oh no!", errors.New("!"))

			testcmd.RunCommand(
				NewPurgeServiceOffering(deps.ui, deps.config, deps.serviceRepo),
				testcmd.NewContext("purge-service-offering", []string{"-f", "-p", "the-provider", "the-service-name"}),
				deps.reqFactory,
			)

			testassert.SliceContains(t, deps.ui.Outputs, testassert.Lines{
				{"FAILED"},
				{"oh no!"},
			})

			assert.Equal(t, deps.serviceRepo.PurgeServiceOfferingCalled, false)
		})

		It("indicates when a service doesn't exist", func() {
			t := mr.T()
			deps := setupDependencies()

			deps.serviceRepo.FindServiceOfferingByLabelAndProviderApiResponse = net.NewNotFoundApiResponse("uh oh cant find it")

			testcmd.RunCommand(
				NewPurgeServiceOffering(deps.ui, deps.config, deps.serviceRepo),
				testcmd.NewContext("purge-service-offering", []string{"-f", "-p", "the-provider", "the-service-name"}),
				deps.reqFactory,
			)

			testassert.SliceContains(t, deps.ui.Outputs, testassert.Lines{
				{"OK"},
				{"Service offering", "does not exist"},
			})

			assert.Equal(t, deps.serviceRepo.PurgeServiceOfferingCalled, false)
		})
	})
}

type commandDependencies struct {
	ui          *testterm.FakeUI
	config      *configuration.Configuration
	serviceRepo *testapi.FakeServiceRepo
	reqFactory  *testreq.FakeReqFactory
}

func setupDependencies() (obj commandDependencies) {
	obj.ui = &testterm.FakeUI{}

	token, _ := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	org := cf.OrganizationFields{}
	org.Name = "my-org"
	space := cf.SpaceFields{}
	space.Name = "my-space"
	obj.config = &configuration.Configuration{
		SpaceFields:        space,
		OrganizationFields: org,
		AccessToken:        token,
	}

	obj.reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}
	obj.serviceRepo = new(testapi.FakeServiceRepo)
	return
}
