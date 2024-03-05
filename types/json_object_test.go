package types_test

import (
	"encoding/json"

	. "code.cloudfoundry.org/cli/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("JSONObject", func() {
	It("marshals as JSON", func() {
		serialized, err := json.Marshal(JSONObject{"foo": "bar"})
		Expect(err).NotTo(HaveOccurred())
		Expect(serialized).To(Equal([]byte(`{"foo":"bar"}`)))
	})

	When("nil", func() {
		It("marshals as an empty object (not as null)", func() {
			var o JSONObject
			serialized, err := json.Marshal(o)
			Expect(err).NotTo(HaveOccurred())
			Expect(serialized).To(Equal([]byte(`{}`)))
		})
	})
})
