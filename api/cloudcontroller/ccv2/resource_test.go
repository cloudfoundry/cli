package ccv2_test

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Resource", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("ResourceMatch", func() {
		Context("when resource matching is successful", func() {
			BeforeEach(func() {
				responseBody := `[
						{
							"fn":   "some-file-1",
							"mode": "744",
							"sha1": "some-sha-1",
							"size": 1
						},
						{
							"fn":   "some-file-3",
							"mode": "744",
							"sha1": "some-sha-3",
							"size": 3
						}
					]`

				// Note: ordering matters with VerifyBody
				expectedRequestBody := []map[string]interface{}{
					map[string]interface{}{
						"fn":   "some-file-1",
						"mode": "744",
						"sha1": "some-sha-1",
						"size": 1,
					},
					map[string]interface{}{
						"fn":   "some-file-2",
						"mode": "744",
						"sha1": "some-sha-2",
						"size": 2,
					},
					map[string]interface{}{
						"fn":   "some-file-3",
						"mode": "744",
						"sha1": "some-sha-3",
						"size": 3,
					},
				}

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPut, "/v2/resource_match"),
						VerifyJSONRepresenting(expectedRequestBody),
						RespondWith(http.StatusCreated, responseBody, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the resources and warnings", func() {
				resourcesToMatch := []Resource{
					{
						Filename: "some-file-1",
						Mode:     0744,
						SHA1:     "some-sha-1",
						Size:     1,
					},
					{
						Filename: "some-file-2",
						Mode:     0744,
						SHA1:     "some-sha-2",
						Size:     2,
					},
					{
						Filename: "some-file-3",
						Mode:     0744,
						SHA1:     "some-sha-3",
						Size:     3,
					},
				}
				matchedResources, warnings, err := client.ResourceMatch(resourcesToMatch)

				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))

				Expect(matchedResources).To(ConsistOf(
					Resource{
						Filename: "some-file-1",
						Mode:     0744,
						SHA1:     "some-sha-1",
						Size:     1,
					},
					Resource{
						Filename: "some-file-3",
						Mode:     0744,
						SHA1:     "some-sha-3",
						Size:     3,
					}))
			})
		})

		Context("when the cc returns an error", func() {
			BeforeEach(func() {
				response := `{
					"code": 10001,
					"description": "Some Error",
					"error_code": "CF-SomeError"
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPut, "/v2/resource_match"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns an error", func() {
				_, warnings, err := client.ResourceMatch(nil)
				Expect(err).To(MatchError(ccerror.V2UnexpectedResponseError{
					ResponseCode: http.StatusTeapot,
					V2ErrorResponse: ccerror.V2ErrorResponse{
						Code:        10001,
						Description: "Some Error",
						ErrorCode:   "CF-SomeError",
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})
})
