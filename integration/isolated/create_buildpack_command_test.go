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

var _ = Describe("create-buildpack command", func() {
	Context("successful creation", func() {
		var (
			dir string
			err error
		)
		BeforeEach(func() {
			LoginCF()

			dir, err = ioutil.TempDir("", "create-buildpack-test")
			Expect(err).ToNot(HaveOccurred())

			filename := "some-file"
			manifestFile := filepath.Join(dir, filename)
			err = ioutil.WriteFile(manifestFile, []byte(""), 0400)
			Expect(err).ToNot(HaveOccurred())

			session := CF("create-buildpack", "some-buildpack", dir, "1")
			Eventually(session).Should(Exit(0))
		})

		AfterEach(func() {
			Eventually(CF("delete-buildpack", "some-buildpack", "-f")).Should(Exit(0))
			Expect(os.RemoveAll(dir)).To(Succeed())
		})

		It("lists the created buildpack with stack", func() {
			session := CF("buildpacks")
			Eventually(session).Should(Exit(0))
			Expect(session.Out).To(Say("some-buildpack.*cflinuxfs2.*1"))

		})
	})
	Context("when the wrong data type is provided as the position argument", func() {
		var filename string

		BeforeEach(func() {
			// The args take a filepath. Creating a real file will avoid the file
			// does not exist error, and trigger the correct error case we are
			// testing.
			filename = "some-file"
			err := ioutil.WriteFile(filename, []byte{}, 0400)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			err := os.Remove(filename)
			Expect(err).NotTo(HaveOccurred())
		})

		It("outputs an error message to the user, provides help text, and exits 1", func() {
			session := CF("create-buildpack", "some-buildpack", "some-file", "not-an-integer")
			Eventually(session.Err).Should(Say("Incorrect usage: Value for POSITION must be integer"))
			Eventually(session).Should(Say("cf create-buildpack BUILDPACK PATH POSITION")) // help
			Eventually(session).Should(Exit(1))
		})
	})

	Context("when a nonexistent file is provided", func() {
		It("outputs an error message to the user and exits 1", func() {
			session := CF("create-buildpack", "some-buildpack", "some-bogus-file", "1")
			Eventually(session.Err).Should(Say("Incorrect Usage: The specified path 'some-bogus-file' does not exist."))
			Eventually(session).Should(Exit(1))
		})
	})

	Context("when a URL is provided as the buildpack", func() {
		BeforeEach(func() {
			LoginCF()
		})

		It("outputs an error message to the user, provides help text, and exits 1", func() {
			session := CF("create-buildpack", "some-buildpack", "https://example.com/bogus.tgz", "1")
			Eventually(session).Should(Say("Failed to create a local temporary zip file for the buildpack"))
			Eventually(session).Should(Say("FAILED"))
			Eventually(session).Should(Say("Couldn't write zip file: zip: not a valid zip file"))
			Eventually(session).Should(Exit(1))
		})
	})
})
