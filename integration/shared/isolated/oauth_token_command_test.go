package isolated

import (
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
			Eventually(session).Should(Say("oauth-token - Retrieve and display the OAuth token for the current session"))
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

	When("the environment is setup correctly and user is logged in with password grant", func() {
		BeforeEach(func() {
			helpers.LoginCF()
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
					conf.ConfigFile.UAAOAuthClient = "foo"
					conf.ConfigFile.UAAOAuthClientSecret = "bar"
				})
			})

			It("displays an error and exits 1", func() {
				session := helpers.CF("oauth-token")

				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say(`Credentials were rejected, please try again\.`))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the refresh token and oauth creds are valid", func() {
			It("refreshes the access token and displays it", func() {
				session := helpers.CF("oauth-token")

				Eventually(session).Should(Say("bearer .+"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the environment is setup correctly and user is logged in with client credentials grant", func() {
		BeforeEach(func() {
			helpers.LoginCFWithClientCredentials()
		})

		When("the oauth client ID and secret combination is invalid", func() {
			BeforeEach(func() {
				helpers.SetConfig(func(conf *configv3.Config) {
					conf.ConfigFile.UAAOAuthClient = "foo"
					conf.ConfigFile.UAAOAuthClientSecret = "bar"
				})
			})

			It("displays an error and exits 1", func() {
				session := helpers.CF("oauth-token")

				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say(`Credentials were rejected, please try again\.`))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the access token is invalid", func() {
			BeforeEach(func() {
				helpers.SetConfig(func(conf *configv3.Config) {
					conf.ConfigFile.AccessToken = "invalid-access-token"
				})
			})

			It("refreshes the access token and displays it", func() {
				session := helpers.CF("oauth-token")

				Eventually(session).Should(Say("bearer .+"))
				Eventually(session).Should(Exit(0))
			})
		})

		When("the oauth creds are valid", func() {
			It("refreshes the access token and displays it", func() {
				session := helpers.CF("oauth-token")

				Eventually(session).Should(Say("bearer .+"))
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
