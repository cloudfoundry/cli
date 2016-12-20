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
		Context("when no API endpoint is set", func() {
			BeforeEach(func() {
				helpers.UnsetAPI()
			})

			It("fails with no API endpoint set message", func() {
				session := helpers.CF("delete-org", "banana")
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("No API endpoint set. Use 'cf login' or 'cf api' to target an endpoint."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when not logged in", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
			})

			It("fails with not logged in message", func() {
				session := helpers.CF("delete-org", "banana")
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("Not logged in. Use 'cf login' to log in."))
				Eventually(session).Should(Exit(1))
			})
		})
	})

	Context("when the org name it not provided", func() {
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
			helpers.CreateOrgAndSpace(orgName, helpers.PrefixedRandomName("space"))
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
					Eventually(session.Out).Should(Say("Really delete the org %s and everything associated with it\\?>", orgName))
					Eventually(session.Out).Should(Say("Deleting org %s as %s...", orgName, username))
					Eventually(session.Out).Should(Say("OK"))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when the user enters 'n'", func() {
				BeforeEach(func() {
					buffer.Write([]byte("n\n"))
				})

				It("does not delete the org", func() {
					session := helpers.CFWithStdin(buffer, "delete-org", orgName)
					Eventually(session.Out).Should(Say("Really delete the org %s and everything associated with it\\?>", orgName))
					Eventually(session.Out).Should(Say("Delete cancelled"))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when the user enters the default input (hits return)", func() {
				BeforeEach(func() {
					buffer.Write([]byte("\n"))
				})

				It("does not delete the org", func() {
					session := helpers.CFWithStdin(buffer, "delete-org", orgName)
					Eventually(session.Out).Should(Say("Really delete the org %s and everything associated with it\\?>", orgName))
					Eventually(session.Out).Should(Say("Delete cancelled"))
					Eventually(session).Should(Exit(0))
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
					Eventually(session.Out).Should(Say("Really delete the org %s and everything associated with it\\?>", orgName))
					Eventually(session.Out).Should(Say("invalid input \\(not y, n, yes, or no\\)"))
					Eventually(session.Out).Should(Say("Really delete the org %s and everything associated with it\\?>", orgName))
					Eventually(session).Should(Exit(0))
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
			})
		})
	})
})
