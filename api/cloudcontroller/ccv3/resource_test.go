package ccv3_test

import (
	"encoding/json"

	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Resource", func() {
	Describe("MarshalJSON", func() {
		It("marshals the json properly", func() {
			resource := Resource{
				Filename: "some-file-1",
				Mode:     0744,
				SHA1:     "some-sha-1",
				Size:     1,
			}
			data, err := json.Marshal(resource)
			Expect(err).ToNot(HaveOccurred())
			Expect(data).To(MatchJSON(`{
				"fn":   "some-file-1",
				"mode": "744",
				"sha1": "some-sha-1",
				"size": 1
			}`))
		})
	})

	Describe("MarshalJSON", func() {
		It("marshals the json properly", func() {
			raw := `{
				"fn":   "some-file-1",
				"mode": "744",
				"sha1": "some-sha-1",
				"size": 1
			}`

			var data Resource
			err := json.Unmarshal([]byte(raw), &data)
			Expect(err).ToNot(HaveOccurred())
			Expect(data).To(Equal(Resource{
				Filename: "some-file-1",
				Mode:     0744,
				SHA1:     "some-sha-1",
				Size:     1,
			}))
		})
	})
})
