package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("delete-space command", func() {
	Describe("help", func() {
		It("shows usage", func() {
			session := helpers.CF("help", "delete-space")
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Say("\\s+delete-space - Delete a space"))
			Eventually(session).Should(Say("USAGE:"))
			Eventually(session).Should(Say("delete-space SPACE \\[-o ORG\\] \\[-f\\]"))
			Eventually(session).Should(Say("OPTIONS:"))
			Eventually(session).Should(Say("\\s+-f\\s+Force deletion without confirmation"))
			Eventually(session).Should(Say("\\s+-o\\s+Delete space within specified org"))
			Eventually(session).Should(Exit(0))
		})
	})

	Context("when the environment is not setup correctly", func() {
		Context("when the org is provided", func() {
			It("fails with the appropriate errors", func() {
				helpers.CheckEnvironmentTargetedCorrectly(false, false, ReadOnlyOrg, "delete-space", "space-name", "-o", ReadOnlyOrg)
			})
		})

		Context("when the org is not provided", func() {
			It("fails with the appropriate errors", func() {
				helpers.CheckEnvironmentTargetedCorrectly(true, false, ReadOnlyOrg, "delete-space", "space-name")
			})
		})
	})

	Context("when the space name it not provided", func() {
		It("displays an error and help", func() {
			session := helpers.CF("delete-space")
			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `SPACE` was not provided"))
			Eventually(session.Out).Should(Say("USAGE"))
			Eventually(session).Should(Exit(1))
		})
	})

	Context("when the space does not exist", func() {
		var orgName string

		BeforeEach(func() {
			helpers.LoginCF()

			orgName = helpers.NewOrgName()
			helpers.CreateOrg(orgName)
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		It("fails and displays space not found", func() {
			username, _ := helpers.GetCredentials()
			session := helpers.CF("delete-space", "-f", "-o", orgName, "please-do-not-exist-in-real-life")
			Eventually(session.Out).Should(Say("Deleting space please-do-not-exist-in-real-life in org %s as %s...", orgName, username))
			Eventually(session.Out).Should(Say("FAILED"))
			Eventually(session.Err).Should(Say("Space 'please-do-not-exist-in-real-life' not found\\."))
			Eventually(session).Should(Exit(1))
		})
	})

	Context("when the space exists", func() {
		var orgName string
		var spaceName string

		BeforeEach(func() {
			helpers.LoginCF()

			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()
			helpers.CreateOrgAndSpace(orgName, spaceName)
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
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

				It("deletes the space", func() {
					username, _ := helpers.GetCredentials()
					session := helpers.CFWithStdin(buffer, "delete-space", spaceName)
					Eventually(session.Out).Should(Say("Really delete the space %s\\? \\[yN\\]", spaceName))
					Eventually(session.Out).Should(Say("Deleting space %s in org %s as %s...", spaceName, orgName, username))
					Eventually(session.Out).Should(Say("OK"))
					Eventually(session).Should(Exit(0))
					Eventually(helpers.CF("space", spaceName)).Should(Exit(1))
				})
			})

			Context("when the user enters 'n'", func() {
				BeforeEach(func() {
					buffer.Write([]byte("n\n"))
				})

				It("does not delete the space", func() {
					session := helpers.CFWithStdin(buffer, "delete-space", spaceName)
					Eventually(session.Out).Should(Say("Really delete the space %s\\? \\[yN\\]", spaceName))
					Eventually(session.Out).Should(Say("Delete cancelled"))
					Eventually(session).Should(Exit(0))
					Eventually(helpers.CF("space", spaceName)).Should(Exit(0))
				})
			})

			Context("when the user enters the default input (hits return)", func() {
				BeforeEach(func() {
					buffer.Write([]byte("\n"))
				})

				It("does not delete the org", func() {
					session := helpers.CFWithStdin(buffer, "delete-space", spaceName)
					Eventually(session.Out).Should(Say("Really delete the space %s\\? \\[yN\\]", spaceName))
					Eventually(session.Out).Should(Say("Delete cancelled"))
					Eventually(session).Should(Exit(0))
					Eventually(helpers.CF("space", spaceName)).Should(Exit(0))
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
					session := helpers.CFWithStdin(buffer, "delete-space", spaceName)
					Eventually(session.Out).Should(Say("Really delete the space %s\\? \\[yN\\]", spaceName))
					Eventually(session.Out).Should(Say("invalid input \\(not y, n, yes, or no\\)"))
					Eventually(session.Out).Should(Say("Really delete the space %s\\? \\[yN\\]", spaceName))
					Eventually(session).Should(Exit(0))
					Eventually(helpers.CF("space", spaceName)).Should(Exit(0))
				})
			})
		})

		Context("when the -f flag is provided", func() {
			It("deletes the space", func() {
				username, _ := helpers.GetCredentials()
				session := helpers.CF("delete-space", spaceName, "-f")
				Eventually(session.Out).Should(Say("Deleting space %s in org %s as %s\\.\\.\\.", spaceName, orgName, username))
				Eventually(session.Out).Should(Say("OK"))
				Eventually(session).Should(Exit(0))
				Eventually(helpers.CF("space", spaceName)).Should(Exit(1))
			})

			Context("when the space was targeted", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("target", "-s", spaceName)).Should(Exit(0))
				})

				It("deletes the space and clears the target", func() {
					username, _ := helpers.GetCredentials()
					session := helpers.CF("delete-space", spaceName, "-f")
					Eventually(session.Out).Should(Say("Deleting space %s in org %s as %s\\.\\.\\.", spaceName, orgName, username))
					Eventually(session.Out).Should(Say("OK"))
					Eventually(session.Out).Should(Say("TIP: No space targeted, use 'cf target -s' to target a space\\."))
					Eventually(session).Should(Exit(0))

					Eventually(helpers.CF("space", spaceName)).Should(Exit(1))

					session = helpers.CF("target")
					Eventually(session.Out).Should(Say("No space targeted, use 'cf target -s SPACE'"))
					Eventually(session).Should(Exit(0))
				})
			})
		})
	})

	Context("when the -o organzation does not exist", func() {
		BeforeEach(func() {
			helpers.LoginCF()
		})

		It("fails and displays org not found", func() {
			username, _ := helpers.GetCredentials()
			session := helpers.CF("delete-space", "-f", "-o", "please-do-not-exist-in-real-life", "please-do-not-exist-in-real-life")
			Eventually(session.Out).Should(Say("Deleting space please-do-not-exist-in-real-life in org please-do-not-exist-in-real-life as %s...", username))
			Eventually(session.Err).Should(Say("Organization 'please-do-not-exist-in-real-life' not found\\."))
			Eventually(session.Out).Should(Say("FAILED"))
			Eventually(session).Should(Exit(1))
		})
	})
})
