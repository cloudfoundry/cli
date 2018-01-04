package ccv3_test

import (
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Task", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("CreateApplicationTask", func() {
		Context("when the application exists", func() {
			var response string

			BeforeEach(func() {
				response = `{
					"sequence_id": 3
				}`
			})

			Context("when the name is empty", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodPost, "/v3/apps/some-app-guid/tasks"),
							VerifyJSON(`{"command":"some command"}`),
							RespondWith(http.StatusAccepted, response, http.Header{"X-Cf-Warnings": {"warning"}}),
						),
					)
				})

				It("creates and returns the task and all warnings", func() {
					task, warnings, err := client.CreateApplicationTask("some-app-guid", Task{Command: "some command"})
					Expect(err).ToNot(HaveOccurred())

					Expect(task).To(Equal(Task{SequenceID: 3}))
					Expect(warnings).To(ConsistOf("warning"))
				})
			})

			Context("when the name is not empty", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodPost, "/v3/apps/some-app-guid/tasks"),
							VerifyJSON(`{"command":"some command", "name":"some-task-name"}`),
							RespondWith(http.StatusAccepted, response, http.Header{"X-Cf-Warnings": {"warning"}}),
						),
					)
				})

				It("creates and returns the task and all warnings", func() {
					task, warnings, err := client.CreateApplicationTask("some-app-guid", Task{Command: "some command", Name: "some-task-name"})
					Expect(err).ToNot(HaveOccurred())

					Expect(task).To(Equal(Task{SequenceID: 3}))
					Expect(warnings).To(ConsistOf("warning"))
				})
			})

			Context("when the disk size is not 0", func() {
				BeforeEach(func() {
					response := `{
						"disk_in_mb": 123,
						"sequence_id": 3
					}`
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodPost, "/v3/apps/some-app-guid/tasks"),
							VerifyJSON(`{"command":"some command", "disk_in_mb": 123}`),
							RespondWith(http.StatusAccepted, response, http.Header{"X-Cf-Warnings": {"warning"}}),
						),
					)
				})

				It("creates and returns the task and all warnings with the provided disk size", func() {
					task, warnings, err := client.CreateApplicationTask("some-app-guid", Task{Command: "some command", DiskInMB: uint64(123)})
					Expect(err).ToNot(HaveOccurred())

					Expect(task).To(Equal(Task{DiskInMB: uint64(123), SequenceID: 3}))
					Expect(warnings).To(ConsistOf("warning"))
				})
			})

			Context("when the memory is not 0", func() {
				BeforeEach(func() {
					response := `{
						"memory_in_mb": 123,
						"sequence_id": 3
					}`
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodPost, "/v3/apps/some-app-guid/tasks"),
							VerifyJSON(`{"command":"some command", "memory_in_mb": 123}`),
							RespondWith(http.StatusAccepted, response, http.Header{"X-Cf-Warnings": {"warning"}}),
						),
					)
				})

				It("creates and returns the task and all warnings with the provided memory", func() {
					task, warnings, err := client.CreateApplicationTask("some-app-guid", Task{Command: "some command", MemoryInMB: uint64(123)})
					Expect(err).ToNot(HaveOccurred())

					Expect(task).To(Equal(Task{MemoryInMB: uint64(123), SequenceID: 3}))
					Expect(warnings).To(ConsistOf("warning"))
				})
			})

		})

		Context("when the cloud controller returns errors and warnings", func() {
			BeforeEach(func() {
				response := `{
					"errors": [
						{
							"code": 10008,
							"detail": "The request is semantically invalid: command presence",
							"title": "CF-UnprocessableEntity"
						},
						{
							"code": 10010,
							"detail": "App not found",
							"title": "CF-ResourceNotFound"
						}
					]
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v3/apps/some-app-guid/tasks"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"warning"}}),
					),
				)
			})

			It("returns the errors and all warnings", func() {
				_, warnings, err := client.CreateApplicationTask("some-app-guid", Task{Command: "some command"})
				Expect(err).To(MatchError(ccerror.V3UnexpectedResponseError{
					ResponseCode: http.StatusTeapot,
					V3ErrorResponse: ccerror.V3ErrorResponse{
						Errors: []ccerror.V3Error{
							{
								Code:   10008,
								Detail: "The request is semantically invalid: command presence",
								Title:  "CF-UnprocessableEntity",
							},
							{
								Code:   10010,
								Detail: "App not found",
								Title:  "CF-ResourceNotFound",
							},
						},
					},
				}))
				Expect(warnings).To(ConsistOf("warning"))
			})
		})
	})

	Describe("GetApplicationTasks", func() {
		Context("when the application exists", func() {
			BeforeEach(func() {
				response1 := fmt.Sprintf(`{
					"pagination": {
						"next": {
							"href": "%s/v3/apps/some-app-guid/tasks?per_page=2&page=2"
						}
					},
					"resources": [
						{
							"guid": "task-1-guid",
							"sequence_id": 1,
							"name": "task-1",
							"command": "some-command",
							"state": "SUCCEEDED",
							"created_at": "2016-11-07T05:59:01Z"
						},
						{
							"guid": "task-2-guid",
							"sequence_id": 2,
							"name": "task-2",
							"command": "some-command",
							"state": "FAILED",
							"created_at": "2016-11-07T06:59:01Z"
						}
					]
				}`, server.URL())
				response2 := `{
					"pagination": {
						"next": null
					},
					"resources": [
						{
							"guid": "task-3-guid",
							"sequence_id": 3,
							"name": "task-3",
							"command": "some-command",
							"state": "RUNNING",
							"created_at": "2016-11-07T07:59:01Z"
						}
					]
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/apps/some-app-guid/tasks", "per_page=2"),
						RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"warning-1"}}),
					),
				)
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/apps/some-app-guid/tasks", "per_page=2&page=2"),
						RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"warning-2"}}),
					),
				)
			})

			It("returns a list of tasks associated with the application and all warnings", func() {
				tasks, warnings, err := client.GetApplicationTasks(
					"some-app-guid",
					Query{Key: PerPage, Values: []string{"2"}},
				)
				Expect(err).ToNot(HaveOccurred())

				Expect(tasks).To(ConsistOf(
					Task{
						GUID:       "task-1-guid",
						SequenceID: 1,
						Name:       "task-1",
						State:      constant.TaskSucceeded,
						CreatedAt:  "2016-11-07T05:59:01Z",
						Command:    "some-command",
					},
					Task{
						GUID:       "task-2-guid",
						SequenceID: 2,
						Name:       "task-2",
						State:      constant.TaskFailed,
						CreatedAt:  "2016-11-07T06:59:01Z",
						Command:    "some-command",
					},
					Task{
						GUID:       "task-3-guid",
						SequenceID: 3,
						Name:       "task-3",
						State:      constant.TaskRunning,
						CreatedAt:  "2016-11-07T07:59:01Z",
						Command:    "some-command",
					},
				))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})

		Context("when the application does not exist", func() {
			BeforeEach(func() {
				response := `{
					"errors": [
						{
							"code": 10010,
							"detail": "App not found",
							"title": "CF-ResourceNotFound"
						}
					]
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/apps/some-app-guid/tasks"),
						RespondWith(http.StatusNotFound, response),
					),
				)
			})

			It("returns a ResourceNotFoundError", func() {
				_, _, err := client.GetApplicationTasks("some-app-guid")
				Expect(err).To(MatchError(ccerror.ApplicationNotFoundError{}))
			})
		})

		Context("when the cloud controller returns errors and warnings", func() {
			BeforeEach(func() {
				response := `{
					"errors": [
						{
							"code": 10008,
							"detail": "The request is semantically invalid: command presence",
							"title": "CF-UnprocessableEntity"
						},
						{
							"code": 10010,
							"detail": "App not found",
							"title": "CF-ResourceNotFound"
						}
					]
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/apps/some-app-guid/tasks"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"warning"}}),
					),
				)
			})

			It("returns the errors and all warnings", func() {
				_, warnings, err := client.GetApplicationTasks("some-app-guid")
				Expect(err).To(MatchError(ccerror.V3UnexpectedResponseError{
					ResponseCode: http.StatusTeapot,
					V3ErrorResponse: ccerror.V3ErrorResponse{
						Errors: []ccerror.V3Error{
							{
								Code:   10008,
								Detail: "The request is semantically invalid: command presence",
								Title:  "CF-UnprocessableEntity",
							},
							{
								Code:   10010,
								Detail: "App not found",
								Title:  "CF-ResourceNotFound",
							},
						},
					},
				}))
				Expect(warnings).To(ConsistOf("warning"))
			})
		})
	})

	Describe("UpdateTask", func() {
		Context("when the request succeeds", func() {
			BeforeEach(func() {
				response := `{
          "guid": "task-3-guid",
          "sequence_id": 3,
          "name": "task-3",
          "command": "some-command",
          "state": "CANCELING",
          "created_at": "2016-11-07T07:59:01Z"
        }`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPut, "/v3/tasks/some-task-guid/cancel"),
						RespondWith(http.StatusAccepted, response, http.Header{"X-Cf-Warnings": {"warning"}}),
					),
				)
			})

			It("returns the task and warnings", func() {
				task, warnings, err := client.UpdateTask("some-task-guid")
				Expect(err).ToNot(HaveOccurred())

				Expect(task).To(Equal(Task{
					GUID:       "task-3-guid",
					SequenceID: 3,
					Name:       "task-3",
					Command:    "some-command",
					State:      constant.TaskCanceling,
					CreatedAt:  "2016-11-07T07:59:01Z",
				}))
				Expect(warnings).To(ConsistOf("warning"))
			})
		})

		Context("when the request fails", func() {
			BeforeEach(func() {
				response := `{
					"errors": [
						{
							"code": 10008,
							"detail": "The request is semantically invalid: command presence",
							"title": "CF-UnprocessableEntity"
						},
						{
							"code": 10010,
							"detail": "App not found",
							"title": "CF-ResourceNotFound"
						}
					]
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPut, "/v3/tasks/some-task-guid/cancel"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"warning"}}),
					),
				)
			})

			It("returns the errors and all warnings", func() {
				_, warnings, err := client.UpdateTask("some-task-guid")
				Expect(err).To(MatchError(ccerror.V3UnexpectedResponseError{
					ResponseCode: http.StatusTeapot,
					V3ErrorResponse: ccerror.V3ErrorResponse{
						Errors: []ccerror.V3Error{
							{
								Code:   10008,
								Detail: "The request is semantically invalid: command presence",
								Title:  "CF-UnprocessableEntity",
							},
							{
								Code:   10010,
								Detail: "App not found",
								Title:  "CF-ResourceNotFound",
							},
						},
					},
				}))
				Expect(warnings).To(ConsistOf("warning"))
			})
		})
	})
})
