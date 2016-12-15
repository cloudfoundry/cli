package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("delete-org command", func() {
	BeforeEach(func() {
		helpers.RunIfExperimental("remove after #133310639")
	})

	Context("when the environment is not setup correctly", func() {
		Context("when no API endpoint is set", func() {
			BeforeEach(func() {
				helpers.UnsetAPI()
			})

			It("fails with no API endpoint set message", func() {
				session := helpers.CF("delete-org", "banana")
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Say("No API endpoint set. Use 'cf login' or 'cf api' to target an endpoint."))
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
				Eventually(session.Out).Should(Say("Not logged in. Use 'cf login' to log in."))
				Eventually(session).Should(Exit(1))
			})
		})
	})

	Context("when the org name it not provided", func() {
		It("displays an error and help", func() {
			session := helpers.CF("delete-org")
			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `ORG` was not provided"))
			Eventually(session).Should(Say("USAGE"))
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
			Eventually(session).Should(Say("Deleting org please-do-not-exist-in-real-life as %s...", username))
			Eventually(session).Should(Say("OK"))
			Eventually(session).Should(Say("Org please-do-not-exist-in-real-life does not exist."))
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

			Context("when the user enters y", func() {
				BeforeEach(func() {
					buffer.Write([]byte("y\n"))
				})

				It("deletes the org", func() {
					username, _ := helpers.GetCredentials()
					session := helpers.CFWithStdin(buffer, "delete-org", orgName)
					Eventually(session).Should(Say("Really delete the org %s and everything associated with it\\?>", orgName))
					Eventually(session).Should(Say("Deleting org %s as %s...", orgName, username))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when the user enters n", func() {
				BeforeEach(func() {
					buffer.Write([]byte("n\n"))
				})

				It("does not delete the org", func() {
					session := helpers.CFWithStdin(buffer, "delete-org", orgName)
					Eventually(session).Should(Say("Really delete the org %s and everything associated with it\\?>", orgName))
					Eventually(session).Should(Say("Delete cancelled"))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when the user enters an invalid answer", func() {
				BeforeEach(func() {
					buffer.Write([]byte("wat\n"))
				})

				It("asks again", func() {
				})
			})
		})

		Context("when the -f flag is provided", func() {
			It("deletes the org", func() {
				username, _ := helpers.GetCredentials()
				session := helpers.CF("delete-org", orgName, "-f")
				Eventually(session).Should(Say("Deleting org %s as %s...", orgName, username))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
