package translatableerror_test

import (
	"code.cloudfoundry.org/cli/command/translatableerror"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ServicePlanNotFoundError", func() {
	It("returns a message with the service offering name", func() {
		offeringName := "some-service-offering"
		planName := "some-plan"
		err := translatableerror.ServicePlanNotFoundError{
			ServiceName: offeringName,
			PlanName:    planName,
		}
		Expect(err.Error()).To(Equal("The plan {{.PlanName}} could not be found for service {{.ServiceName}}"))
	})

	When("there is no service offering name", func() {
		It("returns a generic message", func() {
			err := translatableerror.ServicePlanNotFoundError{
				PlanName: "some-plan",
			}
			Expect(err.Error()).To(Equal("Service plan '{{.PlanName}}' not found."))
		})
	})
})
