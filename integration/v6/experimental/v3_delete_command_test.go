package experimental

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("v3-delete command", func() {
	var (
		orgName   string
		spaceName string
		appName   string
	)

	BeforeEach(func() {
		orgName = helpers.NewOrgName()
		spaceName = helpers.NewSpaceName()
		appName = helpers.PrefixedRandomName("app")
	})

	When("--help flag is set", func() {
		It("Displays command usage to output", func() {
			session := helpers.CF("v3-delete", "--help")
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Say("v3-delete - Delete a V3 App"))
			Eventually(session).Should(Say("USAGE:"))
			Eventually(session).Should(Say(`cf v3-delete APP_NAME \[-f\]`))
			Eventually(session).Should(Say("OPTIONS:"))
			Eventually(session).Should(Say(`\s+-f\s+Force deletion without confirmation`))
			Eventually(session).Should(Exit(0))
		})
	})

	When("the app name is not provided", func() {
		It("tells the user that the app name is required, prints help text, and exits 1", func() {
			session := helpers.CF("v3-delete")

			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `APP_NAME` was not provided"))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	It("displays the experimental warning", func() {
		session := helpers.CF("v3-delete", appName)
		Eventually(session.Err).Should(Say("This command is in EXPERIMENTAL stage and may change without notice"))
		Eventually(session).Should(Exit())
	})

	When("the environment is not setup correctly", func() {
		When("no API endpoint is set", func() {
			BeforeEach(func() {
				helpers.UnsetAPI()
			})

			It("fails with no API endpoint set message", func() {
				session := helpers.CF("v3-delete", appName)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("No API endpoint set. Use 'cf login' or 'cf api' to target an endpoint."))
				Eventually(session).Should(Exit(1))
			})
		})

		When("not logged in", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
			})

			It("fails with not logged in message", func() {
				session := helpers.CF("v3-delete", appName)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("Not logged in. Use 'cf login' or 'cf login --sso' to log in."))
				Eventually(session).Should(Exit(1))
			})
		})

		When("there is no org set", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
				helpers.LoginCF()
			})

			It("fails with no targeted org error message", func() {
				session := helpers.CF("v3-delete", appName)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("No org targeted, use 'cf target -o ORG' to target an org."))
				Eventually(session).Should(Exit(1))
			})
		})

		When("there is no space set", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
				helpers.LoginCF()
				helpers.TargetOrg(ReadOnlyOrg)
			})

			It("fails with no targeted space error message", func() {
				session := helpers.CF("v3-delete", appName)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("No space targeted, use 'cf target -s SPACE' to target a space."))
				Eventually(session).Should(Exit(1))
			})
		})
	})

	When("the environment is setup correctly", func() {
		BeforeEach(func() {
			helpers.SetupCF(orgName, spaceName)
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		When("the app does not exist", func() {
			When("the -f flag is provided", func() {
				It("it displays the app does not exist", func() {
					username, _ := helpers.GetCredentials()
					session := helpers.CF("v3-delete", appName, "-f")
					Eventually(session).Should(Say("Deleting app %s in org %s / space %s as %s...", appName, orgName, spaceName, username))
					Eventually(session).Should(Say("App %s does not exist", appName))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))
				})
			})

			When("the -f flag not is provided", func() {
				var buffer *Buffer

				BeforeEach(func() {
					buffer = NewBuffer()
				})

				When("the user enters 'y'", func() {
					BeforeEach(func() {
						_, err := buffer.Write([]byte("y\n"))
						Expect(err).NotTo(HaveOccurred())
					})

					It("it displays the app does not exist", func() {
						username, _ := helpers.GetCredentials()
						session := helpers.CFWithStdin(buffer, "v3-delete", appName)
						Eventually(session).Should(Say(`Really delete the app %s\? \[yN\]`, appName))
						Eventually(session).Should(Say("Deleting app %s in org %s / space %s as %s...", appName, orgName, spaceName, username))
						Eventually(session).Should(Say("App %s does not exist", appName))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Exit(0))
					})
				})

				When("the user enters 'n'", func() {
					BeforeEach(func() {
						_, err := buffer.Write([]byte("n\n"))
						Expect(err).NotTo(HaveOccurred())
					})

					It("does not delete the app", func() {
						session := helpers.CFWithStdin(buffer, "v3-delete", appName)
						Eventually(session).Should(Say(`Really delete the app %s\? \[yN\]`, appName))
						Eventually(session).Should(Say("Delete cancelled"))
						Eventually(session).Should(Exit(0))
					})
				})

				When("the user enters the default input (hits return)", func() {
					BeforeEach(func() {
						_, err := buffer.Write([]byte("\n"))
						Expect(err).NotTo(HaveOccurred())
					})

					It("does not delete the app", func() {
						session := helpers.CFWithStdin(buffer, "v3-delete", appName)
						Eventually(session).Should(Say(`Really delete the app %s\? \[yN\]`, appName))
						Eventually(session).Should(Say("Delete cancelled"))
						Eventually(session).Should(Exit(0))
					})
				})

				When("the user enters an invalid answer", func() {
					BeforeEach(func() {
						// The second '\n' is intentional. Otherwise the buffer will be
						// closed while the interaction is still waiting for input; it gets
						// an EOF and causes an error.
						_, err := buffer.Write([]byte("wat\n\n"))
						Expect(err).NotTo(HaveOccurred())
					})

					It("asks again", func() {
						session := helpers.CFWithStdin(buffer, "v3-delete", appName)
						Eventually(session).Should(Say(`Really delete the app %s\? \[yN\]`, appName))
						Eventually(session).Should(Say(`invalid input \(not y, n, yes, or no\)`))
						Eventually(session).Should(Say(`Really delete the app %s\? \[yN\]`, appName))
						Eventually(session).Should(Exit(0))
					})
				})
			})
		})

		When("the app exists", func() {
			BeforeEach(func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "v3-push", appName)).Should(Exit(0))
				})
			})

			It("deletes the app", func() {
				session := helpers.CF("v3-delete", appName, "-f")
				username, _ := helpers.GetCredentials()
				Eventually(session).Should(Say("Deleting app %s in org %s / space %s as %s...", appName, orgName, spaceName, username))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))

				Eventually(helpers.CF("v3-app", appName)).Should(Exit(1))
			})
		})
	})
})
