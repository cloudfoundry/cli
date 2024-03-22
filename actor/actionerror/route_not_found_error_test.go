package actionerror_test

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("RouteNotFoundError", func() {
	It("returns an error message referencing domain, host and path", func() {
		err := actionerror.RouteNotFoundError{
			DomainName: "some-domain.com",
			Path:       "mypath",
			Host:       "hostname",
		}
		Expect(err).To(MatchError("Route with host 'hostname', domain 'some-domain.com', and path 'mypath' not found."))
	})

	When("the hostname and path are empty", func() {
		It("return an error with empty hostname and path", func() {
			err := actionerror.RouteNotFoundError{DomainName: "some-domain.com"}
			Expect(err).To(MatchError("Route with host '', domain 'some-domain.com', and path '/' not found."))
		})
	})

	When("the port is specified", func() {
		It("returns an error message referencing domain and port", func() {
			err := actionerror.RouteNotFoundError{
				DomainName: "some-domain.com",
				Port:       1052,
			}
			Expect(err).To(MatchError("Route with domain 'some-domain.com' and port 1052 not found."))
		})
	})
})
