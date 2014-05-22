package service_test

import (
	. "github.com/cloudfoundry/cli/cf/commands/service"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	testapi "github.com/cloudfoundry/cli/testhelpers/api"
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
		config              configuration.Repository
		requirementsFactory *testreq.FakeReqFactory
		cmd                 CreateService
		serviceRepo         *testapi.FakeServiceRepo

		offering1 models.ServiceOffering
		offering2 models.ServiceOffering
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		config = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}
		serviceRepo = &testapi.FakeServiceRepo{}
		cmd = NewCreateService(ui, config, serviceRepo)

		offering1 = models.ServiceOffering{}
		offering1.Label = "cleardb"
		offering1.Plans = []models.ServicePlanFields{{
			Name: "spark",
			Guid: "cleardb-spark-guid",
		}}

		offering2 = models.ServiceOffering{}
		offering2.Label = "postgres"

		serviceRepo.FindServiceOfferingsForSpaceByLabelReturns.ServiceOfferings = []models.ServiceOffering{offering1, offering2}
	})

	var callCreateService = func(args []string) {
		testcmd.RunCommand(cmd, args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("passes when logged in and a space is targeted", func() {
			callCreateService([]string{"cleardb", "spark", "my-cleardb-service"})
			Expect(testcmd.CommandDidPassRequirements).To(BeTrue())
		})

		It("fails when not logged in", func() {
			requirementsFactory.LoginSuccess = false
			callCreateService([]string{"cleardb", "spark", "my-cleardb-service"})
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})

		It("fails when a space is not targeted", func() {
			requirementsFactory.TargetedSpaceSuccess = false
			callCreateService([]string{"cleardb", "spark", "my-cleardb-service"})
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})
	})

	It("successfully creates a service", func() {
		callCreateService([]string{"cleardb", "spark", "my-cleardb-service"})

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Creating service", "my-cleardb-service", "my-org", "my-space", "my-user"},
			[]string{"OK"},
		))
		Expect(serviceRepo.CreateServiceInstanceArgs.Name).To(Equal("my-cleardb-service"))
		Expect(serviceRepo.CreateServiceInstanceArgs.PlanGuid).To(Equal("cleardb-spark-guid"))
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
