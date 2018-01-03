package ccv2_test

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Stack", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("GetStack", func() {
		Context("when the stack is found", func() {
			BeforeEach(func() {
				response := `{
					"metadata": {
						"guid": "some-stack-guid"
					},
					"entity": {
						"name": "some-stack-name",
						"description": "some stack description"
					}
				}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/stacks/some-stack-guid"),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the stack and warnings", func() {
				stack, warnings, err := client.GetStack("some-stack-guid")
				Expect(err).ToNot(HaveOccurred())
				Expect(stack).To(Equal(Stack{
					Description: "some stack description",
					GUID:        "some-stack-guid",
					Name:        "some-stack-name",
				}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
			})
		})

		Context("when the client returns an error", func() {
			BeforeEach(func() {
				response := `{
					"code": 250003,
					"description": "The stack could not be found: some-stack-guid",
					"error_code": "CF-StackNotFound"
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/stacks/some-stack-guid"),
						RespondWith(http.StatusNotFound, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error and warnings", func() {
				_, warnings, err := client.GetStack("some-stack-guid")
				Expect(err).To(MatchError(ccerror.ResourceNotFoundError{
					Message: "The stack could not be found: some-stack-guid",
				}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
			})
		})
	})

	Describe("GetStacks", func() {
		Context("when no errors are encountered", func() {
			Context("when results are paginated", func() {
				BeforeEach(func() {
					response1 := `{
						"next_url": "/v2/stacks?q=some-query:some-value&page=2",
						"resources": [
							{
								"metadata": {
									"guid": "some-stack-guid-1"
								},
								"entity": {
									"name": "some-stack-name-1",
									"description": "some stack description"
								}
							},
							{
								"metadata": {
									"guid": "some-stack-guid-2"
								},
								"entity": {
									"name": "some-stack-name-2",
									"description": "some stack description"
								}
							}
						]
					}`
					response2 := `{
						"next_url": null,
						"resources": [
							{
								"metadata": {
									"guid": "some-stack-guid-3"
								},
								"entity": {
									"name": "some-stack-name-3",
									"description": "some stack description"
								}
							},
							{
								"metadata": {
									"guid": "some-stack-guid-4"
								},
								"entity": {
									"name": "some-stack-name-4",
									"description": "some stack description"
								}
							}
						]
					}`
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v2/stacks", "q=some-query:some-value"),
							RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"warning-1"}}),
						))
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v2/stacks", "q=some-query:some-value&page=2"),
							RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"warning-2"}}),
						))
				})

				It("returns paginated results and all warnings", func() {
					stacks, warnings, err := client.GetStacks(QQuery{
						Filter:   "some-query",
						Operator: EqualOperator,
						Values:   []string{"some-value"},
					})

					Expect(err).NotTo(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
					Expect(stacks).To(Equal([]Stack{
						{
							Description: "some stack description",
							GUID:        "some-stack-guid-1",
							Name:        "some-stack-name-1",
						},
						{
							Description: "some stack description",
							GUID:        "some-stack-guid-2",
							Name:        "some-stack-name-2",
						},
						{
							Description: "some stack description",
							GUID:        "some-stack-guid-3",
							Name:        "some-stack-name-3",
						},
						{
							Description: "some stack description",
							GUID:        "some-stack-guid-4",
							Name:        "some-stack-name-4",
						},
					}))
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
						VerifyRequest(http.MethodGet, "/v2/stacks"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
					))
			})

			It("returns an error and all warnings", func() {
				_, warnings, err := client.GetStacks()

				Expect(err).To(MatchError(ccerror.V2UnexpectedResponseError{
					ResponseCode: http.StatusTeapot,
					V2ErrorResponse: ccerror.V2ErrorResponse{
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
