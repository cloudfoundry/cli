package plan_builder_test

import (
	"errors"

	"github.com/cloudfoundry/cli/cf/actors/plan_builder"
	"github.com/cloudfoundry/cli/cf/api/fakes"
	testorg "github.com/cloudfoundry/cli/cf/api/organizations/fakes"
	"github.com/cloudfoundry/cli/cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Plan builder", func() {
	var (
		builder plan_builder.PlanBuilder

		planRepo       *fakes.FakeServicePlanRepo
		visibilityRepo *fakes.FakeServicePlanVisibilityRepository
		orgRepo        *testorg.FakeOrganizationRepository

		plan1 models.ServicePlanFields
		plan2 models.ServicePlanFields

		org1 models.Organization
		org2 models.Organization
	)

	BeforeEach(func() {
		plan_builder.PlanToOrgsVisibilityMap = nil
		plan_builder.OrgToPlansVisibilityMap = nil
		planRepo = &fakes.FakeServicePlanRepo{}
		visibilityRepo = &fakes.FakeServicePlanVisibilityRepository{}
		orgRepo = &testorg.FakeOrganizationRepository{}
		builder = plan_builder.NewBuilder(planRepo, visibilityRepo, orgRepo)

		plan1 = models.ServicePlanFields{
			Name:                "service-plan1",
			Guid:                "service-plan1-guid",
			ServiceOfferingGuid: "service-guid1",
		}
		plan2 = models.ServicePlanFields{
			Name:                "service-plan2",
			Guid:                "service-plan2-guid",
			ServiceOfferingGuid: "service-guid1",
		}

		planRepo.SearchReturns = map[string][]models.ServicePlanFields{
			"service-guid1": []models.ServicePlanFields{plan1, plan2},
		}
		org1 = models.Organization{}
		org1.Name = "org1"
		org1.Guid = "org1-guid"

		org2 = models.Organization{}
		org2.Name = "org2"
		org2.Guid = "org2-guid"
		visibilityRepo.ListReturns([]models.ServicePlanVisibilityFields{
			{ServicePlanGuid: "service-plan1-guid", OrganizationGuid: "org1-guid"},
			{ServicePlanGuid: "service-plan1-guid", OrganizationGuid: "org2-guid"},
			{ServicePlanGuid: "service-plan2-guid", OrganizationGuid: "org1-guid"},
		}, nil)
		orgRepo.GetManyOrgsByGuidReturns([]models.Organization{org1, org2}, nil)
	})

	Describe(".AttachOrgsToPlans", func() {
		It("returns plans fully populated with the orgnames that have visibility", func() {
			barePlans := []models.ServicePlanFields{plan1, plan2}

			plans, err := builder.AttachOrgsToPlans(barePlans)
			Expect(err).ToNot(HaveOccurred())

			Expect(plans[0].OrgNames).To(Equal([]string{"org1", "org2"}))
		})
	})

	Describe(".AttachOrgToPlans", func() {
		It("returns plans fully populated with the orgnames that have visibility", func() {
			orgRepo.FindByNameReturns(org1, nil)
			barePlans := []models.ServicePlanFields{plan1, plan2}

			plans, err := builder.AttachOrgToPlans(barePlans, "org1")
			Expect(err).ToNot(HaveOccurred())

			Expect(plans[0].OrgNames).To(Equal([]string{"org1"}))
		})
	})

	Describe(".GetPlansForServiceWithOrgs", func() {
		It("returns all the plans for the service with the provided guid", func() {
			plans, err := builder.GetPlansForServiceWithOrgs("service-guid1")
			Expect(err).ToNot(HaveOccurred())

			Expect(len(plans)).To(Equal(2))
			Expect(plans[0].Name).To(Equal("service-plan1"))
			Expect(plans[0].OrgNames).To(Equal([]string{"org1", "org2"}))
			Expect(plans[1].Name).To(Equal("service-plan2"))
		})
	})

	Describe(".GetPlansForManyServicesWithOrgs", func() {
		It("returns all the plans for all service in a list of guids", func() {
			planRepo.ListPlansFromManyServicesReturns = []models.ServicePlanFields{
				plan1, plan2,
			}
			serviceGuids := []string{"service-guid1", "service-guid2"}
			plans, err := builder.GetPlansForManyServicesWithOrgs(serviceGuids)
			Expect(err).ToNot(HaveOccurred())
			Expect(orgRepo.GetManyOrgsByGuidCallCount()).To(Equal(1))
			Expect(orgRepo.GetManyOrgsByGuidArgsForCall(0)).To(ConsistOf("org1-guid", "org2-guid"))

			Expect(len(plans)).To(Equal(2))
			Expect(plans[0].Name).To(Equal("service-plan1"))
			Expect(plans[0].OrgNames).To(Equal([]string{"org1", "org2"}))
			Expect(plans[1].Name).To(Equal("service-plan2"))
		})

		It("returns errors from the service plan repo", func() {
			planRepo.ListPlansFromManyServicesError = errors.New("Error")
			serviceGuids := []string{"service-guid1", "service-guid2"}
			_, err := builder.GetPlansForManyServicesWithOrgs(serviceGuids)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe(".GetPlansForService", func() {
		It("returns all the plans for the service with the provided guid", func() {
			plans, err := builder.GetPlansForService("service-guid1")
			Expect(err).ToNot(HaveOccurred())

			Expect(len(plans)).To(Equal(2))
			Expect(plans[0].Name).To(Equal("service-plan1"))
			Expect(plans[0].OrgNames).To(BeNil())
			Expect(plans[1].Name).To(Equal("service-plan2"))
		})
	})

	Describe(".GetPlansForServiceForOrg", func() {
		It("returns all the plans for the service with the provided guid", func() {
			orgRepo.FindByNameReturns(org1, nil)
			plans, err := builder.GetPlansForServiceForOrg("service-guid1", "org1")
			Expect(err).ToNot(HaveOccurred())

			Expect(len(plans)).To(Equal(2))
			Expect(plans[0].Name).To(Equal("service-plan1"))
			Expect(plans[0].OrgNames).To(Equal([]string{"org1"}))
			Expect(plans[1].Name).To(Equal("service-plan2"))
		})
	})

	Describe(".GetPlansVisibleToOrg", func() {
		It("returns all the plans visible to the named org", func() {
			plans, err := builder.GetPlansVisibleToOrg("org1")
			Expect(err).ToNot(HaveOccurred())

			Expect(len(plans)).To(Equal(2))
			Expect(plans[0].Name).To(Equal("service-plan1"))
			Expect(plans[0].OrgNames).To(Equal([]string{"org1", "org2"}))
		})
	})
})
