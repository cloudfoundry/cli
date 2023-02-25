package resources_test

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"

	. "code.cloudfoundry.org/cli/resources"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("shared to spaces resource", func() {
	DescribeTable(
		"Unmarshaling",
		func(sharedToSpaces SharedToSpacesListWrapper, serialized string) {
			var parsed SharedToSpacesListWrapper
			Expect(json.Unmarshal([]byte(serialized), &parsed)).NotTo(HaveOccurred())
			Expect(parsed).To(Equal(sharedToSpaces))
		},
		Entry("SharedToSpaceGUIDs", SharedToSpacesListWrapper{SharedToSpaceGUIDs: []string{"fake-space-guid", "other-fake-space-guid"}}, `{"data": [{"guid": "fake-space-guid"}, {"guid": "other-fake-space-guid"}]}`),
		Entry("Spaces", SharedToSpacesListWrapper{
			Spaces: []Space{
				{
					GUID: "fake-space-guid",
					Name: "fake-space-name",
					Relationships: map[constant.RelationshipType]Relationship{
						"organization": Relationship{
							GUID: "some-org-guid",
						},
					},
				},
			},
		}, `{
				"included": {
					"spaces": [
						{ 
							"name": "fake-space-name",
							"guid": "fake-space-guid", 
							"relationships": {
							   "organization": {
								  "data": {
									 "guid": "some-org-guid"
								  }
							   }
							}
						}
					]
				}
			}`),
		Entry("Organizations", SharedToSpacesListWrapper{
			Organizations: []Organization{
				{
					GUID: "fake-org-guid",
					Name: "fake-org-name",
				},
			},
		}, `{
				"included": {
					"organizations": [
						{ 
							"name": "fake-org-name",
							"guid": "fake-org-guid"
						}
					]
				}
			}`),
		Entry(
			"everything",
			SharedToSpacesListWrapper{
				SharedToSpaceGUIDs: []string{"fake-space-guid", "other-fake-space-guid"},
				Spaces: []Space{
					{
						GUID:          "fake-space-guid",
						Name:          "fake-space-name",
						Relationships: map[constant.RelationshipType]Relationship{"organization": Relationship{GUID: "fake-org-guid"}}},
					{
						GUID:          "other-fake-space-guid",
						Name:          "other-fake-space-name",
						Relationships: map[constant.RelationshipType]Relationship{"organization": Relationship{GUID: "fake-org-guid"}}}},
				Organizations: []Organization{
					{
						GUID: "fake-org-guid",
						Name: "fake-org-name",
					},
				}},

			`{
				"data": [{"guid":"fake-space-guid"},{"guid":"other-fake-space-guid"}],
				"links": {
					"self": {
						"href": "https://some-url/v3/service_instances/7915bc51-8203-4758-b0e2-f77bfcdc38cb/relationships/shared_spaces"
					}
				},
				"included": {
					"spaces": [
						{
							"name": "fake-space-name",
							"guid": "fake-space-guid",
							"relationships": {
								"organization": {
									"data": {
										"guid": "fake-org-guid"
									}
								}
							}
						},
						{
							"name": "other-fake-space-name",
							"guid": "other-fake-space-guid",
							"relationships": {
								"organization": {
									"data": {
										"guid": "fake-org-guid"
									}
								}
							}
						}
					],
					"organizations": [
						{
							"name": "fake-org-name",
							"guid": "fake-org-guid"
						}
					]
				}
			}`,
		),
	)
})
