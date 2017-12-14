package ccv3_test

import (
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Process", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("GetApplicationProcesses", func() {
		Context("when the application exists", func() {
			BeforeEach(func() {
				response1 := fmt.Sprintf(`
					{
						"pagination": {
							"next": {
								"href": "%s/v3/apps/some-app-guid/processes?page=2"
							}
						},
						"resources": [
							{
								"guid": "process-1-guid",
								"type": "web",
								"memory_in_mb": 32,
								"health_check": {
                  "type": "port",
                  "data": {
                    "timeout": null,
                    "endpoint": null
                  }
                }
							},
							{
								"guid": "process-2-guid",
								"type": "worker",
								"memory_in_mb": 64,
								"health_check": {
                  "type": "http",
                  "data": {
                    "timeout": 60,
                    "endpoint": "/health"
                  }
                }
							}
						]
					}`, server.URL())
				response2 := `
					{
						"pagination": {
							"next": null
						},
						"resources": [
							{
								"guid": "process-3-guid",
								"type": "console",
								"memory_in_mb": 128,
								"health_check": {
                  "type": "process",
                  "data": {
                    "timeout": 90,
                    "endpoint": null
                  }
                }
							}
						]
					}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/apps/some-app-guid/processes"),
						RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"warning-1"}}),
					),
				)
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/apps/some-app-guid/processes", "page=2"),
						RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"warning-2"}}),
					),
				)
			})

			It("returns a list of processes associated with the application and all warnings", func() {
				processes, warnings, err := client.GetApplicationProcesses("some-app-guid")
				Expect(err).ToNot(HaveOccurred())

				Expect(processes).To(ConsistOf(
					Process{
						GUID:        "process-1-guid",
						Type:        constant.ProcessTypeWeb,
						MemoryInMB:  types.NullUint64{Value: 32, IsSet: true},
						HealthCheck: ProcessHealthCheck{Type: "port"},
					},
					Process{
						GUID:       "process-2-guid",
						Type:       "worker",
						MemoryInMB: types.NullUint64{Value: 64, IsSet: true},
						HealthCheck: ProcessHealthCheck{
							Type: "http",
							Data: ProcessHealthCheckData{Endpoint: "/health"},
						},
					},
					Process{
						GUID:        "process-3-guid",
						Type:        "console",
						MemoryInMB:  types.NullUint64{Value: 128, IsSet: true},
						HealthCheck: ProcessHealthCheck{Type: "process"},
					},
				))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})

		Context("when cloud controller returns an error", func() {
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
						VerifyRequest(http.MethodGet, "/v3/apps/some-app-guid/processes"),
						RespondWith(http.StatusNotFound, response),
					),
				)
			})

			It("returns the error", func() {
				_, _, err := client.GetApplicationProcesses("some-app-guid")
				Expect(err).To(MatchError(ccerror.ApplicationNotFoundError{}))
			})
		})
	})

	Describe("GetApplicationProcessByType", func() {
		var (
			process  Process
			warnings []string
			err      error
		)

		JustBeforeEach(func() {
			process, warnings, err = client.GetApplicationProcessByType("some-app-guid", "some-type")
		})

		Context("when the process exists", func() {
			BeforeEach(func() {
				response := `{
					"guid": "process-1-guid",
					"type": "some-type",
					"memory_in_mb": 32,
					"health_check": {
						"type": "http",
						"data": {
							"timeout": 90,
							"endpoint": "/health"
						}
					}
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/apps/some-app-guid/processes/some-type"),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the process and all warnings", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))
				Expect(process).To(Equal(Process{
					GUID:       "process-1-guid",
					Type:       "some-type",
					MemoryInMB: types.NullUint64{Value: 32, IsSet: true},
					HealthCheck: ProcessHealthCheck{
						Type: "http",
						Data: ProcessHealthCheckData{Endpoint: "/health"}},
				}))
			})
		})

		Context("when the application does not exist", func() {
			BeforeEach(func() {
				response := `{
					"errors": [
						{
							"detail": "Application not found",
							"title": "CF-ResourceNotFound",
							"code": 10010
						}
					]
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/apps/some-app-guid/processes/some-type"),
						RespondWith(http.StatusNotFound, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns a ResourceNotFoundError", func() {
				Expect(warnings).To(ConsistOf("this is a warning"))
				Expect(err).To(MatchError(ccerror.ResourceNotFoundError{Message: "Application not found"}))
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
							"code": 10009,
							"detail": "Some CC Error",
							"title": "CF-SomeNewError"
						}
					]
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/apps/some-app-guid/processes/some-type"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error and all warnings", func() {
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
								Code:   10009,
								Detail: "Some CC Error",
								Title:  "CF-SomeNewError",
							},
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

	Describe("PatchApplicationProcessHealthCheck", func() {
		var (
			endpoint string

			process  Process
			warnings []string
			err      error
		)

		JustBeforeEach(func() {
			process, warnings, err = client.PatchApplicationProcessHealthCheck("some-process-guid", "some-type", endpoint)
		})

		Context("when patching the process succeeds", func() {
			Context("and the endpoint is non-empty", func() {
				BeforeEach(func() {
					endpoint = "some-endpoint"
					expectedBody := `{
					"health_check": {
						"type": "some-type",
						"data": {
							"endpoint": "some-endpoint"
						}
					}
				}`
					expectedResponse := `{
					"health_check": {
						"type": "some-type",
						"data": {
							"endpoint": "some-endpoint"
						}
					}
				}`
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodPatch, "/v3/processes/some-process-guid"),
							VerifyJSON(expectedBody),
							RespondWith(http.StatusOK, expectedResponse, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
						),
					)
				})

				It("patches this process's health check", func() {
					Expect(process).To(Equal(Process{
						HealthCheck: ProcessHealthCheck{
							Type: "some-type",
							Data: ProcessHealthCheckData{
								Endpoint: "some-endpoint",
							},
						},
					}))
					Expect(err).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("this is a warning"))
				})
			})

			Context("and the endpoint is empty", func() {
				BeforeEach(func() {
					endpoint = ""
					expectedBody := `{
					"health_check": {
						"type": "some-type",
						"data": {
							"endpoint": null
						}
					}
				}`
					responseBody := `{
					"guid": "some-process-guid",
					"health_check": {
						"type": "some-type",
						"data": {
							"endpoint": null
						}
					}
				}`
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodPatch, "/v3/processes/some-process-guid"),
							VerifyJSON(expectedBody),
							RespondWith(http.StatusOK, responseBody, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
						),
					)
				})

				It("patches this process's health check", func() {
					Expect(process).To(Equal(Process{GUID: "some-process-guid", HealthCheck: ProcessHealthCheck{
						Type: "some-type",
					}}))
					Expect(err).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("this is a warning"))
				})
			})
		})

		Context("when the process does not exist", func() {
			BeforeEach(func() {
				endpoint = "some-endpoint"
				response := `{
					"errors": [
						{
							"detail": "Process not found",
							"title": "CF-ResourceNotFound",
							"code": 10010
						}
					]
				}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPatch, "/v3/processes/some-process-guid"),
						RespondWith(http.StatusNotFound, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns an error and warnings", func() {
				Expect(err).To(MatchError(ccerror.ProcessNotFoundError{}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})

		Context("when the cloud controller returns errors and warnings", func() {
			BeforeEach(func() {
				endpoint = "some-endpoint"
				response := `{
						"errors": [
							{
								"code": 10008,
								"detail": "The request is semantically invalid: command presence",
								"title": "CF-UnprocessableEntity"
							},
							{
								"code": 10009,
								"detail": "Some CC Error",
								"title": "CF-SomeNewError"
							}
						]
					}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPatch, "/v3/processes/some-process-guid"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error and all warnings", func() {
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
								Code:   10009,
								Detail: "Some CC Error",
								Title:  "CF-SomeNewError",
							},
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

	Describe("CreateApplicationProcessScale", func() {
		var passedProcess Process

		Context("when providing all scale options", func() {
			BeforeEach(func() {
				passedProcess = Process{
					Type:       constant.ProcessTypeWeb,
					Instances:  types.NullInt{Value: 2, IsSet: true},
					MemoryInMB: types.NullUint64{Value: 100, IsSet: true},
					DiskInMB:   types.NullUint64{Value: 200, IsSet: true},
				}
				expectedBody := `{
					"instances": 2,
					"memory_in_mb": 100,
					"disk_in_mb": 200
				}`
				response := `{
					"guid": "some-process-guid"
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v3/apps/some-app-guid/processes/web/actions/scale"),
						VerifyJSON(expectedBody),
						RespondWith(http.StatusAccepted, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("scales the application process; returns the scaled process and all warnings", func() {
				process, warnings, err := client.CreateApplicationProcessScale("some-app-guid", passedProcess)
				Expect(process).To(Equal(Process{GUID: "some-process-guid"}))
				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})

		Context("when providing all scale options with 0 values", func() {
			BeforeEach(func() {
				passedProcess = Process{
					Type:       constant.ProcessTypeWeb,
					Instances:  types.NullInt{Value: 0, IsSet: true},
					MemoryInMB: types.NullUint64{Value: 0, IsSet: true},
					DiskInMB:   types.NullUint64{Value: 0, IsSet: true},
				}
				expectedBody := `{
					"instances": 0,
					"memory_in_mb": 0,
					"disk_in_mb": 0
				}`
				response := `{
					"guid": "some-process-guid"
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v3/apps/some-app-guid/processes/web/actions/scale"),
						VerifyJSON(expectedBody),
						RespondWith(http.StatusAccepted, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("scales the application process to 0 values; returns the scaled process and all warnings", func() {
				process, warnings, err := client.CreateApplicationProcessScale("some-app-guid", passedProcess)
				Expect(process).To(Equal(Process{GUID: "some-process-guid"}))
				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})

		Context("when providing only one scale option", func() {
			BeforeEach(func() {
				passedProcess = Process{Type: constant.ProcessTypeWeb, Instances: types.NullInt{Value: 2, IsSet: true}}
				expectedBody := `{
					"instances": 2
				}`
				response := `{
					"guid": "some-process-guid",
					"instances": 2
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v3/apps/some-app-guid/processes/web/actions/scale"),
						VerifyJSON(expectedBody),
						RespondWith(http.StatusAccepted, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("scales the application process; returns the process object and all warnings", func() {
				process, warnings, err := client.CreateApplicationProcessScale("some-app-guid", passedProcess)
				Expect(process).To(Equal(Process{GUID: "some-process-guid", Instances: types.NullInt{Value: 2, IsSet: true}}))
				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})

		Context("when an error is encountered", func() {
			BeforeEach(func() {
				passedProcess = Process{Type: constant.ProcessTypeWeb, Instances: types.NullInt{Value: 2, IsSet: true}}
				response := `{
						"errors": [
							{
								"code": 10008,
								"detail": "The request is semantically invalid: command presence",
								"title": "CF-UnprocessableEntity"
							},
							{
								"code": 10009,
								"detail": "Some CC Error",
								"title": "CF-SomeNewError"
							}
						]
					}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v3/apps/some-app-guid/processes/web/actions/scale"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns an empty process, the error and all warnings", func() {
				process, warnings, err := client.CreateApplicationProcessScale("some-app-guid", passedProcess)
				Expect(process).To(BeZero())
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
								Code:   10009,
								Detail: "Some CC Error",
								Title:  "CF-SomeNewError",
							},
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})
})
