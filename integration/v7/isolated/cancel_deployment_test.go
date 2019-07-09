package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"

	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Cancel Deployment", func() {
	Context("Help", func() {
		It("appears in cf help -a", func() {
			session := helpers.CF("help", "-a")
			Eventually(session).Should(Exit(0))
			Expect(session).To(HaveCommandInCategoryWithDescription("cancel-deployment", "APPS", "Cancel the most recent deployment for an app. Resets the current droplet to the previous deployment's droplet."))
		})

		It("displays the help information", func() {
			session := helpers.CF("cancel-deployment", "--help")
			Eventually(session).Should(Say(`NAME:`))
			Eventually(session).Should(Say(`cancel-deployment - Cancel the most recent deployment for an app. Resets the current droplet to the previous deployment's droplet.\n`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`USAGE:`))
			Eventually(session).Should(Say(`cf cancel-deployment APP_NAME\n`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`EXAMPLES:`))
			Eventually(session).Should(Say(`cf cancel-deployment my-app\n`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`SEE ALSO:`))
			Eventually(session).Should(Say(`app, push`))

			Eventually(session).Should(Exit(0))
		})
	})
})
