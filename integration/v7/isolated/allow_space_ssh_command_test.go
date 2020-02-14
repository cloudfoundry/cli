package isolated

import (
	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = FDescribe("allow-space-ssh command", func() {
	var (
		orgName   string
		spaceName string
	)

	BeforeEach(func() {
		orgName = helpers.NewOrgName()
		spaceName = helpers.NewSpaceName()
	})

	Describe("help", func() {
		When("--help flag is set", func() {
			It("appears in cf help -a", func() {
				session := helpers.CF("help", "-a")
				Eventually(session).Should(Exit(0))
				Expect(session).To(HaveCommandInCategoryWithDescription("allow-space-ssh", "SPACES", "Allow SSH access for the space"))
			})

			It("displays command usage to output", func() {
				session := helpers.CF("allow-space-ssh", "--help")

				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("allow-space-ssh - Allow SSH access for the space"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf allow-space-ssh SPACE"))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("enable-ssh, space-ssh-allowed, ssh, ssh-enabled"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the space name is not provided", func() {
		It("tells the user that the space name is required, prints help text, and exits 1", func() {
			session := helpers.CF("allow-space-ssh")

			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `SPACE` was not provided"))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, false, ReadOnlyOrg, "allow-space-ssh", spaceName)
		})
	})

	When("the environment is set up correctly", func() {
		var userName string

		BeforeEach(func() {
			helpers.SetupCF(orgName, spaceName)
			userName, _ = helpers.GetCredentials()
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		When("the space does not exist", func() {
			It("displays 'space not found' and exits 1", func() {
				invalidSpaceName := "invalid-space-name"
				session := helpers.CF("allow-space-ssh", invalidSpaceName)

				Eventually(session).Should(Say(`Enabling ssh support for space %s as %s\.\.\.`, invalidSpaceName, userName))
				Eventually(session.Err).Should(Say("Space '%s' not found", invalidSpaceName))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the space exists", func() {
			When("when ssh has not been allowed yet", func() {
				It("allows ssh for the space", func() {
					session := helpers.CF("allow-space-ssh", spaceName)

					Eventually(session).Should(Say(`Enabling ssh support for space %s as %s\.\.\.`, spaceName, userName))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))

					session = helpers.CF("space-ssh-allowed", spaceName)
					Eventually(session).Should(Say(`enabled`))
					Eventually(session).Should(Exit(0))
				})
			})

			When("ssh was previously enabled for the space", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("allow-space-ssh", spaceName)).Should(Exit(0))
				})

				It("informs the user and exits 0", func() {
					session := helpers.CF("allow-space-ssh", spaceName)

					Eventually(session).Should(Say(`Enabling ssh support for space %s as %s\.\.\.`, spaceName, userName))
					Eventually(session).Should(Say("ssh support for space '%s' is already enabled.", spaceName))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))

					session = helpers.CF("space-ssh-allowed", spaceName)
					Eventually(session).Should(Say(`enabled`))
					Eventually(session).Should(Exit(0))
				})
			})
		})
	})
})
