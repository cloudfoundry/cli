package experimental

import (
	"net/http"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("login command", func() {
	BeforeEach(func() {
		helpers.SkipIfClientCredentialsTestMode()
		helpers.TurnOnExperimentalLogin()
	})

	AfterEach(func() {
		helpers.TurnOffExperimentalLogin()
	})

	Describe("Minimum Version Check", func() {
		When("the CLI version is lower than the minimum supported version by the CC", func() {
			var server *ghttp.Server

			BeforeEach(func() {
				server = helpers.StartServerWithMinimumCLIVersion("9000.0.0")

				fakeTokenResponse := map[string]string{
					"access_token": "",
					"token_type":   "bearer",
				}
				server.RouteToHandler(http.MethodPost, "/oauth/token",
					ghttp.RespondWithJSONEncoded(http.StatusOK, fakeTokenResponse))
				server.RouteToHandler(http.MethodGet, "/v3/organizations",
					ghttp.RespondWith(http.StatusOK, `{
					 "total_results": 0,
					 "total_pages": 1,
					 "resources": []}`))
			})

			AfterEach(func() {
				server.Close()
			})

			It("displays the warning and exits successfully", func() {
				session := helpers.CF("login", "-a", server.URL(), "--skip-ssl-validation")
				Eventually(session.Err).Should(Say(`Cloud Foundry API version .+ requires CLI version .+\. You are currently on version .+\. To upgrade your CLI, please visit: https://github.com/cloudfoundry/cli#downloads`))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Describe("API Endpoint", func() {
		When("the API endpoint is not set", func() {
			BeforeEach(func() {
				helpers.UnsetAPI()
			})

			When("the user does not provide the -a flag", func() {
				It("prompts the user for an endpoint", func() {
					input := NewBuffer()
					_, err := input.Write([]byte("\n"))
					Expect(err).ToNot(HaveOccurred())
					session := helpers.CFWithStdin(input, "login")
					Eventually(session).Should(Say("API endpoint:"))
					session.Interrupt()
					Eventually(session).Should(Exit())
				})

				When("the API endpoint provided at the prompt is unreachable", func() {
					It("returns an error", func() {
						input := NewBuffer()
						_, err := input.Write([]byte("does.not.exist\n"))
						Expect(err).ToNot(HaveOccurred())
						session := helpers.CFWithStdin(input, "login")
						Eventually(session).Should(Say("API endpoint:"))
						Eventually(session).Should(Say("FAILED"))
						Eventually(session.Err).Should(Say("Request error: "))
						Eventually(session.Err).Should(Say("TIP: If you are behind a firewall and require an HTTP proxy, verify the https_proxy environment variable is correctly set. Else, check your network connection."))
						Eventually(session).Should(Exit(1))
					})
				})
			})

			When("the user provides the -a flag", func() {
				It("sets the API endpoint and does not prompt the user for the API endpoint", func() {
					var session *Session
					if skipSSLValidation {
						session = helpers.CF("login", "-a", apiURL, "--skip-ssl-validation")
					} else {
						session = helpers.CF("login", "-a", apiURL)
					}
					Eventually(session).Should(Say("API endpoint: %s", apiURL))
					// TODO https://www.pivotaltracker.com/story/show/166938709/comments/204492216
					//Consistently(session).ShouldNot(Say("API endpoint:"))
					//session.Interrupt()
					Eventually(session).Should(Exit())

					session = helpers.CF("api")
					Eventually(session).Should(Exit(0))
					Expect(session).Should(Say("api endpoint:   %s", apiURL))
				})

				When("the provided API endpoint is unreachable", func() {
					It("displays an error and fails", func() {
						var session *Session
						if skipSSLValidation {
							session = helpers.CF("login", "-a", "does.not.exist", "--skip-ssl-validation")
						} else {
							session = helpers.CF("login", "-a", "does.not.exist")
						}

						Eventually(session).Should(Say("API endpoint: does.not.exist"))
						Eventually(session).Should(Say("FAILED"))
						Eventually(session.Err).Should(Say("Request error: "))
						Eventually(session.Err).Should(Say("TIP: If you are behind a firewall and require an HTTP proxy, verify the https_proxy environment variable is correctly set. Else, check your network connection."))
						Eventually(session).Should(Exit(1))
					})
				})

				When("the provided API endpoint has trailing slashes", func() {
					It("removes the extra slashes", func() {
						username, password := helpers.GetCredentials()
						apiURLWithSlash := apiURL + "////"
						session := helpers.CF("login", "-a", apiURLWithSlash, "-u", username, "-p", password, "--skip-ssl-validation")
						Eventually(session).Should(Exit(0))

						session = helpers.CF("api")
						Eventually(session).Should(Say("api endpoint:\\s+%s\n", apiURL))
						Eventually(session).Should(Exit(0))
					})
				})
			})
		})

		When("the API endpoint is already set", func() {
			It("does not prompt the user for API endpoint", func() {
				session := helpers.CF("login")
				Consistently(session).ShouldNot(Say("API endpoint>"))
				session.Interrupt()
				Eventually(session).Should(Exit())
			})

			When("the user provides a new API endpoint with the -a flag", func() {
				When("the provided API endpoint is unreachable", func() {
					It("displays an error and does not change the API endpoint", func() {
						var session *Session
						if skipSSLValidation {
							session = helpers.CF("login", "-a", "does.not.exist", "--skip-ssl-validation")
						} else {
							session = helpers.CF("login", "-a", "does.not.exist")
						}
						Eventually(session).Should(Say("API endpoint: does.not.exist"))
						Eventually(session).Should(Say("FAILED"))
						Eventually(session.Err).Should(Say("Request error: "))
						Eventually(session.Err).Should(Say("TIP: If you are behind a firewall and require an HTTP proxy, verify the https_proxy environment variable is correctly set. Else, check your network connection."))
						Eventually(session).Should(Exit(1))

						apiSession := helpers.CF("api")
						Eventually(apiSession).Should(Exit(0))
						Eventually(apiSession).Should(Say("api endpoint:   %s", apiURL))
					})
				})
			})
		})
	})

	Describe("SSO", func() {
		When("--sso-passcode is provided", func() {
			Context("and --sso is also passed", func() {
				It("fails with a useful error message", func() {
					session := helpers.CF("login", "--sso-passcode", "some-passcode", "--sso")
					Eventually(session.Err).Should(Say("Incorrect Usage: The following arguments cannot be used together: --sso-passcode, --sso"))
					Eventually(session).Should(Exit(1))
				})
			})

			Context("and the provided passcode is incorrect", func() {
				It("prompts twice, displays an error and fails", func() {
					input := NewBuffer()
					_, err := input.Write([]byte("bad-passcode-again\nbad-passcode-strikes-back\n"))
					Expect(err).ToNot(HaveOccurred())
					session := helpers.CFWithStdin(input, "login", "--sso-passcode", "some-passcode")
					Eventually(session).Should(Say("API endpoint:\\s+" + helpers.GetAPI()))
					Eventually(session).Should(Say(`Authenticating\.\.\.`))
					Eventually(session.Err).Should(Say(`Invalid passcode`))

					// Leaving out expectation of prompt text, since it comes from UAA (and doesn't show up on Windows)
					Eventually(session).Should(Say(`Authenticating\.\.\.`))
					Eventually(session.Err).Should(Say(`Invalid passcode`))
					Eventually(session).Should(Say(`Authenticating\.\.\.`))
					Eventually(session.Err).Should(Say(`Invalid passcode`))
					Eventually(session).Should(Say(`API endpoint:\s+` + helpers.GetAPI() + `\s+\(API version: \d\.\d{1,3}\.\d{1,3}\)`))
					Eventually(session).Should(Say(`Not logged in. Use 'cf login' to log in.`))
					Eventually(session.Err).Should(Say(`Unable to authenticate`))
					Eventually(session).Should(Say(`FAILED`))

					Eventually(session).Should(Exit(1))
				})
			})

			When("a passcode isn't provided", func() {
				It("prompts the user to try again", func() {
					session := helpers.CF("login", "--sso-passcode")
					Eventually(session.Err).Should(Say("Incorrect Usage: expected argument for flag `--sso-passcode'"))
					Eventually(session).Should(Exit(1))
				})
			})
		})

		When("a user authenticates with valid client credentials", func() {
			BeforeEach(func() {
				clientID, clientSecret := helpers.SkipIfClientCredentialsNotSet()
				session := helpers.CF("auth", clientID, clientSecret, "--client-credentials")
				Eventually(session).Should(Exit(0))
			})

			When("a different user logs in with valid sso passcode", func() {
				It("should fail log in and display an error informing the user they need to log out", func() {
					session := helpers.CF("login", "--sso-passcode", "my-fancy-sso")

					Eventually(session).Should(Say("API endpoint:\\s+" + helpers.GetAPI()))
					Eventually(session).Should(Say(`API endpoint:\s+` + helpers.GetAPI() + `\s+\(API version: \d\.\d{1,3}\.\d{1,3}\)`))
					Eventually(session.Err).Should(Say(`Service account currently logged in\. Use 'cf logout' to log out service account and try again\.`))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))

					//And I am still logged in
					targetSession := helpers.CF("target")
					Eventually(targetSession).Should(Exit(0))
				})
			})
		})
	})

	Describe("Target Space", func() {
		var (
			orgName  string
			username string
			password string
		)

		BeforeEach(func() {
			helpers.LoginCF()
			orgName = helpers.NewOrgName()
			session := helpers.CF("create-org", orgName)
			Eventually(session).Should(Exit(0))
			username, password = helpers.CreateUserInOrgRole(orgName, "OrgManager")
		})

		When("only one space is available to the user", func() {
			var spaceName string

			BeforeEach(func() {
				spaceName = helpers.NewSpaceName()
				session := helpers.CF("create-space", "-o", orgName, spaceName)
				Eventually(session).Should(Exit(0))
				roleSession := helpers.CF("set-space-role", username, orgName, spaceName, "SpaceManager")
				Eventually(roleSession).Should(Exit(0))
			})

			It("logs the user in and targets the org and the space", func() {
				session := helpers.CF("login", "-u", username, "-p", password, "-a", apiURL, "--skip-ssl-validation")
				Eventually(session).Should(Exit(0))

				targetSession := helpers.CF("target")
				Eventually(targetSession).Should(Exit(0))
				Eventually(targetSession).Should(Say(`org:\s+%s`, orgName))
				Eventually(targetSession).Should(Say(`space:\s+%s`, spaceName))
			})

		})

	})

})
