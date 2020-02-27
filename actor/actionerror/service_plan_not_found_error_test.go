package actionerror_test

import (
	"fmt"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ServicePlanNotFoundError", func() {
	It("returns a message with the service offering name and the plan name", func() {
		offeringName := "some-service-offering"
		planName := "some-plan"
		err := actionerror.ServicePlanNotFoundError{
			OfferingName: offeringName,
			PlanName:     planName,
		}
		Expect(err.Error()).To(Equal(fmt.Sprintf("The plan %s could not be found for service %s.", planName, offeringName)))
	})

	When("no service offering name", func() {
		It("returns a message with the plan name", func() {
			err := actionerror.ServicePlanNotFoundError{
				PlanName: "some-plan",
			}
			Expect(err.Error()).To(Equal("Service plan 'some-plan' not found."))
		})
	})

	When("no plan name", func() {
		It("returns a message with the offering name", func() {
			err := actionerror.ServicePlanNotFoundError{
				OfferingName: "some-service",
			}
			Expect(err.Error()).To(Equal("No service plans found for service offering 'some-service'."))
		})
	})

	When("no names", func() {
		It("returns generic message", func() {
			err := actionerror.ServicePlanNotFoundError{}
			Expect(err.Error()).To(Equal("No service plans found."))
		})
	})
})
