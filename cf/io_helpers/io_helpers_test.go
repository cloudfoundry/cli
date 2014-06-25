package io_helpers_test

import (
	"os"

	. "github.com/cloudfoundry/cli/cf/io_helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("io helpers", func() {
	It("will never overflow the pipe", func() {
		str := ""
		for i := 0; i < 75000; i++ {
			str += "abc"
		}

		output := CaptureOutput(func() {
			os.Stdout.Write([]byte(str))
		})

		Expect(output).To(Equal([]string{str}))
	})
})
