package actionerror_test

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	. "github.com/onsi/ginkgo"
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
})
