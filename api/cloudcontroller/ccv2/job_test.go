package ccv2_test

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Job", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("GetJob", func() {
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
						VerifyRequest(http.MethodGet, "/v2/jobs/job-guid"),
						RespondWith(http.StatusOK, jsonResponse, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
					))
			})

			It("returns job with all warnings", func() {
				job, warnings, err := client.GetJob("job-guid")

				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf(Warnings{"warning-1", "warning-2"}))
				Expect(job.GUID).To(Equal("job-guid"))
				Expect(job.Status).To(Equal(JobStatusQueued))
			})
		})

		Context("when the job fails", func() {
			BeforeEach(func() {
				jsonResponse := `{
					"metadata": {
						"guid": "job-guid",
						"created_at": "2016-06-08T16:41:27Z",
						"url": "/v2/jobs/job-guid"
					},
					"entity": {
						"error": "some-error",
						"guid": "job-guid",
						"status": "failed"
					}
				}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/jobs/job-guid"),
						RespondWith(http.StatusOK, jsonResponse, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
					))
			})

			It("returns job with all warnings", func() {
				job, warnings, err := client.GetJob("job-guid")

				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf(Warnings{"warning-1", "warning-2"}))
				Expect(job.GUID).To(Equal("job-guid"))
				Expect(job.Status).To(Equal(JobStatusFailed))
				Expect(job.Error).To(Equal("some-error"))
			})
		})
	})

	Describe("DeleteOrganization", func() {
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
						VerifyRequest(http.MethodDelete, "/v2/organizations/some-org-guid", "recursive=true&async=true"),
						RespondWith(http.StatusAccepted, jsonResponse, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
					))
			})

			It("deletes the org and returns all warnings", func() {
				job, warnings, err := client.DeleteOrganization("some-org-guid")

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
  "description": "The organization could not be found: some-org-guid",
  "error_code": "CF-OrganizationNotFound"
}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodDelete, "/v2/organizations/some-org-guid", "recursive=true&async=true"),
						RespondWith(http.StatusNotFound, response, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
					))
			})

			It("returns an error and all warnings", func() {
				_, warnings, err := client.DeleteOrganization("some-org-guid")

				Expect(err).To(MatchError(cloudcontroller.ResourceNotFoundError{
					Message: "The organization could not be found: some-org-guid",
				}))
				Expect(warnings).To(ConsistOf(Warnings{"warning-1", "warning-2"}))
			})
		})
	})
})
