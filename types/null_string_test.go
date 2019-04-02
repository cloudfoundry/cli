package types_test

import (
	"encoding/json"

	. "code.cloudfoundry.org/cli/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("NullString", func() {
	type SamplePayload struct {
		OptionalField NullString
	}

	Context("JSON marshaling", func() {
		When("the NullString has a value", func() {
			It("marshals to its value", func() {
				toMarshal := SamplePayload{
					OptionalField: NullString{Value: "some-string", IsSet: true},
				}
				bytes, err := json.Marshal(toMarshal)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(bytes)).To(ContainSubstring(`"some-string"`))
			})
		})

		When("the NullString has no value", func() {
			It("marshals to a JSON null", func() {
				toMarshal := SamplePayload{
					OptionalField: NullString{IsSet: false},
				}
				bytes, err := json.Marshal(toMarshal)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(bytes)).To(ContainSubstring(`null`))
			})
		})
	})

	Context("JSON unmashalling", func() {
		When("the JSON has a string value", func() {
			It("unmarhsalls to a NullString with the correct value", func() {
				toUnmarshal := []byte(`{"optionalField":"our-value"}`)
				var samplePayload SamplePayload
				err := json.Unmarshal(toUnmarshal, &samplePayload)
				Expect(err).ToNot(HaveOccurred())
				Expect(samplePayload.OptionalField.IsSet).To(BeTrue())
				Expect(samplePayload.OptionalField.Value).To(Equal("our-value"))
			})
		})

		When("the JSON has a null value", func() {
			It("unmarhsalls to a NullString with no value", func() {
				toUnmarshal := []byte(`{"optionalField":null}`)
				var samplePayload SamplePayload
				err := json.Unmarshal(toUnmarshal, &samplePayload)
				Expect(err).ToNot(HaveOccurred())
				Expect(samplePayload.OptionalField.IsSet).To(BeFalse())
				Expect(samplePayload.OptionalField.Value).To(Equal(""))
			})
		})
	})

})
