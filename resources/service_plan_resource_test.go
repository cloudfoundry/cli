package resources_test

import (
	"encoding/json"

	. "code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("service plan resource", func() {
	DescribeTable(
		"Unmarshaling",
		func(servicePlan ServicePlan, serialized string) {
			var parsed ServicePlan
			Expect(json.Unmarshal([]byte(serialized), &parsed)).NotTo(HaveOccurred())
			Expect(parsed).To(Equal(servicePlan))
		},
		Entry(
			"basic",
			ServicePlan{
				GUID:                "fake-service-plan-guid",
				Name:                "fake-service-plan-name",
				ServiceOfferingGUID: "fake-service-offering-guid",
			},
			`{
				"guid": "fake-service-plan-guid",
				"name": "fake-service-plan-name",
				"relationships": {
					"service_offering": {
						"data": {
							"guid": "fake-service-offering-guid"
						}
					}
				}
			}`,
		),
		Entry(
			"detailed",
			ServicePlan{
				GUID:                "fake-service-plan-guid",
				Name:                "fake-service-plan-name",
				ServiceOfferingGUID: "fake-service-offering-guid",
				Description:         "fake-description",
				Available:           true,
				VisibilityType:      "public",
				Free:                true,
				Costs: []ServicePlanCost{
					{
						Amount:   12.5,
						Currency: "USD",
						Unit:     "month",
					},
					{
						Amount:   3.25,
						Currency: "EUR",
						Unit:     "day",
					},
				},
				SpaceGUID:                  "fake-space-guid",
				MaintenanceInfoDescription: "cool upgrade",
				MaintenanceInfoVersion:     "1.2.3",
				Metadata: &Metadata{
					Labels: map[string]types.NullString{
						"foo": types.NewNullString("bar"),
						"baz": types.NewNullString(),
					},
				},
			},
			`{
				"guid": "fake-service-plan-guid",
				"name": "fake-service-plan-name",
				"description": "fake-description",
				"available": true,
				"visibility_type": "public",
				"free": true,
				"costs": [
					{
						"amount": 12.5,
						"currency": "USD",
						"unit": "month"
					},
					{
						"amount": 3.25,
						"currency": "EUR",
						"unit": "day"
					}
				],
				"maintenance_info": {
					"description": "cool upgrade",
					"version": "1.2.3"
				},
				"metadata": {
					"labels": {
						"foo": "bar",
						"baz": null
					}
				},
				"relationships": {
					"service_offering": {
						"data": {
							"guid": "fake-service-offering-guid"
						}
					},
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
