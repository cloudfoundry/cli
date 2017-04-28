package isolated

import (
	"io/ioutil"
	"os"

	. "code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("create-buildpack command", func() {
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
			Eventually(session.Out).Should(Say("cf create-buildpack BUILDPACK PATH POSITION")) // help
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
			Eventually(session.Out).Should(Say("Failed to create a local temporary zip file for the buildpack"))
			Eventually(session.Out).Should(Say("FAILED"))
			Eventually(session.Out).Should(Say("Couldn't write zip file: zip: not a valid zip file"))
			Eventually(session).Should(Exit(1))
		})
	})
})
