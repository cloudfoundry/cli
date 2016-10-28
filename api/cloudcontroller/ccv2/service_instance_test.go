package ccv2_test

import (
	"net/http"

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
							"type": "managed_service_instance"
						}
					},
					{
						"metadata": {
							"guid": "some-service-guid-2"
						},
						"entity": {
							"name": "some-service-name-2",
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
							"type": "managed_service_instance"
						}
					},
					{
						"metadata": {
							"guid": "some-service-guid-4"
						},
						"entity": {
							"name": "some-service-name-4",
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
				serviceInstances, warnings, err := client.GetServiceInstances([]Query{{
					Filter:   SpaceGUIDFilter,
					Operator: EqualOperator,
					Value:    "some-space-guid",
				}})
				Expect(err).NotTo(HaveOccurred())

				Expect(serviceInstances).To(ConsistOf([]ServiceInstance{
					{Name: "some-service-name-1", GUID: "some-service-guid-1", Type: ManagedService},
					{Name: "some-service-name-2", GUID: "some-service-guid-2", Type: ManagedService},
					{Name: "some-service-name-3", GUID: "some-service-guid-3", Type: ManagedService},
					{Name: "some-service-name-4", GUID: "some-service-guid-4", Type: ManagedService},
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
								"type": "managed_service_instance"
							}
						},
						{
							"metadata": {
								"guid": "some-service-guid-2"
							},
							"entity": {
								"name": "some-service-name-2",
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
								"type": "managed_service_instance"
							}
						},
						{
							"metadata": {
								"guid": "some-service-guid-4"
							},
							"entity": {
								"name": "some-service-name-4",
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
					serviceInstances, warnings, err := client.GetSpaceServiceInstances("some-space-guid", true, []Query{{
						Filter:   NameFilter,
						Operator: EqualOperator,
						Value:    "foobar",
					}})
					Expect(err).NotTo(HaveOccurred())

					Expect(serviceInstances).To(ConsistOf([]ServiceInstance{
						{Name: "some-service-name-1", GUID: "some-service-guid-1", Type: ManagedService},
						{Name: "some-service-name-2", GUID: "some-service-guid-2", Type: UserProvidedService},
						{Name: "some-service-name-3", GUID: "some-service-guid-3", Type: ManagedService},
						{Name: "some-service-name-4", GUID: "some-service-guid-4", Type: UserProvidedService},
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
								"type": "managed_service_instance"
							}
						},
						{
							"metadata": {
								"guid": "some-service-guid-2"
							},
							"entity": {
								"name": "some-service-name-2",
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
					serviceInstances, warnings, err := client.GetSpaceServiceInstances("some-space-guid", false, []Query{{
						Filter:   NameFilter,
						Operator: EqualOperator,
						Value:    "foobar",
					}})
					Expect(err).NotTo(HaveOccurred())

					Expect(serviceInstances).To(ConsistOf([]ServiceInstance{
						{Name: "some-service-name-1", GUID: "some-service-guid-1", Type: ManagedService},
						{Name: "some-service-name-2", GUID: "some-service-guid-2", Type: ManagedService},
					}))
					Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
				})
			})
		})
	})
})
