package utils_test

import (
	"fmt"

	. "github.com/cloudfoundry/cli/utils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ComputeFileSha1", func() {

	It("returns error if file does not exist", func() {
		path := "file/path/to/no/where"

		sha1, err := ComputeFileSha1(path)
		Ω(len(sha1)).To(Equal(0))
		Ω(err).To(HaveOccurred())
	})

	It("returns the sha1 of a file", func() {
		path := "../fixtures/plugins/test_1.exe"

		sha1, err := ComputeFileSha1(path)
		Ω(err).ToNot(HaveOccurred())
		Ω(fmt.Sprintf("%x", sha1)).To(Equal("025b882e665e3fee35653095812bcbd384efd56f"))
	})

})
