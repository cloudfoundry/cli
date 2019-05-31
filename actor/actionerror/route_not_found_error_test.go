package actionerror_test

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RouteNotFoundError", func() {

	Describe("Error", func() {
		When("DomainName is set", func() {
			When("the host and path are specified", func() {
				It("returns an error message referencing domain, host and path", func() {
					err := actionerror.RouteNotFoundError{
						DomainName: "some-domain.com",
						Path:       "mypath",
						Host:       "hostname",
					}
					Expect(err.Error()).To(Equal("Route with host 'hostname', domain 'some-domain.com', and path 'mypath' not found."))
				})
			})

			When("the host is specified", func() {
				It("returns an error message referencing domain and host", func() {
					err := actionerror.RouteNotFoundError{
						DomainName: "some-domain.com",
						Host:       "hostname",
					}
					Expect(err.Error()).To(Equal("Route with host 'hostname' and domain 'some-domain.com' not found."))
				})
			})

			When("the path is specified", func() {
				It("returns an error message referencing domain and path", func() {
					err := actionerror.RouteNotFoundError{
						DomainName: "some-domain.com",
						Path:       "mypath",
					}
					Expect(err.Error()).To(Equal("Route with domain 'some-domain.com' and path 'mypath' not found."))
				})
			})

			When("neither host nor path is specified", func() {
				It("returns an error message referencing domain", func() {
					err := actionerror.RouteNotFoundError{
						DomainName: "some-domain.com",
					}
					Expect(err.Error()).To(Equal("Route with domain 'some-domain.com' not found."))
				})
			})
		})
		When("Domain GUID is set", func() {
			It("returns an error message with the GUID of the missing space", func() {
				err := actionerror.RouteNotFoundError{
					DomainGUID: "some-domain-guid",
					Host:       "hostname",
					Path:       "mypath",
				}
				Expect(err.Error()).To(Equal("Route with host 'hostname', domain guid 'some-domain-guid', and path 'mypath' not found."))
			})
		})
	})
})
