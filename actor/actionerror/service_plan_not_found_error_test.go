package actionerror_test

import (
	"fmt"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ServicePlanNotFoundError", func() {
	It("returns a message with the service offering name", func() {
		offeringName := "some-service-offering"
		planName := "some-plan"
		err := actionerror.ServicePlanNotFoundError{
			ServiceName: offeringName,
			PlanName:    planName,
		}
		Expect(err.Error()).To(Equal(fmt.Sprintf("The plan %s could not be found for service %s", planName, offeringName)))
	})

	When("there is no service offering name", func() {
		It("returns a generic message", func() {
			err := actionerror.ServicePlanNotFoundError{
				PlanName: "some-plan",
			}
			Expect(err.Error()).To(Equal("Service plan 'some-plan' not found."))
		})
	})
})
