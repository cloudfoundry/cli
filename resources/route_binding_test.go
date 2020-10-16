package resources

import (
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("RouteBinding", func() {
	DescribeTable(
		"Marshaling and Unmarshaling",
		func(binding RouteBinding, serialized string) {
			By("marshalling", func() {
				Expect(json.Marshal(binding)).To(MatchJSON(serialized))
			})

			By("unmarshaling", func() {
				var parsed RouteBinding
				Expect(json.Unmarshal([]byte(serialized), &parsed)).NotTo(HaveOccurred())
				Expect(parsed).To(Equal(binding))
			})
		},
		Entry("empty", RouteBinding{}, `{}`),
		Entry(
			"basic",
			RouteBinding{
				GUID:                "fake-guid",
				RouteServiceURL:     "fake-route-service-url",
				ServiceInstanceGUID: "fake-service-instance-guid",
				RouteGUID:           "fake-route-guid",
			},
			`{
				"guid": "fake-guid",
				"route_service_url": "fake-route-service-url",
				"relationships": {
					"service_instance": {
						"data": {
							"guid": "fake-service-instance-guid"
						}
					},
					"route": {
						"data": {
							"guid": "fake-route-guid"
						}
					}
				}
			}`,
		),
		Entry(
			"complete",
			RouteBinding{
				GUID:                "fake-guid",
				RouteServiceURL:     "fake-route-service-url",
				ServiceInstanceGUID: "fake-service-instance-guid",
				RouteGUID:           "fake-route-guid",
				LastOperation: LastOperation{
					Type:        "fake-operation-type",
					State:       "fake-operation-state",
					Description: "fake-operation-description",
				},
			},
			`{
				"guid": "fake-guid",
				"route_service_url": "fake-route-service-url",
				"last_operation": {
					"type": "fake-operation-type",
					"state": "fake-operation-state",
					"description": "fake-operation-description"
				},
				"relationships": {
					"service_instance": {
						"data": {
							"guid": "fake-service-instance-guid"
						}
					},
					"route": {
						"data": {
							"guid": "fake-route-guid"
						}
					}
				}
			}`,
		),
	)
})
