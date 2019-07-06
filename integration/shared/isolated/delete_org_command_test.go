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
		helpers.SkipIfClientCredentialsTestMode()
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(false, false, ReadOnlyOrg, "delete-org", "some-org")
		})
	})

	When("the org name is not provided", func() {
		It("displays an error and help", func() {
			session := helpers.CF("delete-org")
			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `ORG` was not provided"))
			Eventually(session).Should(Say("USAGE"))
			Eventually(session).Should(Exit(1))
		})
	})

	When("the org does not exist", func() {
		BeforeEach(func() {
			helpers.LoginCF()
		})

		It("displays a warning and exits 0", func() {
			username, _ := helpers.GetCredentials()
			session := helpers.CF("delete-org", "-f", "please-do-not-exist-in-real-life")
			Eventually(session).Should(Say("Deleting org please-do-not-exist-in-real-life as %s...", username))
			Eventually(session).Should(Say("Org please-do-not-exist-in-real-life does not exist."))
			Eventually(session).Should(Say("OK"))
			Eventually(session).Should(Exit(0))
		})
	})

	When("the org exists", func() {
		var orgName string

		BeforeEach(func() {
			helpers.LoginCF()

			orgName = helpers.NewOrgName()
			helpers.CreateOrgAndSpace(orgName, helpers.NewSpaceName())
		})

		AfterEach(func() {
			helpers.QuickDeleteOrgIfExists(orgName)
		})

		When("the -f flag not is provided", func() {
			var buffer *Buffer

			BeforeEach(func() {
				buffer = NewBuffer()
			})

			When("the user enters 'y'", func() {
				BeforeEach(func() {
					_, err := buffer.Write([]byte("y\n"))
					Expect(err).ToNot(HaveOccurred())
				})

				It("deletes the org", func() {
					username, _ := helpers.GetCredentials()
					session := helpers.CFWithStdin(buffer, "delete-org", orgName)
					Eventually(session).Should(Say(`Really delete the org %s, including its spaces, apps, service instances, routes, private domains and space-scoped service brokers\?`, orgName))
					Eventually(session).Should(Say("Deleting org %s as %s...", orgName, username))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))
					Eventually(helpers.CF("org", orgName)).Should(Exit(1))
				})
			})

			When("the user enters 'n'", func() {
				BeforeEach(func() {
					_, err := buffer.Write([]byte("n\n"))
					Expect(err).ToNot(HaveOccurred())
				})

				It("does not delete the org", func() {
					session := helpers.CFWithStdin(buffer, "delete-org", orgName)
					Eventually(session).Should(Say(`Really delete the org %s, including its spaces, apps, service instances, routes, private domains and space-scoped service brokers\?`, orgName))
					Eventually(session).Should(Say("Delete cancelled"))
					Eventually(session).Should(Exit(0))
					Eventually(helpers.CF("org", orgName)).Should(Exit(0))
				})
			})

			When("the user enters the default input (hits return)", func() {
				BeforeEach(func() {
					_, err := buffer.Write([]byte("\n"))
					Expect(err).ToNot(HaveOccurred())
				})

				It("does not delete the org", func() {
					session := helpers.CFWithStdin(buffer, "delete-org", orgName)
					Eventually(session).Should(Say(`Really delete the org %s, including its spaces, apps, service instances, routes, private domains and space-scoped service brokers\?`, orgName))
					Eventually(session).Should(Say("Delete cancelled"))
					Eventually(session).Should(Exit(0))
					Eventually(helpers.CF("org", orgName)).Should(Exit(0))
				})
			})

			When("the user enters an invalid answer", func() {
				BeforeEach(func() {
					// The second '\n' is intentional. Otherwise the buffer will be
					// closed while the interaction is still waiting for input; it gets
					// an EOF and causes an error.
					_, err := buffer.Write([]byte("wat\n\n"))
					Expect(err).ToNot(HaveOccurred())
				})

				It("asks again", func() {
					session := helpers.CFWithStdin(buffer, "delete-org", orgName)
					Eventually(session).Should(Say(`Really delete the org %s, including its spaces, apps, service instances, routes, private domains and space-scoped service brokers\?`, orgName))
					Eventually(session).Should(Say(`invalid input \(not y, n, yes, or no\)`))
					Eventually(session).Should(Say(`Really delete the org %s, including its spaces, apps, service instances, routes, private domains and space-scoped service brokers\?`, orgName))
					Eventually(session).Should(Exit(0))
					Eventually(helpers.CF("org", orgName)).Should(Exit(0))
				})
			})
		})

		When("the -f flag is provided", func() {
			It("deletes the org", func() {
				username, _ := helpers.GetCredentials()
				session := helpers.CF("delete-org", orgName, "-f")
				Eventually(session).Should(Say("Deleting org %s as %s...", orgName, username))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))
				Eventually(helpers.CF("org", orgName)).Should(Exit(1))
			})
		})
	})

	When("deleting an org that is targeted", func() {
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
			Eventually(session).Should(Say("No org or space targeted, use 'cf target -o ORG -s SPACE'"))
			Eventually(session).Should(Exit(0))
		})
	})

	When("deleting an org that is not targeted", func() {
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
			Eventually(session).Should(Say(`org:\s+%s`, ReadOnlyOrg))
			Eventually(session).Should(Say(`space:\s+%s`, ReadOnlySpace))
			Eventually(session).Should(Exit(0))
		})
	})
})
