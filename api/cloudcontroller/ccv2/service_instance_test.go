package ccv2_test

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
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
		Context("when the update is successful", func() {
			Context("when setting the minimum", func() { // are we **only** encoding the things we want
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
						GUID: "some-app-guid",
						Name: "some-app-name",
					}))
					Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
				})
			})
		})

		Context("when the create returns an error", func() {
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

	Describe("ServiceInstance", func() {
		Describe("UserProvided", func() {
			Context("when type is USER_PROVIDED_SERVICE", func() {
				It("returns true", func() {
					service := ServiceInstance{Type: UserProvidedService}
					Expect(service.UserProvided()).To(BeTrue())
				})
			})

			Context("when type is MANAGED_SERVICE", func() {
				It("returns false", func() {
					service := ServiceInstance{Type: ManagedService}
					Expect(service.UserProvided()).To(BeFalse())
				})
			})
		})

		Describe("Managed", func() {
			Context("when type is MANAGED_SERVICE", func() {
				It("returns false", func() {
					service := ServiceInstance{Type: ManagedService}
					Expect(service.Managed()).To(BeTrue())
				})
			})

			Context("when type is USER_PROVIDED_SERVICE", func() {
				It("returns true", func() {
					service := ServiceInstance{Type: UserProvidedService}
					Expect(service.Managed()).To(BeFalse())
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
					"service_plan_guid": "some-service-plan-guid",
					"type": "managed_service_instance",
					"tags": [
						"tag-1",
						"tag-2"
					],
					"dashboard_url": "some-dashboard-url",
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

		Context("when service instances exist", func() {
			It("returns the service instance and warnings", func() {
				serviceInstance, warnings, err := client.GetServiceInstance("some-service-guid")
				Expect(err).NotTo(HaveOccurred())

				Expect(serviceInstance).To(Equal(ServiceInstance{
					GUID:            "some-service-guid",
					Name:            "some-service-name",
					SpaceGUID:       "some-space-guid",
					ServicePlanGUID: "some-service-plan-guid",
					Type:            ManagedService,
					Tags:            []string{"tag-1", "tag-2"},
					DashboardURL:    "some-dashboard-url",
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

		Context("when service instances exist", func() {
			It("returns all the queried service instances", func() {
				serviceInstances, warnings, err := client.GetServiceInstances(QQuery{
					Filter:   SpaceGUIDFilter,
					Operator: EqualOperator,
					Values:   []string{"some-space-guid"},
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(serviceInstances).To(ConsistOf([]ServiceInstance{
					{
						Name:      "some-service-name-1",
						GUID:      "some-service-guid-1",
						SpaceGUID: "some-space-guid",
						Type:      ManagedService,
					},
					{
						Name:      "some-service-name-2",
						GUID:      "some-service-guid-2",
						SpaceGUID: "some-space-guid",
						Type:      ManagedService,
					},
					{
						Name:      "some-service-name-3",
						GUID:      "some-service-guid-3",
						SpaceGUID: "some-space-guid",
						Type:      ManagedService,
					},
					{
						Name:      "some-service-name-4",
						GUID:      "some-service-guid-4",
						SpaceGUID: "some-space-guid",
						Type:      ManagedService,
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

			Context("when service instances exist", func() {
				It("returns all the queried service instances", func() {
					serviceInstances, warnings, err := client.GetSpaceServiceInstances("some-space-guid", true, QQuery{
						Filter:   NameFilter,
						Operator: EqualOperator,
						Values:   []string{"foobar"},
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(serviceInstances).To(ConsistOf([]ServiceInstance{
						{Name: "some-service-name-1", GUID: "some-service-guid-1", SpaceGUID: "some-space-guid", Type: ManagedService},
						{Name: "some-service-name-2", GUID: "some-service-guid-2", SpaceGUID: "some-space-guid", Type: UserProvidedService},
						{Name: "some-service-name-3", GUID: "some-service-guid-3", SpaceGUID: "some-space-guid", Type: ManagedService},
						{Name: "some-service-name-4", GUID: "some-service-guid-4", SpaceGUID: "some-space-guid", Type: UserProvidedService},
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

			Context("when service instances exist", func() {
				It("returns all the queried service instances", func() {
					serviceInstances, warnings, err := client.GetSpaceServiceInstances("some-space-guid", false, QQuery{
						Filter:   NameFilter,
						Operator: EqualOperator,
						Values:   []string{"foobar"},
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(serviceInstances).To(ConsistOf([]ServiceInstance{
						{Name: "some-service-name-1", GUID: "some-service-guid-1", SpaceGUID: "some-space-guid", Type: ManagedService},
						{Name: "some-service-name-2", GUID: "some-service-guid-2", SpaceGUID: "some-space-guid", Type: ManagedService},
					}))
					Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
				})
			})
		})
	})
})
