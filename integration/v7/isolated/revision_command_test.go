package isolated

import (
	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("revision command", func() {
	Describe("help", func() {
		When("--help flag is set", func() {
			It("appears in cf help -a", func() {
				session := helpers.CF("help", "-a")
				Eventually(session).Should(Exit(0))
				Expect(session).To(HaveCommandInCategoryWithDescription("revision", "EXPERIMENTAL COMMANDS", "Show details for a specific app revision"))
			})

			It("Displays revision command usage to output", func() {
				session := helpers.CF("revision", "--help")

				Eventually(session).Should(Exit(0))

				Expect(session).To(Say("NAME:"))
				Expect(session).To(Say("revision - Show details for a specific app revision"))
				Expect(session).To(Say("USAGE:"))
				Expect(session).To(Say(`cf revision APP_NAME [--version VERSION]`))
				Expect(session).To(Say("OPTIONS:"))
				Expect(session).To(Say("--version      The integer representing the specific revision to show"))
				Expect(session).To(Say("SEE ALSO:"))
				Expect(session).To(Say("revisions, rollback"))
			})
		})
	})
})
