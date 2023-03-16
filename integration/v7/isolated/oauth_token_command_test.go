package isolated

import (
	"regexp"

	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/util/configv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("oauth-token command", func() {
	Context("help", func() {
		It("displays the help information", func() {
			session := helpers.CF("oauth-token", "--help")

			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Say("oauth-token - Display the OAuth token for the current session and refresh the token if necessary"))
			Eventually(session).Should(Say("USAGE:"))
			Eventually(session).Should(Say("cf oauth-token"))
			Eventually(session).Should(Say("SEE ALSO:"))
			Eventually(session).Should(Say("curl"))
			Eventually(session).Should(Exit(0))
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(false, false, ReadOnlyOrg, "oauth-token")
		})
	})

	When("the current access token is valid", func() {
		BeforeEach(func() {
			helpers.SkipIfClientCredentialsTestMode()
			helpers.LoginCF()
		})

		It("reads the access token from the config file", func() {
			existingAccessToken := helpers.GetConfig().ConfigFile.AccessToken

			session := helpers.CF("oauth-token").Wait()

			output := string(session.Out.Contents())
			Expect(output).To(ContainSubstring(existingAccessToken))
			Expect(output).To(MatchRegexp("bearer .+"))

			Expect(session.Err.Contents()).To(BeEmpty())
			Expect(session.ExitCode()).To(Equal(0))
		})
	})

	When("the current access token needs to be refreshed", func() {
		BeforeEach(func() {
			helpers.SkipIfClientCredentialsTestMode()
			helpers.LoginCF()
			helpers.SetConfig(func(conf *configv3.Config) {
				conf.ConfigFile.AccessToken = "change-me-token"
			})
		})

		It("refreshes the access token and displays it", func() {
			existingAccessToken := helpers.GetConfig().ConfigFile.AccessToken

			session := helpers.CF("oauth-token").Wait()

			output := string(session.Out.Contents())
			Expect(output).ToNot(ContainSubstring(existingAccessToken))
			Expect(output).To(MatchRegexp("bearer .+"))

			Expect(session.Err.Contents()).To(BeEmpty())
			Expect(session.ExitCode()).To(Equal(0))
		})

		When("the refresh token is invalid", func() {
			BeforeEach(func() {
				helpers.SetConfig(func(conf *configv3.Config) {
					conf.ConfigFile.RefreshToken = "invalid-refresh-token"
				})
			})

			It("displays an error and exits 1", func() {
				session := helpers.CF("oauth-token")

				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say(`The token expired, was revoked, or the token ID is incorrect\. Please log back in to re-authenticate\.`))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the oauth client ID and secret combination is invalid", func() {
			BeforeEach(func() {
				helpers.SetConfig(func(conf *configv3.Config) {
					conf.ConfigFile.UAAOAuthClient = "nonexistent-client"
					conf.ConfigFile.UAAOAuthClientSecret = "some-secret"
				})
			})

			It("displays an error and exits 1", func() {
				session := helpers.CF("oauth-token")

				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say(`Credentials were rejected, please try again\.`))
				Eventually(session).Should(Exit(1))
			})
		})
	})

	Context("using client credentials", func() {
		BeforeEach(func() {
			helpers.LoginCFWithClientCredentials()
		})

		When("the access token is valid", func() {
			It("displays the current access token without trying to re-authenticate", func() {
				existingAccessToken := helpers.GetConfig().ConfigFile.AccessToken
				session := helpers.CF("oauth-token")

				Eventually(session).Should(Say(regexp.QuoteMeta(existingAccessToken)))
				Eventually(session).Should(Exit(0))
			})
		})

		When("the access token needs to be refreshed", func() {
			BeforeEach(func() {
				helpers.SetConfig(func(conf *configv3.Config) {
					conf.ConfigFile.AccessToken = "invalid-access-token"
				})
			})

			When("the provided client credentials are valid", func() {
				BeforeEach(func() {
					clientID, clientSecret := helpers.SkipIfClientCredentialsNotSet()

					helpers.SetConfig(func(conf *configv3.Config) {
						conf.ConfigFile.UAAOAuthClient = clientID
						conf.ConfigFile.UAAOAuthClientSecret = clientSecret
					})
				})

				It("re-authenticates and displays the new access token", func() {
					session := helpers.CF("oauth-token")

					Eventually(session).Should(Say("bearer .+"))
					Eventually(session).Should(Exit(0))
				})
			})

			When("the oauth client ID and secret combination is invalid", func() {
				BeforeEach(func() {
					helpers.SetConfig(func(conf *configv3.Config) {
						conf.ConfigFile.UAAOAuthClient = "nonexistent-client"
						conf.ConfigFile.UAAOAuthClientSecret = "some-secret"
					})
				})

				It("displays an error and exits 1", func() {
					session := helpers.CF("oauth-token")

					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say(`Credentials were rejected, please try again\.`))
					Eventually(session).Should(Exit(1))
				})
			})

			When("the client credentials are not present in the config", func() {
				It("displays an error and exits 1", func() {
					session := helpers.CF("oauth-token")

					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say(`Access token is invalid\.`))
					Eventually(session).Should(Exit(1))
				})
			})
		})
	})
})
