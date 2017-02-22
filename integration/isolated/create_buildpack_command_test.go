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

var _ = FDescribe("create-buildpack command", func() {
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
			Eventually(session).Should(Exit(1))
			Expect(session.Err).To(Say("Incorrect usage: Value for POSITION must be integer"))
			Expect(session.Out).To(Say("cf create-buildpack BUILDPACK PATH POSITION")) // help
		})
	})
})
