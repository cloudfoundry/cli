package resources_test

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/resources"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("service instance resource", func() {
	Describe("MarshalJSON", func() {
		It("marshals all fields", func() {
			si := resources.ServiceInstance{
				Type:      resources.UserProvidedServiceInstance,
				Name:      "fake-space-guid",
				SpaceGUID: "fake-space-guid",
			}
			Expect(json.Marshal(si)).To(MatchJSON(`
			{
				"type": "user-provided",
				"name": "fake-space-guid",
				"relationships": {
					"space": {
						"data": {
							"guid": "fake-space-guid"
						}
					}
				}
            }`))
		})
	})

	Describe("UnmarshalJSON", func() {
		It("unmarshals the guid", func() {
			const input = `{"guid": "fake-service-instance-guid"}`
			var si resources.ServiceInstance
			Expect(json.Unmarshal([]byte(input), &si)).NotTo(HaveOccurred())
			Expect(si).To(Equal(resources.ServiceInstance{GUID: "fake-service-instance-guid"}))
		})
	})
})
