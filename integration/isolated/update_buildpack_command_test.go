package isolated

import (
	"io/ioutil"
	"os"
	"path/filepath"

	. "code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("update-buildpack command", func() {
	Context("when the buildpack is not provided", func() {
		It("returns a buildpack argument not provided error", func() {
			session := CF("update-buildpack", "-p", ".")
			Eventually(session).Should(Exit(1))

			Expect(session.Err.Contents()).To(BeEquivalentTo("Incorrect Usage: the required argument `BUILDPACK` was not provided\n\n"))
		})
	})

	Context("when the buildpack's path does not exist", func() {
		It("returns a buildpack does not exist error", func() {
			session := CF("update-buildpack", "some-buildpack", "-p", "this-is-a-bogus-path")

			Eventually(session.Err).Should(Say("Incorrect Usage: The specified path 'this-is-a-bogus-path' does not exist."))
			Eventually(session).Should(Exit(1))
		})
	})

	Context("when the wrong data type is provided as the position argument", func() {
		It("outputs an error message to the user, provides help text, and exits 1", func() {
			session := CF("update-buildpack", "some-buildpack", "-i", "not-an-integer")
			Eventually(session.Err).Should(Say("Incorrect Usage: invalid argument for flag `-i' \\(expected int\\)"))
			Eventually(session.Out).Should(Say("cf update-buildpack BUILDPACK")) // help
			Eventually(session).Should(Exit(1))
		})
	})

	Context("when a URL is provided as the buildpack", func() {
		var dir string

		BeforeEach(func() {
			LoginCF()

			var err error
			dir, err = ioutil.TempDir("", "update-buildpack-test")
			Expect(err).ToNot(HaveOccurred())

			filename := "some-file"
			tempfile := filepath.Join(dir, filename)
			err = ioutil.WriteFile(tempfile, []byte{}, 0400)
			Expect(err).ToNot(HaveOccurred())

			session := CF("create-buildpack", "some-buildpack", dir, "1")
			Eventually(session).Should(Exit(0))
		})

		AfterEach(func() {
			err := os.RemoveAll(dir)
			Expect(err).NotTo(HaveOccurred())
		})

		It("outputs an error message to the user, provides help text, and exits 1", func() {
			session := CF("update-buildpack", "some-buildpack", "-p", "https://example.com/bogus.tgz")
			Eventually(session.Out).Should(Say("Failed to create a local temporary zip file for the buildpack"))
			Eventually(session.Out).Should(Say("FAILED"))
			Eventually(session.Out).Should(Say("Couldn't write zip file: zip: not a valid zip file"))
			Eventually(session).Should(Exit(1))
		})
	})
})
