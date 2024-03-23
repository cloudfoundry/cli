package ccerror_test

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ServiceOfferingNameAmbiguityError", func() {
	It("returns the right error message", func() {
		err := ccerror.ServiceOfferingNameAmbiguityError{
			ServiceOfferingName: "some-service-name",
		}
		Expect(err.Error()).To(Equal("Service 'some-service-name' is provided by multiple service brokers."))
	})

	When("service broker names are specified", func() {
		It("returns the right error message", func() {
			err := ccerror.ServiceOfferingNameAmbiguityError{
				ServiceOfferingName: "some-service-name",
				ServiceBrokerNames:  []string{"a-service-broker", "another-service-broker"},
			}
			Expect(err.Error()).To(Equal("Service 'some-service-name' is provided by multiple service brokers: a-service-broker, another-service-broker"))
		})
	})
})
