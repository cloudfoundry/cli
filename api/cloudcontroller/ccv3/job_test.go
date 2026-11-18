package ccv3_test

import (
	"fmt"
	"net/http"
	"time"

	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccv3/ccv3fakes"
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccv3/constant"

	. "github.com/onsi/ginkgo/v2"
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

		Describe("IsAt", func() {
			It("returns true when status matches", func() {
				job := Job{State: constant.JobComplete}
				Expect(job.IsAt(constant.JobComplete)).To(BeTrue())
				job = Job{State: constant.JobFailed}
				Expect(job.IsAt(constant.JobFailed)).To(BeTrue())
				job = Job{State: constant.JobProcessing}
				Expect(job.IsAt(constant.JobProcessing)).To(BeTrue())
				job = Job{State: constant.JobPolling}
				Expect(job.IsAt(constant.JobPolling)).To(BeTrue())
			})
			It("returns false when status does not match", func() {
				job := Job{State: constant.JobComplete}
				Expect(job.IsAt(constant.JobFailed)).To(BeFalse())
				Expect(job.IsAt(constant.JobProcessing)).To(BeFalse())
				Expect(job.IsAt(constant.JobPolling)).To(BeFalse())
				job = Job{State: constant.JobPolling}
				Expect(job.IsAt(constant.JobFailed)).To(BeFalse())
				Expect(job.IsAt(constant.JobProcessing)).To(BeFalse())
				Expect(job.IsAt(constant.JobComplete)).To(BeFalse())
			})
		})

		DescribeTable("Errors converts JobErrorDetails",
			func(code int, expectedErrType error) {
				rawErr := JobErrorDetails{
					Code:   constant.JobErrorCode(code),
					Detail: fmt.Sprintf("code %d", code),
					Title:  "some-err-title",
				}

				job := Job{
					GUID:      "some-job-guid",
					RawErrors: []JobErrorDetails{rawErr},
				}

				Expect(job.Errors()).To(HaveLen(1))
				Expect(job.Errors()[0]).To(MatchError(expectedErrType))
			},

			Entry("BuildpackNameStackTaken", 290000, ccerror.BuildpackAlreadyExistsForStackError{Message: "code 290000"}),
			Entry("BuildpackInvalid", 290003, ccerror.BuildpackInvalidError{Message: "code 290003"}),
			Entry("BuildpackStacksDontMatch", 390011, ccerror.BuildpackStacksDontMatchError{Message: "code 390011"}),
			Entry("BuildpackStackDoesNotExist", 390012, ccerror.BuildpackStackDoesNotExistError{Message: "code 390012"}),
			Entry("BuildpackZipError", 390013, ccerror.BuildpackZipInvalidError{Message: "code 390013"}),
			Entry("V3JobFailedError", 1111111, ccerror.V3JobFailedError{JobGUID: "some-job-guid", Code: constant.JobErrorCode(1111111), Detail: "code 1111111", Title: "some-err-title"}),
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
						"warnings": [{"detail": "a warning"}, {"detail": "another warning"}],
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
				Expect(warnings).To(ConsistOf(Warnings{"warning-1", "warning-2", "a warning", "another warning"}))
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
				Expect(job.RawErrors[0].Detail).To(Equal("blah blah"))
				Expect(job.RawErrors[0].Title).To(Equal("CF-JobFail"))
				Expect(job.RawErrors[0].Code).To(BeEquivalentTo(1234))
			})
		})
	})

	When("Polling jobs", func() {
		var (
			jobLocation JobURL

			warnings   Warnings
			executeErr error

			jobPollingTimeout time.Duration
			fakeClock         *ccv3fakes.FakeClock
			pollingSleeps     []time.Duration
			fakeNow           time.Time
			fakeGETDuration   time.Duration
		)

		useFakeClock := func() {
			fakeNow = time.Time{}
			fakeGETDuration = 0
			pollingSleeps = nil
			fakeClock.NowCalls(func() time.Time { return fakeNow })
			fakeClock.SleepCalls(func(duration time.Duration) {
				pollingSleeps = append(pollingSleeps, duration)
				fakeNow = fakeNow.Add(duration)
			})
		}

		setInterval := func(interval time.Duration) {
			client, fakeClock = NewTestClient(Config{
				JobPollingTimeout:  time.Minute,
				JobPollingInterval: interval,
			})
			useFakeClock()
		}

		appendHandler := func(state string, warnings Warnings) {
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/some-job-location"),
					http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
						fakeNow = fakeNow.Add(fakeGETDuration)
					}),
					RespondWith(http.StatusAccepted, fmt.Sprintf(`
						{
							"guid": "job-guid",
							"created_at": "2016-06-08T16:41:27Z",
							"updated_at": "2016-06-08T16:41:27Z",
							"operation": "app.delete",
							"state": "%s",
							"links": {
								"self": {
									"href": "/v3/jobs/job-guid"
								}
							}
						}`, state), http.Header{"X-Cf-Warnings": warnings}),
				),
			)
		}

		appendFailureHandler := func(message string, code constant.JobErrorCode, warnings Warnings) {
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
								"code": %d
							} ],
							"links": {
								"self": {
									"href": "/v3/jobs/job-guid"
								}
							}
						}`, message, code), http.Header{"X-Cf-Warnings": warnings}),
				),
			)
		}

		itSkipsEmptyURLs := func() {
			When("the job URL is empty", func() {
				BeforeEach(func() {
					jobLocation = ""
				})

				It("immediately succeeds", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(BeEmpty())
				})
			})
		}

		itFinishesWhenCompleteOrFailed := func() {
			When("the job starts queued and then finishes successfully", func() {
				BeforeEach(func() {
					setInterval(3 * time.Second)
					appendHandler("PROCESSING", Warnings{"warning-1"})
					appendHandler("PROCESSING", Warnings{"warning-2"})
					appendHandler("COMPLETE", Warnings{"warning-3", "warning-4"})
				})

				It("should poll until completion", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning-1", "warning-2", "warning-3", "warning-4"))
				})
			})

			When("the job starts queued and then fails", func() {
				BeforeEach(func() {
					appendHandler("PROCESSING", Warnings{"warning-1"})
					appendHandler("PROCESSING", Warnings{"warning-2", "warning-3"})
				})
				Context("with an error", func() {
					BeforeEach(func() {
						appendFailureHandler("some-message", constant.JobErrorCodeBuildpackAlreadyExistsForStack, Warnings{"warning-4"})
					})
					It("returns the first error", func() {
						Expect(executeErr).To(MatchError(ccerror.BuildpackAlreadyExistsForStackError{
							Message: "some-message",
						}))
						Expect(warnings).To(ConsistOf("warning-1", "warning-2", "warning-3", "warning-4"))
					})
				})
				Context("without an error", func() {
					BeforeEach(func() {
						server.AppendHandlers(
							CombineHandlers(
								VerifyRequest(http.MethodGet, "/some-job-location"),
								RespondWith(http.StatusOK, `{
							"guid": "job-guid",
							"created_at": "2016-06-08T16:41:27Z",
							"updated_at": "2016-06-08T16:41:27Z",
							"operation": "app.delete",
							"state": "FAILED",
							"errors": null,
							"links": {
								"self": {
									"href": "/v3/jobs/job-guid"
								}
							}
						}`, http.Header{"X-Cf-Warnings": []string{"warning-4"}}),
							),
						)
					})

					It("returns the JobFailedNoErrorError", func() {
						Expect(executeErr).To(MatchError(ccerror.JobFailedNoErrorError{
							JobGUID: "job-guid",
						}))
						Expect(warnings).To(ConsistOf("warning-1", "warning-2", "warning-3", "warning-4"))
					})
				})
			})

		}

		itRespectsTimeouts := func() {
			Context("polling timeouts", func() {
				When("the job runs longer than the OverallPollingTimeout", func() {
					BeforeEach(func() {
						jobPollingTimeout = 100 * time.Millisecond
						client, fakeClock = NewTestClient(Config{
							JobPollingTimeout:  jobPollingTimeout,
							JobPollingInterval: 60 * time.Millisecond,
						})
						useFakeClock()

						appendHandler("PROCESSING", Warnings{"warning-1"})
						appendHandler("PROCESSING", Warnings{"warning-2", "warning-3"})
					})

					It("raises a JobTimeoutError", func() {
						Expect(executeErr).To(MatchError(ccerror.JobTimeoutError{
							Timeout: jobPollingTimeout,
							JobGUID: "job-guid",
						}))
						Expect(warnings).To(ConsistOf("warning-1", "warning-2", "warning-3"))
					})

					It("checks the timeout before another GET", func() {
						Expect(server.ReceivedRequests()).To(HaveLen(2))
						Expect(pollingSleeps).To(Equal([]time.Duration{60 * time.Millisecond, 60 * time.Millisecond}))
					})
				})
			})
		}

		BeforeEach(func() {
			setInterval(3 * time.Second)
			jobLocation = JobURL(fmt.Sprintf("%s/some-job-location", server.URL()))
		})

		Describe("PollJob", func() {
			JustBeforeEach(func() {
				warnings, executeErr = client.PollJob(jobLocation)
			})

			itSkipsEmptyURLs()
			itFinishesWhenCompleteOrFailed()
			itRespectsTimeouts()

			Context("polling delays", func() {
				When("using the default interval", func() {
					BeforeEach(func() {
						setInterval(3 * time.Second)
						for range 6 {
							appendHandler("PROCESSING", nil)
						}
						appendHandler("COMPLETE", nil)
					})

					It("starts at 250ms, doubles to the 3s interval, then stays there", func() {
						Expect(executeErr).NotTo(HaveOccurred())
						Expect(server.ReceivedRequests()).To(HaveLen(7))
						Expect(pollingSleeps).To(Equal([]time.Duration{
							250 * time.Millisecond, 500 * time.Millisecond, time.Second,
							1250 * time.Millisecond, 3 * time.Second, 3 * time.Second,
						}))
					})
				})

				When("each GET takes 100ms", func() {
					BeforeEach(func() {
						setInterval(3 * time.Second)
						fakeGETDuration = 100 * time.Millisecond
						for range 5 {
							appendHandler("PROCESSING", nil)
						}
						appendHandler("COMPLETE", nil)
					})

					It("sleeps 150ms, 400ms, 900ms, 1.15s, and 2.9s to keep checks on the grid", func() {
						Expect(executeErr).NotTo(HaveOccurred())
						Expect(pollingSleeps).To(Equal([]time.Duration{
							150 * time.Millisecond, 400 * time.Millisecond, 900 * time.Millisecond,
							1150 * time.Millisecond, 2900 * time.Millisecond,
						}))
					})
				})

				When("the interval is 1s", func() {
					BeforeEach(func() {
						setInterval(time.Second)
						for range 4 {
							appendHandler("PROCESSING", nil)
						}
						appendHandler("COMPLETE", nil)
					})

					It("caps the delay at the interval", func() {
						Expect(executeErr).NotTo(HaveOccurred())
						Expect(pollingSleeps).To(Equal([]time.Duration{250 * time.Millisecond, 500 * time.Millisecond, 250 * time.Millisecond, time.Second}))
					})
				})

				When("the interval is below the initial delay", func() {
					BeforeEach(func() {
						setInterval(100 * time.Millisecond)
						appendHandler("PROCESSING", nil)
						appendHandler("PROCESSING", nil)
						appendHandler("COMPLETE", nil)
					})

					It("uses the interval for every delay", func() {
						Expect(executeErr).NotTo(HaveOccurred())
						Expect(pollingSleeps).To(Equal([]time.Duration{100 * time.Millisecond, 100 * time.Millisecond}))
					})
				})

				When("the job completes on the first GET", func() {
					BeforeEach(func() {
						setInterval(3 * time.Second)
						appendHandler("COMPLETE", nil)
					})

					It("does not sleep", func() {
						Expect(executeErr).NotTo(HaveOccurred())
						Expect(server.ReceivedRequests()).To(HaveLen(1))
						Expect(pollingSleeps).To(BeEmpty())
					})
				})

				When("PollJob is called a second time on the same client", func() {
					BeforeEach(func() {
						setInterval(3 * time.Second)
						appendHandler("PROCESSING", nil)
						appendHandler("COMPLETE", nil)
					})

					It("starts the second call at 250ms", func() {
						appendHandler("PROCESSING", nil)
						appendHandler("COMPLETE", nil)
						Expect(executeErr).NotTo(HaveOccurred())
						_, secondErr := client.PollJob(jobLocation)
						Expect(secondErr).NotTo(HaveOccurred())
						Expect(pollingSleeps).To(Equal([]time.Duration{
							250 * time.Millisecond, 250 * time.Millisecond,
						}))
					})
				})
			})
		})

		DescribeTable("PollJob accounts for elapsed GET time and the next polling grid point",
			func(elapsed, interval, expected time.Duration) {
				setInterval(interval)
				fakeGETDuration = elapsed
				appendHandler("PROCESSING", nil)
				appendHandler("COMPLETE", nil)

				_, executeErr = client.PollJob(jobLocation)
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(pollingSleeps).To(Equal([]time.Duration{expected}))
			},

			Entry("when elapsed time is 1s and interval is 3s", time.Second, 3*time.Second, 750*time.Millisecond),
			Entry("when elapsed time is 8s and interval is 3s", 8*time.Second, 3*time.Second, time.Second),
			Entry("when elapsed time is 1.75s and interval is 10s", 1750*time.Millisecond, 10*time.Second, 2*time.Second),
			Entry("when elapsed time is exactly on the 3s grid point", 3*time.Second, 3*time.Second, 3*time.Second),
			Entry("when elapsed time and interval are zero", time.Duration(0), time.Duration(0), time.Duration(0)),
			Entry("when the interval is negative", time.Duration(0), -time.Second, -time.Second),
			Entry("when the interval is below the initial delay", time.Duration(0), 100*time.Millisecond, 100*time.Millisecond),
		)

		Describe("PollJobForState", func() {
			JustBeforeEach(func() {
				warnings, executeErr = client.PollJobForState(jobLocation, constant.JobPolling)
			})

			itSkipsEmptyURLs()
			itFinishesWhenCompleteOrFailed()
			itRespectsTimeouts()

			When("Job reaches the required state", func() {
				BeforeEach(func() {
					appendHandler("PROCESSING", Warnings{"warning-1"})
					appendHandler("PROCESSING", Warnings{"warning-2"})
					appendHandler("POLLING", Warnings{"warning-3"})
				})

				It("should poll until the desired state", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning-1", "warning-2", "warning-3"))
				})
			})
		})

		Describe("PollJobToEventStream", func() {
			var stream chan PollJobEvent

			JustBeforeEach(func() {
				stream = client.PollJobToEventStream(jobLocation)
			})

			When("the job URL is empty", func() {
				BeforeEach(func() {
					jobLocation = ""
				})

				It("closes the channel without events", func() {
					Consistently(stream).ShouldNot(Receive())
					Eventually(stream).Should(BeClosed())
				})
			})

			When("the job starts queued and then finishes successfully", func() {
				BeforeEach(func() {
					setInterval(3 * time.Second)
					appendHandler("PROCESSING", Warnings{"warning-1"})
					appendHandler("PROCESSING", Warnings{"warning-2"})
					appendHandler("PROCESSING", nil)
					appendHandler("PROCESSING", nil)
					appendHandler("PROCESSING", nil)
					appendHandler("PROCESSING", nil)
					appendHandler("COMPLETE", Warnings{"warning-3", "warning-4"})
				})

				It("closes the stream after the terminal event without sleeping", func() {
					Eventually(stream).Should(Receive(Equal(PollJobEvent{
						State:    constant.JobProcessing,
						Err:      nil,
						Warnings: Warnings{"warning-1"},
						JobGUID:  "job-guid",
					})))
					Eventually(stream).Should(Receive(Equal(PollJobEvent{
						State:    constant.JobProcessing,
						Err:      nil,
						Warnings: Warnings{"warning-2"},
						JobGUID:  "job-guid",
					})))
					for range 4 {
						Eventually(stream).Should(Receive())
					}
					Eventually(stream).Should(Receive(Equal(PollJobEvent{
						State:    constant.JobComplete,
						Err:      nil,
						Warnings: Warnings{"warning-3", "warning-4"},
						JobGUID:  "job-guid",
					})))
					Eventually(stream).Should(BeClosed())
					Expect(pollingSleeps).To(Equal([]time.Duration{
						250 * time.Millisecond, 500 * time.Millisecond, time.Second,
						1250 * time.Millisecond, 3 * time.Second, 3 * time.Second,
					}))
				})
			})

			When("a GET takes longer than the next scheduled polling point", func() {
				BeforeEach(func() {
					setInterval(3 * time.Second)
					fakeGETDuration = time.Second
					appendHandler("PROCESSING", nil)
					appendHandler("COMPLETE", nil)
				})

				It("sleeps until the next polling grid point", func() {
					Eventually(stream).Should(Receive(Equal(PollJobEvent{
						State:   constant.JobProcessing,
						JobGUID: "job-guid",
					})))
					Eventually(stream).Should(Receive(Equal(PollJobEvent{
						State:   constant.JobComplete,
						JobGUID: "job-guid",
					})))
					Eventually(stream).Should(BeClosed())
					Expect(server.ReceivedRequests()).To(HaveLen(2))
					Expect(pollingSleeps).To(Equal([]time.Duration{750 * time.Millisecond}))
				})
			})

			When("the job starts queued and then fails", func() {
				BeforeEach(func() {
					appendHandler("PROCESSING", Warnings{"warning-1"})
					appendHandler("PROCESSING", Warnings{"warning-2", "warning-3"})
					appendFailureHandler("some-message", constant.JobErrorCodeBuildpackAlreadyExistsForStack, Warnings{"warning-4"})
				})

				It("returns the first error", func() {
					Eventually(stream).Should(Receive(Equal(PollJobEvent{
						State:    constant.JobProcessing,
						Err:      nil,
						Warnings: Warnings{"warning-1"},
						JobGUID:  "job-guid",
					})))
					Eventually(stream).Should(Receive(Equal(PollJobEvent{
						State:    constant.JobProcessing,
						Err:      nil,
						Warnings: Warnings{"warning-2", "warning-3"},
						JobGUID:  "job-guid",
					})))
					Eventually(stream).Should(Receive(Equal(PollJobEvent{
						State:    constant.JobFailed,
						Err:      ccerror.BuildpackAlreadyExistsForStackError{Message: "some-message"},
						Warnings: Warnings{"warning-4"},
						JobGUID:  "job-guid",
					})))
					Eventually(stream).Should(BeClosed())
				})
			})

			Context("polling timeouts", func() {
				When("the job runs longer than the OverallPollingTimeout", func() {
					BeforeEach(func() {
						jobPollingTimeout = 100 * time.Millisecond
						client, fakeClock = NewTestClient(Config{
							JobPollingTimeout:  jobPollingTimeout,
							JobPollingInterval: 60 * time.Millisecond,
						})

						useFakeClock()

						appendHandler("PROCESSING", Warnings{"warning-1"})
						appendHandler("PROCESSING", Warnings{"warning-2", "warning-3"})
						appendHandler("PROCESSING", Warnings{"warning-4"})
					})

					It("returns a JobTimeoutError", func() {
						Eventually(stream).Should(Receive(Equal(PollJobEvent{
							State: constant.JobProcessing,
							Err: ccerror.JobTimeoutError{
								Timeout: jobPollingTimeout,
								JobGUID: "job-guid",
							},
							Warnings: Warnings{"warning-4"},
							JobGUID:  "job-guid",
						})))
						Eventually(stream).Should(BeClosed())
					})

					It("checks timeout after GET and sleeps only between events", func() {
						for range 3 {
							Eventually(stream).Should(Receive())
						}
						Eventually(stream).Should(BeClosed())
						Expect(server.ReceivedRequests()).To(HaveLen(3))
						Expect(pollingSleeps).To(Equal([]time.Duration{60 * time.Millisecond, 60 * time.Millisecond}))
					})
				})
			})
		})
	})
})
