package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("auth command", func() {

	BeforeEach(func() {
		helpers.SkipIfClientCredentialsTestMode()
	})

	When("too many arguments are provided", func() {
		It("displays an 'extra argument' error message", func() {
			session := helpers.CF("auth", "some-username", "some-password", "garbage")

			Eventually(session.Err).Should(Say("Incorrect Usage: unexpected argument \"garbage\""))
			Eventually(session).Should(Say("NAME:"))

			Eventually(session).Should(Exit(1))
		})
	})
})
