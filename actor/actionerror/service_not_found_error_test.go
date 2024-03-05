package actionerror_test

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ServiceNotFoundError", func() {
	It("returns a message with the service offering and broker name", func() {
		err := actionerror.ServiceNotFoundError{
			Name:   "some-service-offering",
			Broker: "some-broker",
		}
		Expect(err.Error()).To(Equal("Service offering 'some-service-offering' for service broker 'some-broker' not found."))
	})

	When("there is no broker name", func() {
		It("does not mention the broker in the message", func() {
			err := actionerror.ServiceNotFoundError{
				Name: "some-service-offering",
			}
			Expect(err.Error()).To(Equal("Service offering 'some-service-offering' not found."))
		})
	})

	When("there is no service offering name", func() {
		It("does not mention the service offering in the message", func() {
			err := actionerror.ServiceNotFoundError{
				Broker: "some-broker",
			}
			Expect(err.Error()).To(Equal("No service offerings found for service broker 'some-broker'."))
		})
	})

	When("there are no names", func() {
		It("does not mention any names in the message", func() {
			err := actionerror.ServiceNotFoundError{}
			Expect(err.Error()).To(Equal("No service offerings found."))
		})
	})
})
