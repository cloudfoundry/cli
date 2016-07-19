package formatters_test

import (
	. "code.cloudfoundry.org/cli/cf/formatters"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("memoryLimit formatting", func() {
	It("returns 'unlimited' when limit is -1", func() {
		Expect(InstanceMemoryLimit(-1)).To(Equal("unlimited"))
	})

	It("formats original value to <val>M when limit is not -1", func() {
		Expect(InstanceMemoryLimit(100)).To(Equal("100M"))
	})
})
