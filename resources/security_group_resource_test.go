package resources_test

import (
	"encoding/json"

	. "code.cloudfoundry.org/cli/resources"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Security group resource", func() {
	DescribeTable("UnmarshalJSON",
		func(givenBytes []byte, expectedStruct SecurityGroup) {
			var actualStruct SecurityGroup
			err := json.Unmarshal(givenBytes, &actualStruct)
			Expect(err).ToNot(HaveOccurred())
			Expect(actualStruct).To(Equal(expectedStruct))
		},
		Entry(
			"name and guid",
			[]byte(`{"name": "security-group-name","guid": "security-group-guid"}`),
			SecurityGroup{
				Name: "security-group-name",
				GUID: "security-group-guid",
			}),
	)

	DescribeTable("MarshalJSON",
		func(resource SecurityGroup, expectedBytes []byte) {
			bytes, err := json.Marshal(resource)
			Expect(err).ToNot(HaveOccurred())
			Expect(bytes).To(Equal(expectedBytes))
		},
		Entry(
			"name and guid",
			SecurityGroup{
				Name: "security-group-name",
				GUID: "security-group-guid",
			},
			[]byte(`{"name":"security-group-name","guid":"security-group-guid"}`),
		),

		Entry(
			"name only",
			SecurityGroup{
				Name: "security-group-name",
			},
			[]byte(`{"name":"security-group-name"}`),
		),
	)
})
