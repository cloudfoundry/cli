package lookuptable_test

import (
	"code.cloudfoundry.org/cli/util/lookuptable"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("NameFromGuid", func() {
	It("builds a lookup table", func() {
		input := []struct{ Name, GUID string }{
			{Name: "foo-name", GUID: "foo-guid"},
			{Name: "bar-name", GUID: "bar-guid"},
			{Name: "baz-name", GUID: "baz-guid"},
			{Name: "quz-name", GUID: "quz-guid"},
		}

		Expect(lookuptable.NameFromGUID(input)).To(Equal(map[string]string{
			"foo-guid": "foo-name",
			"bar-guid": "bar-name",
			"baz-guid": "baz-name",
			"quz-guid": "quz-name",
		}))
	})

	When("the input is not a slice", func() {
		It("returns nil", func() {
			input := struct{ Name, GUID string }{
				Name: "foo-name",
				GUID: "foo-guid",
			}

			Expect(lookuptable.NameFromGUID(input)).To(BeNil())
		})
	})

	When("the elements are not structs", func() {
		It("returns nil", func() {
			input := []string{"foo", "bar"}
			Expect(lookuptable.NameFromGUID(input)).To(BeNil())
		})
	})
})
