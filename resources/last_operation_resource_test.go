package resources_test

import (
	"encoding/json"

	. "code.cloudfoundry.org/cli/resources"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("last operation resource", func() {
	DescribeTable(
		"Marshaling and Unmarshaling",
		func(lastOperation LastOperation, serialized string) {
			By("marshaling", func() {
				Expect(json.Marshal(lastOperation)).To(MatchJSON(serialized))
			})

			By("unmarshaling", func() {
				var parsed LastOperation
				Expect(json.Unmarshal([]byte(serialized), &parsed)).NotTo(HaveOccurred())
				Expect(parsed).To(Equal(lastOperation))
			})
		},
		Entry("type", LastOperation{Type: "fake-type"}, `{"type": "fake-type"}`),
		Entry("state", LastOperation{State: "fake-state"}, `{"state": "fake-state"}`),
		Entry("description", LastOperation{Description: "fake-description"}, `{"description": "fake-description"}`),
		Entry("created_at", LastOperation{CreatedAt: "fake-created-at"}, `{"created_at": "fake-created-at"}`),
		Entry("updated_at", LastOperation{UpdatedAt: "fake-updated-at"}, `{"updated_at": "fake-updated-at"}`),
		Entry(
			"everything",
			LastOperation{
				Type:        CreateOperation,
				State:       OperationInProgress,
				Description: "doing stuff",
				CreatedAt:   "yesterday",
				UpdatedAt:   "just now",
			},
			`{
				"type": "create",
				"state": "in progress",
				"description": "doing stuff",
				"created_at": "yesterday",
				"updated_at": "just now"
            }`,
		),
	)

	DescribeTable(
		"OmitJSONry",
		func(lastOperation LastOperation, expected bool) {
			Expect(lastOperation.OmitJSONry()).To(Equal(expected))
		},
		Entry("empty object", LastOperation{}, true),
		Entry("type", LastOperation{Type: CreateOperation}, false),
		Entry("state", LastOperation{State: OperationInProgress}, false),
		Entry("created_at", LastOperation{CreatedAt: "now"}, false),
		Entry("updated_at", LastOperation{UpdatedAt: "now"}, false),
	)
})
