package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("delete-org command", func() {
	Context("when the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(false, false, ReadOnlyOrg, "delete-org", "some-org")
		})
	})

	Context("when the org name is not provided", func() {
		It("displays an error and help", func() {
			session := helpers.CF("delete-org")
			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `ORG` was not provided"))
			Eventually(session.Out).Should(Say("USAGE"))
			Eventually(session).Should(Exit(1))
		})
	})

	Context("when the org does not exist", func() {
		BeforeEach(func() {
			helpers.LoginCF()
		})

		It("displays a warning and exits 0", func() {
			username, _ := helpers.GetCredentials()
			session := helpers.CF("delete-org", "-f", "please-do-not-exist-in-real-life")
			Eventually(session.Out).Should(Say("Deleting org please-do-not-exist-in-real-life as %s...", username))
			Eventually(session.Out).Should(Say("Org please-do-not-exist-in-real-life does not exist."))
			Eventually(session.Out).Should(Say("OK"))
			Eventually(session).Should(Exit(0))
		})
	})

	Context("when the org exists", func() {
		var orgName string

		BeforeEach(func() {
			helpers.LoginCF()

			orgName = helpers.NewOrgName()
			helpers.CreateOrgAndSpace(orgName, helpers.NewSpaceName())
		})

		AfterEach(func() {
			helpers.QuickDeleteOrgIfExists(orgName)
		})

		Context("when the -f flag not is provided", func() {
			var buffer *Buffer

			BeforeEach(func() {
				buffer = NewBuffer()
			})

			Context("when the user enters 'y'", func() {
				BeforeEach(func() {
					buffer.Write([]byte("y\n"))
				})

				It("deletes the org", func() {
					username, _ := helpers.GetCredentials()
					session := helpers.CFWithStdin(buffer, "delete-org", orgName)
					Eventually(session.Out).Should(Say("Really delete the org %s, including its spaces, apps, service instances, routes, private domains and space-scoped service brokers\\?", orgName))
					Eventually(session.Out).Should(Say("Deleting org %s as %s...", orgName, username))
					Eventually(session.Out).Should(Say("OK"))
					Eventually(session).Should(Exit(0))
					Eventually(helpers.CF("org", orgName)).Should(Exit(1))
				})
			})

			Context("when the user enters 'n'", func() {
				BeforeEach(func() {
					buffer.Write([]byte("n\n"))
				})

				It("does not delete the org", func() {
					session := helpers.CFWithStdin(buffer, "delete-org", orgName)
					Eventually(session.Out).Should(Say("Really delete the org %s, including its spaces, apps, service instances, routes, private domains and space-scoped service brokers\\?", orgName))
					Eventually(session.Out).Should(Say("Delete cancelled"))
					Eventually(session).Should(Exit(0))
					Eventually(helpers.CF("org", orgName)).Should(Exit(0))
				})
			})

			Context("when the user enters the default input (hits return)", func() {
				BeforeEach(func() {
					buffer.Write([]byte("\n"))
				})

				It("does not delete the org", func() {
					session := helpers.CFWithStdin(buffer, "delete-org", orgName)
					Eventually(session.Out).Should(Say("Really delete the org %s, including its spaces, apps, service instances, routes, private domains and space-scoped service brokers\\?", orgName))
					Eventually(session.Out).Should(Say("Delete cancelled"))
					Eventually(session).Should(Exit(0))
					Eventually(helpers.CF("org", orgName)).Should(Exit(0))
				})
			})

			Context("when the user enters an invalid answer", func() {
				BeforeEach(func() {
					// The second '\n' is intentional. Otherwise the buffer will be
					// closed while the interaction is still waiting for input; it gets
					// an EOF and causes an error.
					buffer.Write([]byte("wat\n\n"))
				})

				It("asks again", func() {
					session := helpers.CFWithStdin(buffer, "delete-org", orgName)
					Eventually(session.Out).Should(Say("Really delete the org %s, including its spaces, apps, service instances, routes, private domains and space-scoped service brokers\\?", orgName))
					Eventually(session.Out).Should(Say("invalid input \\(not y, n, yes, or no\\)"))
					Eventually(session.Out).Should(Say("Really delete the org %s, including its spaces, apps, service instances, routes, private domains and space-scoped service brokers\\?", orgName))
					Eventually(session).Should(Exit(0))
					Eventually(helpers.CF("org", orgName)).Should(Exit(0))
				})
			})
		})

		Context("when the -f flag is provided", func() {
			It("deletes the org", func() {
				username, _ := helpers.GetCredentials()
				session := helpers.CF("delete-org", orgName, "-f")
				Eventually(session.Out).Should(Say("Deleting org %s as %s...", orgName, username))
				Eventually(session.Out).Should(Say("OK"))
				Eventually(session).Should(Exit(0))
				Eventually(helpers.CF("org", orgName)).Should(Exit(1))
			})
		})
	})

	Context("when deleting an org that is targeted", func() {
		var orgName string

		BeforeEach(func() {
			helpers.LoginCF()

			orgName = helpers.NewOrgName()
			spaceName := helpers.NewSpaceName()
			helpers.CreateOrgAndSpace(orgName, spaceName)
			helpers.TargetOrgAndSpace(orgName, spaceName)
		})

		AfterEach(func() {
			helpers.QuickDeleteOrgIfExists(orgName)
		})

		It("clears the targeted org and space", func() {
			session := helpers.CF("delete-org", orgName, "-f")
			Eventually(session).Should(Exit(0))

			session = helpers.CF("target")
			Eventually(session.Out).Should(Say("No org or space targeted, use 'cf target -o ORG -s SPACE'"))
			Eventually(session).Should(Exit(0))
		})
	})

	Context("when deleting an org that is not targeted", func() {
		var orgName string

		BeforeEach(func() {
			helpers.LoginCF()

			orgName = helpers.NewOrgName()
			helpers.TargetOrgAndSpace(ReadOnlyOrg, ReadOnlySpace)
		})

		It("does not clear the targeted org and space", func() {
			session := helpers.CF("delete-org", orgName, "-f")
			Eventually(session).Should(Exit(0))

			session = helpers.CF("target")
			Eventually(session.Out).Should(Say("org:\\s+%s", ReadOnlyOrg))
			Eventually(session.Out).Should(Say("space:\\s+%s", ReadOnlySpace))
			Eventually(session).Should(Exit(0))
		})
	})
})
