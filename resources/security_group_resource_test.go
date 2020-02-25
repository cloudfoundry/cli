package resources_test

import (
	"encoding/json"

	. "code.cloudfoundry.org/cli/resources"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Security group resource", func() {
	var portsString = "some-Ports"
	var descriptionString = "some-Description"
	var typeInt = 1
	var codeInt = 0
	var logBool = false
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
			"Full rules",
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
	)

	DescribeTable("MarshalJSON",
		func(resource SecurityGroup, expectedBytes []byte) {
			actualBytes, err := json.Marshal(resource)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(actualBytes)).To(Equal(string(expectedBytes)))
		},

		Entry(
			"name and empty rules",
			SecurityGroup{
				Name: "security-group-name",
				Rules: []Rule{
					{
						Protocol:    "udp",
						Destination: "another-destination",
					},
				},
			},
			[]byte(`{"name":"security-group-name","rules":[{"protocol":"udp","destination":"another-destination"}]}`),
		),
		Entry(
			"name and rules",
			SecurityGroup{
				Name: "security-group-name",
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
			[]byte(`{"name":"security-group-name","rules":[{"protocol":"all","destination":"some-Destination","ports":"some-Ports","type":1,"code":0,"description":"some-Description","log":false}]}`),
		),
	)
})
