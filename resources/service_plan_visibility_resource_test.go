package resources_test

import (
	"encoding/json"

	. "code.cloudfoundry.org/cli/resources"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("service plan visibility resource", func() {
	DescribeTable(
		"Marshaling and Unmarshaling",
		func(servicePlanVisibility ServicePlanVisibility, serialized string) {
			By("marshaling", func() {
				Expect(json.Marshal(servicePlanVisibility)).To(MatchJSON(serialized))
			})

			By("unmarshaling", func() {
				var parsed ServicePlanVisibility
				Expect(json.Unmarshal([]byte(serialized), &parsed)).NotTo(HaveOccurred())
				Expect(parsed).To(Equal(servicePlanVisibility))
			})
		},
		Entry("public", ServicePlanVisibility{Type: ServicePlanVisibilityPublic}, `{"type": "public"}`),
		Entry("admin", ServicePlanVisibility{Type: ServicePlanVisibilityAdmin}, `{"type": "admin"}`),
		Entry(
			"get space",
			ServicePlanVisibility{
				Type: ServicePlanVisibilitySpace,
				Space: ServicePlanVisibilityDetail{
					Name: "fake-space-name",
					GUID: "fake-space-guid",
				},
			},
			`{
				"type": "space",
				"space": {
					"name": "fake-space-name",
					"guid": "fake-space-guid"
				}
            }`,
		),
		Entry(
			"set space",
			ServicePlanVisibility{
				Type:  ServicePlanVisibilitySpace,
				Space: ServicePlanVisibilityDetail{GUID: "fake-space-guid"},
			},
			`{
				"type": "space",
				"space": {"guid": "fake-space-guid"}
            }`,
		),
		Entry(
			"get orgs",
			ServicePlanVisibility{
				Type: ServicePlanVisibilityOrganization,
				Organizations: []ServicePlanVisibilityDetail{
					{
						Name: "fake-org-1-name",
						GUID: "fake-org-1-guid",
					},
					{
						Name: "fake-org-2-name",
						GUID: "fake-org-2-guid",
					},
					{
						Name: "fake-org-3-name",
						GUID: "fake-org-3-guid",
					},
				},
			},
			`{
				"type": "organization",
				"organizations": [
					{
						"name": "fake-org-1-name",
						"guid": "fake-org-1-guid"
					},
					{
						"name": "fake-org-2-name",
						"guid": "fake-org-2-guid"
					},
					{
						"name": "fake-org-3-name",
						"guid": "fake-org-3-guid"
					}
				]
            }`,
		),
		Entry(
			"set orgs",
			ServicePlanVisibility{
				Type: ServicePlanVisibilityOrganization,
				Organizations: []ServicePlanVisibilityDetail{
					{GUID: "fake-org-1-guid"},
					{GUID: "fake-org-2-guid"},
					{GUID: "fake-org-3-guid"},
				},
			},
			`{
				"type": "organization",
				"organizations": [
					{"guid": "fake-org-1-guid"},
					{"guid": "fake-org-2-guid"},
					{"guid": "fake-org-3-guid"}
				]
            }`,
		),
	)
})
