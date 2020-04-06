package ccerror_test

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ServiceOfferingNotFoundError", func() {
	It("returns a message with the service offering and broker name", func() {
		err := ccerror.ServiceOfferingNotFoundError{
			ServiceOfferingName: "some-service-offering",
			ServiceBrokerName:   "some-broker",
		}
		Expect(err.Error()).To(Equal("Service offering 'some-service-offering' for service broker 'some-broker' not found."))
	})

	When("there is no broker name", func() {
		It("does not mention the broker in the message", func() {
			err := ccerror.ServiceOfferingNotFoundError{
				ServiceOfferingName: "some-service-offering",
			}
			Expect(err.Error()).To(Equal("Service offering 'some-service-offering' not found."))
		})
	})

	When("there is no service offering name", func() {
		It("does not mention the service offering in the message", func() {
			err := ccerror.ServiceOfferingNotFoundError{
				ServiceBrokerName: "some-broker",
			}
			Expect(err.Error()).To(Equal("No service offerings found for service broker 'some-broker'."))
		})
	})

	When("there are no names", func() {
		It("does not mention any names in the message", func() {
			err := ccerror.ServiceOfferingNotFoundError{}
			Expect(err.Error()).To(Equal("No service offerings found."))
		})
	})
})
