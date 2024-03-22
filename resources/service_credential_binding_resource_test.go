package resources_test

import (
	"encoding/json"

	. "code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
	. "github.com/onsi/ginkgo/v2"
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
		Entry("app guid",
			ServiceCredentialBinding{AppGUID: "fake-app-guid"},
			`{ "relationships": { "app": { "data": { "guid": "fake-app-guid" } } } }`,
		),
		Entry(
			"last operation",
			ServiceCredentialBinding{
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
			"parameters",
			ServiceCredentialBinding{
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
			"everything",
			ServiceCredentialBinding{
				Type:                AppBinding,
				GUID:                "fake-guid",
				Name:                "fake-name",
				AppGUID:             "fake-app-guid",
				ServiceInstanceGUID: "fake-service-instance-guid",
				Parameters: types.NewOptionalObject(map[string]interface{}{
					"foo": "bar",
				}),
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
				},
				"parameters": {
					"foo": "bar"
				}
			}`,
		),
	)
})
