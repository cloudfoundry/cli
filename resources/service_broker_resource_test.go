package resources_test

import (
	"encoding/json"

	. "code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("service broker resource", func() {
	DescribeTable(
		"Marshaling and Unmarshaling",
		func(serviceInstance ServiceBroker, serialized string) {
			By("marshaling", func() {
				Expect(json.Marshal(serviceInstance)).To(MatchJSON(serialized))
			})

			By("unmarshaling", func() {
				var parsed ServiceBroker
				Expect(json.Unmarshal([]byte(serialized), &parsed)).NotTo(HaveOccurred())
				Expect(parsed).To(Equal(serviceInstance))
			})
		},
		Entry("name", ServiceBroker{Name: "fake-name"}, `{"name": "fake-name"}`),
		Entry("guid", ServiceBroker{GUID: "fake-guid"}, `{"guid": "fake-guid"}`),
		Entry("url", ServiceBroker{URL: "https://fake-url.com"}, `{"url": "https://fake-url.com"}`),
		Entry(
			"space guid",
			ServiceBroker{SpaceGUID: "fake-space-guid"},
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
			"authentication",
			ServiceBroker{
				CredentialsType: ServiceBrokerBasicCredentials,
				Username:        "fake-username",
				Password:        "fake-password",
			},
			`{
				"authentication": {
					"type": "basic",
					"credentials": {
						"username": "fake-username",
						"password": "fake-password"
					}
				}
			}`,
		),
		Entry(
			"metadata",
			ServiceBroker{
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
			ServiceBroker{
				Name:            "fake-name",
				GUID:            "fake-guid",
				URL:             "https://fake-url.com",
				SpaceGUID:       "fake-space-guid",
				CredentialsType: ServiceBrokerBasicCredentials,
				Username:        "fake-username",
				Password:        "fake-password",
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
				"authentication": {
					"type": "basic",
					"credentials": {
						"username": "fake-username",
						"password": "fake-password"
					}
				},
				"metadata": {
					"labels": {
						"foo": "bar",
						"baz": null
					}
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
