package resources_test

import (
	"encoding/json"
	"fmt"

	. "code.cloudfoundry.org/cli/cf/api/resources"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ServiceInstanceResource", func() {
	var resource, resourceWithNullLastOp ServiceInstanceResource

	BeforeEach(func() {
		err := json.Unmarshal([]byte(`
    {
      "metadata": {
        "guid": "fake-guid",
        "url": "/v2/service_instances/fake-guid",
        "created_at": "2015-01-13T18:52:08+00:00",
        "updated_at": null
      },
      "entity": {
        "name": "fake service name",
        "credentials": {
        },
        "service_plan_guid": "fake-service-plan-guid",
        "space_guid": "fake-space-guid",
        "dashboard_url": "https://fake/dashboard/url",
        "type": "managed_service_instance",
        "space_url": "/v2/spaces/fake-space-guid",
        "service_plan_url": "/v2/service_plans/fake-service-plan-guid",
        "service_bindings_url": "/v2/service_instances/fake-guid/service_bindings",
        "last_operation": {
          "type": "create",
          "state": "in progress",
          "description": "fake state description",
          "created_at": "fake created at",
          "updated_at": "fake updated at"
        },
				"tags": [ "tag1", "tag2" ],
        "service_plan": {
          "metadata": {
            "guid": "fake-service-plan-guid"
          },
          "entity": {
            "name": "fake-service-plan-name",
            "free": true,
            "description": "fake-description",
            "public": true,
            "active": true,
            "service_guid": "fake-service-guid"
          }
        },
        "service_bindings": [{
          "metadata": {
            "guid": "fake-service-binding-guid",
            "url": "http://fake/url"
          },
          "entity": {
            "app_guid": "fake-app-guid"
          }
        }]
      }
    }`), &resource)

		Expect(err).ToNot(HaveOccurred())

		err = json.Unmarshal([]byte(`
    {
      "metadata": {
        "guid": "fake-guid",
        "url": "/v2/service_instances/fake-guid",
        "created_at": "2015-01-13T18:52:08+00:00",
        "updated_at": null
      },
      "entity": {
        "name": "fake service name",
        "credentials": {
        },
        "service_plan_guid": "fake-service-plan-guid",
        "space_guid": "fake-space-guid",
        "dashboard_url": "https://fake/dashboard/url",
        "type": "managed_service_instance",
        "space_url": "/v2/spaces/fake-space-guid",
        "service_plan_url": "/v2/service_plans/fake-service-plan-guid",
        "service_bindings_url": "/v2/service_instances/fake-guid/service_bindings",
        "last_operation": null,
				"tags": [ "tag1", "tag2" ],
        "service_plan": {
          "metadata": {
            "guid": "fake-service-plan-guid"
          },
          "entity": {
            "name": "fake-service-plan-name",
            "free": true,
            "description": "fake-description",
            "public": true,
            "active": true,
            "service_guid": "fake-service-guid"
          }
        },
        "service_bindings": [{
          "metadata": {
            "guid": "fake-service-binding-guid",
            "url": "http://fake/url"
          },
          "entity": {
            "app_guid": "fake-app-guid"
          }
        }]
      }
    }`), &resourceWithNullLastOp)

		Expect(err).ToNot(HaveOccurred())
	})

	Context("Async brokers", func() {
		var instanceString string

		BeforeEach(func() {
			instanceString = `
						{
							%s,
							"entity": {
								"name": "fake service name",
								"credentials": {
								},
								"service_plan_guid": "fake-service-plan-guid",
								"space_guid": "fake-space-guid",
								"dashboard_url": "https://fake/dashboard/url",
								"type": "managed_service_instance",
								"space_url": "/v2/spaces/fake-space-guid",
								"service_plan_url": "/v2/service_plans/fake-service-plan-guid",
								"service_bindings_url": "/v2/service_instances/fake-guid/service_bindings",
								"last_operation": {
									"type": "create",
									"state": "in progress",
									"description": "fake state description",
									"updated_at": "fake updated at"
								},
								"service_plan": {
									"metadata": {
										"guid": "fake-service-plan-guid"
									},
									"entity": {
										"name": "fake-service-plan-name",
										"free": true,
										"description": "fake-description",
										"public": true,
										"active": true,
										"service_guid": "fake-service-guid"
									}
								},
								"service_bindings": [{
									"metadata": {
										"guid": "fake-service-binding-guid",
										"url": "http://fake/url"
									},
									"entity": {
										"app_guid": "fake-app-guid"
									}
								}]
							}
						}`
		})

		Describe("#ToFields", func() {
			It("unmarshalls the fields of a service instance resource", func() {
				fields := resource.ToFields()

				Expect(fields.GUID).To(Equal("fake-guid"))
				Expect(fields.Name).To(Equal("fake service name"))
				Expect(fields.Tags).To(Equal([]string{"tag1", "tag2"}))
				Expect(fields.DashboardURL).To(Equal("https://fake/dashboard/url"))
				Expect(fields.LastOperation.Type).To(Equal("create"))
				Expect(fields.LastOperation.State).To(Equal("in progress"))
				Expect(fields.LastOperation.Description).To(Equal("fake state description"))
				Expect(fields.LastOperation.CreatedAt).To(Equal("fake created at"))
				Expect(fields.LastOperation.UpdatedAt).To(Equal("fake updated at"))
			})

			Context("When created_at is null", func() {
				It("unmarshalls the service instance resource model", func() {
					var resourceWithNullCreatedAt ServiceInstanceResource
					metadata := `"metadata": {
								"guid": "fake-guid",
								"url": "/v2/service_instances/fake-guid",
								"created_at": null,
								"updated_at": "2015-01-13T18:52:08+00:00"
							}`
					stringWithNullCreatedAt := fmt.Sprintf(instanceString, metadata)

					err := json.Unmarshal([]byte(stringWithNullCreatedAt), &resourceWithNullCreatedAt)
					Expect(err).ToNot(HaveOccurred())
				})
			})

			Context("When created_at is missing", func() {
				It("unmarshalls the service instance resource model", func() {
					var resourceWithMissingCreatedAt ServiceInstanceResource

					metadata := `"metadata": {
								"guid": "fake-guid",
								"url": "/v2/service_instances/fake-guid",
								"updated_at": "2015-01-13T18:52:08+00:00"
							}`
					stringWithMissingCreatedAt := fmt.Sprintf(instanceString, metadata)

					err := json.Unmarshal([]byte(stringWithMissingCreatedAt), &resourceWithMissingCreatedAt)
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})

		Describe("#ToModel", func() {
			It("unmarshalls the service instance resource model", func() {
				instance := resource.ToModel()

				Expect(instance.ServiceInstanceFields.GUID).To(Equal("fake-guid"))
				Expect(instance.ServiceInstanceFields.Name).To(Equal("fake service name"))
				Expect(instance.ServiceInstanceFields.Tags).To(Equal([]string{"tag1", "tag2"}))
				Expect(instance.ServiceInstanceFields.DashboardURL).To(Equal("https://fake/dashboard/url"))
				Expect(instance.ServiceInstanceFields.LastOperation.Type).To(Equal("create"))
				Expect(instance.ServiceInstanceFields.LastOperation.State).To(Equal("in progress"))
				Expect(instance.ServiceInstanceFields.LastOperation.Description).To(Equal("fake state description"))
				Expect(instance.ServiceInstanceFields.LastOperation.CreatedAt).To(Equal("fake created at"))
				Expect(instance.ServiceInstanceFields.LastOperation.UpdatedAt).To(Equal("fake updated at"))

				Expect(instance.ServicePlan.GUID).To(Equal("fake-service-plan-guid"))
				Expect(instance.ServicePlan.Free).To(BeTrue())
				Expect(instance.ServicePlan.Description).To(Equal("fake-description"))
				Expect(instance.ServicePlan.Public).To(BeTrue())
				Expect(instance.ServicePlan.Active).To(BeTrue())
				Expect(instance.ServicePlan.ServiceOfferingGUID).To(Equal("fake-service-guid"))

				Expect(instance.ServiceBindings[0].GUID).To(Equal("fake-service-binding-guid"))
				Expect(instance.ServiceBindings[0].URL).To(Equal("http://fake/url"))
				Expect(instance.ServiceBindings[0].AppGUID).To(Equal("fake-app-guid"))
			})
		})
	})

	Context("Old brokers (no last_operation)", func() {
		Describe("#ToFields", func() {
			It("unmarshalls the fields of a service instance resource", func() {
				fields := resourceWithNullLastOp.ToFields()

				Expect(fields.GUID).To(Equal("fake-guid"))
				Expect(fields.Name).To(Equal("fake service name"))
				Expect(fields.Tags).To(Equal([]string{"tag1", "tag2"}))
				Expect(fields.DashboardURL).To(Equal("https://fake/dashboard/url"))
				Expect(fields.LastOperation.Type).To(Equal(""))
				Expect(fields.LastOperation.State).To(Equal(""))
				Expect(fields.LastOperation.Description).To(Equal(""))
			})
		})

		Describe("#ToModel", func() {
			It("unmarshalls the service instance resource model", func() {
				instance := resourceWithNullLastOp.ToModel()

				Expect(instance.ServiceInstanceFields.GUID).To(Equal("fake-guid"))
				Expect(instance.ServiceInstanceFields.Name).To(Equal("fake service name"))
				Expect(instance.ServiceInstanceFields.Tags).To(Equal([]string{"tag1", "tag2"}))
				Expect(instance.ServiceInstanceFields.DashboardURL).To(Equal("https://fake/dashboard/url"))

				Expect(instance.ServiceInstanceFields.LastOperation.Type).To(Equal(""))
				Expect(instance.ServiceInstanceFields.LastOperation.State).To(Equal(""))
				Expect(instance.ServiceInstanceFields.LastOperation.Description).To(Equal(""))

				Expect(instance.ServicePlan.GUID).To(Equal("fake-service-plan-guid"))
				Expect(instance.ServicePlan.Free).To(BeTrue())
				Expect(instance.ServicePlan.Description).To(Equal("fake-description"))
				Expect(instance.ServicePlan.Public).To(BeTrue())
				Expect(instance.ServicePlan.Active).To(BeTrue())
				Expect(instance.ServicePlan.ServiceOfferingGUID).To(Equal("fake-service-guid"))

				Expect(instance.ServiceBindings[0].GUID).To(Equal("fake-service-binding-guid"))
				Expect(instance.ServiceBindings[0].URL).To(Equal("http://fake/url"))
				Expect(instance.ServiceBindings[0].AppGUID).To(Equal("fake-app-guid"))
			})
		})
	})
})
