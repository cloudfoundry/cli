package formatters_test

import (
	. "github.com/cloudfoundry/cli/cf/formatters"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("memoryLimit formatting", func() {
	It("returns 'Unlimited' when limit is -1", func() {
		Expect(InstanceMemoryLimit(-1)).To(Equal("Unlimited"))
	})

	It("formats original value to <val>M when limit is not -1", func() {
		Expect(InstanceMemoryLimit(100)).To(Equal("100M"))
	})
})
