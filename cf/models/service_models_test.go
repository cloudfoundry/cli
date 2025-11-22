package models_test

import (
	"github.com/cloudfoundry/cli/cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Service Models", func() {
	Describe("ServiceBindingRequest", func() {
		It("stores binding request information", func() {
			params := map[string]interface{}{
				"key1": "value1",
				"key2": 123,
			}

			request := models.ServiceBindingRequest{
				AppGuid:             "app-guid",
				ServiceInstanceGuid: "service-instance-guid",
				Params:              params,
			}

			Expect(request.AppGuid).To(Equal("app-guid"))
			Expect(request.ServiceInstanceGuid).To(Equal("service-instance-guid"))
			Expect(request.Params).To(HaveKeyWithValue("key1", "value1"))
			Expect(request.Params).To(HaveKeyWithValue("key2", 123))
		})

		It("handles empty params", func() {
			request := models.ServiceBindingRequest{
				AppGuid:             "app-guid",
				ServiceInstanceGuid: "service-instance-guid",
				Params:              nil,
			}

			Expect(request.Params).To(BeNil())
		})
	})

	Describe("ServiceBindingFields", func() {
		It("stores binding fields", func() {
			binding := models.ServiceBindingFields{
				Guid:    "binding-guid",
				Url:     "http://binding.url",
				AppGuid: "app-guid",
			}

			Expect(binding.Guid).To(Equal("binding-guid"))
			Expect(binding.Url).To(Equal("http://binding.url"))
			Expect(binding.AppGuid).To(Equal("app-guid"))
		})
	})

	Describe("LastOperationFields", func() {
		It("stores last operation information", func() {
			lastOp := models.LastOperationFields{
				Type:        "create",
				State:       "succeeded",
				Description: "Service created successfully",
				CreatedAt:   "2015-01-01T00:00:00Z",
				UpdatedAt:   "2015-01-01T01:00:00Z",
			}

			Expect(lastOp.Type).To(Equal("create"))
			Expect(lastOp.State).To(Equal("succeeded"))
			Expect(lastOp.Description).To(Equal("Service created successfully"))
			Expect(lastOp.CreatedAt).To(Equal("2015-01-01T00:00:00Z"))
			Expect(lastOp.UpdatedAt).To(Equal("2015-01-01T01:00:00Z"))
		})

		It("handles different states", func() {
			inProgress := models.LastOperationFields{State: "in progress"}
			succeeded := models.LastOperationFields{State: "succeeded"}
			failed := models.LastOperationFields{State: "failed"}

			Expect(inProgress.State).To(Equal("in progress"))
			Expect(succeeded.State).To(Equal("succeeded"))
			Expect(failed.State).To(Equal("failed"))
		})
	})

	Describe("ServiceInstanceCreateRequest", func() {
		It("stores create request information", func() {
			params := map[string]interface{}{"config": "value"}
			tags := []string{"tag1", "tag2"}

			request := models.ServiceInstanceCreateRequest{
				Name:      "my-service",
				SpaceGuid: "space-guid",
				PlanGuid:  "plan-guid",
				Params:    params,
				Tags:      tags,
			}

			Expect(request.Name).To(Equal("my-service"))
			Expect(request.SpaceGuid).To(Equal("space-guid"))
			Expect(request.PlanGuid).To(Equal("plan-guid"))
			Expect(request.Params).To(HaveKeyWithValue("config", "value"))
			Expect(request.Tags).To(Equal([]string{"tag1", "tag2"}))
		})

		It("handles optional fields", func() {
			request := models.ServiceInstanceCreateRequest{
				Name:      "my-service",
				SpaceGuid: "space-guid",
			}

			Expect(request.PlanGuid).To(BeEmpty())
			Expect(request.Params).To(BeNil())
			Expect(request.Tags).To(BeNil())
		})
	})

	Describe("ServiceInstanceUpdateRequest", func() {
		It("stores update request information", func() {
			params := map[string]interface{}{"new-config": "new-value"}
			tags := []string{"new-tag"}

			request := models.ServiceInstanceUpdateRequest{
				PlanGuid: "new-plan-guid",
				Params:   params,
				Tags:     tags,
			}

			Expect(request.PlanGuid).To(Equal("new-plan-guid"))
			Expect(request.Params).To(HaveKeyWithValue("new-config", "new-value"))
			Expect(request.Tags).To(Equal([]string{"new-tag"}))
		})

		It("handles empty tags", func() {
			request := models.ServiceInstanceUpdateRequest{
				Tags: []string{},
			}

			Expect(len(request.Tags)).To(Equal(0))
		})
	})

	Describe("ServiceInstanceFields", func() {
		It("stores service instance fields", func() {
			lastOp := models.LastOperationFields{
				Type:  "update",
				State: "succeeded",
			}
			params := map[string]interface{}{"key": "value"}

			instance := models.ServiceInstanceFields{
				Guid:             "instance-guid",
				Name:             "my-service",
				LastOperation:    lastOp,
				SysLogDrainUrl:   "syslog://drain.url",
				ApplicationNames: []string{"app1", "app2"},
				Params:           params,
				DashboardUrl:     "http://dashboard.url",
			}

			Expect(instance.Guid).To(Equal("instance-guid"))
			Expect(instance.Name).To(Equal("my-service"))
			Expect(instance.LastOperation.Type).To(Equal("update"))
			Expect(instance.SysLogDrainUrl).To(Equal("syslog://drain.url"))
			Expect(instance.ApplicationNames).To(Equal([]string{"app1", "app2"}))
			Expect(instance.Params).To(HaveKeyWithValue("key", "value"))
			Expect(instance.DashboardUrl).To(Equal("http://dashboard.url"))
		})
	})

	Describe("ServiceInstance", func() {
		It("embeds ServiceInstanceFields", func() {
			instance := models.ServiceInstance{
				ServiceInstanceFields: models.ServiceInstanceFields{
					Guid: "instance-guid",
					Name: "my-service",
				},
			}

			Expect(instance.Guid).To(Equal("instance-guid"))
			Expect(instance.Name).To(Equal("my-service"))
		})

		It("has service bindings", func() {
			instance := models.ServiceInstance{
				ServiceInstanceFields: models.ServiceInstanceFields{
					Guid: "instance-guid",
				},
				ServiceBindings: []models.ServiceBindingFields{
					{Guid: "binding-1-guid", AppGuid: "app-1-guid"},
					{Guid: "binding-2-guid", AppGuid: "app-2-guid"},
				},
			}

			Expect(len(instance.ServiceBindings)).To(Equal(2))
			Expect(instance.ServiceBindings[0].Guid).To(Equal("binding-1-guid"))
		})

		It("has service keys", func() {
			instance := models.ServiceInstance{
				ServiceInstanceFields: models.ServiceInstanceFields{
					Guid: "instance-guid",
				},
				ServiceKeys: []models.ServiceKeyFields{
					{Guid: "key-1-guid", Name: "key-1"},
					{Guid: "key-2-guid", Name: "key-2"},
				},
			}

			Expect(len(instance.ServiceKeys)).To(Equal(2))
			Expect(instance.ServiceKeys[0].Name).To(Equal("key-1"))
		})

		It("has service plan", func() {
			instance := models.ServiceInstance{
				ServiceInstanceFields: models.ServiceInstanceFields{
					Guid: "instance-guid",
				},
				ServicePlan: models.ServicePlanFields{
					Guid: "plan-guid",
					Name: "plan-name",
				},
			}

			Expect(instance.ServicePlan.Guid).To(Equal("plan-guid"))
			Expect(instance.ServicePlan.Name).To(Equal("plan-name"))
		})

		It("has service offering", func() {
			instance := models.ServiceInstance{
				ServiceInstanceFields: models.ServiceInstanceFields{
					Guid: "instance-guid",
				},
				ServiceOffering: models.ServiceOfferingFields{
					Guid:  "offering-guid",
					Label: "mysql",
				},
			}

			Expect(instance.ServiceOffering.Guid).To(Equal("offering-guid"))
			Expect(instance.ServiceOffering.Label).To(Equal("mysql"))
		})

		Describe("IsUserProvided", func() {
			It("returns true when service plan guid is empty", func() {
				instance := models.ServiceInstance{
					ServiceInstanceFields: models.ServiceInstanceFields{
						Guid: "instance-guid",
					},
					ServicePlan: models.ServicePlanFields{
						Guid: "",
					},
				}

				Expect(instance.IsUserProvided()).To(BeTrue())
			})

			It("returns false when service plan guid is present", func() {
				instance := models.ServiceInstance{
					ServiceInstanceFields: models.ServiceInstanceFields{
						Guid: "instance-guid",
					},
					ServicePlan: models.ServicePlanFields{
						Guid: "plan-guid",
					},
				}

				Expect(instance.IsUserProvided()).To(BeFalse())
			})
		})
	})

	Describe("ServiceOfferingFields", func() {
		It("stores service offering fields", func() {
			offering := models.ServiceOfferingFields{
				Guid:             "offering-guid",
				BrokerGuid:       "broker-guid",
				Label:            "postgresql",
				Provider:         "provider-name",
				Version:          "9.5",
				Description:      "PostgreSQL database service",
				DocumentationUrl: "http://docs.postgresql.org",
			}

			Expect(offering.Guid).To(Equal("offering-guid"))
			Expect(offering.BrokerGuid).To(Equal("broker-guid"))
			Expect(offering.Label).To(Equal("postgresql"))
			Expect(offering.Provider).To(Equal("provider-name"))
			Expect(offering.Version).To(Equal("9.5"))
			Expect(offering.Description).To(Equal("PostgreSQL database service"))
			Expect(offering.DocumentationUrl).To(Equal("http://docs.postgresql.org"))
		})
	})

	Describe("ServiceOffering", func() {
		It("embeds ServiceOfferingFields", func() {
			offering := models.ServiceOffering{
				ServiceOfferingFields: models.ServiceOfferingFields{
					Guid:  "offering-guid",
					Label: "mysql",
				},
			}

			Expect(offering.Guid).To(Equal("offering-guid"))
			Expect(offering.Label).To(Equal("mysql"))
		})

		It("has plans", func() {
			offering := models.ServiceOffering{
				ServiceOfferingFields: models.ServiceOfferingFields{
					Guid:  "offering-guid",
					Label: "mysql",
				},
				Plans: []models.ServicePlanFields{
					{Guid: "plan-1-guid", Name: "small"},
					{Guid: "plan-2-guid", Name: "medium"},
					{Guid: "plan-3-guid", Name: "large"},
				},
			}

			Expect(len(offering.Plans)).To(Equal(3))
			Expect(offering.Plans[0].Name).To(Equal("small"))
			Expect(offering.Plans[1].Name).To(Equal("medium"))
			Expect(offering.Plans[2].Name).To(Equal("large"))
		})
	})

	Describe("ServiceOfferings", func() {
		It("implements sort.Interface for Len", func() {
			offerings := models.ServiceOfferings{
				{ServiceOfferingFields: models.ServiceOfferingFields{Label: "mysql"}},
				{ServiceOfferingFields: models.ServiceOfferingFields{Label: "postgresql"}},
			}

			Expect(offerings.Len()).To(Equal(2))
		})

		It("implements sort.Interface for Swap", func() {
			offerings := models.ServiceOfferings{
				{ServiceOfferingFields: models.ServiceOfferingFields{Label: "mysql"}},
				{ServiceOfferingFields: models.ServiceOfferingFields{Label: "postgresql"}},
			}

			offerings.Swap(0, 1)
			Expect(offerings[0].Label).To(Equal("postgresql"))
			Expect(offerings[1].Label).To(Equal("mysql"))
		})

		It("implements sort.Interface for Less", func() {
			offerings := models.ServiceOfferings{
				{ServiceOfferingFields: models.ServiceOfferingFields{Label: "postgresql"}},
				{ServiceOfferingFields: models.ServiceOfferingFields{Label: "mysql"}},
			}

			Expect(offerings.Less(1, 0)).To(BeTrue())  // mysql < postgresql
			Expect(offerings.Less(0, 1)).To(BeFalse()) // postgresql > mysql
		})

		It("can be sorted alphabetically", func() {
			offerings := models.ServiceOfferings{
				{ServiceOfferingFields: models.ServiceOfferingFields{Label: "redis"}},
				{ServiceOfferingFields: models.ServiceOfferingFields{Label: "mysql"}},
				{ServiceOfferingFields: models.ServiceOfferingFields{Label: "postgresql"}},
			}

			// After sorting, order should be: mysql, postgresql, redis
			Expect(offerings.Less(1, 0)).To(BeTrue())  // mysql < redis
			Expect(offerings.Less(2, 0)).To(BeTrue())  // postgresql < redis
			Expect(offerings.Less(1, 2)).To(BeTrue())  // mysql < postgresql
		})
	})
})
