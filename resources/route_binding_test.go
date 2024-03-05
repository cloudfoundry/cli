package resources

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/types"

	. "github.com/onsi/ginkgo/v2"
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
		Entry("guid", RouteBinding{GUID: "fake-guid"}, `{"guid": "fake-guid"}`),
		Entry("route service url", RouteBinding{RouteServiceURL: "fake-route-service-url"}, `{"route_service_url": "fake-route-service-url"}`),
		Entry(
			"route guid",
			RouteBinding{RouteGUID: "fake-route-guid"},
			`{
				"relationships": {
					"route": {
						"data": {
							"guid": "fake-route-guid"
						}
					}
				}
			}`,
		),
		Entry(
			"service instance guid",
			RouteBinding{ServiceInstanceGUID: "fake-service-instance-guid"},
			`{
				"relationships": {
					"service_instance": {
						"data": {
							"guid": "fake-service-instance-guid"
						}
					}
				}
			}`,
		),
		Entry(
			"parameters",
			RouteBinding{
				Parameters: types.NewOptionalObject(map[string]interface{}{
					"foo": "bar",
				}),
			},
			`{
				"parameters": {
					"foo": "bar"
				}
			}`,
		),
		Entry(
			"last operation",
			RouteBinding{
				LastOperation: LastOperation{
					Type:        "fake-operation-type",
					State:       "fake-operation-state",
					Description: "fake-operation-description",
				},
			},
			`{
				"last_operation": {
					"type": "fake-operation-type",
					"state": "fake-operation-state",
					"description": "fake-operation-description"
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
				Parameters: types.NewOptionalObject(map[string]interface{}{
					"foo": "bar",
				}),
			},
			`{
				"guid": "fake-guid",
				"route_service_url": "fake-route-service-url",
				"parameters": {"foo": "bar"},
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
