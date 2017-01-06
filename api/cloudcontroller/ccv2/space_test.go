package ccv2_test

import (
	"net/http"

	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Space", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("GetSpaces", func() {
		Context("when no errors are encountered", func() {
			Context("when results are paginated", func() {
				BeforeEach(func() {
					response1 := `{
					"next_url": "/v2/spaces?q=some-query:some-value&page=2",
					"resources": [
						{
							"metadata": {
								"guid": "space-guid-1"
							},
							"entity": {
								"name": "space-1",
								"allow_ssh": false
							}
						},
						{
							"metadata": {
								"guid": "space-guid-2"
							},
							"entity": {
								"name": "space-2",
								"allow_ssh": true
							}
						}
					]
				}`
					response2 := `{
					"next_url": null,
					"resources": [
						{
							"metadata": {
								"guid": "space-guid-3"
							},
							"entity": {
								"name": "space-3",
								"allow_ssh": false
							}
						},
						{
							"metadata": {
								"guid": "space-guid-4"
							},
							"entity": {
								"name": "space-4",
								"allow_ssh": true
							}
						}
					]
				}`
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v2/spaces", "q=some-query:some-value"),
							RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"warning-1"}}),
						))
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v2/spaces", "q=some-query:some-value&page=2"),
							RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"warning-2"}}),
						))
				})

				It("returns paginated results and all warnings", func() {
					spaces, warnings, err := client.GetSpaces([]Query{{
						Filter:   "some-query",
						Operator: EqualOperator,
						Value:    "some-value",
					}})

					Expect(err).NotTo(HaveOccurred())
					Expect(spaces).To(Equal([]Space{
						{
							GUID:     "space-guid-1",
							Name:     "space-1",
							AllowSSH: false,
						},
						{
							GUID:     "space-guid-2",
							Name:     "space-2",
							AllowSSH: true,
						},
						{
							GUID:     "space-guid-3",
							Name:     "space-3",
							AllowSSH: false,
						},
						{
							GUID:     "space-guid-4",
							Name:     "space-4",
							AllowSSH: true,
						},
					}))
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				})
			})
		})

		Context("when an error is encountered", func() {
			BeforeEach(func() {
				response := `{
  "code": 10001,
  "description": "Some Error",
  "error_code": "CF-SomeError"
}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/spaces"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
					))
			})

			It("returns an error and all warnings", func() {
				_, warnings, err := client.GetSpaces(nil)

				Expect(err).To(MatchError(UnexpectedResponseError{
					ResponseCode: http.StatusTeapot,
					CCErrorResponse: CCErrorResponse{
						Code:        10001,
						Description: "Some Error",
						ErrorCode:   "CF-SomeError",
					},
				}))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})
	})
})
