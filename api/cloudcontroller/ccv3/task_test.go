package ccv3_test

import (
	"net/http"

	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Task", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("RunTask", func() {
		Context("when the task was successfully created", func() {
			BeforeEach(func() {
				response := `{
  "sequence_id": 3
}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v3/apps/some-app-guid/tasks", ""),
						RespondWith(http.StatusAccepted, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the task and warnings", func() {
				task, warnings, err := client.RunTask("some-app-guid", "some command")
				Expect(err).ToNot(HaveOccurred())
				Expect(task).To(Equal(Task{SequenceID: 3}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})

		Context("when the API returns an error and warnings", func() {
			BeforeEach(func() {
				response := `{
  "errors": [
    {
      "code": 10008,
      "detail": "The request is semantically invalid: command presence",
      "title": "CF-UnprocessableEntity"
    }
  ]
}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v3/apps/some-app-guid/tasks", ""),
						RespondWith(http.StatusUnprocessableEntity, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the errors and warnings", func() {
				expectedErr := UnexpectedResponseError{
					ResponseCode: 422,
					CCErrorResponse: CCErrorResponse{[]CCError{{
						Code:   10008,
						Detail: "The request is semantically invalid: command presence",
						Title:  "CF-UnprocessableEntity",
					}}},
				}
				_, warnings, err := client.RunTask("some-app-guid", "some command")
				Expect(err).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})
})
