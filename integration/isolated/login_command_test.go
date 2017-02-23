package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("login command", func() {
	var buffer *Buffer

	BeforeEach(func() {
		buffer = NewBuffer()
		buffer.Write([]byte("\n"))
	})

	Context("when the API endpoint is not set", func() {
		BeforeEach(func() {
			helpers.UnsetAPI()
			buffer = NewBuffer()
			buffer.Write([]byte("\n"))
		})

		It("prompts the user for an endpoint", func() {
			session := helpers.CFWithStdin(buffer, "login")
			Eventually(session.Out).Should(Say("API endpoint>"))
		})
	})

	Context("when --sso-passcode flag is given", func() {
		Context("when a passcode isn't provided", func() {
			It("prompts the user to try again", func() {
				session := helpers.CFWithStdin(buffer, "login", "--sso-passcode")
				Eventually(session.Err).Should(Say("Incorrect Usage: expected argument for flag `--sso-passcode'"))
			})
		})

		Context("when the provided passcode is invalid", func() {
			It("prompts the user to try again", func() {
				session := helpers.CFWithStdin(buffer, "login", "--sso-passcode", "bad-passcode")
				Eventually(session.Out).Should(Say("Authenticating..."))
				Eventually(session.Out).Should(Say("Credentials were rejected, please try again."))
			})
		})
	})

	Context("when both --sso and --sso-passcode flags are provided", func() {
		It("errors with invalid use", func() {
			session := helpers.CFWithStdin(buffer, "login", "--sso", "--sso-passcode", "some-passcode")
			Eventually(session.Out).Should(Say("Incorrect usage: --sso-passcode flag cannot be used with --sso"))
			Eventually(session).Should(Exit(1))

		})
	})
})
