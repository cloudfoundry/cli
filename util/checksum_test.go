package util_test

import (
	"fmt"
	"io/ioutil"
	"os"

	. "code.cloudfoundry.org/cli/util"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Sha1Checksum", func() {

	var (
		checksum Sha1Checksum
	)

	Describe("ComputeFileSha1", func() {
		Context("If file does not exist", func() {
			It("returns error", func() {
				checksum = NewSha1Checksum("file/path/to/no/where")

				sha1, err := checksum.ComputeFileSha1()
				Expect(len(sha1)).To(Equal(0))
				Expect(err).To(HaveOccurred())
			})
		})

		Context("If file does exist", func() {
			var (
				f   *os.File
				err error
			)

			BeforeEach(func() {
				f, err = ioutil.TempFile("", "sha1_test_")
				Expect(err).NotTo(HaveOccurred())
				defer f.Close()
				f.Write([]byte("abc"))
			})

			AfterEach(func() {
				os.RemoveAll(f.Name())
			})

			It("returns the sha1 of a file", func() {
				checksum = NewSha1Checksum(f.Name())

				sha1, err := checksum.ComputeFileSha1()
				Expect(err).NotTo(HaveOccurred())
				Expect(len(sha1)).To(Equal(20))
				Expect(fmt.Sprintf("%x", sha1)).To(Equal("a9993e364706816aba3e25717850c26c9cd0d89d"))
			})
		})

	})

	Describe("CheckSha1", func() {
		Context("file doesn't exist", func() {
			It("returns false", func() {
				checksum = NewSha1Checksum("file/path/to/no/where")

				sha1, err := checksum.ComputeFileSha1()
				Expect(len(sha1)).To(Equal(0))
				Expect(err).To(HaveOccurred())
			})
		})

		Context("If file does exist", func() {
			var (
				f   *os.File
				err error
			)

			BeforeEach(func() {
				f, err = ioutil.TempFile("", "sha1_test_")
				Expect(err).NotTo(HaveOccurred())
				defer f.Close()
				f.Write([]byte("abc"))
			})

			AfterEach(func() {
				os.RemoveAll(f.Name())
			})

			It("returns false if sha1 doesn't match", func() {
				checksum = NewSha1Checksum(f.Name())

				Expect(checksum.CheckSha1("skj33933dabs2292391223aa393fjs92")).To(BeFalse())
			})

			It("returns true if sha1 matches", func() {
				checksum = NewSha1Checksum(f.Name())

				Expect(checksum.CheckSha1("a9993e364706816aba3e25717850c26c9cd0d89d")).To(BeTrue())
			})
		})

	})
})
