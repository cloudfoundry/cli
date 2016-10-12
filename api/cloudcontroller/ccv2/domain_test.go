package ccv2_test

import (
	"net/http"

	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Domain", func() {
	var client *CloudControllerClient

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("GetSharedDomain", func() {
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
					VerifyRequest("GET", "/v2/shared_domains/shared-domain-guid"),
					RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
				),
			)
		})

		Context("when the shared domain exists", func() {
			It("returns the shared domain entity", func() {
				domain, warnings, err := client.GetSharedDomain("shared-domain-guid")
				Expect(err).NotTo(HaveOccurred())
				Expect(domain).To(Equal(Domain{Name: "shared-domain-1.com", GUID: "shared-domain-guid"}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
			})
		})
	})

	Describe("GetPrivateDomain", func() {
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
					VerifyRequest("GET", "/v2/private_domains/private-domain-guid"),
					RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
				),
			)
		})

		Context("when the private domain exists", func() {
			It("returns the private domain entity", func() {
				domain, warnings, err := client.GetPrivateDomain("private-domain-guid")
				Expect(err).NotTo(HaveOccurred())
				Expect(domain).To(Equal(Domain{Name: "private-domain-1.com", GUID: "private-domain-guid"}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
			})
		})
	})
})
