package resources_test

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/types"

	. "code.cloudfoundry.org/cli/resources"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
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
		Entry("empty", ServiceInstance{}, `{}`),
		Entry("type", ServiceInstance{Type: "fake-type"}, `{"type": "fake-type"}`),
		Entry("name", ServiceInstance{Name: "fake-name"}, `{"name": "fake-name"}`),
		Entry("guid", ServiceInstance{GUID: "fake-guid"}, `{"guid": "fake-guid"}`),
		Entry("tags", ServiceInstance{Tags: types.NewOptionalStringSlice("foo", "bar")}, `{"tags": ["foo", "bar"]}`),
		Entry("tags empty", ServiceInstance{Tags: types.NewOptionalStringSlice()}, `{"tags": []}`),
		Entry("syslog", ServiceInstance{SyslogDrainURL: types.NewOptionalString("https://fake-syslog.com")}, `{"syslog_drain_url": "https://fake-syslog.com"}`),
		Entry("syslog empty", ServiceInstance{SyslogDrainURL: types.NewOptionalString("")}, `{"syslog_drain_url": ""}`),
		Entry("route", ServiceInstance{RouteServiceURL: types.NewOptionalString("https://fake-route.com")}, `{"route_service_url": "https://fake-route.com"}`),
		Entry("route empty", ServiceInstance{RouteServiceURL: types.NewOptionalString("https://fake-route.com")}, `{"route_service_url": "https://fake-route.com"}`),
		Entry("dashboard", ServiceInstance{DashboardURL: types.NewOptionalString("https://fake-dashboard.com")}, `{"dashboard_url": "https://fake-dashboard.com"}`),
		Entry("dashboard empty", ServiceInstance{DashboardURL: types.NewOptionalString("https://fake-dashboard.com")}, `{"dashboard_url": "https://fake-dashboard.com"}`),
		Entry(
			"credentials",
			ServiceInstance{
				Credentials: types.NewOptionalObject(map[string]interface{}{
					"foo": "bar",
					"baz": false,
				}),
			},
			`{
				"credentials": {
					"foo": "bar",
					"baz": false
				}
			}`,
		),
		Entry(
			"credentials empty",
			ServiceInstance{
				Credentials: types.NewOptionalObject(map[string]interface{}{}),
			},
			`{
				"credentials": {}
			}`,
		),
		Entry(
			"last operation",
			ServiceInstance{
				LastOperation: LastOperation{Type: CreateOperation, State: OperationInProgress},
			},
			`{
				"last_operation": {
					"type": "create",
					"state": "in progress"
				}
			}`,
		),
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
			"plan guid",
			ServiceInstance{PlanGUID: "fake-plan-guid"},
			`{
				"relationships": {
					"service_plan": {
						"data": {
							"guid": "fake-plan-guid"
						}
					}
				}
            }`,
		),
		Entry(
			"service offering guid",
			ServiceInstance{ServiceOfferingGUID: "fake-service-offering-guid"},
			`{
				"relationships": {
					"service_offering": {
						"data": {
							"guid": "fake-service-offering-guid"
						}
					}
				}
            }`,
		),
		Entry(
			"everything",
			ServiceInstance{
				Type:                UserProvidedServiceInstance,
				GUID:                "fake-guid",
				Name:                "fake-space-guid",
				SpaceGUID:           "fake-space-guid",
				ServiceOfferingGUID: "fake-service-offering-guid",
				PlanGUID:            "fake-plan-guid",
				Tags:                types.NewOptionalStringSlice("foo", "bar"),
				SyslogDrainURL:      types.NewOptionalString("https://fake-syslog.com"),
				RouteServiceURL:     types.NewOptionalString("https://fake-route.com"),
				DashboardURL:        types.NewOptionalString("https://fake-dashboard.com"),
				Credentials: types.NewOptionalObject(map[string]interface{}{
					"foo": "bar",
					"baz": false,
				}),
			},
			`{
				"type": "user-provided",
				"guid": "fake-guid",
				"name": "fake-space-guid",
				"tags": ["foo", "bar"],
				"syslog_drain_url": "https://fake-syslog.com",
				"route_service_url": "https://fake-route.com",
				"dashboard_url": "https://fake-dashboard.com",
				"credentials": {
					"foo": "bar",
					"baz": false
				},
				"relationships": {
					"space": {
						"data": {
							"guid": "fake-space-guid"
						}
					},
					"service_offering": {
						"data": {
							"guid": "fake-service-offering-guid"
						}
					},
					"service_plan": {
						"data": {
							"guid": "fake-plan-guid"
						}
					}
				}
            }`,
		),
	)
})
