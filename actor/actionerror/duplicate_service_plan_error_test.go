package actionerror_test

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DuplicateServicePlanError", func() {
	Describe("Error", func() {
		When("Only Name is specified", func() {
			It("returns the right error message", func() {
				err := actionerror.DuplicateServicePlanError{
					Name: "some-service-plan-name",
				}
				Expect(err.Error()).To(
					Equal("Service plan 'some-service-plan-name' is provided by multiple service offerings. Specify an offering by using the '-e' flag."))
			})
		})
		When("Name and ServiceOfferingName are specified", func() {
			It("returns the right error message", func() {
				err := actionerror.DuplicateServicePlanError{
					Name:                "some-service-plan-name",
					ServiceOfferingName: "some-service-offering-name",
				}
				Expect(err.Error()).To(
					Equal("Service plan 'some-service-plan-name' is provided by multiple service offerings. Service offering 'some-service-offering-name' is provided by multiple service brokers. Specify a broker name by using the '-b' flag."))
			})
		})
		When("Name and ServiceBrokerName are specified", func() {
			It("returns the right error message", func() {
				err := actionerror.DuplicateServicePlanError{
					Name:              "some-service-plan-name",
					ServiceBrokerName: "some-broker-name",
				}
				Expect(err.Error()).To(
					Equal("Service plan 'some-service-plan-name' is provided by multiple service offerings. Specify an offering by using the '-e' flag."))
			})
		})
	})
})
