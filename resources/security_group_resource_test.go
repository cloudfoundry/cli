package resources_test

import (
	"encoding/json"

	. "code.cloudfoundry.org/cli/resources"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Security group resource", func() {
	var (
		portsString       = "some-Ports"
		descriptionString = "some-Description"
		typeInt           = 1
		codeInt           = 0
		logBool           = false
	)

	DescribeTable("UnmarshalJSON",
		func(givenBytes []byte, expectedStruct SecurityGroup) {
			var actualStruct SecurityGroup
			err := json.Unmarshal(givenBytes, &actualStruct)
			Expect(err).ToNot(HaveOccurred())
			Expect(actualStruct).To(Equal(expectedStruct))
		},
		Entry(
			"name and guid",
			[]byte(`{"name": "security-group-name","guid": "security-group-guid","rules":[]}`),
			SecurityGroup{
				Name:  "security-group-name",
				GUID:  "security-group-guid",
				Rules: []Rule{},
			},
		),
		Entry(
			"full rules",
			[]byte(`{"name": "security-group-name",
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
]}`),
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
			"relationships",
			[]byte(`{
				"name": "security-group-name",
				"guid": "security-group-guid",
				"rules": [],
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
				Rules:             []Rule{},
				StagingSpaceGUIDs: []string{"space-guid-1", "space-guid-2"},
				RunningSpaceGUIDs: []string{"space-guid-3"},
			},
		),
	)

	DescribeTable("MarshalJSON",
		func(resource SecurityGroup, expectedBytes []byte) {
			actualBytes, err := json.Marshal(resource)
			Expect(err).ToNot(HaveOccurred())
			Expect(actualBytes).To(Equal(expectedBytes))
		},

		Entry(
			"name and empty rules",
			SecurityGroup{
				Name: "security-group-name",
				GUID: "security-group-guid",
				Rules: []Rule{
					{
						Protocol:    "udp",
						Destination: "another-destination",
					},
				},
			},
			[]byte(`{"guid":"security-group-guid","name":"security-group-name","rules":[{"protocol":"udp","destination":"another-destination"}]}`),
		),
		Entry(
			"name and rules",
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
			[]byte(`{"guid":"security-group-guid","name":"security-group-name","rules":[{"protocol":"all","destination":"some-Destination","ports":"some-Ports","type":1,"code":0,"description":"some-Description","log":false}]}`),
		),
		Entry(
			"relationships",
			SecurityGroup{
				Name:              "security-group-name",
				GUID:              "security-group-guid",
				StagingSpaceGUIDs: []string{"space-guid-1", "space-guid-2"},
				RunningSpaceGUIDs: []string{"space-guid-3"},
			},
			[]byte(`{"guid":"security-group-guid","name":"security-group-name","relationships":{"running_spaces":{"data":[{"guid":"space-guid-3"}]},"staging_spaces":{"data":[{"guid":"space-guid-1"},{"guid":"space-guid-2"}]}},"rules":[]}`),
		),
	)
})
