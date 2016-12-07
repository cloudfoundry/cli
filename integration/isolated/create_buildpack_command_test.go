package isolated

import (
	. "code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("create-buildpack command", func() {
	Context("when the wrong data type is provided as the position argument", func() {
		It("outputs an error message to the user, provides help text, and exits 1", func() {
			session := CF("create-buildpack", "some-buildpack", "some-path", "not-an-integer")
			Eventually(session).Should(Exit(1))
			Expect(session.Err).To(Say("Incorrect usage: Value for POSITION must be integer"))
			Expect(session.Out).To(Say("cf create-buildpack BUILDPACK PATH POSITION")) // help
		})
	})
})
