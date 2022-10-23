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
			Eventually(session).Should(Say(`\s+delete-space - Delete a space`))
			Eventually(session).Should(Say("USAGE:"))
			Eventually(session).Should(Say(`delete-space SPACE \[-o ORG\] \[-f\]`))
			Eventually(session).Should(Say("OPTIONS:"))
			Eventually(session).Should(Say(`\s+-f\s+Force deletion without confirmation`))
			Eventually(session).Should(Say(`\s+-o\s+Delete space within specified org`))
			Eventually(session).Should(Exit(0))
		})
	})

	When("the environment is not setup correctly", func() {
		When("the org is provided", func() {
			It("fails with the appropriate errors", func() {
				helpers.CheckEnvironmentTargetedCorrectly(false, false, ReadOnlyOrg, "delete-space", "space-name", "-o", ReadOnlyOrg)
			})
		})

		When("the org is not provided", func() {
			It("fails with the appropriate errors", func() {
				helpers.CheckEnvironmentTargetedCorrectly(true, false, ReadOnlyOrg, "delete-space", "space-name")
			})
		})
	})

	When("the space name it not provided", func() {
		It("displays an error and help", func() {
			session := helpers.CF("delete-space")
			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `SPACE` was not provided"))
			Eventually(session).Should(Say("USAGE"))
			Eventually(session).Should(Exit(1))
		})
	})

	When("the space does not exist", func() {
		var orgName string

		BeforeEach(func() {
			helpers.LoginCF()

			orgName = helpers.NewOrgName()
			helpers.CreateOrg(orgName)
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		It("passes and displays space not found", func() {
			username, _ := helpers.GetCredentials()
			session := helpers.CF("delete-space", "-f", "-o", orgName, "please-do-not-exist-in-real-life")
			Eventually(session).Should(Say("Deleting space please-do-not-exist-in-real-life in org %s as %s...", orgName, username))
			Eventually(session).Should(Say("OK"))
			Eventually(session.Err).Should(Say(`Space 'please-do-not-exist-in-real-life' does not exist.`))
			Eventually(session).Should(Exit(0))
		})
	})

	When("the space exists", func() {
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

				It("deletes the space", func() {
					username, _ := helpers.GetCredentials()
					session := helpers.CFWithStdin(buffer, "delete-space", spaceName)
					Eventually(session).Should(Say(`This action impacts all resources scoped to this space, including apps, service instances, and space-scoped service brokers.`))
					Eventually(session).Should(Say(`Really delete the space %s\? \[yN\]`, spaceName))
					Eventually(session).Should(Say("Deleting space %s in org %s as %s...", spaceName, orgName, username))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))
					Eventually(helpers.CF("space", spaceName)).Should(Exit(1))
				})
			})

			When("the user enters 'n'", func() {
				BeforeEach(func() {
					_, err := buffer.Write([]byte("n\n"))
					Expect(err).ToNot(HaveOccurred())
				})

				It("does not delete the space", func() {
					session := helpers.CFWithStdin(buffer, "delete-space", spaceName)
					Eventually(session).Should(Say(`Really delete the space %s\? \[yN\]`, spaceName))
					Eventually(session).Should(Say("'%s' has not been deleted.", spaceName))
					Eventually(session).Should(Exit(0))
					Eventually(helpers.CF("space", spaceName)).Should(Exit(0))
				})
			})

			When("the user enters the default input (hits return)", func() {
				BeforeEach(func() {
					_, err := buffer.Write([]byte("\n"))
					Expect(err).ToNot(HaveOccurred())
				})

				It("does not delete the org", func() {
					session := helpers.CFWithStdin(buffer, "delete-space", spaceName)
					Eventually(session).Should(Say(`Really delete the space %s\? \[yN\]`, spaceName))
					Eventually(session).Should(Say("'%s' has not been deleted.", spaceName))
					Eventually(session).Should(Exit(0))
					Eventually(helpers.CF("space", spaceName)).Should(Exit(0))
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
					session := helpers.CFWithStdin(buffer, "delete-space", spaceName)
					Eventually(session).Should(Say(`Really delete the space %s\? \[yN\]`, spaceName))
					Eventually(session).Should(Say(`invalid input \(not y, n, yes, or no\)`))
					Eventually(session).Should(Say(`Really delete the space %s\? \[yN\]`, spaceName))
					Eventually(session).Should(Exit(0))
					Eventually(helpers.CF("space", spaceName)).Should(Exit(0))
				})
			})
		})

		When("the -f flag is provided", func() {
			It("deletes the space", func() {
				username, _ := helpers.GetCredentials()
				session := helpers.CF("delete-space", spaceName, "-f")
				Eventually(session).Should(Say(`Deleting space %s in org %s as %s\.\.\.`, spaceName, orgName, username))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))
				Eventually(helpers.CF("space", spaceName)).Should(Exit(1))
			})

			When("the space was targeted", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("target", "-s", spaceName)).Should(Exit(0))
				})

				It("deletes the space and clears the target", func() {
					username, _ := helpers.GetCredentials()
					session := helpers.CF("delete-space", spaceName, "-f")
					Eventually(session).Should(Say(`Deleting space %s in org %s as %s\.\.\.`, spaceName, orgName, username))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Say(`TIP: No space targeted, use 'cf target -s' to target a space\.`))
					Eventually(session).Should(Exit(0))

					Eventually(helpers.CF("space", spaceName)).Should(Exit(1))

					session = helpers.CF("target")
					Eventually(session).Should(Say("No space targeted, use 'cf target -s SPACE'"))
					Eventually(session).Should(Exit(0))
				})
			})
		})
	})

	When("the -o organization does not exist", func() {
		BeforeEach(func() {
			helpers.LoginCF()
		})

		It("fails and displays org not found", func() {
			username, _ := helpers.GetCredentials()
			session := helpers.CF("delete-space", "-f", "-o", "please-do-not-exist-in-real-life", "please-do-not-exist-in-real-life")
			Eventually(session).Should(Say("Deleting space please-do-not-exist-in-real-life in org please-do-not-exist-in-real-life as %s...", username))
			Eventually(session.Err).Should(Say(`Organization 'please-do-not-exist-in-real-life' not found\.`))
			Eventually(session).Should(Say("FAILED"))
			Eventually(session).Should(Exit(1))
		})
	})
})
