package isolated

import (
	. "code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("unbind-service command", func() {
	Context("when the buildpack is not provided", func() {
		It("returns a buildpack argument not provided error", func() {
			session := CF("update-buildpack", "-p", ".")
			Eventually(session).Should(Exit(1))

			Expect(session.Err.Contents()).To(BeEquivalentTo("Incorrect Usage: the required argument `BUILDPACK` was not provided\n\n"))
		})
	})

	Context("when the wrong data type is provided as the position argument", func() {
		It("outputs an error message to the user, provides help text, and exits 1", func() {
			session := CF("update-buildpack", "some-buildpack", "-i", "not-an-integer")
			Eventually(session).Should(Exit(1))
			Expect(session.Err).To(Say("Incorrect Usage: invalid argument for flag `-i' \\(expected int\\)"))
			Expect(session.Out).To(Say("cf update-buildpack BUILDPACK")) // help
		})
	})
})
