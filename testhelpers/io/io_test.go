package io_test

import (
	"os"
	"strings"

	. "github.com/cloudfoundry/cli/testhelpers/io"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("io helpers", func() {
	It("will never overflow the pipe", func() {
		characters := make([]string, 0, 75000)
		for i := 0; i < 75000; i++ {
			characters = append(characters, "z")
		}

		str := strings.Join(characters, "")

		output := CaptureOutput(func() {
			os.Stdout.Write([]byte(str))
		})

		Expect(output).To(Equal([]string{str}))
	})
})
