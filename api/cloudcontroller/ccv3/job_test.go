package ccv3_test

import (
	"fmt"
	"net/http"
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/ccv3fakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Job", func() {
	var client *Client

	Describe("Job", func() {
		DescribeTable("IsComplete",
			func(status constant.JobState, expected bool) {
				job := Job{State: status}
				Expect(job.IsComplete()).To(Equal(expected))
			},

			Entry("when failed, it returns false", constant.JobFailed, false),
			Entry("when complete, it returns true", constant.JobComplete, true),
			Entry("when processing, it returns false", constant.JobProcessing, false),
		)

		DescribeTable("HasFailed",
			func(status constant.JobState, expected bool) {
				job := Job{State: status}
				Expect(job.HasFailed()).To(Equal(expected))
			},

			Entry("when failed, it returns true", constant.JobFailed, true),
			Entry("when complete, it returns false", constant.JobComplete, false),
			Entry("when processing, it returns false", constant.JobProcessing, false),
		)
	})

	Describe("GetJob", func() {
		var (
			jobLocation JobURL

			job        Job
			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			client, _ = NewTestClient()
			jobLocation = JobURL(fmt.Sprintf("%s/some-job-location", server.URL()))
		})

		JustBeforeEach(func() {
			job, warnings, executeErr = client.GetJob(jobLocation)
		})

		When("no errors are encountered", func() {
			BeforeEach(func() {
				jsonResponse := `{
						"guid": "job-guid",
						"created_at": "2016-06-08T16:41:27Z",
						"updated_at": "2016-06-08T16:41:27Z",
						"operation": "app.delete",
						"state": "PROCESSING",
						"links": {
							"self": {
								"href": "/v3/jobs/job-guid"
							}
						}
					}
				}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/some-job-location"),
						RespondWith(http.StatusOK, jsonResponse, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
					))
			})

			It("returns job with all warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf(Warnings{"warning-1", "warning-2"}))
				Expect(job.GUID).To(Equal("job-guid"))
				Expect(job.State).To(Equal(constant.JobProcessing))
			})
		})

		When("the job fails", func() {
			BeforeEach(func() {
				jsonResponse := `{
						"guid": "job-guid",
						"created_at": "2016-06-08T16:41:27Z",
						"updated_at": "2016-06-08T16:41:27Z",
						"operation": "delete",
						"state": "FAILED",
						"errors": [
							{
								"detail": "blah blah",
								"title": "CF-JobFail",
								"code": 1234
							}
						],
						"links": {
							"self": {
								"href": "/v3/jobs/job-guid"
							}
						}
					}
				}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/some-job-location"),
						RespondWith(http.StatusOK, jsonResponse, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
					))
			})

			It("returns job with all warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf(Warnings{"warning-1", "warning-2"}))
				Expect(job.GUID).To(Equal("job-guid"))
				Expect(job.State).To(Equal(constant.JobFailed))
				Expect(job.Errors[0].Detail).To(Equal("blah blah"))
				Expect(job.Errors[0].Title).To(Equal("CF-JobFail"))
				Expect(job.Errors[0].Code).To(BeEquivalentTo(1234))
			})
		})
	})

	Describe("PollJob", func() {
		var (
			jobLocation JobURL

			warnings   Warnings
			executeErr error

			startTime time.Time
		)

		BeforeEach(func() {
			client, _ = NewTestClient(Config{JobPollingTimeout: time.Minute})
			jobLocation = JobURL(fmt.Sprintf("%s/some-job-location", server.URL()))
		})

		JustBeforeEach(func() {
			startTime = time.Now()
			warnings, executeErr = client.PollJob(jobLocation)
		})

		When("the job starts queued and then finishes successfully", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/some-job-location"),
						RespondWith(http.StatusAccepted, `{
							"guid": "job-guid",
							"created_at": "2016-06-08T16:41:27Z",
							"updated_at": "2016-06-08T16:41:27Z",
							"operation": "app.delete",
							"state": "PROCESSING",
							"links": {
								"self": {
									"href": "/v3/jobs/job-guid"
								}
							}
						}`, http.Header{"X-Cf-Warnings": {"warning-1"}}),
					))

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/some-job-location"),
						RespondWith(http.StatusAccepted, `{
							"guid": "job-guid",
							"created_at": "2016-06-08T16:41:27Z",
							"updated_at": "2016-06-08T16:41:27Z",
							"operation": "app.delete",
							"state": "PROCESSING",
							"links": {
								"self": {
									"href": "/v3/jobs/job-guid"
								}
							}
						}`, http.Header{"X-Cf-Warnings": {"warning-2"}}),
					))

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/some-job-location"),
						RespondWith(http.StatusAccepted, `{
							"guid": "job-guid",
							"created_at": "2016-06-08T16:41:27Z",
							"updated_at": "2016-06-08T16:41:27Z",
							"operation": "app.delete",
							"state": "COMPLETE",
							"links": {
								"self": {
									"href": "/v3/jobs/job-guid"
								}
							}
						}`, http.Header{"X-Cf-Warnings": {"warning-3, warning-4"}}),
					))
			})

			It("should poll until completion", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("warning-1", "warning-2", "warning-3", "warning-4"))
			})
		})

		When("the job starts queued and then fails", func() {
			var jobFailureMessage string

			BeforeEach(func() {
				jobFailureMessage = "I am a banana!!!"

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/some-job-location"),
						RespondWith(http.StatusAccepted, `{
							"guid": "job-guid",
							"created_at": "2016-06-08T16:41:27Z",
							"updated_at": "2016-06-08T16:41:27Z",
							"operation": "app.delete",
							"state": "PROCESSING",
							"links": {
								"self": {
									"href": "/v3/jobs/job-guid"
								}
							}
						}`, http.Header{"X-Cf-Warnings": {"warning-1"}}),
					))

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/some-job-location"),
						RespondWith(http.StatusAccepted, `{
							"guid": "job-guid",
							"created_at": "2016-06-08T16:41:27Z",
							"updated_at": "2016-06-08T16:41:27Z",
							"operation": "app.delete",
							"state": "PROCESSING",
							"links": {
								"self": {
									"href": "/v3/jobs/job-guid"
								}
							}
						}`, http.Header{"X-Cf-Warnings": {"warning-2, warning-3"}}),
					))

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/some-job-location"),
						RespondWith(http.StatusOK, fmt.Sprintf(`{
							"guid": "job-guid",
							"created_at": "2016-06-08T16:41:27Z",
							"updated_at": "2016-06-08T16:41:27Z",
							"operation": "app.delete",
							"state": "FAILED",
							"errors": [ {
								"detail": "%s",
								"title": "CF-AppBitsUploadInvalid",
								"code": 160001
							} ],
							"links": {
								"self": {
									"href": "/v3/jobs/job-guid"
								}
							}
						}`, jobFailureMessage), http.Header{"X-Cf-Warnings": {"warning-4"}}),
					))
			})

			It("returns a JobFailedError", func() {
				Expect(executeErr).To(MatchError(ccerror.V3JobFailedError{
					JobGUID: "job-guid",
					Code:    160001,
					Detail:  jobFailureMessage,
					Title:   "CF-AppBitsUploadInvalid",
				}))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2", "warning-3", "warning-4"))
			})
		})

		Context("polling timeouts", func() {
			When("the job runs longer than the OverallPollingTimeout", func() {
				var (
					jobPollingTimeout time.Duration
					fakeClock         *ccv3fakes.FakeClock
				)

				BeforeEach(func() {
					jobPollingTimeout = 100 * time.Millisecond
					client, fakeClock = NewTestClient(Config{
						JobPollingTimeout: jobPollingTimeout,
					})

					clockTime := time.Now()
					fakeClock.NowReturnsOnCall(0, clockTime)
					fakeClock.NowReturnsOnCall(1, clockTime)
					fakeClock.NowReturnsOnCall(2, clockTime.Add(60*time.Millisecond))
					fakeClock.NowReturnsOnCall(3, clockTime.Add(60*time.Millisecond*2))

					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/some-job-location"),
							RespondWith(http.StatusAccepted, `{
							"guid": "job-guid",
							"created_at": "2016-06-08T16:41:27Z",
							"updated_at": "2016-06-08T16:41:27Z",
							"operation": "app.delete",
							"state": "PROCESSING",
							"links": {
								"self": {
									"href": "/v3/jobs/job-guid"
								}
							}
						}`, http.Header{"X-Cf-Warnings": {"warning-1"}}),
						))

					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/some-job-location"),
							RespondWith(http.StatusAccepted, `{
							"guid": "job-guid",
							"created_at": "2016-06-08T16:41:27Z",
							"updated_at": "2016-06-08T16:41:27Z",
							"operation": "app.delete",
							"state": "PROCESSING",
							"links": {
								"self": {
									"href": "/v3/jobs/job-guid"
								}
							}
						}`, http.Header{"X-Cf-Warnings": {"warning-2, warning-3"}}),
						))

					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/some-job-location"),
							RespondWith(http.StatusAccepted, `{
							"guid": "job-guid",
							"created_at": "2016-06-08T16:41:27Z",
							"updated_at": "2016-06-08T16:41:27Z",
							"operation": "app.delete",
							"state": "FINISHED",
							"links": {
								"self": {
									"href": "/v3/jobs/job-guid"
								}
							}
						}`, http.Header{"X-Cf-Warnings": {"warning-4"}}),
						))
				})

				It("raises a JobTimeoutError", func() {
					Expect(executeErr).To(MatchError(ccerror.JobTimeoutError{
						Timeout: jobPollingTimeout,
						JobGUID: "job-guid",
					}))
					Expect(warnings).To(ConsistOf("warning-1", "warning-2", "warning-3"))
				})

				// Fuzzy test to ensure that the overall function time isn't [far]
				// greater than the OverallPollingTimeout. Since this is partially
				// dependent on the speed of the system, the expectation is that the
				// function *should* never exceed three times the timeout.
				It("does not run [too much] longer than the timeout", func() {
					endTime := time.Now()
					Expect(executeErr).To(HaveOccurred())

					// If the jobPollingTimeout is less than the PollingInterval,
					// then the margin may be too small, we should not allow the
					// jobPollingTimeout to be set to less than the PollingInterval
					Expect(endTime).To(BeTemporally("~", startTime, 3*jobPollingTimeout))
				})
			})
		})
	})
})
