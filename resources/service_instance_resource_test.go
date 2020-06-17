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
		Entry("type", ServiceInstance{Type: "fake-type"}, `{"type": "fake-type"}`),
		Entry("name", ServiceInstance{Name: "fake-name"}, `{"name": "fake-name"}`),
		Entry("guid", ServiceInstance{GUID: "fake-guid"}, `{"guid": "fake-guid"}`),
		Entry("tags", ServiceInstance{Tags: types.NewOptionalStringSlice("foo", "bar")}, `{"tags": ["foo", "bar"]}`),
		Entry("tags empty", ServiceInstance{Tags: types.NewOptionalStringSlice()}, `{"tags": []}`),
		Entry("syslog", ServiceInstance{SyslogDrainURL: types.NewOptionalString("https://fake-syslog.com")}, `{"syslog_drain_url": "https://fake-syslog.com"}`),
		Entry("syslog empty", ServiceInstance{SyslogDrainURL: types.NewOptionalString("")}, `{"syslog_drain_url": ""}`),
		Entry("route", ServiceInstance{RouteServiceURL: types.NewOptionalString("https://fake-route.com")}, `{"route_service_url": "https://fake-route.com"}`),
		Entry("route empty", ServiceInstance{RouteServiceURL: types.NewOptionalString("https://fake-route.com")}, `{"route_service_url": "https://fake-route.com"}`),
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
				Type:            UserProvidedServiceInstance,
				GUID:            "fake-guid",
				Name:            "fake-space-guid",
				SpaceGUID:       "fake-space-guid",
				Tags:            types.NewOptionalStringSlice("foo", "bar"),
				SyslogDrainURL:  types.NewOptionalString("https://fake-syslog.com"),
				RouteServiceURL: types.NewOptionalString("https://fake-route.com"),
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
				"credentials": {
					"foo": "bar",
					"baz": false
				},
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
