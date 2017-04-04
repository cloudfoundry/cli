package ccv2_test

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Space Quotas", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("GetSpaceQuota", func() {
		Context("when no errors are encountered", func() {
			BeforeEach(func() {
				response := `{
					"metadata": {
						"guid": "space-quota-guid",
						"url": "/v2/space_quota_definitions/space-quota-guid",
						"updated_at": null
					},
					"entity": {
						"name": "space-quota"
					}
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/space_quota_definitions/space-quota-guid"),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the Space Quota", func() {
				spaceQuota, warnings, err := client.GetSpaceQuota("space-quota-guid")
				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
				Expect(spaceQuota).To(Equal(SpaceQuota{
					Name: "space-quota",
					GUID: "space-quota-guid",
				}))
			})
		})

		Context("when the request returns an error", func() {
			BeforeEach(func() {
				response := `{
					"code": 210002,
					"description": "The space quota could not be found: some-space-quota-guid",
					"error_code": "CF-AppNotFound"
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/space_quota_definitions/some-space-quota-guid"),
						RespondWith(http.StatusNotFound, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error and warnings", func() {
				_, warnings, err := client.GetSpaceQuota("some-space-quota-guid")
				Expect(err).To(MatchError(ccerror.ResourceNotFoundError{Message: "The space quota could not be found: some-space-quota-guid"}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
			})
		})
	})
})
