package actionerror_test

import (
	"fmt"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ServicePlanNotFoundError", func() {
	const (
		offeringName = "some-service-offering"
		planName     = "some-plan"
		brokerName   = "some-broker"
	)

	It("returns a message with the service offering name and the plan name", func() {
		err := actionerror.ServicePlanNotFoundError{
			OfferingName: offeringName,
			PlanName:     planName,
		}
		Expect(err.Error()).To(Equal(fmt.Sprintf("The plan %s could not be found for service offering %s.", planName, offeringName)))
	})

	It("returns a message with the service offering name, the plan name and broker name", func() {
		err := actionerror.ServicePlanNotFoundError{
			OfferingName:      offeringName,
			PlanName:          planName,
			ServiceBrokerName: brokerName,
		}
		Expect(err.Error()).To(Equal(fmt.Sprintf(
			"The plan %s could not be found for service offering %s and broker %s.",
			planName,
			offeringName,
			brokerName,
		)))
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
