package resources_test

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/types"

	. "code.cloudfoundry.org/cli/resources"
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
		Entry("upgrade available", ServiceInstance{UpgradeAvailable: types.NewOptionalBoolean(false)}, `{"upgrade_available": false}`),
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
			"parameters",
			ServiceInstance{
				Parameters: types.NewOptionalObject(map[string]interface{}{
					"tomato": "potato",
					"baz":    true,
				}),
			},
			`{
				"parameters": {
					"tomato": "potato",
					"baz": true
				}
			}`,
		),
		Entry(
			"parameters empty",
			ServiceInstance{
				Parameters: types.NewOptionalObject(map[string]interface{}{}),
			},
			`{
				"parameters": {}
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
			ServiceInstance{ServicePlanGUID: "fake-plan-guid"},
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
			"maintenance info version",
			ServiceInstance{MaintenanceInfoVersion: "3.2.1"},
			`{
				"maintenance_info": {
					"version": "3.2.1"
				}
			}`,
		),
		Entry(
			"metadata",
			ServiceInstance{
				Metadata: &Metadata{
					Labels: map[string]types.NullString{
						"foo": types.NewNullString("bar"),
						"baz": types.NewNullString(),
					},
				},
			},
			`{
				"metadata": {
					"labels": {
						"foo": "bar",
						"baz": null
					}
				}
			}`,
		),
		Entry(
			"everything",
			ServiceInstance{
				Type:                   UserProvidedServiceInstance,
				GUID:                   "fake-guid",
				Name:                   "fake-space-guid",
				SpaceGUID:              "fake-space-guid",
				ServicePlanGUID:        "fake-service-plan-guid",
				Tags:                   types.NewOptionalStringSlice("foo", "bar"),
				SyslogDrainURL:         types.NewOptionalString("https://fake-syslog.com"),
				RouteServiceURL:        types.NewOptionalString("https://fake-route.com"),
				DashboardURL:           types.NewOptionalString("https://fake-dashboard.com"),
				UpgradeAvailable:       types.NewOptionalBoolean(true),
				MaintenanceInfoVersion: "1.0.0",
				Credentials: types.NewOptionalObject(map[string]interface{}{
					"foo": "bar",
					"baz": false,
				}),
				Parameters: types.NewOptionalObject(map[string]interface{}{
					"tomato": "potato",
					"baz":    true,
				}),
				LastOperation: LastOperation{
					Type:  "create",
					State: "in progress",
				},
				Metadata: &Metadata{
					Labels: map[string]types.NullString{
						"foo": types.NewNullString("bar"),
						"baz": types.NewNullString(),
					},
				},
			},
			`{
				"type": "user-provided",
				"guid": "fake-guid",
				"name": "fake-space-guid",
				"tags": ["foo", "bar"],
				"syslog_drain_url": "https://fake-syslog.com",
				"route_service_url": "https://fake-route.com",
				"dashboard_url": "https://fake-dashboard.com",
				"maintenance_info": {
					"version": "1.0.0"
				},
				"upgrade_available": true,
				"credentials": {
					"foo": "bar",
					"baz": false
				},
				"parameters": {
					"tomato": "potato",
					"baz": true
				},
				"last_operation": {
					"type": "create",
					"state": "in progress"
				},
				"metadata": {
					"labels": {
						"foo": "bar",
						"baz": null
					}
				},
				"relationships": {
					"service_plan": {
						"data": {
							"guid": "fake-service-plan-guid"
						}
					},
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
