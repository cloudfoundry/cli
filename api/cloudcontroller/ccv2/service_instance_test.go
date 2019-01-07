package ccv2_test

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Service Instance", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("Bind", func() {
		When("the update is successful", func() {
			When("setting the minimum", func() { // are we **only** encoding the things we want
				BeforeEach(func() {
					response := `
						{
							"metadata": {
								"guid": "some-app-guid"
							},
							"entity": {
								"name": "some-app-name",
								"space_guid": "some-space-guid"
							}
						}`
					requestBody := map[string]string{
						"name":       "some-app-name",
						"space_guid": "some-space-guid",
					}
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodPost, "/v2/apps"),
							VerifyJSONRepresenting(requestBody),
							RespondWith(http.StatusCreated, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
						),
					)
				})

				It("returns the created object and warnings", func() {
					app, warnings, err := client.CreateApplication(Application{
						Name:      "some-app-name",
						SpaceGUID: "some-space-guid",
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(app).To(Equal(Application{
						GUID:      "some-app-guid",
						Name:      "some-app-name",
						SpaceGUID: "some-space-guid",
					}))
					Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
				})
			})
		})

		When("the create returns an error", func() {
			BeforeEach(func() {
				response := `
					{
						"description": "Request invalid due to parse error: Field: name, Error: Missing field name, Field: space_guid, Error: Missing field space_guid",
						"error_code": "CF-MessageParseError",
						"code": 1001
					}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v2/apps"),
						RespondWith(http.StatusBadRequest, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error and warnings", func() {
				_, warnings, err := client.CreateApplication(Application{})
				Expect(err).To(MatchError(ccerror.BadRequestError{Message: "Request invalid due to parse error: Field: name, Error: Missing field name, Field: space_guid, Error: Missing field space_guid"}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
			})
		})
	})

	Describe("CreateServiceInstance", func() {
		When("creating service instance succeeds", func() {
			var (
				spaceGUID       string
				servicePlanGUID string
				serviceInstance string
				parameters      map[string]interface{}
				tags            []string
			)

			BeforeEach(func() {
				response := `{
					"metadata": {
						"guid": "service-instance-guid"
					},
					"entity": {
						"name": "my-service-instance",
						"service_plan_guid": "service-plan-guid",
						"space_guid": "space-guid",
						"dashboard_url": "http://dashboard.url",
						"type": "managed_service_instance",
						"last_operation": {
							"type": "create",
							"state": "in progress",
							"description": "",
							"updated_at": "2016-06-08T16:41:26Z",
							"created_at": "2016-06-08T16:41:29Z"
						},
						"tags": ["a-tag", "another-tag"]
					}
				}`
				spaceGUID = "some-space-guid"
				servicePlanGUID = "some-plan-guid"
				serviceInstance = "service-instance-name"
				parameters = map[string]interface{}{
					"param1": "some-value",
					"param2": "another-value",
				}
				tags = []string{"a-tag, another-tag"}
				requestBody := map[string]interface{}{
					"name":              serviceInstance,
					"service_plan_guid": servicePlanGUID,
					"space_guid":        spaceGUID,
					"parameters":        parameters,
					"tags":              tags,
				}
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v2/service_instances", "accepts_incomplete=true"),
						VerifyJSONRepresenting(requestBody),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"warning-1,warning-2"}}),
					),
				)
			})

			It("returns the service instance and all warnings", func() {
				serviceInstance, warnings, err := client.CreateServiceInstance(spaceGUID, servicePlanGUID, serviceInstance, parameters, tags)
				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf(Warnings{"warning-1", "warning-2"}))
				Expect(serviceInstance).To(Equal(ServiceInstance{
					GUID:            "service-instance-guid",
					Name:            "my-service-instance",
					SpaceGUID:       "space-guid",
					ServicePlanGUID: "service-plan-guid",
					ServiceGUID:     "",
					Type:            "managed_service_instance",
					Tags:            []string{"a-tag", "another-tag"},
					DashboardURL:    "http://dashboard.url",
					LastOperation: LastOperation{
						Type:        "create",
						State:       "in progress",
						Description: "",
						UpdatedAt:   "2016-06-08T16:41:26Z",
						CreatedAt:   "2016-06-08T16:41:29Z",
					},
				}))
			})
		})

		When("the endpoint returns an error", func() {
			BeforeEach(func() {
				response := `{
								"code": 10003,
								"description": "You are not authorized to perform the requested action"
							}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v2/service_instances"),
						RespondWith(http.StatusForbidden, response, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
					))
			})

			It("returns all warnings and propagates the error", func() {
				_, warnings, err := client.CreateServiceInstance("space-GUID", "service-plan-GUID", "service-instance", map[string]interface{}{}, []string{})
				Expect(err).To(MatchError(ccerror.ForbiddenError{
					Message: "You are not authorized to perform the requested action",
				}))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})
	})

	Describe("ServiceInstance", func() {
		Describe("Managed", func() {
			When("type is MANAGED_SERVICE", func() {
				It("returns false", func() {
					service := ServiceInstance{Type: constant.ServiceInstanceTypeManagedService}
					Expect(service.Managed()).To(BeTrue())
				})
			})

			When("type is USER_PROVIDED_SERVICE", func() {
				It("returns true", func() {
					service := ServiceInstance{Type: constant.ServiceInstanceTypeUserProvidedService}
					Expect(service.Managed()).To(BeFalse())
				})
			})
		})

		Describe("UserProvided", func() {
			When("type is USER_PROVIDED_SERVICE", func() {
				It("returns true", func() {
					service := ServiceInstance{Type: constant.ServiceInstanceTypeUserProvidedService}
					Expect(service.UserProvided()).To(BeTrue())
				})
			})

			When("type is MANAGED_SERVICE", func() {
				It("returns false", func() {
					service := ServiceInstance{Type: constant.ServiceInstanceTypeManagedService}
					Expect(service.UserProvided()).To(BeFalse())
				})
			})
		})
	})

	Describe("GetServiceInstance", func() {
		BeforeEach(func() {
			response := `{
				"metadata": {
					"guid": "some-service-guid"
				},
				"entity": {
					"name": "some-service-name",
					"space_guid": "some-space-guid",
					"service_guid": "some-service-guid",
					"service_plan_guid": "some-service-plan-guid",
					"type": "managed_service_instance",
					"tags": [
						"tag-1",
						"tag-2"
					],
					"dashboard_url": "some-dashboard-url",
					"route_service_url": "some-route-service-url",
					"last_operation": {
						"type": "create",
						"state": "succeeded",
						"description": "service broker-provided description",
						"updated_at": "updated-at-time",
						"created_at": "created-at-time"
					}
				}
			}`

			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/v2/service_instances/some-service-guid"),
					RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
				),
			)
		})

		When("service instances exist", func() {
			It("returns the service instance and warnings", func() {
				serviceInstance, warnings, err := client.GetServiceInstance("some-service-guid")
				Expect(err).NotTo(HaveOccurred())

				Expect(serviceInstance).To(Equal(ServiceInstance{
					GUID:            "some-service-guid",
					Name:            "some-service-name",
					SpaceGUID:       "some-space-guid",
					ServiceGUID:     "some-service-guid",
					ServicePlanGUID: "some-service-plan-guid",
					Type:            constant.ServiceInstanceTypeManagedService,
					Tags:            []string{"tag-1", "tag-2"},
					DashboardURL:    "some-dashboard-url",
					RouteServiceURL: "some-route-service-url",
					LastOperation: LastOperation{
						Type:        "create",
						State:       "succeeded",
						Description: "service broker-provided description",
						UpdatedAt:   "updated-at-time",
						CreatedAt:   "created-at-time",
					},
				}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
			})
		})
	})

	Describe("GetServiceInstances", func() {
		BeforeEach(func() {
			response1 := `{
				"next_url": "/v2/service_instances?q=space_guid:some-space-guid&page=2",
				"resources": [
					{
						"metadata": {
							"guid": "some-service-guid-1"
						},
						"entity": {
							"name": "some-service-name-1",
							"space_guid": "some-space-guid",
							"service_guid": "some-service-guid",
							"type": "managed_service_instance"
						}
					},
					{
						"metadata": {
							"guid": "some-service-guid-2"
						},
						"entity": {
							"name": "some-service-name-2",
							"space_guid": "some-space-guid",
							"type": "managed_service_instance"
						}
					}
				]
			}`

			response2 := `{
				"next_url": null,
				"resources": [
					{
						"metadata": {
							"guid": "some-service-guid-3"
						},
						"entity": {
							"name": "some-service-name-3",
							"space_guid": "some-space-guid",
							"type": "managed_service_instance"
						}
					},
					{
						"metadata": {
							"guid": "some-service-guid-4"
						},
						"entity": {
							"name": "some-service-name-4",
							"space_guid": "some-space-guid",
							"type": "managed_service_instance"
						}
					}
				]
			}`

			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/v2/service_instances", "q=space_guid:some-space-guid"),
					RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
				),
			)

			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/v2/service_instances", "q=space_guid:some-space-guid&page=2"),
					RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"this is another warning"}}),
				),
			)
		})

		When("service instances exist", func() {
			It("returns all the queried service instances", func() {
				serviceInstances, warnings, err := client.GetServiceInstances(Filter{
					Type:     constant.SpaceGUIDFilter,
					Operator: constant.EqualOperator,
					Values:   []string{"some-space-guid"},
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(serviceInstances).To(ConsistOf([]ServiceInstance{
					{
						Name:        "some-service-name-1",
						GUID:        "some-service-guid-1",
						SpaceGUID:   "some-space-guid",
						ServiceGUID: "some-service-guid",
						Type:        constant.ServiceInstanceTypeManagedService,
					},
					{
						Name:      "some-service-name-2",
						GUID:      "some-service-guid-2",
						SpaceGUID: "some-space-guid",
						Type:      constant.ServiceInstanceTypeManagedService,
					},
					{
						Name:      "some-service-name-3",
						GUID:      "some-service-guid-3",
						SpaceGUID: "some-space-guid",
						Type:      constant.ServiceInstanceTypeManagedService,
					},
					{
						Name:      "some-service-name-4",
						GUID:      "some-service-guid-4",
						SpaceGUID: "some-space-guid",
						Type:      constant.ServiceInstanceTypeManagedService,
					},
				}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning", "this is another warning"}))
			})
		})
	})

	Describe("GetSpaceServiceInstances", func() {
		Context("including user provided services", func() {
			BeforeEach(func() {
				response1 := `{
					"next_url": "/v2/spaces/some-space-guid/service_instances?return_user_provided_service_instances=true&q=name:foobar&page=2",
					"resources": [
						{
							"metadata": {
								"guid": "some-service-guid-1"
							},
							"entity": {
								"name": "some-service-name-1",
								"space_guid": "some-space-guid",
					"service_guid": "some-service-guid",
								"type": "managed_service_instance"
							}
						},
						{
							"metadata": {
								"guid": "some-service-guid-2"
							},
							"entity": {
								"name": "some-service-name-2",
								"space_guid": "some-space-guid",
								"type": "user_provided_service_instance"
							}
						}
					]
				}`

				response2 := `{
					"next_url": null,
					"resources": [
						{
							"metadata": {
								"guid": "some-service-guid-3"
							},
							"entity": {
								"name": "some-service-name-3",
								"space_guid": "some-space-guid",
								"type": "managed_service_instance"
							}
						},
						{
							"metadata": {
								"guid": "some-service-guid-4"
							},
							"entity": {
								"name": "some-service-name-4",
								"space_guid": "some-space-guid",
								"type": "user_provided_service_instance"
							}
						}
					]
				}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/spaces/some-space-guid/service_instances", "return_user_provided_service_instances=true&q=name:foobar"),
						RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/spaces/some-space-guid/service_instances", "return_user_provided_service_instances=true&q=name:foobar&page=2"),
						RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"this is another warning"}}),
					),
				)
			})

			When("service instances exist", func() {
				It("returns all the queried service instances", func() {
					serviceInstances, warnings, err := client.GetSpaceServiceInstances("some-space-guid", true, Filter{
						Type:     constant.NameFilter,
						Operator: constant.EqualOperator,
						Values:   []string{"foobar"},
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(serviceInstances).To(ConsistOf([]ServiceInstance{
						{Name: "some-service-name-1", GUID: "some-service-guid-1", SpaceGUID: "some-space-guid", ServiceGUID: "some-service-guid", Type: constant.ServiceInstanceTypeManagedService},
						{Name: "some-service-name-2", GUID: "some-service-guid-2", SpaceGUID: "some-space-guid", Type: constant.ServiceInstanceTypeUserProvidedService},
						{Name: "some-service-name-3", GUID: "some-service-guid-3", SpaceGUID: "some-space-guid", Type: constant.ServiceInstanceTypeManagedService},
						{Name: "some-service-name-4", GUID: "some-service-guid-4", SpaceGUID: "some-space-guid", Type: constant.ServiceInstanceTypeUserProvidedService},
					}))
					Expect(warnings).To(ConsistOf(Warnings{"this is a warning", "this is another warning"}))
				})
			})
		})

		Context("excluding user provided services", func() {
			BeforeEach(func() {
				response := `{
					"next_url": null,
					"resources": [
						{
							"metadata": {
								"guid": "some-service-guid-1"
							},
							"entity": {
								"name": "some-service-name-1",
								"space_guid": "some-space-guid",
								"type": "managed_service_instance"
							}
						},
						{
							"metadata": {
								"guid": "some-service-guid-2"
							},
							"entity": {
								"name": "some-service-name-2",
								"space_guid": "some-space-guid",
								"type": "managed_service_instance"
							}
						}
					]
				}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/spaces/some-space-guid/service_instances", "q=name:foobar"),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			When("service instances exist", func() {
				It("returns all the queried service instances", func() {
					serviceInstances, warnings, err := client.GetSpaceServiceInstances("some-space-guid", false, Filter{
						Type:     constant.NameFilter,
						Operator: constant.EqualOperator,
						Values:   []string{"foobar"},
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(serviceInstances).To(ConsistOf([]ServiceInstance{
						{Name: "some-service-name-1", GUID: "some-service-guid-1", SpaceGUID: "some-space-guid", Type: constant.ServiceInstanceTypeManagedService},
						{Name: "some-service-name-2", GUID: "some-service-guid-2", SpaceGUID: "some-space-guid", Type: constant.ServiceInstanceTypeManagedService},
					}))
					Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
				})
			})
		})
	})

	Describe("GetUserProvidedServiceInstances", func() {
		var (
			serviceInstances []ServiceInstance
			warnings         Warnings
			executeErr       error
		)

		JustBeforeEach(func() {
			serviceInstances, warnings, executeErr = client.GetUserProvidedServiceInstances(Filter{
				Type:     constant.SpaceGUIDFilter,
				Operator: constant.EqualOperator,
				Values:   []string{"some-space-guid"},
			})
		})

		When("getting user provided service instances errors", func() {
			BeforeEach(func() {
				response := `{
					"code": 1,
					"description": "some error description",
					"error_code": "CF-SomeError"
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/user_provided_service_instances", "q=space_guid:some-space-guid"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
					),
				)
			})

			It("returns the error and all warnings", func() {
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

		When("getting user provided service instances succeeds", func() {
			BeforeEach(func() {
				response1 := `{
				"next_url": "/v2/user_provided_service_instances?q=space_guid:some-space-guid&page=2",
				"resources": [
					{
						"metadata": {
							"guid": "some-service-guid-1"
						},
						"entity": {
							"name": "some-service-name-1",
							"route_service_url": "some-route-service-url",
							"space_guid": "some-space-guid",
							"type": "user_provided_service_instance"
						}
					},
					{
						"metadata": {
							"guid": "some-service-guid-2"
						},
						"entity": {
							"name": "some-service-name-2",
							"route_service_url": "some-route-service-url",
							"space_guid": "some-space-guid",
							"type": "user_provided_service_instance"
						}
					}
				]
			}`

				response2 := `{
				"next_url": null,
				"resources": [
					{
						"metadata": {
							"guid": "some-service-guid-3"
						},
						"entity": {
							"name": "some-service-name-3",
							"route_service_url": "some-route-service-url",
							"space_guid": "some-space-guid",
							"type": "user_provided_service_instance"
						}
					},
					{
						"metadata": {
							"guid": "some-service-guid-4"
						},
						"entity": {
							"name": "some-service-name-4",
							"route_service_url": "some-route-service-url",
							"space_guid": "some-space-guid",
							"type": "user_provided_service_instance"
						}
					}
				]
			}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/user_provided_service_instances", "q=space_guid:some-space-guid"),
						RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/user_provided_service_instances", "q=space_guid:some-space-guid&page=2"),
						RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"this is another warning"}}),
					),
				)
			})

			It("returns all the queried service instances", func() {
				Expect(executeErr).NotTo(HaveOccurred())

				Expect(serviceInstances).To(ConsistOf([]ServiceInstance{
					{
						Name:            "some-service-name-1",
						GUID:            "some-service-guid-1",
						SpaceGUID:       "some-space-guid",
						RouteServiceURL: "some-route-service-url",
						Type:            constant.ServiceInstanceTypeUserProvidedService,
					},
					{
						Name:            "some-service-name-2",
						GUID:            "some-service-guid-2",
						SpaceGUID:       "some-space-guid",
						RouteServiceURL: "some-route-service-url",
						Type:            constant.ServiceInstanceTypeUserProvidedService,
					},
					{
						Name:            "some-service-name-3",
						GUID:            "some-service-guid-3",
						SpaceGUID:       "some-space-guid",
						RouteServiceURL: "some-route-service-url",
						Type:            constant.ServiceInstanceTypeUserProvidedService,
					},
					{
						Name:            "some-service-name-4",
						GUID:            "some-service-guid-4",
						SpaceGUID:       "some-space-guid",
						RouteServiceURL: "some-route-service-url",
						Type:            constant.ServiceInstanceTypeUserProvidedService,
					},
				}))

				Expect(warnings).To(ConsistOf(Warnings{"this is a warning", "this is another warning"}))
			})
		})
	})
})
