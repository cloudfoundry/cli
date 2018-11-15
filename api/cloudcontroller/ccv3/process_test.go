package ccv3_test

import (
	"encoding/json"
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("Process", func() {
	var client *Client

	BeforeEach(func() {
		client, _ = NewTestClient()
	})

	Describe("Process", func() {
		Describe("MarshalJSON", func() {
			var (
				process      Process
				processBytes []byte
				err          error
			)

			BeforeEach(func() {
				process = Process{}
			})

			JustBeforeEach(func() {
				processBytes, err = process.MarshalJSON()
				Expect(err).ToNot(HaveOccurred())
			})

			When("instances is provided", func() {
				BeforeEach(func() {
					process = Process{
						Instances: types.NullInt{Value: 0, IsSet: true},
					}
				})

				It("sets the instances to be set", func() {
					Expect(string(processBytes)).To(MatchJSON(`{"instances": 0}`))
				})
			})

			When("memory is provided", func() {
				BeforeEach(func() {
					process = Process{
						MemoryInMB: types.NullUint64{Value: 0, IsSet: true},
					}
				})

				It("sets the memory to be set", func() {
					Expect(string(processBytes)).To(MatchJSON(`{"memory_in_mb": 0}`))
				})
			})

			When("disk is provided", func() {
				BeforeEach(func() {
					process = Process{
						DiskInMB: types.NullUint64{Value: 0, IsSet: true},
					}
				})

				It("sets the disk to be set", func() {
					Expect(string(processBytes)).To(MatchJSON(`{"disk_in_mb": 0}`))
				})
			})

			When("health check type http is provided", func() {
				BeforeEach(func() {
					process = Process{
						HealthCheckType:     "http",
						HealthCheckEndpoint: "some-endpoint",
					}
				})

				It("sets the health check type to http and has an endpoint", func() {
					Expect(string(processBytes)).To(MatchJSON(`{"health_check":{"type":"http", "data": {"endpoint": "some-endpoint"}}}`))
				})
			})

			When("health check type port is provided", func() {
				BeforeEach(func() {
					process = Process{
						HealthCheckType: "port",
					}
				})

				It("sets the health check type to port", func() {
					Expect(string(processBytes)).To(MatchJSON(`{"health_check":{"type":"port", "data": {"endpoint": null}}}`))
				})
			})

			When("health check type process is provided", func() {
				BeforeEach(func() {
					process = Process{
						HealthCheckType: "process",
					}
				})

				It("sets the health check type to process", func() {
					Expect(string(processBytes)).To(MatchJSON(`{"health_check":{"type":"process", "data": {"endpoint": null}}}`))
				})
			})

			When("process has no fields provided", func() {
				BeforeEach(func() {
					process = Process{}
				})

				It("sets the health check type to process", func() {
					Expect(string(processBytes)).To(MatchJSON(`{}`))
				})
			})
		})

		Describe("UnmarshalJSON", func() {
			var (
				process      Process
				processBytes []byte
				err          error
			)
			BeforeEach(func() {
				processBytes = []byte("{}")
			})

			JustBeforeEach(func() {
				err = json.Unmarshal(processBytes, &process)
				Expect(err).ToNot(HaveOccurred())
			})
			When("health check type http is provided", func() {
				BeforeEach(func() {
					processBytes = []byte(`{"health_check":{"type":"http", "data": {"endpoint": "some-endpoint"}}}`)
				})

				It("sets the health check type to http and has an endpoint", func() {
					Expect(process).To(MatchFields(IgnoreExtras, Fields{
						"HealthCheckType":     Equal("http"),
						"HealthCheckEndpoint": Equal("some-endpoint"),
					}))
				})
			})

			When("health check type port is provided", func() {
				BeforeEach(func() {
					processBytes = []byte(`{"health_check":{"type":"port", "data": {"endpoint": null}}}`)
				})

				It("sets the health check type to port", func() {
					Expect(process).To(MatchFields(IgnoreExtras, Fields{
						"HealthCheckType": Equal("port"),
					}))
				})
			})

			When("health check type process is provided", func() {
				BeforeEach(func() {
					processBytes = []byte(`{"health_check":{"type":"process", "data": {"endpoint": null}}}`)
				})

				It("sets the health check type to process", func() {
					Expect(process).To(MatchFields(IgnoreExtras, Fields{
						"HealthCheckType": Equal("process"),
					}))
				})
			})
		})
	})

	Describe("CreateApplicationProcessScale", func() {
		var passedProcess Process

		When("providing all scale options", func() {
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
				Expect(process).To(MatchFields(IgnoreExtras, Fields{"GUID": Equal("some-process-guid")}))
				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})

		When("providing all scale options with 0 values", func() {
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
				Expect(process).To(MatchFields(IgnoreExtras, Fields{"GUID": Equal("some-process-guid")}))
				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})

		When("providing only one scale option", func() {
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
				Expect(process).To(MatchFields(IgnoreExtras, Fields{
					"GUID":      Equal("some-process-guid"),
					"Instances": Equal(types.NullInt{Value: 2, IsSet: true}),
				}))
				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})

		When("an error is encountered", func() {
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
				Expect(err).To(MatchError(ccerror.MultiError{
					ResponseCode: http.StatusTeapot,
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
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
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

		When("the process exists", func() {
			BeforeEach(func() {
				response := `{
					"guid": "process-1-guid",
					"type": "some-type",
					"command": "start-command-1",
					"instances": 22,
					"memory_in_mb": 32,
					"disk_in_mb": 1024,
					"health_check": {
						"type": "http",
						"data": {
							"timeout": 90,
							"endpoint": "/health",
							"invocation_timeout": 42
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
				Expect(process).To(MatchAllFields(Fields{
					"GUID":                         Equal("process-1-guid"),
					"Type":                         Equal("some-type"),
					"Command":                      Equal("start-command-1"),
					"Instances":                    Equal(types.NullInt{Value: 22, IsSet: true}),
					"MemoryInMB":                   Equal(types.NullUint64{Value: 32, IsSet: true}),
					"DiskInMB":                     Equal(types.NullUint64{Value: 1024, IsSet: true}),
					"HealthCheckType":              Equal("http"),
					"HealthCheckEndpoint":          Equal("/health"),
					"HealthCheckInvocationTimeout": Equal(42),
				}))
			})
		})

		When("the application does not exist", func() {
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

		When("the cloud controller returns errors and warnings", func() {
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
				Expect(err).To(MatchError(ccerror.MultiError{
					ResponseCode: http.StatusTeapot,
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
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

	Describe("GetApplicationProcesses", func() {
		When("the application exists", func() {
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
								"command": "[PRIVATE DATA HIDDEN IN LISTS]",
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
								"command": "[PRIVATE DATA HIDDEN IN LISTS]",
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
								"command": "[PRIVATE DATA HIDDEN IN LISTS]",
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
						GUID:            "process-1-guid",
						Type:            constant.ProcessTypeWeb,
						Command:         "[PRIVATE DATA HIDDEN IN LISTS]",
						MemoryInMB:      types.NullUint64{Value: 32, IsSet: true},
						HealthCheckType: "port",
					},
					Process{
						GUID:                "process-2-guid",
						Type:                "worker",
						Command:             "[PRIVATE DATA HIDDEN IN LISTS]",
						MemoryInMB:          types.NullUint64{Value: 64, IsSet: true},
						HealthCheckType:     "http",
						HealthCheckEndpoint: "/health",
					},
					Process{
						GUID:            "process-3-guid",
						Type:            "console",
						Command:         "[PRIVATE DATA HIDDEN IN LISTS]",
						MemoryInMB:      types.NullUint64{Value: 128, IsSet: true},
						HealthCheckType: "process",
					},
				))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})

		When("cloud controller returns an error", func() {
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

	Describe("UpdateProcess", func() {
		var (
			inputProcess Process

			process  Process
			warnings []string
			err      error
		)

		BeforeEach(func() {
			inputProcess = Process{
				GUID: "some-process-guid",
			}
		})

		JustBeforeEach(func() {
			process, warnings, err = client.UpdateProcess(inputProcess)
		})

		When("patching the process succeeds", func() {
			Context("and the command is set", func() {
				BeforeEach(func() {
					inputProcess.Command = "some-command"

					expectedBody := `{
						"command": "some-command"
					}`

					expectedResponse := `{
						"command": "some-command"
					}`

					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodPatch, "/v3/processes/some-process-guid"),
							VerifyJSON(expectedBody),
							RespondWith(http.StatusOK, expectedResponse, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
						),
					)
				})

				It("patches this process's command", func() {
					Expect(process).To(MatchFields(IgnoreExtras, Fields{
						"Command": Equal("some-command"),
					}))
					Expect(err).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("this is a warning"))
				})
			})

			Context("and the endpoint is set", func() {
				BeforeEach(func() {
					inputProcess.HealthCheckEndpoint = "some-endpoint"
					inputProcess.HealthCheckType = "some-type"

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
							"endpoint": "some-endpoint",
							"invocation_timeout": null
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
					Expect(process).To(MatchFields(IgnoreExtras, Fields{
						"HealthCheckType":     Equal("some-type"),
						"HealthCheckEndpoint": Equal("some-endpoint"),
					}))
					Expect(err).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("this is a warning"))
				})
			})

			Context("and invocation timeout is set", func() {
				BeforeEach(func() {
					inputProcess.HealthCheckInvocationTimeout = 42
					inputProcess.HealthCheckType = "some-type"

					expectedBody := `{
					"health_check": {
						"type": "some-type",
						"data": {
							"endpoint": null,
							"invocation_timeout": 42
						}
					}
				}`
					expectedResponse := `{
					"health_check": {
						"type": "some-type",
						"data": {
							"endpoint": null,
							"invocation_timeout": 42
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
						HealthCheckType:              "some-type",
						HealthCheckEndpoint:          "",
						HealthCheckInvocationTimeout: 42,
					}))
					Expect(err).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("this is a warning"))
				})
			})

			Context("and the endpoint and timeout are not set", func() {
				BeforeEach(func() {
					inputProcess.HealthCheckType = "some-type"

					expectedBody := `{
					"health_check": {
						"type": "some-type",
						"data": {
							"endpoint": null
						}
					}
				}`
					responseBody := `{
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
					Expect(process).To(MatchFields(IgnoreExtras, Fields{
						"HealthCheckType":     Equal("some-type"),
						"HealthCheckEndpoint": BeEmpty(),
					}))
					Expect(err).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("this is a warning"))
				})
			})
		})

		When("the process does not exist", func() {
			BeforeEach(func() {
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

		When("the cloud controller returns errors and warnings", func() {
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
						VerifyRequest(http.MethodPatch, "/v3/processes/some-process-guid"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error and all warnings", func() {
				Expect(err).To(MatchError(ccerror.MultiError{
					ResponseCode: http.StatusTeapot,
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
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})
})
