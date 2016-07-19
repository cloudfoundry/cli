package models_test

import (
	. "code.cloudfoundry.org/cli/cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ServicePlanFields", func() {
	var servicePlanFields ServicePlanFields

	BeforeEach(func() {
		servicePlanFields = ServicePlanFields{
			GUID:                "I-am-a-guid",
			Name:                "BestServicePlanEver",
			Free:                false,
			Public:              true,
			Description:         "A Plan For Service",
			Active:              true,
			ServiceOfferingGUID: "service-offering-guid",
			OrgNames:            []string{"org1", "org2"},
		}
	})

	Describe(".OrgHasVisibility", func() {
		Context("when the service plan is public", func() {
			It("returns true", func() {
				Expect(servicePlanFields.OrgHasVisibility("anyOrg")).To(BeTrue())
			})
		})

		Context("when the service plan is not public", func() {
			BeforeEach(func() {
				servicePlanFields.Public = false
			})

			It("returns true if the orgname is in the list of orgs that have visibility", func() {
				Expect(servicePlanFields.OrgHasVisibility("org1")).To(BeTrue())
			})

			It("returns false if the orgname is not in the list of orgs that have visibility", func() {
				Expect(servicePlanFields.OrgHasVisibility("org-that-has-no-visibility")).To(BeFalse())
			})
		})

	})
})
