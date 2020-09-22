package resources_test

import (
	"encoding/json"

	. "code.cloudfoundry.org/cli/resources"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("service credential binding resource", func() {
	DescribeTable(
		"Marshaling and Unmarshaling",
		func(binding ServiceCredentialBinding, serialized string) {
			By("marshaling", func() {
				Expect(json.Marshal(binding)).To(MatchJSON(serialized))
			})

			By("unmarshaling", func() {
				var parsed ServiceCredentialBinding
				Expect(json.Unmarshal([]byte(serialized), &parsed)).NotTo(HaveOccurred())
				Expect(parsed).To(Equal(binding))
			})
		},
		Entry("empty", ServiceCredentialBinding{}, `{}`),
		Entry("type", ServiceCredentialBinding{Type: "fake-type"}, `{"type": "fake-type"}`),
		Entry("name", ServiceCredentialBinding{Name: "fake-name"}, `{"name": "fake-name"}`),
		Entry("guid", ServiceCredentialBinding{GUID: "fake-guid"}, `{"guid": "fake-guid"}`),
		Entry("service instance guid guid",
			ServiceCredentialBinding{ServiceInstanceGUID: "fake-instance-guid"},
			`{ "relationships": { "service_instance": { "data": { "guid": "fake-instance-guid" } } } }`,
		),
		Entry("App guid",
			ServiceCredentialBinding{AppGUID: "fake-app-guid"},
			`{ "relationships": { "app": { "data": { "guid": "fake-app-guid" } } } }`,
		),
		Entry(
			"everything",
			ServiceCredentialBinding{
				Type:                AppBinding,
				GUID:                "fake-guid",
				Name:                "fake-name",
				AppGUID:             "fake-app-guid",
				ServiceInstanceGUID: "fake-service-instance-guid",
			},
			`{
				"type": "app",
				"guid": "fake-guid",
				"name": "fake-name",
				"relationships": {
					"service_instance": {
						"data": {
							"guid": "fake-service-instance-guid"
						}
					},
					"app": {
						"data": {
							"guid": "fake-app-guid"
						}
					}
				}
			}`,
		),
	)
})
