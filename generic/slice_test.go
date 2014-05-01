package generic_test

import (
	"github.com/cloudfoundry/cli/generic"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func init() {
	Describe("IsSliceable", func() {
		It("should return false when the type cannot be sliced", func() {
			Expect(generic.IsSliceable("bad slicing")).To(BeFalse())
		})

		It("should return true if the type can be sliced", func() {
			Expect(generic.IsSliceable([]string{"a string"})).To(BeTrue())
		})

		It("should return true if the type can be sliced", func() {
			Expect(generic.IsSliceable([]interface{}{1, 2, 3})).To(BeTrue())
		})
	})
}
