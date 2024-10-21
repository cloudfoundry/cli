package resources_test

import (
	"encoding/json"

	. "code.cloudfoundry.org/cli/v7/resources"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("service instance resource", func() {
	DescribeTable(
		"Marshaling and Unmarshaling",
		func(serviceInstance ServiceInstance, serialized string) {
			By("marshaling", func() {
				Expect(json.Marshal(serviceInstance)).To(MatchJSON(serialized))
			})

			By("unmarshaling", func() {
				var parsed ServiceInstance
				Expect(json.Unmarshal([]byte(serialized), &parsed)).NotTo(HaveOccurred())
				Expect(parsed).To(Equal(serviceInstance))
			})
		},
		Entry("type", ServiceInstance{Type: "fake-type"}, `{"type": "fake-type"}`),
		Entry("name", ServiceInstance{Name: "fake-name"}, `{"name": "fake-name"}`),
		Entry("guid", ServiceInstance{GUID: "fake-guid"}, `{"guid": "fake-guid"}`),
		Entry(
			"space guid",
			ServiceInstance{SpaceGUID: "fake-space-guid"},
			`{
				"relationships": {
					"space": {
						"data": {
							"guid": "fake-space-guid"
						}
					}
				}
            }`,
		),
		Entry(
			"everything",
			ServiceInstance{
				Type:      UserProvidedServiceInstance,
				GUID:      "fake-guid",
				Name:      "fake-space-guid",
				SpaceGUID: "fake-space-guid",
			},
			`{
				"type": "user-provided",
				"guid": "fake-guid",
				"name": "fake-space-guid",
				"relationships": {
					"space": {
						"data": {
							"guid": "fake-space-guid"
						}
					}
				}
            }`,
		),
	)
})
