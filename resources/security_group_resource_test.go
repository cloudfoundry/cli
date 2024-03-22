package resources_test

import (
	"encoding/json"

	. "code.cloudfoundry.org/cli/resources"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Security group resource", func() {
	var (
		portsString       = "some-Ports"
		descriptionString = "some-Description"
		typeInt           = 1
		codeInt           = 0
		logBool           = false
		trueBool          = true
		falseBool         = false
	)

	DescribeTable("UnmarshalJSON",
		func(givenBytes []byte, expectedStruct SecurityGroup) {
			var actualStruct SecurityGroup
			err := json.Unmarshal(givenBytes, &actualStruct)
			Expect(err).ToNot(HaveOccurred())
			Expect(actualStruct).To(Equal(expectedStruct))
		},
		Entry(
			"name, guid, and rules",
			[]byte(`{
				"name": "security-group-name",
				"guid": "security-group-guid",
				"rules":[
					{
						"protocol":"all",
						"destination":"some-Destination",
						"ports":"some-Ports",
						"type":1,
						"code":0,
						"description":"some-Description",
						"log":false
					}
				]
			}`),
			SecurityGroup{
				Name: "security-group-name",
				GUID: "security-group-guid",
				Rules: []Rule{
					{
						Protocol:    "all",
						Destination: "some-Destination",
						Ports:       &portsString,
						Type:        &typeInt,
						Code:        &codeInt,
						Description: &descriptionString,
						Log:         &logBool,
					},
				},
			},
		),
		Entry(
			"globally enabled",
			[]byte(`{
				"name": "security-group-name",
				"guid": "security-group-guid",
				"globally_enabled": {
					"running": true,
					"staging": false
				}
			}`),
			SecurityGroup{
				Name:                   "security-group-name",
				GUID:                   "security-group-guid",
				StagingGloballyEnabled: &falseBool,
				RunningGloballyEnabled: &trueBool,
			},
		),
		Entry(
			"relationships",
			[]byte(`{
				"name": "security-group-name",
				"guid": "security-group-guid",
				"relationships": {
					"staging_spaces": {
					  "data": [
						{ "guid": "space-guid-1" },
						{ "guid": "space-guid-2" }
					  ]
					},
					"running_spaces": {
					  "data": [
						{ "guid": "space-guid-3" }
					  ]
					}
				}
			}`),
			SecurityGroup{
				Name:              "security-group-name",
				GUID:              "security-group-guid",
				StagingSpaceGUIDs: []string{"space-guid-1", "space-guid-2"},
				RunningSpaceGUIDs: []string{"space-guid-3"},
			},
		),
	)

	DescribeTable("MarshalJSON",
		func(resource SecurityGroup, expected string) {
			actualBytes, err := json.Marshal(resource)
			Expect(err).ToNot(HaveOccurred())
			Expect(actualBytes).To(MatchJSON(expected))
		},
		Entry(
			"name and empty rules",
			SecurityGroup{
				Name:                   "security-group-name",
				GUID:                   "security-group-guid",
				RunningGloballyEnabled: &trueBool,
				StagingGloballyEnabled: &falseBool,
				Rules: []Rule{
					{
						Protocol:    "udp",
						Destination: "another-destination",
					},
				},
			},
			`{"globally_enabled":{"running":true,"staging":false},"guid":"security-group-guid","name":"security-group-name","rules":[{"protocol":"udp","destination":"another-destination"}]}`,
		),
		Entry(
			"only rules",
			SecurityGroup{
				Rules: []Rule{
					{
						Protocol:    "udp",
						Destination: "another-destination",
					},
				},
			},
			`{"rules":[{"protocol":"udp","destination":"another-destination"}]}`,
		),
		Entry(
			"name and rules",
			SecurityGroup{
				Name:                   "security-group-name",
				GUID:                   "security-group-guid",
				RunningGloballyEnabled: &trueBool,
				StagingGloballyEnabled: &falseBool,
				Rules: []Rule{
					{
						Protocol:    "all",
						Destination: "some-Destination",
						Ports:       &portsString,
						Type:        &typeInt,
						Code:        &codeInt,
						Description: &descriptionString,
						Log:         &logBool,
					},
				},
			},
			`{"globally_enabled":{"running":true,"staging":false},"guid":"security-group-guid","name":"security-group-name","rules":[{"protocol":"all","destination":"some-Destination","ports":"some-Ports","type":1,"code":0,"description":"some-Description","log":false}]}`,
		),
		Entry(
			"globally enabled",
			SecurityGroup{
				Name:                   "security-group-name",
				GUID:                   "security-group-guid",
				RunningGloballyEnabled: &trueBool,
				StagingGloballyEnabled: &falseBool,
				StagingSpaceGUIDs:      []string{"space-guid-1", "space-guid-2"},
				RunningSpaceGUIDs:      []string{"space-guid-3"},
			},
			`{"globally_enabled":{"running":true,"staging":false},"guid":"security-group-guid","name":"security-group-name","relationships":{"running_spaces":{"data":[{"guid":"space-guid-3"}]},"staging_spaces":{"data":[{"guid":"space-guid-1"},{"guid":"space-guid-2"}]}}}`,
		),
		Entry(
			"relationships",
			SecurityGroup{
				Name:                   "security-group-name",
				GUID:                   "security-group-guid",
				RunningGloballyEnabled: &trueBool,
				StagingGloballyEnabled: &falseBool,
				StagingSpaceGUIDs:      []string{"space-guid-1", "space-guid-2"},
				RunningSpaceGUIDs:      []string{"space-guid-3"},
			},
			`{"globally_enabled":{"running":true,"staging":false},"guid":"security-group-guid","name":"security-group-name","relationships":{"running_spaces":{"data":[{"guid":"space-guid-3"}]},"staging_spaces":{"data":[{"guid":"space-guid-1"},{"guid":"space-guid-2"}]}}}`,
		),
	)
})
