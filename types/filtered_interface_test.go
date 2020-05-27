package types_test

import (
	. "code.cloudfoundry.org/cli/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("FilteredInterface", func() {
	var nullInterface FilteredInterface

	Describe("UnmarshalJSON", func() {
		When("a string value is provided", func() {
			It("parses a out a valid FilteredInterface", func() {
				err := nullInterface.UnmarshalJSON([]byte(`"some-string"`))
				Expect(err).ToNot(HaveOccurred())
				Expect(nullInterface).To(Equal(FilteredInterface{Value: "some-string", IsSet: true}))
			})
		})

		When("an empty string value is provided", func() {
			It("parses a out a valid FilteredInterface", func() {
				err := nullInterface.UnmarshalJSON([]byte(`""`))
				Expect(err).ToNot(HaveOccurred())
				Expect(nullInterface).To(Equal(FilteredInterface{Value: "", IsSet: true}))
			})
		})

		When("an integer value is provided", func() {
			It("returns a string", func() {
				err := nullInterface.UnmarshalJSON([]byte(`28`))
				Expect(err).ToNot(HaveOccurred())
				Expect(nullInterface).To(Equal(FilteredInterface{Value: float64(28), IsSet: true}))
			})
		})

		When("a FilteredInterface is set to some boolean", func() {
			It("returns a string", func() {
				err := nullInterface.UnmarshalJSON([]byte(`true`))
				Expect(err).ToNot(HaveOccurred())
				Expect(nullInterface).To(Equal(FilteredInterface{Value: true, IsSet: true}))
			})
		})

		When("JSON values provided", func() {
			It("parses a out a valid FilteredInterface", func() {
				err := nullInterface.UnmarshalJSON([]byte(`{"json":"{\"key\":\"value\"}"}`))
				Expect(err).ToNot(HaveOccurred())
				Expect(nullInterface).To(Equal(FilteredInterface{
					Value: map[string]interface{}{"json": "{\"key\":\"value\"}"},
					IsSet: true}))
			})
		})

		When("an empty JSON is provided", func() {
			It("parses a out a valid FilteredInterface", func() {
				err := nullInterface.UnmarshalJSON([]byte("null"))
				Expect(err).ToNot(HaveOccurred())
				Expect(nullInterface).To(Equal(FilteredInterface{Value: nil, IsSet: true}))
			})
		})
	})

	Describe("MarshalJSON", func() {
		When("a FilteredInterface is set to some string", func() {
			It("returns a string", func() {
				nullInterface = FilteredInterface{Value: "some-string", IsSet: true}
				marshalled, err := nullInterface.MarshalJSON()
				Expect(err).ToNot(HaveOccurred())
				Expect(marshalled).To(Equal([]byte(`"some-string"`)))
			})
		})

		When("a FilteredInterface is set to some number", func() {
			It("returns a string", func() {
				nullInterface = FilteredInterface{Value: 28, IsSet: true}
				marshalled, err := nullInterface.MarshalJSON()
				Expect(err).ToNot(HaveOccurred())
				Expect(marshalled).To(Equal([]byte(`28`)))
			})
		})

		When("a FilteredInterface is set to some boolean", func() {
			It("returns a string", func() {
				nullInterface = FilteredInterface{Value: true, IsSet: true}
				marshalled, err := nullInterface.MarshalJSON()
				Expect(err).ToNot(HaveOccurred())
				Expect(marshalled).To(Equal([]byte(`true`)))
			})
		})

		When("a FilteredInterface is set to some object", func() {
			It("returns a flattened string", func() {
				nullInterface = FilteredInterface{
					Value: map[string]interface{}{
						"json": map[string]interface{}{
							"key": "value",
						},
					},
					IsSet: true,
				}
				marshalled, err := nullInterface.MarshalJSON()
				Expect(err).ToNot(HaveOccurred())
				Expect(marshalled).To(Equal([]byte(`{"json":{"key":"value"}}`)))
			})
		})

		When("a FilteredInterface is not set", func() {
			It("returns null", func() {
				nullInterface = FilteredInterface{IsSet: false}
				marshalled, err := nullInterface.MarshalJSON()
				Expect(err).ToNot(HaveOccurred())
				Expect(marshalled).To(Equal([]byte(`null`)))
			})
		})
	})
})
