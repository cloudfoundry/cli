package service_test

import (
	"github.com/cloudfoundry/cli/cf/actors/service_builder/fakes"
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	. "github.com/cloudfoundry/cli/cf/commands/service"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

var _ = Describe("create-service command", func() {
	var (
		ui                  *testterm.FakeUI
		config              core_config.Repository
		requirementsFactory *testreq.FakeReqFactory
		cmd                 CreateService
		serviceRepo         *testapi.FakeServiceRepo
		serviceBuilder      *fakes.FakeServiceBuilder

		offering1 models.ServiceOffering
		offering2 models.ServiceOffering
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		config = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}
		serviceRepo = &testapi.FakeServiceRepo{}
		serviceBuilder = &fakes.FakeServiceBuilder{}
		cmd = NewCreateService(ui, config, serviceRepo, serviceBuilder)

		offering1 = models.ServiceOffering{}
		offering1.Label = "cleardb"
		offering1.Plans = []models.ServicePlanFields{{
			Name: "spark",
			Guid: "cleardb-spark-guid",
			Free: true,
		}, {
			Name: "expensive",
			Guid: "luxury-guid",
			Free: false,
		}}

		offering2 = models.ServiceOffering{}
		offering2.Label = "postgres"

		serviceBuilder.GetServicesByNameForSpaceWithPlansReturns(models.ServiceOfferings{offering1, offering2}, nil)
	})

	var callCreateService = func(args []string) bool {
		return testcmd.RunCommand(cmd, args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("passes when logged in and a space is targeted", func() {
			Expect(callCreateService([]string{"cleardb", "spark", "my-cleardb-service"})).To(BeTrue())
		})

		It("fails when not logged in", func() {
			requirementsFactory.LoginSuccess = false
			Expect(callCreateService([]string{"cleardb", "spark", "my-cleardb-service"})).To(BeFalse())
		})

		It("fails when a space is not targeted", func() {
			requirementsFactory.TargetedSpaceSuccess = false
			Expect(callCreateService([]string{"cleardb", "spark", "my-cleardb-service"})).To(BeFalse())
		})
	})

	It("successfully creates a service", func() {
		callCreateService([]string{"cleardb", "spark", "my-cleardb-service"})

		spaceGuid, serviceName := serviceBuilder.GetServicesByNameForSpaceWithPlansArgsForCall(0)
		Expect(spaceGuid).To(Equal(config.SpaceFields().Guid))
		Expect(serviceName).To(Equal("cleardb"))
		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Creating service", "my-cleardb-service", "my-org", "my-space", "my-user"},
			[]string{"OK"},
		))
		Expect(serviceRepo.CreateServiceInstanceArgs.Name).To(Equal("my-cleardb-service"))
		Expect(serviceRepo.CreateServiceInstanceArgs.PlanGuid).To(Equal("cleardb-spark-guid"))
	})

	Describe("warning the user about paid services", func() {
		It("does not warn the user when the service is free", func() {
			callCreateService([]string{"cleardb", "spark", "my-free-cleardb-service"})
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Creating service", "my-free-cleardb-service", "my-org", "my-space", "my-user"},
				[]string{"OK"},
			))
			Expect(ui.Outputs).NotTo(ContainSubstrings([]string{"will incurr a cost"}))
		})

		It("warns the user when the service is not free", func() {
			callCreateService([]string{"cleardb", "expensive", "my-expensive-cleardb-service"})
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Creating service", "my-expensive-cleardb-service", "my-org", "my-space", "my-user"},
				[]string{"OK"},
				[]string{"Attention: The plan `expensive` of service `cleardb` is not free.  The instance `my-expensive-cleardb-service` will incur a cost.  Contact your administrator if you think this is in error."},
			))
		})
	})

	It("warns the user when the service already exists with the same service plan", func() {
		serviceRepo.CreateServiceInstanceReturns.Error = errors.NewModelAlreadyExistsError("Service", "my-cleardb-service")

		callCreateService([]string{"cleardb", "spark", "my-cleardb-service"})

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Creating service", "my-cleardb-service"},
			[]string{"OK"},
			[]string{"my-cleardb-service", "already exists"},
		))
		Expect(serviceRepo.CreateServiceInstanceArgs.Name).To(Equal("my-cleardb-service"))
		Expect(serviceRepo.CreateServiceInstanceArgs.PlanGuid).To(Equal("cleardb-spark-guid"))
	})

	Context("When there are multiple services with the same label", func() {
		It("finds the plan even if it has to search multiple services", func() {
			offering2.Label = "cleardb"

			serviceRepo.CreateServiceInstanceReturns.Error = errors.NewModelAlreadyExistsError("Service", "my-cleardb-service")
			callCreateService([]string{"cleardb", "spark", "my-cleardb-service"})

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Creating service", "my-cleardb-service", "my-org", "my-space", "my-user"},
				[]string{"OK"},
			))
			Expect(serviceRepo.CreateServiceInstanceArgs.Name).To(Equal("my-cleardb-service"))
			Expect(serviceRepo.CreateServiceInstanceArgs.PlanGuid).To(Equal("cleardb-spark-guid"))
		})
	})
})
