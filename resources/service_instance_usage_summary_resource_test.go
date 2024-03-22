package resources_test

import (
	"encoding/json"

	. "code.cloudfoundry.org/cli/resources"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("service instance usage resource", func() {
	DescribeTable(
		"Unmarshaling",
		func(serviceInstanceUsageSummaryList ServiceInstanceUsageSummaryList, serialized string) {
			var parsed ServiceInstanceUsageSummaryList
			Expect(json.Unmarshal([]byte(serialized), &parsed)).NotTo(HaveOccurred())
			Expect(parsed).To(Equal(serviceInstanceUsageSummaryList))
		},
		Entry("space_guid", ServiceInstanceUsageSummaryList{UsageSummary: []ServiceInstanceUsageSummary{{SpaceGUID: "fake-space-guid"}}}, `{"usage_summary":[{"space": {"guid": "fake-space-guid"}}]}`),
		Entry("bound_app_count", ServiceInstanceUsageSummaryList{UsageSummary: []ServiceInstanceUsageSummary{{BoundAppCount: 2}}}, `{"usage_summary":[{"bound_app_count": 2}]}`),
		Entry(
			"everything",
			ServiceInstanceUsageSummaryList{
				UsageSummary: []ServiceInstanceUsageSummary{
					{
						SpaceGUID:     "fake-space-guid",
						BoundAppCount: 4,
					},
					{
						SpaceGUID:     "other-fake-space-guid",
						BoundAppCount: 3,
					},
				},
			},
			`{
				"usage_summary": [
					{
					   "space": {
							"guid": "fake-space-guid"
						},
						"bound_app_count": 4
					},
				   {
						"space": {
							"guid": "other-fake-space-guid"
						},
						"bound_app_count": 3
					}
				]
			}`,
		),
	)
})
