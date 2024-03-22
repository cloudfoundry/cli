package flag_test

import (
	. "code.cloudfoundry.org/cli/command/flag"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("OptionalString", func() {
	var optionalString OptionalString

	BeforeEach(func() {
		optionalString = OptionalString{}
	})

	Describe("default value", func() {
		It("is unset by default", func() {
			Expect(optionalString.IsSet).To(BeFalse())
		})

		It("has an empty value", func() {
			Expect(optionalString.Value).To(BeEmpty())
		})
	})

	Describe("UnmarshalFlag", func() {
		BeforeEach(func() {
			err := optionalString.UnmarshalFlag("some string")
			Expect(err).NotTo(HaveOccurred())
		})

		It("is set", func() {
			Expect(optionalString.IsSet).To(BeTrue())
		})

		It("has the right value", func() {
			Expect(optionalString.Value).To(Equal("some string"))
		})
	})
})
