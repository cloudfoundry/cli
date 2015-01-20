package resources_test

import (
	"encoding/json"

	. "github.com/cloudfoundry/cli/cf/api/resources"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ServiceInstanceResource", func() {
	var resource ServiceInstanceResource

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
        "gateway_data": null,
        "dashboard_url": "https://fake/dashboard/url",
        "type": "managed_service_instance",
        "space_url": "/v2/spaces/fake-space-guid",
        "service_plan_url": "/v2/service_plans/fake-service-plan-guid",
        "service_bindings_url": "/v2/service_instances/fake-guid/service_bindings",
        "state": "creating",
        "state_description": "fake state description",
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
	})

	Describe("#ToFields", func() {
		It("unmarshalls the fields of a service instance resource", func() {
			fields := resource.ToFields()

			Expect(fields.Guid).To(Equal("fake-guid"))
			Expect(fields.Name).To(Equal("fake service name"))
			Expect(fields.DashboardUrl).To(Equal("https://fake/dashboard/url"))
			Expect(fields.State).To(Equal("creating"))
			Expect(fields.StateDescription).To(Equal("fake state description"))
		})
	})

	Describe("#ToModel", func() {
		It("unmarshalls the service instance resource model", func() {
			instance := resource.ToModel()

			Expect(instance.ServiceInstanceFields.Guid).To(Equal("fake-guid"))
			Expect(instance.ServiceInstanceFields.Name).To(Equal("fake service name"))
			Expect(instance.ServiceInstanceFields.DashboardUrl).To(Equal("https://fake/dashboard/url"))
			Expect(instance.ServiceInstanceFields.State).To(Equal("creating"))
			Expect(instance.ServiceInstanceFields.StateDescription).To(Equal("fake state description"))

			Expect(instance.ServicePlan.Guid).To(Equal("fake-service-plan-guid"))
			Expect(instance.ServicePlan.Free).To(BeTrue())
			Expect(instance.ServicePlan.Description).To(Equal("fake-description"))
			Expect(instance.ServicePlan.Public).To(BeTrue())
			Expect(instance.ServicePlan.Active).To(BeTrue())
			Expect(instance.ServicePlan.ServiceOfferingGuid).To(Equal("fake-service-guid"))

			Expect(instance.ServiceBindings[0].Guid).To(Equal("fake-service-binding-guid"))
			Expect(instance.ServiceBindings[0].Url).To(Equal("http://fake/url"))
			Expect(instance.ServiceBindings[0].AppGuid).To(Equal("fake-app-guid"))
		})
	})
})
