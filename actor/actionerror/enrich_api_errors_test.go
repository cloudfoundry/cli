package actionerror_test

import (
	"errors"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"

	"code.cloudfoundry.org/cli/actor/actionerror"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("service access actions", func() {
	It("enriches a ServiceOfferingNameAmbiguityError", func() {
		err := actionerror.EnrichAPIErrors(ccerror.ServiceOfferingNameAmbiguityError{
			ServiceOfferingName: "foo",
			ServiceBrokerNames:  []string{"bar", "baz", "qux"},
		})

		Expect(err).To(MatchError("Service 'foo' is provided by multiple service brokers: bar, baz, qux\nSpecify a broker by using the '-b' flag."))
	})

	It("handles nil", func() {
		Expect(actionerror.EnrichAPIErrors(nil)).To(BeNil())
	})

	It("passes though other errors", func() {
		Expect(actionerror.EnrichAPIErrors(errors.New("foo"))).To(MatchError(errors.New("foo")))
	})
})
