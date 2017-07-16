// generated from codetemplates/delete_async_by_guid_test.go.template

package ccv2_test

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("DeleteSpace", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestClient()
	})

	Context("when no errors are encountered", func() {
		BeforeEach(func() {
			jsonResponse := `{
				"metadata": {
					"guid": "job-guid",
					"created_at": "2016-06-08T16:41:27Z",
					"url": "/v2/jobs/job-guid"
				},
				"entity": {
					"guid": "job-guid",
					"status": "queued"
				}
			}`

			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodDelete, "/v2/spaces/some-space-guid", "recursive=true&async=true"),
					RespondWith(http.StatusAccepted, jsonResponse, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
				))
		})

		It("deletes the Space and returns all warnings", func() {
			job, warnings, err := client.DeleteSpace("some-space-guid")

			Expect(err).NotTo(HaveOccurred())
			Expect(warnings).To(ConsistOf(Warnings{"warning-1", "warning-2"}))
			Expect(job.GUID).To(Equal("job-guid"))
			Expect(job.Status).To(Equal(JobStatusQueued))
		})
	})

	Context("when an error is encountered", func() {
		BeforeEach(func() {
			response := `{
"code": 30003,
"description": "The Space could not be found: some-space-guid",
"error_code": "CF-SpaceNotFound"
}`
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodDelete, "/v2/spaces/some-space-guid", "recursive=true&async=true"),
					RespondWith(http.StatusNotFound, response, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
				))
		})

		It("returns an error and all warnings", func() {
			_, warnings, err := client.DeleteSpace("some-space-guid")

			Expect(err).To(MatchError(ccerror.ResourceNotFoundError{
				Message: "The Space could not be found: some-space-guid",
			}))
			Expect(warnings).To(ConsistOf(Warnings{"warning-1", "warning-2"}))
		})
	})
})
