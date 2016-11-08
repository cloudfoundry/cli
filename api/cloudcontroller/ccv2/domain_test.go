package ccv2_test

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Domain", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("GetSharedDomain", func() {
		Context("when the shared domain exists", func() {
			BeforeEach(func() {
				response := `{
						"metadata": {
							"guid": "shared-domain-guid",
							"updated_at": null
						},
						"entity": {
							"name": "shared-domain-1.com"
						}
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/shared_domains/shared-domain-guid"),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the shared domain and all warnings", func() {
				domain, warnings, err := client.GetSharedDomain("shared-domain-guid")
				Expect(err).NotTo(HaveOccurred())
				Expect(domain).To(Equal(Domain{Name: "shared-domain-1.com", GUID: "shared-domain-guid"}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
			})
		})

		Context("when the shared domain does not exist", func() {
			BeforeEach(func() {
				response := `{
					"code": 130002,
					"description": "The domain could not be found: shared-domain-guid",
					"error_code": "CF-DomainNotFound"
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/shared_domains/shared-domain-guid"),
						RespondWith(http.StatusNotFound, response),
					),
				)
			})

			It("returns an error", func() {
				domain, _, err := client.GetSharedDomain("shared-domain-guid")
				Expect(err).To(MatchError(cloudcontroller.ResourceNotFoundError{
					Message: "The domain could not be found: shared-domain-guid",
				}))
				Expect(domain).To(Equal(Domain{}))
			})
		})
	})

	Describe("GetPrivateDomain", func() {
		Context("when the private domain exists", func() {
			BeforeEach(func() {
				response := `{
						"metadata": {
							"guid": "private-domain-guid",
							"updated_at": null
						},
						"entity": {
							"name": "private-domain-1.com"
						}
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/private_domains/private-domain-guid"),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the private domain and all warnings", func() {
				domain, warnings, err := client.GetPrivateDomain("private-domain-guid")
				Expect(err).NotTo(HaveOccurred())
				Expect(domain).To(Equal(Domain{Name: "private-domain-1.com", GUID: "private-domain-guid"}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
			})
		})

		Context("when the private domain does not exist", func() {
			BeforeEach(func() {
				response := `{
					"code": 130002,
					"description": "The domain could not be found: private-domain-guid",
					"error_code": "CF-DomainNotFound"
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/private_domains/private-domain-guid"),
						RespondWith(http.StatusNotFound, response),
					),
				)
			})

			It("returns an error", func() {
				domain, _, err := client.GetPrivateDomain("private-domain-guid")
				Expect(err).To(MatchError(cloudcontroller.ResourceNotFoundError{
					Message: "The domain could not be found: private-domain-guid",
				}))
				Expect(domain).To(Equal(Domain{}))
			})
		})
	})
})
