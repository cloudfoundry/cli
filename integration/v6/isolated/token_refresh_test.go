package isolated

import (
	"fmt"

	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/util/configv3"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Token Refreshing", func() {
	Describe("password grant type", func() {
		BeforeEach(func() {
			helpers.LoginCF()
			helpers.SkipIfClientCredentialsTestMode()
		})

		Describe("config file backwards compatibility", func() {
			// If we write "password" as the grant type, versions of the CLI before 6.44.0 will not be
			// able to use their refresh token correctly.
			When("logging in with rewritten cf auth", func() {
				BeforeEach(func() {
					helpers.LoginCF()
				})

				It("persists an empty string as the grant type in config.json", func() {
					c := helpers.GetConfig()
					Expect(c.UAAGrantType()).To(Equal(""))
				})
			})

			When("logging in with un-rewritten cf login", func() {
				BeforeEach(func() {
					helpers.TurnOffExperimentalLogin()
					u, p := helpers.GetCredentials()
					session := helpers.CF("login", "-u", u, "-p", p)
					Eventually(session).Should(Exit(0))
				})

				It("persists an empty string as the grant type in config.json", func() {
					c := helpers.GetConfig()
					Expect(c.UAAGrantType()).To(Equal(""))
				})
			})

			When("logging in with rewritten cf login", func() {
				BeforeEach(func() {
					helpers.TurnOnExperimentalLogin()
					u, p := helpers.GetCredentials()
					session := helpers.CF("login", "-u", u, "-p", p)
					Eventually(session).Should(Exit(0))
				})

				AfterEach(func() {
					helpers.TurnOffExperimentalLogin()
				})

				It("persists an empty string as the grant type in config.json", func() {
					c := helpers.GetConfig()
					Expect(c.UAAGrantType()).To(Equal(""))
				})
			})
		})

		When("the token is invalid", func() {
			When("password is explicitly stored as the grant type", func() {
				BeforeEach(func() {
					helpers.SetConfig(func(config *configv3.Config) {
						config.ConfigFile.AccessToken = helpers.ExpiredAccessToken()
						config.ConfigFile.TargetedOrganization.GUID = "fake-org"
						config.ConfigFile.TargetedSpace.GUID = "fake-space"
						config.ConfigFile.UAAGrantType = "password"
					})
				})

				When("running a v6 command", func() {
					When("the cloud controller client encounters an invalid token response", func() {
						It("refreshes the token", func() {
							session := helpers.CF("unbind-service", "app", "service")
							Eventually(session.Err).Should(Say("App 'app' not found"))
							Eventually(session).Should(Exit(1))
						})
					})

					When("the UAA client encounters an invalid token response", func() {
						It("refreshes the token", func() {
							username, _ := helpers.GetCredentials()
							session := helpers.CF("create-user", username, helpers.NewPassword())
							Eventually(session.Err).Should(Say(fmt.Sprintf("user %s already exists", username)))
							Eventually(session).Should(Exit(0))
						})
					})
				})

				When("running an unrefactored v6 command", func() {
					It("refreshes the token", func() {
						session := helpers.CF("stack", "some-stack")
						Eventually(session).Should(Say("Stack some-stack not found"))
						Eventually(session).Should(Exit(1))
					})
				})
			})

			When("no grant type is explicitly stored", func() {
				BeforeEach(func() {
					helpers.SetConfig(func(config *configv3.Config) {
						config.ConfigFile.AccessToken = helpers.ExpiredAccessToken()
						config.ConfigFile.TargetedOrganization.GUID = "fake-org"
						config.ConfigFile.TargetedSpace.GUID = "fake-space"
						config.ConfigFile.UAAGrantType = ""
					})
				})

				When("running a v6 command", func() {
					When("the cloud controller client encounters an invalid token response", func() {
						It("refreshes the token", func() {
							session := helpers.CF("unbind-service", "app", "service")
							Eventually(session.Err).Should(Say("App 'app' not found"))
							Eventually(session).Should(Exit(1))
						})
					})

					When("the UAA client encounters an invalid token response", func() {
						It("refreshes the token", func() {
							username, _ := helpers.GetCredentials()
							session := helpers.CF("create-user", username, helpers.NewPassword())
							Eventually(session.Err).Should(Say(fmt.Sprintf("user %s already exists", username)))
							Eventually(session).Should(Exit(0))
						})
					})
				})

				When("running an unrefactored v6 command", func() {
					It("refreshes the token", func() {
						session := helpers.CF("stack", "some-stack")
						Eventually(session).Should(Say("Stack some-stack not found"))
						Eventually(session).Should(Exit(1))
					})
				})
			})
		})
	})

	Describe("client grant type", func() {
		BeforeEach(func() {
			helpers.LoginCFWithClientCredentials()
		})

		When("the token is invalid", func() {
			BeforeEach(func() {
				helpers.SetConfig(func(config *configv3.Config) {
					config.ConfigFile.AccessToken = helpers.ExpiredAccessToken()
					config.ConfigFile.TargetedOrganization.GUID = "fake-org"
					config.ConfigFile.TargetedSpace.GUID = "fake-space"
				})
			})

			When("running a v6 refactored command", func() {
				When("the cloud controller client encounters an invalid token response", func() {
					It("displays an error and exits 1", func() {
						session := helpers.CF("unbind-service", "app", "service")
						Eventually(session.Err).Should(Say(`Credentials were rejected, please try again\.`))
						Eventually(session).Should(Exit(1))
					})
				})

				When("the UAA client encounters an invalid token response", func() {
					It("displays an error and exits 1", func() {
						username := helpers.NewUsername()
						session := helpers.CF("create-user", username, helpers.NewPassword())
						Eventually(session.Err).Should(Say(`Credentials were rejected, please try again\.`))
						Eventually(session).Should(Exit(1))
					})
				})
			})

			// Client Credentials do not refresh token anymore
			When("running a v6 unrefactored command", func() {
				When("the cloud controller client encounters an invalid token response", func() {
					It("displays an error and exits 1", func() {
						session := helpers.CF("quotas")
						// Unrefactored code doesn't show correct username
						Eventually(session).Should(Say("Getting quotas as"))
						Eventually(session).Should(Say("Bad credentials"))
						Eventually(session).Should(Exit(1))
					})
				})
			})
		})

		When("the CLI has authenticated with --client-credentials", func() {
			When("the user has manually stored the client credentials in the config file and the token is expired", func() {
				BeforeEach(func() {
					clientID, clientSecret := helpers.SkipIfClientCredentialsNotSet()

					helpers.SetConfig(func(config *configv3.Config) {
						config.ConfigFile.UAAGrantType = "client_credentials"
						config.ConfigFile.UAAOAuthClient = clientID
						config.ConfigFile.UAAOAuthClientSecret = clientSecret
					})

					helpers.SetConfig(func(config *configv3.Config) {
						config.ConfigFile.AccessToken = helpers.ExpiredAccessToken()
					})
				})

				It("automatically gets a new access token", func() {
					Eventually(helpers.CF("orgs")).Should(Exit(0))
				})
			})
		})
	})
})
