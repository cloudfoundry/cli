package ccv2_test

import (
	"fmt"
	"net/http"
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Job", func() {
	var client *Client

	Describe("Job", func() {
		DescribeTable("Finished",
			func(status JobStatus, expected bool) {
				job := Job{Status: status}
				Expect(job.Finished()).To(Equal(expected))
			},

			Entry("when failed, it returns false", JobStatusFailed, false),
			Entry("when finished, it returns true", JobStatusFinished, true),
			Entry("when queued, it returns false", JobStatusQueued, false),
			Entry("when running, it returns false", JobStatusRunning, false),
		)

		DescribeTable("Failed",
			func(status JobStatus, expected bool) {
				job := Job{Status: status}
				Expect(job.Failed()).To(Equal(expected))
			},

			Entry("when failed, it returns true", JobStatusFailed, true),
			Entry("when finished, it returns false", JobStatusFinished, false),
			Entry("when queued, it returns false", JobStatusQueued, false),
			Entry("when running, it returns false", JobStatusRunning, false),
		)
	})

	Describe("PollJob", func() {
		BeforeEach(func() {
			client = NewTestClient(Config{JobPollingTimeout: time.Minute})
		})

		Context("when the job starts queued and then finishes successfully", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/jobs/some-job-guid"),
						RespondWith(http.StatusAccepted, `{
							"metadata": {
								"guid": "some-job-guid",
								"created_at": "2016-06-08T16:41:27Z",
								"url": "/v2/jobs/some-job-guid"
							},
							"entity": {
								"guid": "some-job-guid",
								"status": "queued"
							}
						}`, http.Header{"X-Cf-Warnings": {"warning-1"}}),
					))

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/jobs/some-job-guid"),
						RespondWith(http.StatusAccepted, `{
							"metadata": {
								"guid": "some-job-guid",
								"created_at": "2016-06-08T16:41:28Z",
								"url": "/v2/jobs/some-job-guid"
							},
							"entity": {
								"guid": "some-job-guid",
								"status": "running"
							}
						}`, http.Header{"X-Cf-Warnings": {"warning-2, warning-3"}}),
					))

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/jobs/some-job-guid"),
						RespondWith(http.StatusAccepted, `{
							"metadata": {
								"guid": "some-job-guid",
								"created_at": "2016-06-08T16:41:29Z",
								"url": "/v2/jobs/some-job-guid"
							},
							"entity": {
								"guid": "some-job-guid",
								"status": "finished"
							}
						}`, http.Header{"X-Cf-Warnings": {"warning-4"}}),
					))
			})

			It("should poll until completion", func() {
				warnings, err := client.PollJob(Job{GUID: "some-job-guid"})
				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("warning-1", "warning-2", "warning-3", "warning-4"))
			})
		})

		Context("when the job starts queued and then fails", func() {
			var jobFailureMessage string
			BeforeEach(func() {
				jobFailureMessage = "I am a banana!!!"

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/jobs/some-job-guid"),
						RespondWith(http.StatusAccepted, `{
							"metadata": {
								"guid": "some-job-guid",
								"created_at": "2016-06-08T16:41:27Z",
								"url": "/v2/jobs/some-job-guid"
							},
							"entity": {
								"guid": "some-job-guid",
								"status": "queued"
							}
						}`, http.Header{"X-Cf-Warnings": {"warning-1"}}),
					))

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/jobs/some-job-guid"),
						RespondWith(http.StatusAccepted, `{
							"metadata": {
								"guid": "some-job-guid",
								"created_at": "2016-06-08T16:41:28Z",
								"url": "/v2/jobs/some-job-guid"
							},
							"entity": {
								"guid": "some-job-guid",
								"status": "running"
							}
						}`, http.Header{"X-Cf-Warnings": {"warning-2, warning-3"}}),
					))

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/jobs/some-job-guid"),
						RespondWith(http.StatusAccepted, fmt.Sprintf(`{
							"metadata": {
								"guid": "some-job-guid",
								"created_at": "2016-06-08T16:41:29Z",
								"url": "/v2/jobs/some-job-guid"
							},
							"entity": {
								"error": "%s",
								"guid": "job-guid",
								"status": "failed"
							}
						}`, jobFailureMessage), http.Header{"X-Cf-Warnings": {"warning-4"}}),
					))
			})

			It("returns a JobFailedError", func() {
				warnings, err := client.PollJob(Job{GUID: "some-job-guid"})
				Expect(err).To(MatchError(JobFailedError{
					JobGUID: "some-job-guid",
					Message: jobFailureMessage,
				}))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2", "warning-3", "warning-4"))
			})
		})

		Context("when retrieving the job errors", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/jobs/some-job-guid"),
						RespondWith(http.StatusAccepted, `{
							INVALID YAML
						}`, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
					))
			})

			It("returns the CC error", func() {
				warnings, err := client.PollJob(Job{GUID: "some-job-guid"})
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				Expect(err.Error()).To(MatchRegexp("invalid character"))
			})
		})

		Describe("JobPollingTimeout", func() {
			Context("when the job runs longer than the OverallPollingTimeout", func() {
				var jobPollingTimeout time.Duration

				BeforeEach(func() {
					jobPollingTimeout = 100 * time.Millisecond
					client = NewTestClient(Config{
						JobPollingTimeout:  jobPollingTimeout,
						JobPollingInterval: 60 * time.Millisecond,
					})

					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v2/jobs/some-job-guid"),
							RespondWith(http.StatusAccepted, `{
							"metadata": {
								"guid": "some-job-guid",
								"created_at": "2016-06-08T16:41:27Z",
								"url": "/v2/jobs/some-job-guid"
							},
							"entity": {
								"guid": "some-job-guid",
								"status": "queued"
							}
						}`, http.Header{"X-Cf-Warnings": {"warning-1"}}),
						))

					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v2/jobs/some-job-guid"),
							RespondWith(http.StatusAccepted, `{
							"metadata": {
								"guid": "some-job-guid",
								"created_at": "2016-06-08T16:41:28Z",
								"url": "/v2/jobs/some-job-guid"
							},
							"entity": {
								"guid": "some-job-guid",
								"status": "running"
							}
						}`, http.Header{"X-Cf-Warnings": {"warning-2, warning-3"}}),
						))

					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v2/jobs/some-job-guid"),
							RespondWith(http.StatusAccepted, `{
							"metadata": {
								"guid": "some-job-guid",
								"created_at": "2016-06-08T16:41:29Z",
								"url": "/v2/jobs/some-job-guid"
							},
							"entity": {
								"guid": "some-job-guid",
								"status": "finished"
							}
						}`, http.Header{"X-Cf-Warnings": {"warning-4"}}),
						))
				})

				It("raises a JobTimeoutError", func() {
					_, err := client.PollJob(Job{GUID: "some-job-guid"})

					Expect(err).To(MatchError(JobTimeoutError{
						Timeout: jobPollingTimeout,
						JobGUID: "some-job-guid",
					}))
				})

				// Fuzzy test to ensure that the overall function time isn't [far]
				// greater than the OverallPollingTimeout. Since this is partially
				// dependant on the speed of the system, the expectation is that the
				// function *should* never exceed twice the timeout.
				It("does not run [too much] longer than the timeout", func() {
					startTime := time.Now()
					client.PollJob(Job{GUID: "some-job-guid"})
					endTime := time.Now()

					// If the jobPollingTimeout is less than the PollingInterval,
					// then the margin may be too small, we should install not allow the
					// jobPollingTimeout to be set to less than the PollingInterval
					Expect(endTime).To(BeTemporally("~", startTime, 2*jobPollingTimeout))
				})
			})
		})
	})

	Describe("GetJob", func() {
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
