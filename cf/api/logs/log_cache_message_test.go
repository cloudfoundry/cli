package logs_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("log messages", func() {

	Describe("ToSimpleLog", func() {
		It("returns the message", func() {
			Expect(true).To(Equal(false))
		})
	})

})
