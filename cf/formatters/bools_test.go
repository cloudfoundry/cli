package formatters_test

import (
	. "github.com/cloudfoundry/cli/cf/formatters"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("bool formatting", func() {
	Describe("Allowed", func() {
		It("is 'allowed' when true", func() {
			Expect(Allowed(true)).To(Equal("allowed"))
		})

		It("is 'disallowed' when false", func() {
			Expect(Allowed(false)).To(Equal("disallowed"))
		})
	})
})
