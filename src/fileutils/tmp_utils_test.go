package fileutils_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
	"fileutils"
)


func TestFileutils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Fileutils Suite")
}

var _ = Describe("tmp_utils", func() {
	Describe("SetTmpPathPrefix", func() {
		It("setting tmp path prefix should actually set it", func() {
			prefix := "howdie"
			fileutils.SetTmpPathPrefix(prefix)
			Expect(prefix).To(Equal(fileutils.TmpPathPrefix()))
		})
        })
})
