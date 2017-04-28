package isolated

import (
	. "code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("scale command", func() {
	Context("when the wrong data type is provided to -i", func() {
		It("outputs an error message to the user, provides help text, and exits 1", func() {
			session := CF("scale", "some-app", "-i", "not-an-integer")
			Eventually(session.Err).Should(Say("Incorrect Usage: invalid argument for flag `-i' \\(expected int\\)"))
			Eventually(session.Out).Should(Say("cf scale APP_NAME")) // help
			Eventually(session).Should(Exit(1))
		})
	})
})
