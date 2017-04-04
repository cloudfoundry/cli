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
})
