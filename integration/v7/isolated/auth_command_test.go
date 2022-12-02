package isolated

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	"code.cloudfoundry.org/cli/api/uaa/uaaversion"
	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/util/configv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("auth command", func() {
	BeforeEach(func() {
		helpers.SkipIfClientCredentialsTestMode()
	})

	Context("Help", func() {
		It("displays the help information", func() {
			session := helpers.CF("auth", "--help")
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Say("auth - Authenticate non-interactively\n\n"))

			Eventually(session).Should(Say("USAGE:"))
			Eventually(session).Should(Say("cf auth USERNAME PASSWORD\n"))
			Eventually(session).Should(Say("cf auth CLIENT_ID CLIENT_SECRET --client-credentials\n\n"))

			Eventually(session).Should(Say("ENVIRONMENT VARIABLES:"))
			Eventually(session).Should(Say(`CF_USERNAME=user\s+Authenticating user. Overridden if USERNAME argument is provided.`))
			Eventually(session).Should(Say(`CF_PASSWORD=password\s+Password associated with user. Overridden if PASSWORD argument is provided.`))

			Eventually(session).Should(Say("WARNING:"))
			Eventually(session).Should(Say("Providing your password as a command line option is highly discouraged"))
			Eventually(session).Should(Say("Your password may be visible to others and may be recorded in your shell history\n"))
			Eventually(session).Should(Say("Consider using the CF_PASSWORD environment variable instead\n\n"))

			Eventually(session).Should(Say("EXAMPLES:"))
			Eventually(session).Should(Say("cf auth name@example\\.com \"my password\" \\(use quotes for passwords with a space\\)"))
			Eventually(session).Should(Say("cf auth name@example\\.com \\\"\\\\\"password\\\\\"\\\" \\(escape quotes if used in password\\)\n\n"))

			Eventually(session).Should(Say("OPTIONS:"))
			Eventually(session).Should(Say("--client-credentials\\s+Use \\(non-user\\) service account \\(also called client credentials\\)\n"))
			Eventually(session).Should(Say("--origin\\s+Indicates the identity provider to be used for authentication\n\n"))

			Eventually(session).Should(Say("SEE ALSO:"))
			Eventually(session).Should(Say("api, login, target"))

			Eventually(session).Should(Exit(0))
		})
	})

	When("no positional arguments are provided", func() {
		Context("and no env variables are provided", func() {
			It("errors-out with the help information", func() {
				envWithoutLoginInfo := map[string]string{
					"CF_USERNAME": "",
					"CF_PASSWORD": "",
				}
				session := helpers.CFWithEnv(envWithoutLoginInfo, "auth")
				Eventually(session.Err).Should(Say("Username and password not provided."))
				Eventually(session).Should(Say("NAME:"))

				Eventually(session).Should(Exit(1))
			})
		})

		When("env variables are provided", func() {
			It("authenticates the user", func() {
				username, password := helpers.GetCredentials()
				env := map[string]string{
					"CF_USERNAME": username,
					"CF_PASSWORD": password,
				}
				session := helpers.CFWithEnv(env, "auth")

				Eventually(session).Should(Say("API endpoint: %s", helpers.GetAPI()))
				Eventually(session).Should(Say(`Authenticating\.\.\.`))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Say("Use 'cf target' to view or set your target org and space"))

				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("only a username is provided", func() {
		It("errors-out with a password required error and the help information", func() {
			envWithoutLoginInfo := map[string]string{
				"CF_USERNAME": "",
				"CF_PASSWORD": "",
			}
			session := helpers.CFWithEnv(envWithoutLoginInfo, "auth", "some-user")
			Eventually(session.Err).Should(Say("Password not provided."))
			Eventually(session).Should(Say("NAME:"))

			Eventually(session).Should(Exit(1))
		})
	})

	When("only a password is provided", func() {
		It("errors-out with a username required error and the help information", func() {
			env := map[string]string{
				"CF_USERNAME": "",
				"CF_PASSWORD": "some-pass",
			}
			session := helpers.CFWithEnv(env, "auth")
			Eventually(session.Err).Should(Say("Username not provided."))
			Eventually(session).Should(Say("NAME:"))

			Eventually(session).Should(Exit(1))
		})
	})

	When("extra input is given", func() {
		It("displays an 'unknown flag' error message", func() {
			session := helpers.CF("auth", "some-username", "some-password", "-a", "api.bosh-lite.com")

			Eventually(session.Err).Should(Say("Incorrect Usage: unknown flag `a'"))
			Eventually(session).Should(Say("NAME:"))

			Eventually(session).Should(Exit(1))
		})
	})

	When("the API endpoint is not set", func() {
		BeforeEach(func() {
			helpers.UnsetAPI()
		})

		It("displays an error message", func() {
			session := helpers.CF("auth", "some-username", "some-password")

			Eventually(session).Should(Say("FAILED"))
			Eventually(session.Err).Should(Say(`No API endpoint set\. Use 'cf login' or 'cf api' to target an endpoint\.`))

			Eventually(session).Should(Exit(1))
		})
	})

	When("no flags are set (logging in with password grant type)", func() {
		When("the user provides an invalid username/password combo", func() {
			BeforeEach(func() {
				helpers.LoginCF()
				helpers.TargetOrgAndSpace(ReadOnlyOrg, ReadOnlySpace)
			})

			It("clears the cached tokens and target info, then displays an error message", func() {
				session := helpers.CF("auth", "some-username", "some-password")

				Eventually(session).Should(Say("API endpoint: %s", helpers.GetAPI()))
				Eventually(session).Should(Say(`Authenticating\.\.\.`))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say(`Credentials were rejected, please try again\.`))
				Eventually(session).Should(Exit(1))

				// Verify that the user is not logged-in
				targetSession1 := helpers.CF("target")
				Eventually(targetSession1.Err).Should(Say(`Not logged in\. Use 'cf login' or 'cf login --sso' to log in\.`))
				Eventually(targetSession1).Should(Say("FAILED"))
				Eventually(targetSession1).Should(Exit(1))

				// Verify that neither org nor space is targeted
				helpers.LoginCF()
				targetSession2 := helpers.CF("target")
				Eventually(targetSession2).Should(Say("No org or space targeted, use 'cf target -o ORG -s SPACE'"))
				Eventually(targetSession2).Should(Exit(0))
			})
		})

		When("the username and password are valid", func() {
			It("authenticates the user", func() {
				username, password := helpers.GetCredentials()
				session := helpers.CF("auth", username, password)

				Eventually(session).Should(Say("API endpoint: %s", helpers.GetAPI()))
				Eventually(session).Should(Say(`Authenticating\.\.\.`))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Say("Use 'cf target' to view or set your target org and space"))

				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the 'client-credentials' flag is set", func() {
		When("the user provides an invalid client id/secret combo", func() {
			BeforeEach(func() {
				helpers.LoginCF()
				helpers.TargetOrgAndSpace(ReadOnlyOrg, ReadOnlySpace)
			})

			It("clears the cached tokens and target info, then displays an error message", func() {
				session := helpers.CF("auth", "some-client-id", "some-client-secret", "--client-credentials")

				Eventually(session).Should(Say("API endpoint: %s", helpers.GetAPI()))
				Eventually(session).Should(Say(`Authenticating\.\.\.`))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say(`Credentials were rejected, please try again\.`))
				Eventually(session).Should(Exit(1))

				// Verify that the user is not logged-in
				targetSession1 := helpers.CF("target")
				Eventually(targetSession1.Err).Should(Say(`Not logged in\. Use 'cf login' or 'cf login --sso' to log in\.`))
				Eventually(targetSession1).Should(Say("FAILED"))
				Eventually(targetSession1).Should(Exit(1))

				// Verify that neither org nor space is targeted
				helpers.LoginCF()
				targetSession2 := helpers.CF("target")
				Eventually(targetSession2).Should(Say("No org or space targeted, use 'cf target -o ORG -s SPACE'"))
				Eventually(targetSession2).Should(Exit(0))
			})
		})

		When("the client id and client secret are valid", func() {
			It("authenticates the user", func() {
				clientID, clientSecret := helpers.SkipIfClientCredentialsNotSet()
				session := helpers.CF("auth", clientID, clientSecret, "--client-credentials")

				Eventually(session).Should(Say("API endpoint: %s", helpers.GetAPI()))
				Eventually(session).Should(Say(`Authenticating\.\.\.`))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Say("Use 'cf target' to view or set your target org and space"))

				Eventually(session).Should(Exit(0))
			})

			It("writes the client id but does not write the client secret to the config file", func() {
				clientID, clientSecret := helpers.SkipIfClientCredentialsNotSet()
				session := helpers.CF("auth", clientID, clientSecret, "--client-credentials")
				Eventually(session).Should(Exit(0))

				rawConfig, err := ioutil.ReadFile(filepath.Join(homeDir, ".cf", "config.json"))
				Expect(err).NotTo(HaveOccurred())

				Expect(string(rawConfig)).ToNot(ContainSubstring(clientSecret))

				var configFile configv3.JSONConfig
				err = json.Unmarshal(rawConfig, &configFile)

				Expect(err).NotTo(HaveOccurred())
				Expect(configFile.UAAOAuthClient).To(Equal(clientID))
				Expect(configFile.UAAOAuthClientSecret).To(BeEmpty())
				Expect(configFile.UAAGrantType).To(Equal("client_credentials"))
			})
		})
	})

	When("a user authenticates with valid client credentials", func() {
		BeforeEach(func() {
			clientID, clientSecret := helpers.SkipIfClientCredentialsNotSet()
			session := helpers.CF("auth", clientID, clientSecret, "--client-credentials")
			Eventually(session).Should(Exit(0))
		})

		When("a different user authenticates with valid password credentials", func() {
			It("should fail authentication and display an error informing the user they need to log out", func() {
				username, password := helpers.GetCredentials()
				session := helpers.CF("auth", username, password)

				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say(`Service account currently logged in\. Use 'cf logout' to log out service account and try again\.`))
				Eventually(session).Should(Exit(1))
			})
		})

	})

	When("the origin flag is set", func() {
		When("the UAA version is too low to use the --origin flag", func() {
			BeforeEach(func() {
				helpers.SkipIfUAAVersionAtLeast(uaaversion.MinUAAClientVersion)
			})
			It("prints an error message", func() {
				session := helpers.CF("auth", "some-username", "some-password", "--client-credentials", "--origin", "garbage")
				Eventually(session.Err).Should(Say("Option '--origin' requires UAA API version 4.19.0 or higher. Update your Cloud Foundry instance."))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the UAA version is recent enough to support the flag", func() {
			BeforeEach(func() {
				helpers.SkipIfUAAVersionLessThan(uaaversion.MinUAAClientVersion)
			})
			When("--client-credentials is also set", func() {
				It("displays the appropriate error message", func() {
					session := helpers.CF("auth", "some-username", "some-password", "--client-credentials", "--origin", "garbage")

					Eventually(session.Err).Should(Say("Incorrect Usage: The following arguments cannot be used together: --client-credentials, --origin"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("a user authenticates with valid user credentials for that origin", func() {
				var (
					username string
					password string
				)

				BeforeEach(func() {
					username, password = helpers.SkipIfOIDCCredentialsNotSet()
				})

				It("authenticates the user", func() {
					session := helpers.CF("auth", username, password, "--origin", "cli-oidc-provider")

					Eventually(session).Should(Say("API endpoint: %s", helpers.GetAPI()))
					Eventually(session).Should(Say(`Authenticating\.\.\.`))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Say("Use 'cf target' to view or set your target org and space"))
					Eventually(session).Should(Exit(0))
				})
			})

			When("the user provides the default origin and valid credentials", func() {
				It("authenticates the user", func() {
					username, password := helpers.GetCredentials()
					session := helpers.CF("auth", username, password, "--origin", "uaa")

					Eventually(session).Should(Say("API endpoint: %s", helpers.GetAPI()))
					Eventually(session).Should(Say(`Authenticating\.\.\.`))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Say("Use 'cf target' to view or set your target org and space"))
					Eventually(session).Should(Exit(0))
				})
			})

			When("when the user provides an invalid origin", func() {
				It("returns an error", func() {
					session := helpers.CF("auth", "some-user", "some-password", "--origin", "EA")
					Eventually(session.Err).Should(Say("The origin provided is invalid."))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				})
			})
		})
	})

	Describe("Authenticating as a user, through a custom client", func() {
		var (
			session *Session
		)

		BeforeEach(func() {
			customClientID, customClientSecret := helpers.SkipIfCustomClientCredentialsNotSet()

			helpers.LoginCF()

			helpers.SetConfig(func(config *configv3.Config) {
				config.ConfigFile.UAAOAuthClient = customClientID
				config.ConfigFile.UAAOAuthClientSecret = customClientSecret
				config.ConfigFile.UAAGrantType = ""
			})
		})

		It("fails and shows a helpful message", func() {
			username, password := helpers.CreateUser()

			session = helpers.CF("auth", username, password)

			Eventually(session).Should(Exit(1))
			Eventually(session.Err).Should(Say("Error: Support for manually writing your client credentials to config.json has been removed. For similar functionality please use `cf auth --client-credentials`."))
		})
	})
})
