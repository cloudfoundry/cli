package actionerror_test

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Duplicate Service Error", func() {
	It("returns the right error message", func() {
		err := actionerror.DuplicateServiceError{
			Name: "some-service-name",
		}
		Expect(err.Error()).To(
			Equal("Service 'some-service-name' is provided by multiple service brokers. Specify a broker by using the '-b' flag."))
	})

	When("service broker names are specified", func() {
		It("returns the right error message", func() {
			err := actionerror.DuplicateServiceError{
				Name:           "some-service-name",
				ServiceBrokers: []string{"a-service-broker", "another-service-broker"},
			}
			Expect(err.Error()).To(
				Equal("Service 'some-service-name' is provided by multiple service brokers.\n" +
					"Specify a broker from available brokers 'a-service-broker', 'another-service-broker' by using the '-b' flag."))
		})
	})
})
