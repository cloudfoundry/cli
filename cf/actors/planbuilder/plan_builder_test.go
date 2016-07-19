package planbuilder_test

import (
	"errors"

	"code.cloudfoundry.org/cli/cf/actors/planbuilder"
	"code.cloudfoundry.org/cli/cf/api/apifakes"
	"code.cloudfoundry.org/cli/cf/api/organizations/organizationsfakes"
	"code.cloudfoundry.org/cli/cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Plan builder", func() {
	var (
		builder planbuilder.PlanBuilder

		planRepo       *apifakes.OldFakeServicePlanRepo
		visibilityRepo *apifakes.FakeServicePlanVisibilityRepository
		orgRepo        *organizationsfakes.FakeOrganizationRepository

		plan1 models.ServicePlanFields
		plan2 models.ServicePlanFields

		org1 models.Organization
		org2 models.Organization
	)

	BeforeEach(func() {
		planbuilder.PlanToOrgsVisibilityMap = nil
		planbuilder.OrgToPlansVisibilityMap = nil
		planRepo = new(apifakes.OldFakeServicePlanRepo)
		visibilityRepo = new(apifakes.FakeServicePlanVisibilityRepository)
		orgRepo = new(organizationsfakes.FakeOrganizationRepository)
		builder = planbuilder.NewBuilder(planRepo, visibilityRepo, orgRepo)

		plan1 = models.ServicePlanFields{
			Name:                "service-plan1",
			GUID:                "service-plan1-guid",
			ServiceOfferingGUID: "service-guid1",
		}
		plan2 = models.ServicePlanFields{
			Name:                "service-plan2",
			GUID:                "service-plan2-guid",
			ServiceOfferingGUID: "service-guid1",
		}

		planRepo.SearchReturns = map[string][]models.ServicePlanFields{
			"service-guid1": {plan1, plan2},
		}
		org1 = models.Organization{}
		org1.Name = "org1"
		org1.GUID = "org1-guid"

		org2 = models.Organization{}
		org2.Name = "org2"
		org2.GUID = "org2-guid"
		visibilityRepo.ListReturns([]models.ServicePlanVisibilityFields{
			{ServicePlanGUID: "service-plan1-guid", OrganizationGUID: "org1-guid"},
			{ServicePlanGUID: "service-plan1-guid", OrganizationGUID: "org2-guid"},
			{ServicePlanGUID: "service-plan2-guid", OrganizationGUID: "org1-guid"},
		}, nil)
		orgRepo.GetManyOrgsByGUIDReturns([]models.Organization{org1, org2}, nil)
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
			serviceGUIDs := []string{"service-guid1", "service-guid2"}
			plans, err := builder.GetPlansForManyServicesWithOrgs(serviceGUIDs)
			Expect(err).ToNot(HaveOccurred())
			Expect(orgRepo.GetManyOrgsByGUIDCallCount()).To(Equal(1))
			Expect(orgRepo.GetManyOrgsByGUIDArgsForCall(0)).To(ConsistOf("org1-guid", "org2-guid"))

			Expect(len(plans)).To(Equal(2))
			Expect(plans[0].Name).To(Equal("service-plan1"))
			Expect(plans[0].OrgNames).To(Equal([]string{"org1", "org2"}))
			Expect(plans[1].Name).To(Equal("service-plan2"))
		})

		It("returns errors from the service plan repo", func() {
			planRepo.ListPlansFromManyServicesError = errors.New("Error")
			serviceGUIDs := []string{"service-guid1", "service-guid2"}
			_, err := builder.GetPlansForManyServicesWithOrgs(serviceGUIDs)
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
