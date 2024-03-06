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
		Entry("description", ServiceOffering{Description: "once upon a time"}, `{"description": "once upon a time"}`),
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
