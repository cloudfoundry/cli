package ccv3_test

import (
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("Process", func() {
	var client *Client

	BeforeEach(func() {
		client, _ = NewTestClient()
	})
	Describe("GetProcess", func() {
		var (
			process  resources.Process
			warnings []string
			err      error
		)

		JustBeforeEach(func() {
			process, warnings, err = client.GetProcess("some-process-guid")
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
					"relationships": {
						"app": {
							"data": {
								"guid": "some-app-guid"
							}
						}
					},
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
						VerifyRequest(http.MethodGet, "/v3/processes/some-process-guid"),
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
					"AppGUID":                      Equal("some-app-guid"),
					"Command":                      Equal(types.FilteredString{IsSet: true, Value: "start-command-1"}),
					"Instances":                    Equal(types.NullInt{Value: 22, IsSet: true}),
					"MemoryInMB":                   Equal(types.NullUint64{Value: 32, IsSet: true}),
					"DiskInMB":                     Equal(types.NullUint64{Value: 1024, IsSet: true}),
					"HealthCheckType":              Equal(constant.HTTP),
					"HealthCheckEndpoint":          Equal("/health"),
					"HealthCheckInvocationTimeout": BeEquivalentTo(42),
					"HealthCheckTimeout":           BeEquivalentTo(90),
				}))
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
						VerifyRequest(http.MethodGet, "/v3/processes/some-process-guid"),
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

	Describe("CreateApplicationProcessScale", func() {
		var passedProcess resources.Process

		When("providing all scale options", func() {
			BeforeEach(func() {
				passedProcess = resources.Process{
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
				passedProcess = resources.Process{
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
				passedProcess = resources.Process{Type: constant.ProcessTypeWeb, Instances: types.NullInt{Value: 2, IsSet: true}}
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
				passedProcess = resources.Process{Type: constant.ProcessTypeWeb, Instances: types.NullInt{Value: 2, IsSet: true}}
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
			process  resources.Process
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
					"relationships": {
						"app": {
							"data": {
								"guid": "some-app-guid"
							}
						}
					},
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
					"AppGUID":                      Equal("some-app-guid"),
					"Command":                      Equal(types.FilteredString{IsSet: true, Value: "start-command-1"}),
					"Instances":                    Equal(types.NullInt{Value: 22, IsSet: true}),
					"MemoryInMB":                   Equal(types.NullUint64{Value: 32, IsSet: true}),
					"DiskInMB":                     Equal(types.NullUint64{Value: 1024, IsSet: true}),
					"HealthCheckType":              Equal(constant.HTTP),
					"HealthCheckEndpoint":          Equal("/health"),
					"HealthCheckInvocationTimeout": BeEquivalentTo(42),
					"HealthCheckTimeout":           BeEquivalentTo(90),
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
					resources.Process{
						GUID:               "process-1-guid",
						Type:               constant.ProcessTypeWeb,
						Command:            types.FilteredString{IsSet: true, Value: "[PRIVATE DATA HIDDEN IN LISTS]"},
						MemoryInMB:         types.NullUint64{Value: 32, IsSet: true},
						HealthCheckType:    constant.Port,
						HealthCheckTimeout: 0,
					},
					resources.Process{
						GUID:                "process-2-guid",
						Type:                "worker",
						Command:             types.FilteredString{IsSet: true, Value: "[PRIVATE DATA HIDDEN IN LISTS]"},
						MemoryInMB:          types.NullUint64{Value: 64, IsSet: true},
						HealthCheckType:     constant.HTTP,
						HealthCheckEndpoint: "/health",
						HealthCheckTimeout:  60,
					},
					resources.Process{
						GUID:               "process-3-guid",
						Type:               "console",
						Command:            types.FilteredString{IsSet: true, Value: "[PRIVATE DATA HIDDEN IN LISTS]"},
						MemoryInMB:         types.NullUint64{Value: 128, IsSet: true},
						HealthCheckType:    constant.Process,
						HealthCheckTimeout: 90,
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

	Describe("GetNewApplicationProcesses", func() {
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
								"guid": "old-web-process-guid",
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
								"guid": "new-web-process-guid",
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
								"guid": "worker-process-guid",
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
								"guid": "console-process-guid",
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

				deploymentResponse := `
					{
						"state": "DEPLOYING",
						"new_processes": [
							{
								"guid": "new-web-process-guid", 
								"type": "web"
							}
						]
					}
				`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/deployments/some-deployment-guid"),
						RespondWith(http.StatusOK, deploymentResponse, http.Header{"X-CF-Warnings": {"warning-1"}}),
					),
				)
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/apps/some-app-guid/processes"),
						RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"warning-2"}}),
					),
				)
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/apps/some-app-guid/processes", "page=2"),
						RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"warning-3"}}),
					),
				)
			})

			It("returns a list of processes associated with the application and all warnings", func() {
				processes, warnings, err := client.GetNewApplicationProcesses("some-app-guid", "some-deployment-guid")
				Expect(err).ToNot(HaveOccurred())

				Expect(processes).To(ConsistOf(
					resources.Process{
						GUID:               "new-web-process-guid",
						Type:               constant.ProcessTypeWeb,
						Command:            types.FilteredString{IsSet: true, Value: "[PRIVATE DATA HIDDEN IN LISTS]"},
						MemoryInMB:         types.NullUint64{Value: 32, IsSet: true},
						HealthCheckType:    constant.Port,
						HealthCheckTimeout: 0,
					},
					resources.Process{
						GUID:                "worker-process-guid",
						Type:                "worker",
						Command:             types.FilteredString{IsSet: true, Value: "[PRIVATE DATA HIDDEN IN LISTS]"},
						MemoryInMB:          types.NullUint64{Value: 64, IsSet: true},
						HealthCheckType:     constant.HTTP,
						HealthCheckEndpoint: "/health",
						HealthCheckTimeout:  60,
					},
					resources.Process{
						GUID:               "console-process-guid",
						Type:               "console",
						Command:            types.FilteredString{IsSet: true, Value: "[PRIVATE DATA HIDDEN IN LISTS]"},
						MemoryInMB:         types.NullUint64{Value: 128, IsSet: true},
						HealthCheckType:    constant.Process,
						HealthCheckTimeout: 90,
					},
				))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2", "warning-3"))
			})
		})

		When("cloud controller returns an error", func() {
			BeforeEach(func() {
				response := `{
					"errors": [
						{
							"code": 10010,
							"detail": "Deployment not found",
							"title": "CF-ResourceNotFound"
						}
					]
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/deployments/some-deployment-guid"),
						RespondWith(http.StatusNotFound, response),
					),
				)
			})

			It("returns the error", func() {
				_, _, err := client.GetNewApplicationProcesses("some-app-guid", "some-deployment-guid")
				Expect(err).To(MatchError(ccerror.DeploymentNotFoundError{}))
			})
		})
	})

	Describe("UpdateProcess", func() {
		var (
			inputProcess resources.Process

			process  resources.Process
			warnings []string
			err      error
		)

		BeforeEach(func() {
			inputProcess = resources.Process{
				GUID: "some-process-guid",
			}
		})

		JustBeforeEach(func() {
			process, warnings, err = client.UpdateProcess(inputProcess)
		})

		When("patching the process succeeds", func() {
			When("the command is set", func() {
				When("the start command is an arbitrary command", func() {
					BeforeEach(func() {
						inputProcess.Command = types.FilteredString{IsSet: true, Value: "some-command"}

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

					It("patches this process's command with the provided command", func() {
						Expect(err).ToNot(HaveOccurred())
						Expect(warnings).To(ConsistOf("this is a warning"))
						Expect(process).To(MatchFields(IgnoreExtras, Fields{
							"Command": Equal(types.FilteredString{IsSet: true, Value: "some-command"}),
						}))
					})
				})

				When("the start command reset", func() {
					BeforeEach(func() {
						inputProcess.Command = types.FilteredString{IsSet: true}

						expectedBody := `{
							"command": null
						}`

						expectedResponse := `{
							"command": "some-default-command"
						}`

						server.AppendHandlers(
							CombineHandlers(
								VerifyRequest(http.MethodPatch, "/v3/processes/some-process-guid"),
								VerifyJSON(expectedBody),
								RespondWith(http.StatusOK, expectedResponse, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
							),
						)
					})

					It("patches this process's command with 'null' and returns the default command", func() {
						Expect(err).ToNot(HaveOccurred())
						Expect(warnings).To(ConsistOf("this is a warning"))
						Expect(process).To(MatchFields(IgnoreExtras, Fields{
							"Command": Equal(types.FilteredString{IsSet: true, Value: "some-default-command"}),
						}))
					})
				})
			})

			When("the endpoint is set", func() {
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
					Expect(err).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("this is a warning"))
					Expect(process).To(MatchFields(IgnoreExtras, Fields{
						"HealthCheckType":     Equal(constant.HealthCheckType("some-type")),
						"HealthCheckEndpoint": Equal("some-endpoint"),
					}))
				})
			})

			When("the invocation timeout is set", func() {
				BeforeEach(func() {
					inputProcess.HealthCheckInvocationTimeout = 42
					inputProcess.HealthCheckType = "some-type"

					expectedBody := `{
					"health_check": {
						"type": "some-type",
						"data": {
							"invocation_timeout": 42
						}
					}
				}`
					expectedResponse := `{
					"health_check": {
						"type": "some-type",
						"data": {
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
					Expect(err).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("this is a warning"))
					Expect(process).To(Equal(resources.Process{
						HealthCheckType:              "some-type",
						HealthCheckInvocationTimeout: 42,
					}))
				})
			})

			When("the health check timeout is set", func() {
				BeforeEach(func() {
					inputProcess.HealthCheckTimeout = 77
					inputProcess.HealthCheckType = "some-type"

					expectedBody := `{
					"health_check": {
						"type": "some-type",
						"data": {
							"timeout": 77
						}
					}
				}`
					expectedResponse := `{
					"health_check": {
						"type": "some-type",
						"data": {
							"timeout": 77
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
					Expect(err).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("this is a warning"))
					Expect(process).To(Equal(resources.Process{
						HealthCheckType:     "some-type",
						HealthCheckEndpoint: "",
						HealthCheckTimeout:  77,
					}))
				})
			})

			When("the endpoint and timeout are not set", func() {
				BeforeEach(func() {
					inputProcess.HealthCheckType = "some-type"

					expectedBody := `{
					"health_check": {
						"type": "some-type",
						"data": {}
					}
				}`
					responseBody := `{
					"health_check": {
						"type": "some-type"
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
					Expect(err).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("this is a warning"))
					Expect(process).To(MatchFields(IgnoreExtras, Fields{
						"HealthCheckType": Equal(constant.HealthCheckType("some-type")),
					}))
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
