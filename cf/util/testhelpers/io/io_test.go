package io_test

import (
	"os"
	"strings"

	. "code.cloudfoundry.org/cli/cf/util/testhelpers/io"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("io helpers", func() {
	It("will never overflow the pipe", func() {
		str := strings.Repeat("z", 75000)
		output := CaptureOutput(func() {
			os.Stdout.Write([]byte(str))
		})

		Expect(output).To(Equal([]string{str}))
	})
})
