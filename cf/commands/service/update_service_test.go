package service_test

import (
	"errors"

	testplanbuilder "github.com/cloudfoundry/cli/cf/actors/plan_builder/fakes"
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	. "github.com/cloudfoundry/cli/cf/commands/service"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

var _ = Describe("update-service command", func() {
	var (
		ui                  *testterm.FakeUI
		config              core_config.Repository
		requirementsFactory *testreq.FakeReqFactory
		serviceRepo         *testapi.FakeServiceRepo
		planBuilder         *testplanbuilder.FakePlanBuilder
		offering1           models.ServiceOffering
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		config = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}
		serviceRepo = &testapi.FakeServiceRepo{}
		planBuilder = &testplanbuilder.FakePlanBuilder{}

		offering1 = models.ServiceOffering{}
		offering1.Label = "cleardb"
		offering1.Plans = []models.ServicePlanFields{{
			Name: "spark",
			Guid: "cleardb-spark-guid",
		}, {
			Name: "flare",
			Guid: "cleardb-flare-guid",
		},
		}

		//serviceRepo.FindServiceOfferingsForSpaceByLabelReturns.ServiceOfferings = []models.ServiceOffering{offering1}
	})

	var callUpdateService = func(args []string) {
		cmd := NewUpdateService(ui, config, serviceRepo, planBuilder)
		testcmd.RunCommand(cmd, args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("passes when logged in and a space is targeted", func() {
			callUpdateService([]string{"cleardb", "spark", "my-cleardb-service"})
			Expect(testcmd.CommandDidPassRequirements).To(BeTrue())
		})

		It("fails when there are 0 arguments", func() {
			callUpdateService([]string{})
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})

		It("fails when not logged in", func() {
			requirementsFactory.LoginSuccess = false
			callUpdateService([]string{"cleardb", "spark", "my-cleardb-service"})
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})

		It("fails when a space is not targeted", func() {
			requirementsFactory.TargetedSpaceSuccess = false
			callUpdateService([]string{"cleardb", "spark", "my-cleardb-service"})
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})
	})
	Context("when no flags are passed", func() {
		Context("when there is an err finding the instance", func() {
			It("returns an error", func() {
				serviceRepo.FindInstanceByNameErr = true

				callUpdateService([]string{"some-stupid-not-real-instance"})

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Error finding instance"},
					[]string{"FAILED"},
				))
			})
		})
		Context("when the instance exists", func() {
			It("prints a user indicating it is a no-op", func() {
				callUpdateService([]string{"my-service"})

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"OK"},
					[]string{"No changes were made"},
				))
			})
		})
	})
	Context("when the plan flag is passed", func() {
		BeforeEach(func() {
			serviceInstance := models.ServiceInstance{
				ServiceInstanceFields: models.ServiceInstanceFields{
					Name: "my-service-instance",
					Guid: "my-service-instance-guid",
				},
				ServiceOffering: models.ServiceOfferingFields{
					Label: "murkydb",
					Guid:  "murkydb-guid",
				},
			}

			servicePlans := []models.ServicePlanFields{{
				Name: "spark",
				Guid: "murkydb-spark-guid",
			}, {
				Name: "flare",
				Guid: "murkydb-flare-guid",
			},
			}
			serviceRepo.FindInstanceByNameServiceInstance = serviceInstance
			planBuilder.GetPlansForServiceForOrgReturns(servicePlans, nil)

		})
		It("successfully updates a service", func() {
			callUpdateService([]string{"-p", "flare", "my-service-instance"})

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Updating service", "my-service", "as", "my-user", "..."},
				[]string{"OK"},
			))
			Expect(serviceRepo.FindInstanceByNameName).To(Equal("my-service-instance"))
			serviceGuid, orgName := planBuilder.GetPlansForServiceForOrgArgsForCall(0)
			Expect(serviceGuid).To(Equal("murkydb-guid"))
			Expect(orgName).To(Equal("my-org"))
			Expect(serviceRepo.UpdateServiceInstanceArgs.InstanceGuid).To(Equal("my-service-instance-guid"))
			Expect(serviceRepo.UpdateServiceInstanceArgs.PlanGuid).To(Equal("murkydb-flare-guid"))
		})

		Context("when there is an err finding the instance", func() {
			It("returns an error", func() {
				serviceRepo.FindInstanceByNameErr = true

				callUpdateService([]string{"-p", "flare", "some-stupid-not-real-instance"})

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Error finding instance"},
					[]string{"FAILED"},
				))
			})
		})
		Context("when there is an err finding service plans", func() {
			It("returns an error", func() {
				planBuilder.GetPlansForServiceForOrgReturns(nil, errors.New("Error fetching plans"))

				callUpdateService([]string{"-p", "flare", "some-stupid-not-real-instance"})

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Error fetching plans"},
					[]string{"FAILED"},
				))
			})
		})
		Context("when the plan specified does not exist in the service offering", func() {
			It("returns an error", func() {
				callUpdateService([]string{"-p", "not-a-real-plan", "instance-without-service-offering"})

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Plan does not exist for the murkydb service"},
					[]string{"FAILED"},
				))
			})
		})
		Context("when there is an error updating the service instance", func() {
			It("returns an error", func() {
				serviceRepo.UpdateServiceInstanceReturnsErr = true
				callUpdateService([]string{"-p", "flare", "my-service-instance"})

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Error updating service instance"},
					[]string{"FAILED"},
				))
			})
		})
	})
})
