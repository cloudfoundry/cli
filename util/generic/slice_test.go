package generic_test

import (
	"code.cloudfoundry.org/cli/util/generic"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("IsSliceable", func() {
	It("returns false if the type is nil", func() {
		Expect(generic.IsSliceable(nil)).To(BeFalse())
	})

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
