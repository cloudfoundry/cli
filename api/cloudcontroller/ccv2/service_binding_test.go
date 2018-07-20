package ccv2_test

import (
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Service Binding", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("CreateServiceBinding", func() {
		var (
			appGUID           string
			serviceGUID       string
			bindingName       string
			acceptsIncomplete bool
			parameters        map[string]interface{}

			serviceBinding ServiceBinding
			warnings       Warnings
			executeErr     error
		)

		BeforeEach(func() {
			appGUID = "some-app-guid"
			serviceGUID = "some-service-instance-guid"
			parameters = map[string]interface{}{
				"the-service-broker": "wants this object",
			}
		})

		JustBeforeEach(func() {
			serviceBinding, warnings, executeErr = client.CreateServiceBinding(appGUID, serviceGUID, bindingName, acceptsIncomplete, parameters)
		})

		Context("when the create is successful", func() {
			Context("when a service binding name is provided", func() {
				BeforeEach(func() {
					bindingName = "some-binding-name"
					acceptsIncomplete = false

					expectedRequestBody := map[string]interface{}{
						"service_instance_guid": "some-service-instance-guid",
						"app_guid":              "some-app-guid",
						"name":                  "some-binding-name",
						"parameters": map[string]interface{}{
							"the-service-broker": "wants this object",
						},
					}
					response := `
						{
							"metadata": {
								"guid": "some-service-binding-guid"
							}
						}`
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodPost, "/v2/service_bindings", "accepts_incomplete=false"),
							VerifyJSONRepresenting(expectedRequestBody),
							RespondWith(http.StatusCreated, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
						),
					)
				})

				It("returns the created object and warnings", func() {
					Expect(executeErr).NotTo(HaveOccurred())

					Expect(serviceBinding).To(Equal(ServiceBinding{GUID: "some-service-binding-guid"}))
					Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
				})
			})

			Context("when a service binding name is not provided", func() {
				BeforeEach(func() {
					bindingName = ""
					acceptsIncomplete = false

					expectedRequestBody := map[string]interface{}{
						"service_instance_guid": "some-service-instance-guid",
						"app_guid":              "some-app-guid",
						"parameters": map[string]interface{}{
							"the-service-broker": "wants this object",
						},
					}
					response := `
						{
							"metadata": {
								"guid": "some-service-binding-guid"
							}
						}`
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodPost, "/v2/service_bindings", "accepts_incomplete=false"),
							VerifyJSONRepresenting(expectedRequestBody),
							RespondWith(http.StatusCreated, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
						),
					)
				})

				It("returns the created object and warnings", func() {
					Expect(executeErr).NotTo(HaveOccurred())

					Expect(serviceBinding).To(Equal(ServiceBinding{GUID: "some-service-binding-guid"}))
					Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
				})
			})

			Context("when accepts_incomplete is true", func() {
				BeforeEach(func() {
					bindingName = "some-binding-name"
					acceptsIncomplete = true

					expectedRequestBody := map[string]interface{}{
						"service_instance_guid": "some-service-instance-guid",
						"app_guid":              "some-app-guid",
						"name":                  "some-binding-name",
						"parameters": map[string]interface{}{
							"the-service-broker": "wants this object",
						},
					}
					response := `
						{
							"metadata": {
								"guid": "some-service-binding-guid"
							}
						}`
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodPost, "/v2/service_bindings", "accepts_incomplete=true"),
							VerifyJSONRepresenting(expectedRequestBody),
							RespondWith(http.StatusCreated, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
						),
					)
				})

				It("returns the created object and warnings", func() {
					Expect(executeErr).NotTo(HaveOccurred())

					Expect(serviceBinding).To(Equal(ServiceBinding{GUID: "some-service-binding-guid"}))
					Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
				})
			})
		})

		Context("when the create returns an error", func() {
			BeforeEach(func() {
				response := `
				{
					  "description": "The app space binding to service is taken: some-app-guid some-service-instance-guid",
						  "error_code": "CF-ServiceBindingAppServiceTaken",
							  "code": 90003
							}
			`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v2/service_bindings"),
						RespondWith(http.StatusBadRequest, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error and warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.ServiceBindingTakenError{Message: "The app space binding to service is taken: some-app-guid some-service-instance-guid"}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
			})
		})
	})

	Describe("DeleteServiceBinding", func() {
		var (
			serviceBindingGUID string
			acceptsIncomplete  bool

			serviceBinding ServiceBinding
			warnings       Warnings
			executeErr     error
		)

		BeforeEach(func() {
			serviceBindingGUID = "some-service-binding-guid"
		})

		JustBeforeEach(func() {
			serviceBinding, warnings, executeErr = client.DeleteServiceBinding(serviceBindingGUID, acceptsIncomplete)
		})

		Context("when the service binding exist", func() {
			Context("when accepts_incomplete is true", func() {
				BeforeEach(func() {
					acceptsIncomplete = true
					response := fmt.Sprintf(`{
						 "metadata": {
								"guid": "%s"
						 },
						 "entity": {
								"app_guid": "63af8eb4-6ac6-4baa-b97d-da32d473131c",
								"service_instance_guid": "637a1734-3eec-408e-aeaa-bfc577d893b7",
								"last_operation": {
									 "state": "in progress"
								}
						 }
					 }`, serviceBindingGUID)
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodDelete, "/v2/service_bindings/some-service-binding-guid", "accepts_incomplete=true"),
							RespondWith(http.StatusAccepted, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
						),
					)
				})

				It("returns the service binding and warnings", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
					Expect(serviceBinding).To(Equal(ServiceBinding{
						GUID:                serviceBindingGUID,
						AppGUID:             "63af8eb4-6ac6-4baa-b97d-da32d473131c",
						ServiceInstanceGUID: "637a1734-3eec-408e-aeaa-bfc577d893b7",
						LastOperation: LastOperation{
							State: constant.LastOperationInProgress,
						},
					}))
				})
			})

			Context("when accepts_incomplete is false", func() {
				BeforeEach(func() {
					acceptsIncomplete = false
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodDelete, "/v2/service_bindings/some-service-binding-guid"),
							RespondWith(http.StatusNoContent, "{}", http.Header{"X-Cf-Warnings": {"this is a warning"}}),
						))
				})

				It("deletes the service binding", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
				})
			})
		})

		Context("when the service binding does not exist", func() {
			BeforeEach(func() {
				acceptsIncomplete = false
				response := `{
				"code": 90004,
				"description": "The service binding could not be found: some-service-binding-guid",
				"error_code": "CF-ServiceBindingNotFound"
			}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodDelete, "/v2/service_bindings/some-service-binding-guid"),
						RespondWith(http.StatusNotFound, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns a not found error", func() {
				Expect(executeErr).To(MatchError(ccerror.ResourceNotFoundError{
					Message: "The service binding could not be found: some-service-binding-guid",
				}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
			})
		})
	})

	Describe("GetServiceBinding", func() {
		var (
			serviceBinding ServiceBinding
			warnings       Warnings
			executeErr     error
		)

		JustBeforeEach(func() {
			serviceBinding, warnings, executeErr = client.GetServiceBinding("some-service-binding-guid")
		})

		Context("when the cc returns an error", func() {
			BeforeEach(func() {
				response := `{
					"code": 1,
					"description": "some error description",
					"error_code": "CF-SomeError"
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/service_bindings/some-service-binding-guid"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
					),
				)
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError(ccerror.V2UnexpectedResponseError{
					V2ErrorResponse: ccerror.V2ErrorResponse{
						Code:        1,
						Description: "some error description",
						ErrorCode:   "CF-SomeError",
					},
					ResponseCode: http.StatusTeapot,
				}))

				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})

		Context("when there are no errors", func() {
			Context("and entity.last_operation is not present", func() {
				BeforeEach(func() {
					response := `{
						"metadata": {
							"guid": "service-binding-guid-1"
						},
						"entity": {
							"app_guid":"app-guid-1",
							"service_instance_guid": "service-instance-guid-1"
						}
					}`
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v2/service_bindings/some-service-binding-guid"),
							RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
						),
					)
				})

				It("returns the service binding", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))

					Expect(serviceBinding).To(Equal(ServiceBinding{
						GUID:                "service-binding-guid-1",
						AppGUID:             "app-guid-1",
						ServiceInstanceGUID: "service-instance-guid-1",
					}))
				})
			})

			Context("and entity.last_operation is present", func() {
				BeforeEach(func() {
					response := `{
						"metadata": {
							"guid": "service-binding-guid-1"
						},
						"entity": {
							"app_guid":"app-guid-1",
							"service_instance_guid": "service-instance-guid-1",
							"last_operation": {
								 "state": "succeeded"
							}
						}
					}`
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v2/service_bindings/some-service-binding-guid"),
							RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
						),
					)
				})

				It("returns the service binding", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))

					Expect(serviceBinding).To(Equal(ServiceBinding{
						GUID:                "service-binding-guid-1",
						AppGUID:             "app-guid-1",
						ServiceInstanceGUID: "service-instance-guid-1",
						LastOperation:       LastOperation{State: constant.LastOperationSucceeded},
					}))
				})
			})
		})
	})

	Describe("GetServiceBindings", func() {
		BeforeEach(func() {
			response1 := `{
				"next_url": "/v2/service_bindings?q=app_guid:some-app-guid&page=2",
				"resources": [
					{
						"metadata": {
							"guid": "service-binding-guid-1"
						},
						"entity": {
							"app_guid":"app-guid-1",
							"service_instance_guid": "service-instance-guid-1"
						}
					},
					{
						"metadata": {
							"guid": "service-binding-guid-2"
						},
						"entity": {
							"app_guid":"app-guid-2",
							"service_instance_guid": "service-instance-guid-2"
						}
					}
				]
			}`
			response2 := `{
				"next_url": null,
				"resources": [
					{
						"metadata": {
							"guid": "service-binding-guid-3"
						},
						"entity": {
							"app_guid":"app-guid-3",
							"service_instance_guid": "service-instance-guid-3"
						}
					},
					{
						"metadata": {
							"guid": "service-binding-guid-4"
						},
						"entity": {
							"app_guid":"app-guid-4",
							"service_instance_guid": "service-instance-guid-4"
						}
					}
				]
			}`
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/v2/service_bindings", "q=app_guid:some-app-guid"),
					RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
				),
			)
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/v2/service_bindings", "q=app_guid:some-app-guid&page=2"),
					RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"this is another warning"}}),
				),
			)
		})

		Context("when service bindings exist", func() {
			It("returns all the queried service bindings", func() {
				serviceBindings, warnings, err := client.GetServiceBindings(Filter{
					Type:     constant.AppGUIDFilter,
					Operator: constant.EqualOperator,
					Values:   []string{"some-app-guid"},
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(serviceBindings).To(ConsistOf([]ServiceBinding{
					{GUID: "service-binding-guid-1", AppGUID: "app-guid-1", ServiceInstanceGUID: "service-instance-guid-1"},
					{GUID: "service-binding-guid-2", AppGUID: "app-guid-2", ServiceInstanceGUID: "service-instance-guid-2"},
					{GUID: "service-binding-guid-3", AppGUID: "app-guid-3", ServiceInstanceGUID: "service-instance-guid-3"},
					{GUID: "service-binding-guid-4", AppGUID: "app-guid-4", ServiceInstanceGUID: "service-instance-guid-4"},
				}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning", "this is another warning"}))
			})
		})
	})

	Describe("GetServiceInstanceServiceBindings", func() {
		Context("when there are service bindings", func() {
			BeforeEach(func() {
				response1 := `{
					"next_url": "/v2/service_instances/some-service-instance-guid/service_bindings?page=2",
					"resources": [
						{
							"metadata": {
								"guid": "service-binding-guid-1"
							},
							"entity": {
								"app_guid":"app-guid-1",
								"service_instance_guid": "service-instance-guid-1"
							}
						},
						{
							"metadata": {
								"guid": "service-binding-guid-2"
							},
							"entity": {
								"app_guid":"app-guid-2",
								"service_instance_guid": "service-instance-guid-2"
							}
						}
					]
				}`
				response2 := `{
					"next_url": null,
					"resources": [
						{
							"metadata": {
								"guid": "service-binding-guid-3"
							},
							"entity": {
								"app_guid":"app-guid-3",
								"service_instance_guid": "service-instance-guid-3"
							}
						},
						{
							"metadata": {
								"guid": "service-binding-guid-4"
							},
							"entity": {
								"app_guid":"app-guid-4",
								"service_instance_guid": "service-instance-guid-4"
							}
						}
					]
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/service_instances/some-service-instance-guid/service_bindings"),
						RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/service_instances/some-service-instance-guid/service_bindings", "page=2"),
						RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"this is another warning"}}),
					),
				)
			})

			It("returns all the service bindings and all warnings", func() {
				serviceBindings, warnings, err := client.GetServiceInstanceServiceBindings("some-service-instance-guid")
				Expect(err).NotTo(HaveOccurred())
				Expect(serviceBindings).To(Equal([]ServiceBinding{
					{GUID: "service-binding-guid-1", AppGUID: "app-guid-1", ServiceInstanceGUID: "service-instance-guid-1"},
					{GUID: "service-binding-guid-2", AppGUID: "app-guid-2", ServiceInstanceGUID: "service-instance-guid-2"},
					{GUID: "service-binding-guid-3", AppGUID: "app-guid-3", ServiceInstanceGUID: "service-instance-guid-3"},
					{GUID: "service-binding-guid-4", AppGUID: "app-guid-4", ServiceInstanceGUID: "service-instance-guid-4"},
				}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning", "this is another warning"}))
			})
		})

		Context("when there are no service bindings", func() {
			BeforeEach(func() {
				response1 := `{
					"next_url": null,
					"resources": []
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/service_instances/some-service-instance-guid/service_bindings"),
						RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})
			It("returns an empty list of service bindings and all warnings", func() {
				serviceBindings, warnings, err := client.GetServiceInstanceServiceBindings("some-service-instance-guid")
				Expect(err).NotTo(HaveOccurred())
				Expect(serviceBindings).To(HaveLen(0))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
			})
		})

		Context("when an error is encountered", func() {
			BeforeEach(func() {
				response := `{
				 "description": "Unknown request",
				 "error_code": "CF-NotFound",
				 "code": 10000
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/service_instances/some-service-instance-guid/service_bindings"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns an error and all warnings", func() {
				_, warnings, err := client.GetServiceInstanceServiceBindings("some-service-instance-guid")
				Expect(err).To(MatchError(ccerror.V2UnexpectedResponseError{
					V2ErrorResponse: ccerror.V2ErrorResponse{
						Code:        10000,
						Description: "Unknown request",
						ErrorCode:   "CF-NotFound",
					},
					RequestIDs:   nil,
					ResponseCode: 418,
				}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
			})
		})
	})

	Describe("GetUserProvidedServiceInstanceServiceBindings", func() {
		Context("when there are service bindings", func() {
			BeforeEach(func() {
				response1 := `{
					"next_url": "/v2/user_provided_service_instances/some-user-provided-service-instance-guid/service_bindings?page=2",
					"resources": [
						{
							"metadata": {
								"guid": "service-binding-guid-1"
							},
							"entity": {
								"app_guid":"app-guid-1",
								"service_instance_guid": "user-provided-service-instance-guid-1"
							}
						},
						{
							"metadata": {
								"guid": "service-binding-guid-2"
							},
							"entity": {
								"app_guid":"app-guid-2",
								"service_instance_guid": "user-provided-service-instance-guid-2"
							}
						}
					]
				}`
				response2 := `{
					"next_url": null,
					"resources": [
						{
							"metadata": {
								"guid": "service-binding-guid-3"
							},
							"entity": {
								"app_guid":"app-guid-3",
								"service_instance_guid": "user-provided-service-instance-guid-3"
							}
						},
						{
							"metadata": {
								"guid": "service-binding-guid-4"
							},
							"entity": {
								"app_guid":"app-guid-4",
								"service_instance_guid": "user-provided-service-instance-guid-4"
							}
						}
					]
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/user_provided_service_instances/some-user-provided-service-instance-guid/service_bindings"),
						RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/user_provided_service_instances/some-user-provided-service-instance-guid/service_bindings", "page=2"),
						RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"this is another warning"}}),
					),
				)
			})

			It("returns all the service bindings and all warnings", func() {
				serviceBindings, warnings, err := client.GetUserProvidedServiceInstanceServiceBindings("some-user-provided-service-instance-guid")
				Expect(err).NotTo(HaveOccurred())
				Expect(serviceBindings).To(Equal([]ServiceBinding{
					{GUID: "service-binding-guid-1", AppGUID: "app-guid-1", ServiceInstanceGUID: "user-provided-service-instance-guid-1"},
					{GUID: "service-binding-guid-2", AppGUID: "app-guid-2", ServiceInstanceGUID: "user-provided-service-instance-guid-2"},
					{GUID: "service-binding-guid-3", AppGUID: "app-guid-3", ServiceInstanceGUID: "user-provided-service-instance-guid-3"},
					{GUID: "service-binding-guid-4", AppGUID: "app-guid-4", ServiceInstanceGUID: "user-provided-service-instance-guid-4"},
				}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning", "this is another warning"}))
			})
		})

		Context("when there are no service bindings", func() {
			BeforeEach(func() {
				response := `{
					"next_url": null,
					"resources": []
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/user_provided_service_instances/some-user-provided-service-instance-guid/service_bindings"),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns an empty list of service bindings and all warnings", func() {
				serviceBindings, warnings, err := client.GetUserProvidedServiceInstanceServiceBindings("some-user-provided-service-instance-guid")
				Expect(err).NotTo(HaveOccurred())
				Expect(serviceBindings).To(HaveLen(0))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
			})
		})

		Context("when an error is encountered", func() {
			BeforeEach(func() {
				response := `{
				 "description": "Unknown request",
				 "error_code": "CF-NotFound",
				 "code": 10000
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/user_provided_service_instances/some-user-provided-service-instance-guid/service_bindings"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns an error and all warnings", func() {
				_, warnings, err := client.GetUserProvidedServiceInstanceServiceBindings("some-user-provided-service-instance-guid")
				Expect(err).To(MatchError(ccerror.V2UnexpectedResponseError{
					V2ErrorResponse: ccerror.V2ErrorResponse{
						Code:        10000,
						Description: "Unknown request",
						ErrorCode:   "CF-NotFound",
					},
					RequestIDs:   nil,
					ResponseCode: 418,
				}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
			})
		})
	})
})
