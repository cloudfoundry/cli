package resources_test

import (
	"encoding/json"

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
		Entry("tags", ServiceInstance{Tags: []string{"foo", "bar"}}, `{"tags": ["foo", "bar"]}`),
		Entry("syslog", ServiceInstance{SyslogDrainURL: "https://fake-sylogg.com"}, `{"syslog_drain_url": "https://fake-sylogg.com"}`),
		Entry("route", ServiceInstance{RouteServiceURL: "https://fake-route.com"}, `{"route_service_url": "https://fake-route.com"}`),
		Entry(
			"credentials",
			ServiceInstance{
				Credentials: map[string]interface{}{
					"foo": "bar",
					"baz": 42,
				},
			},
			`{
				"credentials": {
					"foo": "bar",
					"baz": 42
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
			"everything",
			ServiceInstance{
				Type:            UserProvidedServiceInstance,
				GUID:            "fake-guid",
				Name:            "fake-space-guid",
				SpaceGUID:       "fake-space-guid",
				Tags:            []string{"foo", "bar"},
				SyslogDrainURL:  "https://fake-sylogg.com",
				RouteServiceURL: "https://fake-route.com",
				Credentials: map[string]interface{}{
					"foo": "bar",
					"baz": 42,
				},
			},
			`{
				"type": "user-provided",
				"guid": "fake-guid",
				"name": "fake-space-guid",
				"tags": ["foo", "bar"],
				"syslog_drain_url": "https://fake-sylogg.com",
				"route_service_url": "https://fake-route.com",
				"credentials": {
					"foo": "bar",
					"baz": 42
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
