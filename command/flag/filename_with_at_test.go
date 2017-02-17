package flag_test

import (
	"io/ioutil"
	"os"

	. "code.cloudfoundry.org/cli/command/flag"
	flags "github.com/jessevdk/go-flags"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("FilenameWithAt", func() {
	var filenameWithAt FilenameWithAt

	Describe("Complete", func() {
		BeforeEach(func() {
			filenameWithAt = FilenameWithAt("")
		})

		Context("when the match is empty", func() {
			It("returns no completions", func() {
				Expect(filenameWithAt.Complete("")).To(BeEmpty())
			})
		})

		Context("when the match doesn't start with @", func() {
			It("returns no completions", func() {
				Expect(filenameWithAt.Complete("a@b")).To(BeEmpty())
			})
		})

		Context("when the match starts with @", func() {
			var tempDir string

			BeforeEach(func() {
				var err error

				tempDir, err = ioutil.TempDir("", "")
				Expect(err).ToNot(HaveOccurred())

				err = os.Chdir(tempDir)
				Expect(err).ToNot(HaveOccurred())

				for _, filename := range []string{"abc", "abd", "efg"} {
					err = ioutil.WriteFile(filename, []byte{}, 0400)
					Expect(err).ToNot(HaveOccurred())
				}
			})

			AfterEach(func() {
				err := os.RemoveAll(tempDir)
				Expect(err).ToNot(HaveOccurred())
			})

			Context("when there are no characters after the @", func() {
				It("returns all files", func() {
					Expect(filenameWithAt.Complete("@")).To(ConsistOf(
						flags.Completion{Item: "@abc"},
						flags.Completion{Item: "@abd"},
						flags.Completion{Item: "@efg"},
					))
				})
			})

			Context("when there are characters after the @", func() {
				It("returns matching files", func() {
					Expect(filenameWithAt.Complete("@a")).To(ConsistOf(
						flags.Completion{Item: "@abc"},
						flags.Completion{Item: "@abd"},
					))
				})
			})
		})
	})
})
