package resources_test

import (
	"encoding/json"

	. "code.cloudfoundry.org/cli/v8/resources"
	"code.cloudfoundry.org/cli/v8/types"
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
		Entry("created_at", ServiceCredentialBinding{CreatedAt: "fake-created-at"}, `{"created_at": "fake-created-at"}`),
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
		Entry("created_at", ServiceCredentialBinding{CreatedAt: "fake-created-at"}, `{"created_at": "fake-created-at"}`),
		Entry(
			"strategy",
			ServiceCredentialBinding{
				Strategy: SingleBindingStrategy,
			},
			`{
				"strategy": "single"
			}`,
		),
		Entry(
			"everything",
			ServiceCredentialBinding{
				Type:                AppBinding,
				GUID:                "fake-guid",
				Name:                "fake-name",
				CreatedAt:           "fake-created-at",
				AppGUID:             "fake-app-guid",
				ServiceInstanceGUID: "fake-service-instance-guid",
				Parameters: types.NewOptionalObject(map[string]interface{}{
					"foo": "bar",
				}),
				Strategy:  MultipleBindingStrategy,
				CreatedAt: "fake-created-at",
			},
			`{
				"type": "app",
				"guid": "fake-guid",
				"name": "fake-name",
				"created_at": "fake-created-at",
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
				},
				"strategy": "multiple",
				"created_at": "fake-created-at"
			}`,
		),
	)
})
