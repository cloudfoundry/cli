package resources_test

import (
	"encoding/json"

	. "code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("service offering resource", func() {
	DescribeTable(
		"Unmarshaling",
		func(serviceInstance ServiceOffering, serialized string) {
			var parsed ServiceOffering
			Expect(json.Unmarshal([]byte(serialized), &parsed)).NotTo(HaveOccurred())
			Expect(parsed).To(Equal(serviceInstance))
		},
		Entry("name", ServiceOffering{Name: "fake-name"}, `{"name": "fake-name"}`),
		Entry("guid", ServiceOffering{GUID: "fake-guid"}, `{"guid": "fake-guid"}`),
		Entry("shareable", ServiceOffering{AllowsInstanceSharing: true}, `{"shareable": true}`),
		Entry("description", ServiceOffering{Description: "once upon a time"}, `{"description": "once upon a time"}`),
		Entry("documentation_url", ServiceOffering{DocumentationURL: "https://docs.com"}, `{"documentation_url": "https://docs.com"}`),
		Entry("tags", ServiceOffering{Tags: types.NewOptionalStringSlice("foo", "bar")}, `{"tags": ["foo", "bar"]}`),
		Entry("tags empty", ServiceOffering{Tags: types.NewOptionalStringSlice()}, `{"tags": []}`),
		Entry(
			"service broker guid",
			ServiceOffering{ServiceBrokerGUID: "fake-service-broker-guid"},
			`{
				"relationships": {
					"service_broker": {
						"data": {
							"guid": "fake-service-broker-guid"
						}
					}
				}
            }`,
		),
		Entry(
			"metadata",
			ServiceOffering{
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
			ServiceOffering{
				Name:              "fake-name",
				GUID:              "fake-guid",
				Description:       "once upon a time",
				DocumentationURL:  "https://docs.com",
				ServiceBrokerGUID: "fake-service-broker-guid",
				Metadata: &Metadata{
					Labels: map[string]types.NullString{
						"foo": types.NewNullString("bar"),
						"baz": types.NewNullString(),
					},
				},
			},
			`{
				"name": "fake-name",
				"guid": "fake-guid",
				"url": "https://fake-url.com",
				"description": "once upon a time",
				"documentation_url": "https://docs.com",
				"metadata": {
					"labels": {
						"foo": "bar",
						"baz": null
					}
				},
				"relationships": {
					"service_broker": {
						"data": {
							"guid": "fake-service-broker-guid"
						}
					}
				}
            }`,
		),
	)
})
