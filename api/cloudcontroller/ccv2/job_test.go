package ccv2_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/ccv2fakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/wrapper"
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
				Expect(err).To(MatchError(ccerror.JobFailedError{
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

					Expect(err).To(MatchError(ccerror.JobTimeoutError{
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
					Expect(endTime).To(BeTemporally("~", startTime, 3*jobPollingTimeout))
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

				Expect(err).To(MatchError(ccerror.ResourceNotFoundError{
					Message: "The organization could not be found: some-org-guid",
				}))
				Expect(warnings).To(ConsistOf(Warnings{"warning-1", "warning-2"}))
			})
		})
	})

	Describe("UploadApplicationPackage", func() {
		BeforeEach(func() {
			client = NewTestClient()
		})

		Context("when the upload is successful", func() {
			var (
				resources  []Resource
				reader     io.Reader
				readerBody []byte
			)

			BeforeEach(func() {
				resources = []Resource{
					{Filename: "foo"},
					{Filename: "bar"},
				}

				readerBody = []byte("hello world")
				reader = bytes.NewReader(readerBody)

				verifyHeaderAndBody := func(_ http.ResponseWriter, req *http.Request) {
					contentType := req.Header.Get("Content-Type")
					Expect(contentType).To(MatchRegexp("multipart/form-data; boundary=[\\w\\d]+"))

					defer req.Body.Close()
					reader := multipart.NewReader(req.Body, contentType[30:])

					// Verify that matched resources are sent properly
					resourcesPart, err := reader.NextPart()
					Expect(err).NotTo(HaveOccurred())

					Expect(resourcesPart.FormName()).To(Equal("resources"))

					defer resourcesPart.Close()
					expectedJSON, err := json.Marshal(resources)
					Expect(err).NotTo(HaveOccurred())
					Expect(ioutil.ReadAll(resourcesPart)).To(MatchJSON(expectedJSON))

					// Verify that the application bits are sent properly
					resourcesPart, err = reader.NextPart()
					Expect(err).NotTo(HaveOccurred())

					Expect(resourcesPart.FormName()).To(Equal("application"))
					Expect(resourcesPart.FileName()).To(Equal("application.zip"))

					defer resourcesPart.Close()
					Expect(ioutil.ReadAll(resourcesPart)).To(Equal(readerBody))
				}

				response := `{
					"metadata": {
						"guid": "job-guid",
						"url": "/v2/jobs/job-guid"
					},
					"entity": {
						"guid": "job-guid",
						"status": "queued"
					}
				}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPut, "/v2/apps/some-app-guid/bits", "async=true"),
						verifyHeaderAndBody,
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the created job and warnings", func() {
				job, warnings, err := client.UploadApplicationPackage("some-app-guid", resources, reader, int64(len(readerBody)))
				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))
				Expect(job).To(Equal(Job{
					GUID:   "job-guid",
					Status: JobStatusQueued,
				}))
			})
		})

		Context("when the CC returns an error", func() {
			BeforeEach(func() {
				response := `{
					"code": 30003,
					"description": "Banana",
					"error_code": "CF-Banana"
				}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPut, "/v2/apps/some-app-guid/bits", "async=true"),
						RespondWith(http.StatusNotFound, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error", func() {
				_, warnings, err := client.UploadApplicationPackage("some-app-guid", []Resource{}, bytes.NewReader(nil), 0)
				Expect(err).To(MatchError(ccerror.ResourceNotFoundError{Message: "Banana"}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})

		Context("when passed a nil resources", func() {
			It("returns a NilObjectError", func() {
				_, _, err := client.UploadApplicationPackage("some-app-guid", nil, bytes.NewReader(nil), 0)
				Expect(err).To(MatchError(ccerror.NilObjectError{Object: "existingResources"}))
			})
		})

		Context("when passed a nil reader", func() {
			It("returns a NilObjectError", func() {
				_, _, err := client.UploadApplicationPackage("some-app-guid", []Resource{}, nil, 0)
				Expect(err).To(MatchError(ccerror.NilObjectError{Object: "newResources"}))
			})
		})

		Context("when an error is returned from the new resources reader", func() {
			var (
				fakeReader  *ccv2fakes.FakeReader
				expectedErr error
			)

			BeforeEach(func() {
				expectedErr = errors.New("some read error")
				fakeReader = new(ccv2fakes.FakeReader)
				fakeReader.ReadReturns(0, expectedErr)

				server.AppendHandlers(
					VerifyRequest(http.MethodPut, "/v2/apps/some-app-guid/bits", "async=true"),
				)
			})

			It("returns the error", func() {
				_, _, err := client.UploadApplicationPackage("some-app-guid", []Resource{}, fakeReader, 3)
				Expect(err).To(MatchError(expectedErr))
			})
		})

		Context("when a retryable error occurs", func() {
			BeforeEach(func() {
				wrapper := &wrapper.CustomWrapper{
					CustomMake: func(connection cloudcontroller.Connection, request *cloudcontroller.Request, response *cloudcontroller.Response) error {
						defer GinkgoRecover() // Since this will be running in a thread

						if strings.HasSuffix(request.URL.String(), "/v2/apps/some-app-guid/bits?async=true") {
							defer request.Body.Close()
							return request.ResetBody()
						}
						return connection.Make(request, response)
					},
				}

				client = NewTestClient(Config{Wrappers: []ConnectionWrapper{wrapper}})
			})

			It("returns the PipeSeekError", func() {
				_, _, err := client.UploadApplicationPackage("some-app-guid", []Resource{}, strings.NewReader("hello world"), 3)
				Expect(err).To(MatchError(ccerror.PipeSeekError{}))
			})
		})

		Context("when an http error occurs mid-transfer", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("some read error")

				wrapper := &wrapper.CustomWrapper{
					CustomMake: func(connection cloudcontroller.Connection, request *cloudcontroller.Request, response *cloudcontroller.Response) error {
						defer GinkgoRecover() // Since this will be running in a thread

						if strings.HasSuffix(request.URL.String(), "/v2/apps/some-app-guid/bits?async=true") {
							defer request.Body.Close()
							request.Body.Read(make([]byte, 32*1024))
							return expectedErr
						}
						return connection.Make(request, response)
					},
				}

				client = NewTestClient(Config{Wrappers: []ConnectionWrapper{wrapper}})
			})

			It("returns the http error", func() {
				_, _, err := client.UploadApplicationPackage("some-app-guid", []Resource{}, strings.NewReader(strings.Repeat("a", 33*1024)), 3)
				Expect(err).To(MatchError(expectedErr))
			})
		})
	})
})
