package isolated

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"

	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/util/configv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("login command", func() {
	BeforeEach(func() {
		helpers.SkipIfClientCredentialsTestMode()
	})

	Describe("Help Text", func() {
		When("--help flag is set", func() {
			It("displays the command usage", func() {
				session := helpers.CF("login", "--help")
				Eventually(session).Should(Exit(0))

				Expect(session).Should(Say("NAME:\n"))
				Expect(session).Should(Say("login - Log user in"))

				Expect(session).Should(Say("USAGE:\n"))
				Expect(session).Should(Say(`cf login \[-a API_URL\] \[-u USERNAME\] \[-p PASSWORD\] \[-o ORG\] \[-s SPACE\] \[--sso | --sso-passcode PASSCODE\] \[--origin ORIGIN\]`))

				Expect(session).Should(Say("WARNING:\n"))
				Expect(session).Should(Say("Providing your password as a command line option is highly discouraged\n"))
				Expect(session).Should(Say("Your password may be visible to others and may be recorded in your shell history\n"))

				Expect(session).Should(Say("EXAMPLES:\n"))
				Expect(session).Should(Say(regexp.QuoteMeta("cf login (omit username and password to login interactively -- cf will prompt for both)")))
				Expect(session).Should(Say(regexp.QuoteMeta("cf login -u name@example.com -p pa55woRD (specify username and password as arguments)")))
				Expect(session).Should(Say(regexp.QuoteMeta("cf login -u name@example.com -p \"my password\" (use quotes for passwords with a space)")))
				Expect(session).Should(Say(regexp.QuoteMeta("cf login -u name@example.com -p \"\\\"password\\\"\" (escape quotes if used in password)")))
				Expect(session).Should(Say(regexp.QuoteMeta("cf login --sso (cf will provide a url to obtain a one-time passcode to login)")))
				Expect(session).Should(Say(regexp.QuoteMeta("cf login --origin ldap")))

				Expect(session).Should(Say("ALIAS:\n"))
				Expect(session).Should(Say("l"))

				Expect(session).Should(Say("OPTIONS:\n"))
				Expect(session).Should(Say(`-a\s+API endpoint \(e.g. https://api\.example\.com\)`))
				Expect(session).Should(Say(`-o\s+Org`))
				Expect(session).Should(Say(`-p\s+Password`))
				Expect(session).Should(Say(`-s\s+Space`))
				Expect(session).Should(Say(`--skip-ssl-validation\s+Skip verification of the API endpoint\. Not recommended\!`))
				Expect(session).Should(Say(`--sso\s+Prompt for a one-time passcode to login`))
				Expect(session).Should(Say(`--sso-passcode\s+One-time passcode`))
				Expect(session).Should(Say(`-u\s+Username`))

				Expect(session).Should(Say("SEE ALSO:\n"))
				Expect(session).Should(Say("api, auth, target"))
			})
		})
	})

	Describe("API Endpoint", func() {
		var (
			username string
			password string
		)

		When("the API endpoint is not set", func() {
			BeforeEach(func() {
				helpers.UnsetAPI()
				username, password = helpers.GetCredentials()
			})

			When("the user does not provide the -a flag", func() {
				It("prompts the user for an endpoint", func() {
					session := helpers.CF("login")
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

				When("the provided API endpoint is reachable", func() {

					var session *Session

					BeforeEach(func() {
						session = helpers.CF("login", "-a", apiURL, "-u", username, "-p", password, "--skip-ssl-validation")
						Eventually(session).Should(Say("API endpoint: %s", apiURL))
						Eventually(session).Should(Exit(0))
					})
					It("writes fields to the config file when targeting an API", func() {
						rawConfig, err := ioutil.ReadFile(filepath.Join(homeDir, ".cf", "config.json"))
						Expect(err).NotTo(HaveOccurred())

						var configFile configv3.JSONConfig
						err = json.Unmarshal(rawConfig, &configFile)
						Expect(err).NotTo(HaveOccurred())

						Expect(configFile.ConfigVersion).To(Equal(configv3.CurrentConfigVersion))
						Expect(configFile.Target).To(Equal(apiURL))
						Expect(configFile.APIVersion).To(MatchRegexp(`\d+\.\d+\.\d+`))
						Expect(configFile.AuthorizationEndpoint).ToNot(BeEmpty())
						Expect(configFile.DopplerEndpoint).To(MatchRegexp("^wss://"))
						Expect(configFile.LogCacheEndpoint).To(MatchRegexp(".*log-cache.*"))

					})

					It("sets the API endpoint to the provided value", func() {
						session = helpers.CF("api")
						Eventually(session).Should(Exit(0))
						Expect(session).Should(Say("API endpoint:   %s", apiURL))
					})
				})

				When("the provided API endpoint is unreachable", func() {
					It("displays an error and fails", func() {
						session := helpers.CF("login", "-a", "does.not.exist", "--skip-ssl-validation")

						Eventually(session).Should(Say("API endpoint: does.not.exist"))
						Eventually(session).Should(Say("FAILED"))
						Eventually(session.Err).Should(Say("Request error: "))
						Eventually(session.Err).Should(Say("TIP: If you are behind a firewall and require an HTTP proxy, verify the https_proxy environment variable is correctly set. Else, check your network connection."))
						Eventually(session).Should(Exit(1))
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
				It("attempts the new API endpoint, and does not update the API endpoint unless it succeeds", func() {
					session := helpers.CF("login", "-a", "does.not.exist", "--skip-ssl-validation")
					Eventually(session).Should(Say("API endpoint: does.not.exist"))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))

					apiSession := helpers.CF("api")
					Eventually(apiSession).Should(Exit(0))
					Eventually(apiSession).Should(Say("API endpoint:   %s", apiURL))
				})
			})
		})
	})

	Describe("API URL scheme", func() {
		When("no scheme is included in the API endpoint", func() {
			var (
				hostname  string
				serverURL *url.URL
				err       error
				session   *Session
			)

			BeforeEach(func() {
				serverURL, err = url.Parse(helpers.GetAPI())
				Expect(err).NotTo(HaveOccurred())

				hostname = serverURL.Hostname()
				username, password := helpers.GetCredentials()

				session = helpers.CF("login", "-u", username, "-p", password, "-a", hostname, "--skip-ssl-validation")
			})

			It("displays the API endpoint as https once targeted", func() {
				Eventually(session).Should(Say("API endpoint: %s", hostname))
				Eventually(session).Should(Say(`API version:\s+\d\.\d{1,3}\.\d{1,3}`))
				Eventually(session).Should(Exit(0))
			})

		})

		When("the API endpoint's scheme is http", func() {
			var httpURL string

			BeforeEach(func() {
				apiURL, err := url.Parse(helpers.GetAPI())
				Expect(err).NotTo(HaveOccurred())
				apiURL.Scheme = "http"

				httpURL = apiURL.String()
			})

			It("shows a warning to the user and logs in successfully", func() {
				username, password := helpers.GetCredentials()
				session := helpers.CF("login", "-u", username, "-p", password, "-a", httpURL, "--skip-ssl-validation")

				Eventually(session).Should(Say("API endpoint: %s", httpURL))
				Eventually(session.Err).Should(Say("Warning: Insecure http API endpoint detected: secure https API endpoints are recommended"))
				Eventually(session).Should(Exit(0))
			})
		})

		When("the API endpoint's scheme is https", func() {
			// This test is somewhat redundant because the integration test setup will have already logged in successfully with certificates at this point
			// In the interest of test coverage however, we have decided to keep it in.
			When("the OS provides a valid SSL Certificate (Unix: SSL_CERT_FILE or SSL_CERT_DIR Environment variables) (Windows: Import-Certificate call)", func() {
				BeforeEach(func() {
					if skipSSLValidation {
						Skip("SKIP_SSL_VALIDATION is enabled")
					}
				})

				It("trusts the cert and allows the users to log in", func() {
					username, password := helpers.GetCredentials()
					session := helpers.CF("login", "-u", username, "-p", password, "-a", helpers.GetAPI())
					Eventually(session).Should(Say("API endpoint: %s", apiURL))
					Eventually(session).Should(Exit())

					session = helpers.CF("api")
					Eventually(session).Should(Exit(0))
					Expect(session).Should(Say("API endpoint:   %s", apiURL))
				})
			})

			When("the SSL Certificate is invalid", func() {
				var server *ghttp.Server

				BeforeEach(func() {
					cliVersion := "1.0.0"
					server = helpers.StartMockServerWithMinimumCLIVersion(cliVersion)
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

				It("errors when --skip-ssl-validation is not provided", func() {
					session := helpers.CF("login", "-a", server.URL())
					Eventually(session).Should(Say("API endpoint: %s", server.URL()))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("Invalid SSL Cert for %s", server.URL()))
					Eventually(session.Err).Should(Say("TIP: Use 'cf login --skip-ssl-validation' to continue with an insecure API endpoint"))
					Eventually(session).Should(Exit(1))
				})

				It("doesn't complain about an invalid cert when we specify --skip-ssl-validation", func() {
					session := helpers.CF("login", "-a", server.URL(), "--skip-ssl-validation")
					Eventually(session).Should(Exit(0))

					Expect(string(session.Err.Contents())).Should(Not(ContainSubstring("Invalid SSL Cert for %s", server.URL())))
				})

				When("targeted with --skip-ssl-validation", func() {
					BeforeEach(func() {
						Eventually(helpers.CF("api", server.URL(), "--skip-ssl-validation")).Should(Exit(0))
					})

					When("logging in without --skip-ssl-validation", func() {
						It("displays a helpful error message and exits 1", func() {
							session := helpers.CF("login", "-a", server.URL())
							Eventually(session).Should(Say("API endpoint: %s", server.URL()))
							Eventually(session).Should(Say("FAILED"))
							Eventually(session.Err).Should(Say("Invalid SSL Cert for %s", server.URL()))
							Eventually(session.Err).Should(Say("TIP: Use 'cf login --skip-ssl-validation' to continue with an insecure API endpoint"))
							Eventually(session).Should(Exit(1))
						})
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
					Eventually(session).Should(Say(`API endpoint:\s+` + helpers.GetAPI()))
					Eventually(session).Should(Say(`API version:\s+\d\.\d{1,3}\.\d{1,3}`))
					Eventually(session).Should(Say(`Not logged in. Use 'cf login' or 'cf login --sso' to log in.`))
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
	})

	Describe("Choosing the identity provider", func() {
		When("the user provides the --origin flag", func() {
			It("logs in successfully", func() {
				username, password := helpers.GetCredentials()
				session := helpers.CF("login", "-u", username, "-p", password, "--origin", "uaa")
				Eventually(session).Should(Say("API endpoint:\\s+" + helpers.GetAPI()))
				Eventually(session).Should(Say(`Authenticating\.\.\.`))
				Eventually(session).Should(Say(`OK`))
				Eventually(session).Should(Say(`API endpoint:\s+` + helpers.GetAPI()))
				Eventually(session).Should(Say(`API version:\s+\d\.\d{1,3}\.\d{1,3}`))
				Eventually(session).Should(Say("user:\\s+" + username))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Describe("User Credentials", func() {
		var (
			username string
			password string
		)

		BeforeEach(func() {
			username, password = helpers.GetCredentials()
		})

		It("prompts the user for email and password", func() {
			buffer := NewBuffer()
			_, err := buffer.Write([]byte(fmt.Sprintf("%s\n%s\n", username, password)))
			Expect(err).ToNot(HaveOccurred())
			session := helpers.CFWithStdin(buffer, "login")
			Eventually(session).Should(Say("Email:"))
			Eventually(session).Should(Say("\n"))
			Eventually(session).Should(Say("Password:"))
			Eventually(session).Should(Say("\n\n"))
			Eventually(session).Should(Exit(0))
		})

		When("the user's account has been locked due to too many failed attempts", func() {
			BeforeEach(func() {
				helpers.LoginCF()
				username, _ = helpers.CreateUser()
				helpers.LogoutCF()
			})

			It("displays a helpful error and does not reprompt", func() {
				input := NewBuffer()
				_, err := input.Write([]byte("garbage\ngarbage\ngarbage\n"))
				Expect(err).ToNot(HaveOccurred())
				session := helpers.CFWithStdin(input, "login", "-u", username)
				Eventually(session).Should(Exit(1))

				input = NewBuffer()
				_, err = input.Write([]byte("garbage\ngarbage\ngarbage\n"))
				Expect(err).ToNot(HaveOccurred())
				session = helpers.CFWithStdin(input, "login", "-u", username)
				Eventually(session).Should(Exit(1))

				input = NewBuffer()
				_, err = input.Write([]byte("garbage\ngarbage\ngarbage\n"))
				Expect(err).NotTo(HaveOccurred())
				session = helpers.CFWithStdin(input, "login", "-u", username)
				Eventually(session).Should(Say(`Password`))
				Eventually(session.Err).Should(Say(`Your account has been locked because of too many failed attempts to login\.`))
				Consistently(session).ShouldNot(Say(`Password`))
				Eventually(session.Err).Should(Say(`Unable to authenticate.`))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the -u flag is provided", func() {
			It("prompts the user for their password", func() {
				buffer := NewBuffer()
				_, err := buffer.Write([]byte(fmt.Sprintf("%s\n", password)))
				Expect(err).ToNot(HaveOccurred())
				session := helpers.CFWithStdin(buffer, "login", "-u", username)
				Eventually(session).Should(Say("Password: "))
				Eventually(session).Should(Exit(0))
			})
		})

		When("the user provides the -p flag", func() {
			It("prompts the user for their email and logs in successfully", func() {
				input := NewBuffer()
				_, err := input.Write([]byte(username + "\n"))
				Expect(err).ToNot(HaveOccurred())
				session := helpers.CFWithStdin(input, "login", "-p", password)
				Eventually(session).Should(Say("Email: "))
				Eventually(session).Should(Exit(0))
			})

			When("the password flag is given incorrectly", func() {
				It("Prompts the user two more times before exiting with an error", func() {
					input := NewBuffer()
					_, err := input.Write([]byte(username + "\n" + "bad-password\n" + "bad-password\n"))
					Expect(err).ToNot(HaveOccurred())
					session := helpers.CFWithStdin(input, "login", "-p", "bad-password")
					Eventually(session).Should(Say("Email: "))
					Eventually(session.Err).Should(Say("Credentials were rejected, please try again."))
					Eventually(session).Should(Say("Password: "))
					Eventually(session.Err).Should(Say("Credentials were rejected, please try again."))
					Eventually(session).Should(Say("Not logged in. Use 'cf login' or 'cf login --sso' to log in."))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("Unable to authenticate."))
					Eventually(session).Should(Exit(1))
				})
			})
		})

		When("the user provides the -p and -u flags", func() {
			Context("and the credentials are correct", func() {
				It("logs in successfully", func() {
					session := helpers.CF("login", "-p", password, "-u", username)
					Eventually(session).Should(Say("API endpoint:\\s+" + helpers.GetAPI()))
					Eventually(session).Should(Say(`Authenticating\.\.\.`))
					Eventually(session).Should(Say(`OK`))
					Eventually(session).Should(Say(`API endpoint:\s+` + helpers.GetAPI()))
					Eventually(session).Should(Say(`API version:\s+\d\.\d{1,3}\.\d{1,3}`))
					Eventually(session).Should(Say("user:\\s+" + username))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("and the credentials are incorrect", func() {
				It("prompts twice, displays an error and fails", func() {
					input := NewBuffer()
					_, err := input.Write([]byte("garbage\ngarbage\n"))
					Expect(err).ToNot(HaveOccurred())
					session := helpers.CFWithStdin(input, "login", "-p", "nope", "-u", "faker")
					Eventually(session).Should(Say("API endpoint:\\s+" + helpers.GetAPI()))
					Eventually(session).Should(Say(`Authenticating\.\.\.`))
					Eventually(session.Err).Should(Say(`Credentials were rejected, please try again.`))
					Eventually(session).Should(Say(`Password:`))
					Eventually(session).Should(Say(`Authenticating\.\.\.`))
					Eventually(session.Err).Should(Say(`Credentials were rejected, please try again.`))
					Eventually(session).Should(Say(`Password:`))
					Eventually(session).Should(Say(`Authenticating\.\.\.`))
					Eventually(session.Err).Should(Say(`Credentials were rejected, please try again.`))
					Eventually(session).Should(Say(`API endpoint:\s+` + helpers.GetAPI()))
					Eventually(session).Should(Say(`API version:\s+\d\.\d{1,3}\.\d{1,3}`))
					Eventually(session).Should(Say(`Not logged in. Use 'cf login' or 'cf login --sso' to log in.`))
					Eventually(session.Err).Should(Say(`Unable to authenticate.`))
					Eventually(session).Should(Say(`FAILED`))

					Eventually(session).Should(Exit(1))
				})

				Context("and the user was previously logged in", func() {
					BeforeEach(func() {
						helpers.LoginCF()
					})

					It("logs them out", func() {
						session := helpers.CF("login", "-p", "nope", "-u", "faker")
						Eventually(session).Should(Say(`Not logged in. Use 'cf login' or 'cf login --sso' to log in.`))
						Eventually(session).Should(Exit())

						orgsSession := helpers.CF("orgs")
						Eventually(orgsSession.Err).Should(Say(`Not logged in. Use 'cf login' or 'cf login --sso' to log in.`))
						Eventually(orgsSession).Should(Exit())
					})
				})
			})
		})

		When("MFA is enabled", func() {
			var (
				mfaCode string
				server  *ghttp.Server
			)

			BeforeEach(func() {
				password = "some-password"
				mfaCode = "123456"
				server = helpers.StartAndTargetMockServerWithAPIVersions(helpers.DefaultV2Version, helpers.DefaultV3Version)
				helpers.AddMfa(server, password, mfaCode)
			})

			AfterEach(func() {
				server.Close()
			})

			When("correct MFA code and credentials are provided", func() {
				BeforeEach(func() {
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

				It("logs in the user", func() {
					input := NewBuffer()
					_, err := input.Write([]byte(username + "\n" + password + "\n" + mfaCode + "\n"))
					Expect(err).ToNot(HaveOccurred())
					session := helpers.CFWithStdin(input, "login")
					Eventually(session).Should(Say("Email: "))
					Eventually(session).Should(Say("\n"))
					Eventually(session).Should(Say("Password:"))
					Eventually(session).Should(Say("\n"))
					Eventually(session).Should(Say("MFA Code \\( Register at %[1]s \\)", server.URL()))
					Eventually(session).Should(Exit(0))
				})
			})

			When("incorrect MFA code and credentials are provided", func() {
				It("fails", func() {
					input := NewBuffer()
					wrongMfaCode := mfaCode + "foo"
					_, err := input.Write([]byte(username + "\n" + password + "\n" + wrongMfaCode + "\n" + password + "\n" + wrongMfaCode + "\n"))
					Expect(err).ToNot(HaveOccurred())
					session := helpers.CFWithStdin(input, "login")
					Eventually(session).Should(Say("Password: "))
					Eventually(session).Should(Say("MFA Code \\( Register at %[1]s \\)", server.URL()))
					Eventually(session).Should(Say("Password: "))
					Eventually(session).Should(Say("MFA Code \\( Register at %[1]s \\)", server.URL()))
					Eventually(session).Should(Say("Not logged in. Use 'cf login' or 'cf login --sso' to log in."))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("Unable to authenticate."))

					Eventually(session).Should(Exit(1))
				})
			})
		})
	})

	Describe("Client Credentials", func() {
		var (
			username     string
			password     string
			clientID     string
			clientSecret string
		)

		When("the user has manually added client credentials to the config file", func() {
			BeforeEach(func() {
				clientID, clientSecret = helpers.SkipIfCustomClientCredentialsNotSet()
				username, password = helpers.GetCredentials()
				helpers.SetConfig(func(config *configv3.Config) {
					config.ConfigFile.UAAOAuthClient = clientID
					config.ConfigFile.UAAOAuthClientSecret = clientSecret
					config.ConfigFile.UAAGrantType = ""
				})
			})

			It("returns an unsupported error", func() {
				errorMessage := "Error: Support for manually writing your client credentials to config.json has been removed. For similar functionality please use `cf auth --client-credentials`."

				session := helpers.CF("login", "-u", username, "-p", password)
				Eventually(session.Err).Should(Say(errorMessage))
				Eventually(session).Should(Exit(1))
			})
		})

		When("already logged in with client credentials", func() {
			BeforeEach(func() {
				clientID, clientSecret = helpers.SkipIfClientCredentialsNotSet()
				username, password = helpers.GetCredentials()

				session := helpers.CF("auth", clientID, clientSecret, "--client-credentials")
				Eventually(session).Should(Exit(0))
			})

			It("should fail log in and display an error informing the user they need to log out", func() {
				session := helpers.CF("login", "-p", password, "-u", username)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("Service account currently logged in. Use 'cf logout' to log out service account and try again."))
				Eventually(session).Should(Exit(1))

				// And I am still logged in
				targetSession := helpers.CF("target")
				Eventually(targetSession).Should(Exit(0))
			})
		})
	})

	Describe("Target Organization", func() {
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

		When("there is only one org available to the user", func() {
			It("logs the user in and targets the organization automatically", func() {
				session := helpers.CF("login", "-u", username, "-p", password, "-a", apiURL, "--skip-ssl-validation")
				Eventually(session).Should(Exit(0))

				targetSession := helpers.CF("target")
				Eventually(targetSession).Should(Exit(0))
				Eventually(targetSession).Should(Say(`org:\s+%s`, orgName))
			})
		})

		When("the -o flag is not passed", func() {
			When("there are multiple orgs available to the user", func() {
				BeforeEach(func() {
					orgName = helpers.NewOrgName()
					createOrgSession := helpers.CF("create-org", orgName)
					Eventually(createOrgSession).Should(Exit(0))
					setOrgRoleSession := helpers.CF("set-org-role", username, orgName, "OrgManager")
					Eventually(setOrgRoleSession).Should(Exit(0))
				})

				When("there are more than 50 orgs", func() {
					var server *ghttp.Server

					BeforeEach(func() {
						server = helpers.StartAndTargetMockServerWithAPIVersions(helpers.DefaultV2Version, helpers.DefaultV3Version)
						helpers.AddLoginRoutes(server)
						helpers.AddFiftyOneOrgs(server)
						// handle request for spaces under "org20"
						helpers.AddEmptyPaginatedResponse(server, "/v3/spaces?order_by=name&organization_guids=f6653aac-938e-4469-9a66-56a02796412b")
					})

					AfterEach(func() {
						server.Close()
					})

					It("displays a message and prompts the user for the org name", func() {
						input := NewBuffer()
						_, wErr := input.Write([]byte(fmt.Sprintf("%s\n", "org20"))) // "org20" is one of the orgs in the test fixture
						Expect(wErr).ToNot(HaveOccurred())

						session := helpers.CFWithStdin(input, "login", "-u", username, "-p", password, "--skip-ssl-validation")

						Eventually(session).Should(Say("Select an org:"))
						Eventually(session).Should(Say("There are too many options to display; please type in the name."))
						Eventually(session).Should(Say("\n\n"))
						Eventually(session).Should(Say(regexp.QuoteMeta(`Org (enter to skip):`)))
						Eventually(session).Should(Say(`Targeted org org20\.`))

						Eventually(session).Should(Exit(0))
					})
				})

				When("user selects an organization by using numbered list", func() {
					It("prompts the user for org and targets the selected org", func() {
						input := NewBuffer()
						_, err := input.Write([]byte("1\n"))
						Expect(err).ToNot(HaveOccurred())

						session := helpers.CFWithStdin(input, "login", "-u", username, "-p", password, "-a", apiURL, "--skip-ssl-validation")
						Eventually(session).Should(Exit(0))

						re := regexp.MustCompile("1\\. (?P<OrgName>.*)\n")
						matches := re.FindStringSubmatch(string(session.Out.Contents()))
						Expect(matches).To(HaveLen((2)))
						expectedOrgName := matches[1]

						targetSession := helpers.CF("target")
						Eventually(targetSession).Should(Exit(0))
						Eventually(targetSession).Should(Say(`org:\s+%s`, expectedOrgName))
					})

					When("the user selects a number greater than the number of orgs", func() {
						It("prompts the user until a valid number is entered", func() {
							input := NewBuffer()
							_, err := input.Write([]byte("3\n"))
							Expect(err).ToNot(HaveOccurred())

							session := helpers.CFWithStdin(input, "login", "-u", username, "-p", password)

							Eventually(session).Should(Say(regexp.QuoteMeta("Select an org:")))
							Eventually(session).Should(Say(regexp.QuoteMeta(`Org (enter to skip):`)))
							Eventually(session).Should(Say(regexp.QuoteMeta(`Org (enter to skip):`)))

							session.Interrupt()
							Eventually(session).Should(Exit())
						})
					})
				})

				When("user selects an organization by org name", func() {
					It("prompts the user for an org and then targets the selected org", func() {
						input := NewBuffer()
						_, err := input.Write([]byte(fmt.Sprintf("%s\n", orgName)))
						Expect(err).ToNot(HaveOccurred())

						session := helpers.CFWithStdin(input, "login", "-u", username, "-p", password, "-a", apiURL, "--skip-ssl-validation")
						Eventually(session).Should(Say(`\d\. %s`, orgName))
						Eventually(session).Should(Exit(0))

						targetSession := helpers.CF("target")
						Eventually(targetSession).Should(Exit(0))
						Eventually(targetSession).Should(Say(`org:\s+%s`, orgName))
					})
				})

				When("user does not select an organization", func() {
					It("successfully logs in but does not target any org", func() {
						input := NewBuffer()
						_, err := input.Write([]byte("\n"))
						Expect(err).ToNot(HaveOccurred())

						session := helpers.CFWithStdin(input, "login", "-u", username, "-p", password, "-a", apiURL, "--skip-ssl-validation")
						Eventually(session).Should(Say(`Org \(enter to skip\):`))
						Consistently(session).ShouldNot(Say(`Org \(enter to skip\):`))
						Eventually(session).Should(Exit(0))

						targetSession := helpers.CF("target")
						Eventually(targetSession).Should(Exit(0))
						Eventually(targetSession).Should(Say("No org or space targeted, use 'cf target -o ORG -s SPACE'"))
					})
				})

				When("the user enters an invalid organization at the prompt", func() {
					It("displays an error message and does not target the org", func() {
						orgName = "invalid-org"
						input := NewBuffer()
						_, err := input.Write([]byte(fmt.Sprintf("%s\n", orgName)))
						Expect(err).ToNot(HaveOccurred())

						session := helpers.CFWithStdin(input, "login", "-u", username, "-p", password, "--skip-ssl-validation")
						Eventually(session).Should(Exit(1))
						Eventually(session).Should(Say("FAILED"))
						Eventually(session.Err).Should(Say("Organization '%s' not found", orgName))

						targetSession := helpers.CF("target")
						Eventually(targetSession).Should(Exit(0))
						Eventually(targetSession).Should(Say(`user:\s+%s`, username))
						Eventually(targetSession).ShouldNot(Say(`org:\s+%s`, orgName))
						Eventually(targetSession).Should(Say("No org or space targeted, use 'cf target -o ORG -s SPACE'"))
					})
				})
			})
		})

		When("the -o flag is not passed, and the -s flag is passed", func() {
			var (
				spaceName     string
				secondOrgName string
				thirdOrgName  string
			)

			BeforeEach(func() {
				orgName = helpers.NewOrgName()
				createOrgSession := helpers.CF("create-org", orgName)
				Eventually(createOrgSession).Should(Exit(0))
				setOrgRoleSession := helpers.CF("set-org-role", username, orgName, "OrgManager")
				Eventually(setOrgRoleSession).Should(Exit(0))
				spaceName = helpers.NewSpaceName()
				createSpaceSession := helpers.CF("create-space", "-o", orgName, spaceName)
				Eventually(createSpaceSession).Should(Exit(0))
				roleSession := helpers.CF("set-space-role", username, orgName, spaceName, "SpaceManager")
				Eventually(roleSession).Should(Exit(0))

				secondOrgName = helpers.NewOrgName()
				createOrgSession = helpers.CF("create-org", secondOrgName)
				Eventually(createOrgSession).Should(Exit(0))
				setOrgRoleSession = helpers.CF("set-org-role", username, secondOrgName, "OrgManager")
				Eventually(setOrgRoleSession).Should(Exit(0))
				secondSpaceName := helpers.NewSpaceName()
				createSpaceSession = helpers.CF("create-space", "-o", secondOrgName, secondSpaceName)
				Eventually(createSpaceSession).Should(Exit(0))
				roleSession = helpers.CF("set-space-role", username, secondOrgName, secondSpaceName, "SpaceManager")
				Eventually(roleSession).Should(Exit(0))
			})

			Context("only one org has a correctly named space", func() {
				It("automatically targets the org associated with the space", func() {
					session := helpers.CF("login", "-u", username, "-p", password, "-s", spaceName, "-a", apiURL, "--skip-ssl-validation")
					Eventually(session).Should(Exit(0))

					targetSession := helpers.CF("target")
					Eventually(targetSession).Should(Exit(0))
					Eventually(targetSession).Should(Say(`org:\s+%s`, orgName))
					Eventually(targetSession).Should(Say(`space:\s+%s`, spaceName))
				})
			})

			Context("multiple orgs have the correctly named space", func() {
				BeforeEach(func() {
					thirdOrgName = helpers.NewOrgName()
					createOrgSession := helpers.CF("create-org", thirdOrgName)
					Eventually(createOrgSession).Should(Exit(0))
					setOrgRoleSession := helpers.CF("set-org-role", username, thirdOrgName, "OrgManager")
					Eventually(setOrgRoleSession).Should(Exit(0))
					createSpaceSession := helpers.CF("create-space", "-o", thirdOrgName, spaceName)
					Eventually(createSpaceSession).Should(Exit(0))
					roleSession := helpers.CF("set-space-role", username, thirdOrgName, spaceName, "SpaceManager")
					Eventually(roleSession).Should(Exit(0))
				})

				It("prompts the user to select from the matching orgs", func() {
					input := NewBuffer()
					_, err := input.Write([]byte("1\n"))
					Expect(err).ToNot(HaveOccurred())

					session := helpers.CFWithStdin(input, "login", "-u", username, "-p", password, "-s", spaceName, "-a", apiURL, "--skip-ssl-validation")
					Eventually(session).Should(Exit(0))

					re := regexp.MustCompile("[0-9]+\\. (?P<OrgName>.*)")
					matches := re.FindStringSubmatch(string(session.Out.Contents()))
					Expect(matches).To(HaveLen((2)))
					expectedOrgName := matches[1]

					targetSession := helpers.CF("target")
					Eventually(targetSession).Should(Exit(0))
					Eventually(targetSession).Should(Say(`org:\s+%s`, expectedOrgName))
					Eventually(targetSession).Should(Say(`space:\s+%s`, spaceName))
				})

				When("the user enters an organization that doesn't contain the space at the prompt", func() {
					It("targets the org, then displays an error message and does not target the space", func() {
						input := NewBuffer()
						_, err := input.Write([]byte(fmt.Sprintf("%s\n", secondOrgName)))
						Expect(err).ToNot(HaveOccurred())

						session := helpers.CFWithStdin(input, "login", "-u", username, "-p", password, "-s", spaceName, "-a", apiURL, "--skip-ssl-validation")

						Eventually(session).Should(Exit(1))
						Eventually(session).Should(Say("FAILED"))
						Eventually(session.Err).Should(Say("Organization '%s' containing space '%s' not found.", secondOrgName, spaceName))

						targetSession := helpers.CF("target")
						Eventually(targetSession).Should(Exit(0))
						Eventually(targetSession).Should(Say(`user:\s+%s`, username))
						Eventually(targetSession).ShouldNot(Say(`org:\s+%s`, orgName))
						Eventually(targetSession).Should(Say("No org or space targeted, use 'cf target -o ORG -s SPACE'"))
					})
				})
			})
		})

		When("the -o flag is passed", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
			})

			When("the organization is valid", func() {
				It("targets the organization that was passed as an argument", func() {
					session := helpers.CF("login", "-u", username, "-p", password, "-o", orgName)

					Eventually(session).Should(Exit(0))
					Eventually(session).Should(Say(`org:\s+%s`, orgName))

					targetSession := helpers.CF("target")
					Eventually(targetSession).Should(Exit(0))
					Eventually(targetSession).Should(Say(`org:\s+%s`, orgName))
				})
			})

			When("the organization is invalid", func() {
				It("logs in the user, displays an error message, and does not target any organization", func() {
					orgName = "invalid-org"
					session := helpers.CF("login", "-u", username, "-p", password, "-o", orgName)

					Eventually(session).Should(Exit(1))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("Organization '%s' not found", orgName))

					targetSession := helpers.CF("target")
					Eventually(targetSession).Should(Exit(0))
					Eventually(targetSession).Should(Say(`user:\s+%s`, username))
					Eventually(targetSession).ShouldNot(Say(`org:\s+%s`, orgName))
					Eventually(targetSession).Should(Say("No org or space targeted, use 'cf target -o ORG -s SPACE'"))
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
			username, password = helpers.CreateUser()
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

		When("multiple spaces are available to the user", func() {
			var (
				spaceName  string
				spaceName2 string
			)

			BeforeEach(func() {
				spaceName = helpers.NewSpaceName()
				session := helpers.CF("create-space", "-o", orgName, spaceName)
				Eventually(session).Should(Exit(0))
				roleSession := helpers.CF("set-space-role", username, orgName, spaceName, "SpaceManager")
				Eventually(roleSession).Should(Exit(0))

				spaceName2 = helpers.NewSpaceName()
				session2 := helpers.CF("create-space", "-o", orgName, spaceName2)
				Eventually(session2).Should(Exit(0))
				roleSession2 := helpers.CF("set-space-role", username, orgName, spaceName2, "SpaceManager")
				Eventually(roleSession2).Should(Exit(0))
			})

			When("the -s flag is passed", func() {
				BeforeEach(func() {
					orgName2 := helpers.NewOrgName()
					session := helpers.CF("create-org", orgName2)
					Eventually(session).Should(Exit(0))
					session = helpers.CF("set-org-role", username, orgName2, "OrgManager")
					Eventually(session).Should(Exit(0))
				})

				It("targets the org and the space", func() {
					stdin := NewBuffer()
					_, writeErr := stdin.Write([]byte(orgName + "\n"))
					Expect(writeErr).ToNot(HaveOccurred())
					session := helpers.CFWithStdin(stdin, "login", "-u", username, "-p", password, "-a", apiURL, "-s", spaceName, "--skip-ssl-validation")

					Eventually(session).Should(Say(`Targeted org\s+%s\.`, orgName))
					Eventually(session).Should(Say(`\nTargeted space\s+%s\.`, spaceName))

					Eventually(session).Should(Say(`org:\s+%s`, orgName))
					Eventually(session).Should(Say(`space:\s+%s`, spaceName))
					Eventually(session).Should(Exit(0))

					sessionOutput := string(session.Out.Contents())
					Expect(sessionOutput).To(MatchRegexp(`\S\nAPI`))

					targetSession := helpers.CF("target")
					Eventually(targetSession).Should(Exit(0))
					Eventually(targetSession).Should(Say(`org:\s+%s`, orgName))
					Eventually(targetSession).Should(Say(`space:\s+%s`, spaceName))
				})

				When("the space name is invalid", func() {
					BeforeEach(func() {
						spaceName = "invalid-space-name"
					})

					It("the command fails and displays an error message. It targets the org but not the space.", func() {
						stdin := NewBuffer()
						_, writeErr := stdin.Write([]byte(orgName + "\n"))
						Expect(writeErr).ToNot(HaveOccurred())
						session := helpers.CFWithStdin(stdin, "login", "-u", username, "-p", password, "-a", apiURL, "-o", orgName, "-s", spaceName, "--skip-ssl-validation")
						Eventually(session).Should(Exit(1))
						Eventually(session).Should(Say("FAILED"))
						Eventually(session.Err).Should(Say("Space '%s' not found", spaceName))

						targetSession := helpers.CF("target")
						Eventually(targetSession).Should(Exit(0))
						Eventually(targetSession).Should(Say(`org:\s+%s`, orgName))
						Eventually(targetSession).Should(Say("No space targeted, use 'cf target -s SPACE'"))
					})
				})
			})

			When("the -s flag is not passed", func() {
				It("prompts the user to pick their space by list position", func() {
					input := NewBuffer()
					_, err := input.Write([]byte("1\n"))
					Expect(err).ToNot(HaveOccurred())

					session := helpers.CFWithStdin(input, "login", "-u", username, "-p", password, "-a", apiURL, "--skip-ssl-validation")
					Eventually(session).Should(Exit(0))

					re := regexp.MustCompile("1\\. (?P<SpaceName>.*)\n")
					submatches := re.FindStringSubmatch(string(session.Out.Contents()))
					Expect(submatches).ToNot(BeEmpty(), "missing numbered space list")
					expectedSpaceName := submatches[1]

					targetSession := helpers.CF("target")
					Eventually(targetSession).Should(Exit(0))
					Eventually(targetSession).Should(Say(`space:\s+%s`, expectedSpaceName))
				})

				It("reprompts the user if an invalid number is entered", func() {
					input := NewBuffer()
					_, err := input.Write([]byte("4\n"))
					Expect(err).ToNot(HaveOccurred())

					session := helpers.CFWithStdin(input, "login", "-u", username, "-p", password, "-a", apiURL, "--skip-ssl-validation")
					Eventually(session).Should(Say(regexp.QuoteMeta("Space (enter to skip):")))
					Eventually(session).Should(Say(regexp.QuoteMeta("Space (enter to skip):")))
					session.Interrupt()
					Eventually(session).Should(Exit())
				})

				It("allows the user to pick their space by name", func() {
					input := NewBuffer()
					_, err := input.Write([]byte(spaceName + "\n"))
					Expect(err).ToNot(HaveOccurred())

					session := helpers.CFWithStdin(input, "login", "-u", username, "-p", password, "-a", apiURL, "--skip-ssl-validation")
					Eventually(session).Should(Exit(0))

					targetSession := helpers.CF("target")
					Eventually(targetSession).Should(Exit(0))
					Eventually(targetSession).Should(Say(`space:\s+%s`, spaceName))
				})

				It("allows the user to skip picking a space", func() {
					input := NewBuffer()
					_, err := input.Write([]byte("\n"))
					Expect(err).ToNot(HaveOccurred())

					session := helpers.CFWithStdin(input, "login", "-u", username, "-p", password, "-a", apiURL, "--skip-ssl-validation")
					Eventually(session).Should(Exit(0))

					targetSession := helpers.CF("target")
					Eventually(targetSession).Should(Exit(0))
					Eventually(targetSession).Should(Say(`No space targeted, use 'cf target -s SPACE'`))
				})

				When("the input space name is invalid", func() {
					BeforeEach(func() {
						spaceName = "invalid-space-name"
					})

					It("the command fails and displays an error message. It does not target the space.", func() {
						input := NewBuffer()
						_, err := input.Write([]byte(spaceName + "\n"))
						Expect(err).ToNot(HaveOccurred())

						session := helpers.CFWithStdin(input, "login", "-u", username, "-p", password, "-a", apiURL, "--skip-ssl-validation")
						Eventually(session).Should(Exit(1))
						Eventually(session).Should(Say("FAILED"))
						Eventually(session.Err).Should(Say("Space '%s' not found", spaceName))

						targetSession := helpers.CF("target")
						Eventually(targetSession).Should(Exit(0))
						Eventually(targetSession).Should(Say(`org:\s+%s`, orgName))
						Eventually(targetSession).Should(Say("No space targeted, use 'cf target -s SPACE'"))
					})
				})

				When("there are more than 50 spaces", func() {
					var server *ghttp.Server
					BeforeEach(func() {
						server = helpers.StartAndTargetMockServerWithAPIVersions(helpers.DefaultV2Version, helpers.DefaultV3Version)
						helpers.AddLoginRoutes(server)
						helpers.AddFiftyOneSpaces(server)
					})

					AfterEach(func() {
						server.Close()
					})

					It("displays a message and prompts the user for the space name", func() {
						input := NewBuffer()
						_, wErr := input.Write([]byte(fmt.Sprintf("%s\n", "test-space-1")))
						Expect(wErr).ToNot(HaveOccurred())

						session := helpers.CFWithStdin(input, "login", "-u", username, "-p", password, "--skip-ssl-validation")

						Eventually(session).Should(Say("Select a space:"))
						Eventually(session).Should(Say("There are too many options to display; please type in the name."))
						Eventually(session).Should(Say("\n\n"))
						Eventually(session).Should(Say(regexp.QuoteMeta(`Space (enter to skip):`)))
						Eventually(session).Should(Say(`Targeted space test-space-1\.`))

						Eventually(session).Should(Exit(0))
					})
				})
			})
		})
	})

	Describe("Full interactive happy path", func() {
		var (
			orgName1   string
			orgName2   string
			spaceName1 string
			spaceName2 string
			username   string
			password   string
		)

		BeforeEach(func() {
			helpers.LoginCF()
			orgName1 = helpers.NewOrgName()
			orgName2 = helpers.NewOrgName()
			spaceName1 = helpers.NewSpaceName()
			spaceName2 = helpers.NewSpaceName()

			Eventually(helpers.CF("create-org", orgName1)).Should(Exit(0))
			Eventually(helpers.CF("create-org", orgName2)).Should(Exit(0))
			Eventually(helpers.CF("create-space", "-o", orgName1, spaceName1)).Should(Exit(0))
			Eventually(helpers.CF("create-space", "-o", orgName1, spaceName2)).Should(Exit(0))

			username, password = helpers.CreateUser()
			Eventually(helpers.CF("set-org-role", username, orgName1, "OrgManager")).Should(Exit(0))
			Eventually(helpers.CF("set-org-role", username, orgName2, "OrgManager")).Should(Exit(0))
			Eventually(helpers.CF("set-space-role", username, orgName1, spaceName1, "SpaceManager")).Should(Exit(0))
			Eventually(helpers.CF("set-space-role", username, orgName1, spaceName2, "SpaceManager")).Should(Exit(0))

			helpers.UnsetAPI()
		})

		When("there are multiple orgs and spaces available to a user", func() {
			It("prompts for username, password, org, and space. Then logs in and targets correctly", func() {
				buffer := NewBuffer()
				_, err := buffer.Write([]byte(fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n", apiURL, username, password, orgName1, spaceName2)))
				Expect(err).ToNot(HaveOccurred())

				session := helpers.CFWithStdin(buffer, "login", "--skip-ssl-validation")
				Eventually(session).Should(Say("API endpoint:"))
				Eventually(session).Should(Say("\n"))
				Eventually(session).Should(Say("Email:"))
				Eventually(session).Should(Say("\n"))
				Eventually(session).Should(Say("Password:"))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Say("\n\n"))
				Eventually(session).Should(Say("Select an org:"))
				Eventually(session).Should(Say("\n\n"))
				Eventually(session).Should(Say(regexp.QuoteMeta(`Org (enter to skip):`)))
				Eventually(session).Should(Say(fmt.Sprintf("Targeted org %s", orgName1)))
				Eventually(session).Should(Say("\n\n"))
				Eventually(session).Should(Say("Select a space:"))
				Eventually(session).Should(Say("\n\n"))
				Eventually(session).Should(Say(regexp.QuoteMeta(`Space (enter to skip):`)))
				Eventually(session).Should(Say(fmt.Sprintf(`Targeted space %s\.`, spaceName2)))
				Eventually(session).Should(Say("\n\n"))
				Eventually(session).Should(Say(`org:\s+%s`, orgName1))
				Eventually(session).Should(Say(`space:\s+%s`, spaceName2))
				Eventually(session).Should(Exit(0))

				targetSession := helpers.CF("target")
				Eventually(targetSession).Should(Exit(0))
				Eventually(targetSession).Should(Say(`org:\s+%s`, orgName1))
				Eventually(targetSession).Should(Say(`space:\s+%s`, spaceName2))
			})
		})
	})
})
