package isolated

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"runtime"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("login command", func() {
	Describe("Help Text", func() {
		When("--help flag is set", func() {
			It("displays the command usage", func() {
				session := helpers.CF("login", "--help")
				Eventually(session).Should(Exit(0))

				Expect(session).Should(Say("NAME:\n"))
				Expect(session).Should(Say("login - Log user in"))

				Expect(session).Should(Say("USAGE:\n"))
				Expect(session).Should(Say(`cf login \[-a API_URL\] \[-u USERNAME\] \[-p PASSWORD\] \[-o ORG\] \[-s SPACE\] \[--sso | --sso-passcode PASSCODE\]`))

				Expect(session).Should(Say("WARNING:\n"))
				Expect(session).Should(Say("Providing your password as a command line option is highly discouraged\n"))
				Expect(session).Should(Say("Your password may be visible to others and may be recorded in your shell history\n"))

				Expect(session).Should(Say("EXAMPLES:\n"))
				Expect(session).Should(Say(regexp.QuoteMeta("cf login (omit username and password to login interactively -- cf will prompt for both)")))
				Expect(session).Should(Say(regexp.QuoteMeta("cf login -u name@example.com -p pa55woRD (specify username and password as arguments)")))
				Expect(session).Should(Say(regexp.QuoteMeta("cf login -u name@example.com -p \"my password\" (use quotes for passwords with a space)")))
				Expect(session).Should(Say(regexp.QuoteMeta("cf login -u name@example.com -p \"\\\"password\\\"\" (escape quotes if used in password)")))
				Expect(session).Should(Say(regexp.QuoteMeta("cf login --sso (cf will provide a url to obtain a one-time passcode to login)")))

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

	Describe("Minimum Version Check", func() {
		When("the api version is less than the minimum supported version", func() {
			var server *ghttp.Server

			BeforeEach(func() {
				server = helpers.StartServerWithAPIVersions("1.0.0", "")
				server.RouteToHandler(http.MethodPost, "/oauth/token",
					ghttp.RespondWithJSONEncoded(http.StatusOK, struct{}{}))
				server.RouteToHandler(http.MethodGet, "/v2/organizations",
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
				Eventually(session).Should(Say("Your API version is no longer supported. Upgrade to a newer version of the API."))
				Eventually(session).Should(Exit(0))
			})
		})

		When("the CLI version is lower than the minimum supported version by the CC", func() {
			var server *ghttp.Server

			BeforeEach(func() {
				server = helpers.StartServerWithMinimumCLIVersion("9000.0.0")
				server.RouteToHandler(http.MethodPost, "/oauth/token",
					ghttp.RespondWithJSONEncoded(http.StatusOK, struct{}{}))
				server.RouteToHandler(http.MethodGet, "/v2/organizations",
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
				Eventually(session).Should(Say(`Cloud Foundry API version .+ requires CLI version .+\.  You are currently on version .+\. To upgrade your CLI, please visit: https://github.com/cloudfoundry/cli#downloads`))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Describe("API Endpoint", func() {
		BeforeEach(func() {
			helpers.TurnOnExperimental()
		})

		AfterEach(func() {
			helpers.TurnOffExperimental()
		})

		When("the API endpoint is not set", func() {
			BeforeEach(func() {
				helpers.UnsetAPI()
			})

			When("the user does not provide the -a flag", func() {
				It("prompts the user for an endpoint", func() {
					input := NewBuffer()
					input.Write([]byte("\n"))
					session := helpers.CFWithStdin(input, "login")
					Eventually(session).Should(Say("API endpoint:"))
					session.Interrupt()
					Eventually(session).Should(Exit())
				})

				When("the API endpoint provided at the prompt is unreachable", func() {
					It("returns an error", func() {
						input := NewBuffer()
						input.Write([]byte("does.not.exist\n"))
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
					// TODO This currently because we dont have the user/password prompt implemented. Uncomment this line after we implement the other prompts
					// Consistently(session).ShouldNot(Say("API endpoint:"))
					// session.Interrupt()
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

	Describe("SSL Validation", func() {
		When("no scheme is included in the API endpoint", func() {
			var (
				hostname  string
				serverURL *url.URL
				err       error
			)

			BeforeEach(func() {
				serverURL, err = url.Parse(helpers.GetAPI())
				Expect(err).NotTo(HaveOccurred())

				hostname = serverURL.Hostname()
			})

			It("defaults to https", func() {
				username, password := helpers.GetCredentials()
				var session *Session
				if skipSSLValidation {
					session = helpers.CF("login", "-u", username, "-p", password, "-a", hostname, "--skip-ssl-validation")
				} else {
					session = helpers.CF("login", "-u", username, "-p", password, "-a", hostname)
				}

				Eventually(session).Should(Say("API endpoint: %s", hostname))
				Eventually(session).Should(Say(`API endpoint:\s+` + helpers.GetAPI() + `\s+\(API version: \d\.\d{1,3}\.\d{1,3}\)`))
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

			It("shows a warning to the user", func() {
				username, password := helpers.GetCredentials()
				session := helpers.CF("login", "-u", username, "-p", password, "-a", httpURL, "--skip-ssl-validation")

				Eventually(session).Should(Say("API endpoint: %s", httpURL))
				Eventually(session).Should(Say("Warning: Insecure http API endpoint detected: secure https API endpoints are recommended"))
				Eventually(session).Should(Exit(0))
			})
		})

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
				Expect(session).Should(Say("api endpoint:   %s", apiURL))
			})
		})

		When("the SSL Certificate is invalid", func() {
			var (
				server    *ghttp.Server
				serverURL *url.URL
				err       error
			)

			BeforeEach(func() {
				cliVersion := "1.0.0"
				server = helpers.StartServerWithMinimumCLIVersion(cliVersion)
				serverURL, err = url.Parse(server.URL())
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				server.Close()
			})

			It("displays a helpful error message and exits 1", func() {
				session := helpers.CF("login", "-a", serverURL.String())
				Eventually(session).Should(Say("API endpoint: %s", serverURL))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Say("Invalid SSL Cert for %s:%s", serverURL.Hostname(), serverURL.Port()))
				Eventually(session).Should(Say("TIP: Use 'cf login --skip-ssl-validation' to continue with an insecure API endpoint"))
				Eventually(session).Should(Exit(1))
			})
		})
	})

	Describe("SSO", func() {
		When("--sso-passcode is provided", func() {
			Context("and --sso is also passed", func() {
				It("fails with a useful error message", func() {
					session := helpers.CF("login", "--sso-passcode", "some-passcode", "--sso")
					Eventually(session).Should(Say("Incorrect usage: --sso-passcode flag cannot be used with --sso"))
					Eventually(session).Should(Exit(1))
				})
			})

			Context("and the provided passcode is incorrect", func() {
				It("prompts twice, displays an error and fails", func() {
					session := helpers.CF("login", "--sso-passcode", "some-passcode")
					Eventually(session).Should(Say("API endpoint:\\s+" + helpers.GetAPI()))
					Eventually(session).Should(Say(`Authenticating\.\.\.`))
					Eventually(session).Should(Say(`Credentials were rejected, please try again.`))

					// Leaving out expectation of prompt text, since it comes from UAA (and doesn't show up on Windows)
					Eventually(session).Should(Say(`Authenticating\.\.\.`))
					Eventually(session).Should(Say(`Credentials were rejected, please try again.`))
					Eventually(session).Should(Say(`Authenticating\.\.\.`))
					Eventually(session).Should(Say(`Credentials were rejected, please try again.`))
					Eventually(session).Should(Say(`API endpoint:\s+` + helpers.GetAPI() + `\s+\(API version: \d\.\d{1,3}\.\d{1,3}\)`))
					Eventually(session).Should(Say(`Not logged in. Use 'cf login' to log in.`))
					Eventually(session).Should(Say(`FAILED`))
					Eventually(session).Should(Say(`Unable to authenticate`))

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
				// Following test is desired, but not current behavior.
				XIt("should fail log in and display an error informing the user they need to log out", func() {
					session := helpers.CF("login", "--sso-passcode", "my-fancy-sso")

					Eventually(session).Should(Say("API endpoint:\\s+" + helpers.GetAPI()))
					Eventually(session).Should(Say(`API endpoint:\s+` + helpers.GetAPI() + `\s+\(API version: \d\.\d{1,3}\.\d{1,3}\)`))
					// The following message is a bit strange in the output. Consider removing?
					Eventually(session).Should(Say("Not logged in. Use 'cf login' to log in."))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Say(`Service account currently logged in\. Use 'cf logout' to log out service account and try again\.`))
					Eventually(session).Should(Exit(1))
				})
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
				session := helpers.CF("login", "-u", username, "-p", password, "-a", apiURL)
				Eventually(session).Should(Exit(0))

				targetSession := helpers.CF("target")
				Eventually(targetSession).Should(Exit(0))
				Eventually(targetSession).Should(Say(`org:\s+%s`, orgName))
			})
		})

		When("there are multiple orgs available to the user", func() {
			BeforeEach(func() {
				orgName = helpers.NewOrgName()
				createOrgSession := helpers.CF("create-org", orgName)
				Eventually(createOrgSession).Should(Exit(0))
				setOrgRoleSession := helpers.CF("set-org-role", username, orgName, "OrgManager")
				Eventually(setOrgRoleSession).Should(Exit(0))
			})

			When("user selects an organization by using numbered list", func() {
				It("prompt the user for org and target the selected org", func() {
					input := NewBuffer()
					input.Write([]byte("1\n"))
					var session *Session
					if skipSSLValidation {
						session = helpers.CFWithStdin(input, "login", "-u", username, "-p", password, "-a", apiURL, "--skip-ssl-validation")
					} else {
						session = helpers.CFWithStdin(input, "login", "-u", username, "-p", password, "-a", apiURL)
					}

					Eventually(session).Should(Exit(0))

					re := regexp.MustCompile("1\\. (?P<OrgName>.*)\n")
					expectedOrgName := re.FindStringSubmatch(string(session.Out.Contents()))[1]

					targetSession := helpers.CF("target")
					Eventually(targetSession).Should(Exit(0))
					Eventually(targetSession).Should(Say(`org:\s+%s`, expectedOrgName))
				})
			})

			When("user selects an organization by org name", func() {
				It("prompt the user for org and target the selected org", func() {
					input := NewBuffer()
					input.Write([]byte(fmt.Sprintf("%s\n", orgName)))

					var session *Session
					if skipSSLValidation {
						session = helpers.CFWithStdin(input, "login", "-u", username, "-p", password, "-a", apiURL, "--skip-ssl-validation")
					} else {
						session = helpers.CFWithStdin(input, "login", "-u", username, "-p", password, "-a", apiURL)
					}

					Eventually(session).Should(Exit(0))

					targetSession := helpers.CF("target")
					Eventually(targetSession).Should(Exit(0))
					Eventually(targetSession).Should(Say(`org:\s+%s`, orgName))
				})
			})

			When("user does not select an organization", func() {
				It("succesfully logs in but does not target any org", func() {
					input := NewBuffer()
					input.Write([]byte("\n"))

					var session *Session
					if skipSSLValidation {
						session = helpers.CFWithStdin(input, "login", "-u", username, "-p", password, "-a", apiURL, "--skip-ssl-validation")
					} else {
						session = helpers.CFWithStdin(input, "login", "-u", username, "-p", password, "-a", apiURL)
					}

					Eventually(session).Should(Exit(0))

					targetSession := helpers.CF("target")
					Eventually(targetSession).Should(Exit(0))
					Eventually(targetSession).Should(Say("No org or space targeted, use 'cf target -o ORG -s SPACE'"))
				})
			})

			When("there are more than 50 orgs", func() {
				BeforeEach(func() {
					helpers.CreateOrgs(50, username)
				})

				It("displays a message and prompt the user for the org name", func() {
					input := NewBuffer()
					input.Write([]byte(fmt.Sprintf("%s\n", orgName)))

					session := helpers.CFWithStdin(input, "login", "-u", username, "-p", password)

					Eventually(session).Should(Exit(0))
					Expect(session).Should(Say("There are too many options to display, please type in the name."))

					targetSession := helpers.CF("target")
					Eventually(targetSession).Should(Exit(0))
					Eventually(targetSession).Should(Say(`org:\s+%s`, orgName))
				})
			})
		})

		When("the -o flag is passed", func() {
			It("targets the organization that was passed as an argument", func() {
				session := helpers.CF("login", "-u", username, "-p", password, "-o", orgName)

				Eventually(session).Should(Exit(0))

				targetSession := helpers.CF("target")
				Eventually(targetSession).Should(Exit(0))
				Eventually(targetSession).Should(Say(`org:\s+%s`, orgName))
			})
		})
	})

	Describe("User Credentials", func() {
		// TODO: Figure out a way to pass in password when we don't have a tty
		XIt("prompts the user for email and password", func() {
			username, password := helpers.GetCredentials()
			buffer := NewBuffer()
			buffer.Write([]byte(fmt.Sprintf("%s\n%s\n", username, password)))
			session := helpers.CFWithStdin(buffer, "login")
			Eventually(session).Should(Say("Email> "))
			Eventually(session).Should(Say("Password> "))
			Eventually(session).Should(Exit(0))
		})

		When("the user provides the -p flag", func() {
			It("prompts the user for their email and logs in successfully", func() {
				username, password := helpers.GetCredentials()
				input := NewBuffer()
				input.Write([]byte(username + "\n"))
				session := helpers.CFWithStdin(input, "login", "-p", password)
				Eventually(session).Should(Say("Email> "))
				Eventually(session).Should(Exit(0))
			})
		})

		When("the user provides the -p and -u flags", func() {
			Context("and the credentials are correct", func() {
				It("logs in successfully", func() {
					username, password := helpers.GetCredentials()
					session := helpers.CF("login", "-p", password, "-u", username)
					Eventually(session).Should(Say("API endpoint:\\s+" + helpers.GetAPI()))
					Eventually(session).Should(Say(`Authenticating\.\.\.`))
					Eventually(session).Should(Say(`OK`))
					Eventually(session).Should(Say(`API endpoint:\s+` + helpers.GetAPI() + `\s+\(API version: \d\.\d{1,3}\.\d{1,3}\)`))
					Eventually(session).Should(Say("User:\\s+" + username))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("and the credentials are incorrect", func() {
				BeforeEach(func() {
					if runtime.GOOS == "windows" {
						// TODO: Don't skip this test on windows.
						Skip("Skipping on Windows until refactor of cf login.")
					}
				})

				It("prompts twice, displays an error and fails", func() {
					username, password := helpers.GetCredentials()
					badPassword := password + "_wrong"
					session := helpers.CF("login", "-p", badPassword, "-u", username)
					Eventually(session).Should(Say("API endpoint:\\s+" + helpers.GetAPI()))
					Eventually(session).Should(Say(`Authenticating\.\.\.`))
					Eventually(session).Should(Say(`Credentials were rejected, please try again.`))
					Eventually(session).Should(Say(`Password>`))
					Eventually(session).Should(Say(`Authenticating\.\.\.`))
					Eventually(session).Should(Say(`Credentials were rejected, please try again.`))
					Eventually(session).Should(Say(`Password>`))
					Eventually(session).Should(Say(`Authenticating\.\.\.`))
					Eventually(session).Should(Say(`Credentials were rejected, please try again.`))
					Eventually(session).Should(Say(`API endpoint:\s+` + helpers.GetAPI() + `\s+\(API version: \d\.\d{1,3}\.\d{1,3}\)`))
					Eventually(session).Should(Say(`Not logged in. Use 'cf login' to log in.`))
					Eventually(session).Should(Say(`FAILED`))
					Eventually(session).Should(Say(`Unable to authenticate`))

					Eventually(session).Should(Exit(1))
				})

				Context("and the user was previously logged in", func() {
					BeforeEach(func() {
						helpers.LoginCF()
					})

					It("logs them out", func() {
						username, password := helpers.GetCredentials()
						badPassword := password + "_wrong"
						session := helpers.CF("login", "-p", badPassword, "-u", username)
						Eventually(session).Should(Say(`Not logged in. Use 'cf login' to log in.`))
						Eventually(session).Should(Exit())

						orgsSession := helpers.CF("orgs")
						Eventually(orgsSession.Err).Should(Say(`Not logged in. Use 'cf login' to log in.`))
						Eventually(orgsSession).Should(Exit())
					})
				})
			})

			When("already logged in with client credentials", func() {
				BeforeEach(func() {
					clientID, clientSecret := helpers.SkipIfClientCredentialsNotSet()
					session := helpers.CF("auth", clientID, clientSecret, "--client-credentials")
					Eventually(session).Should(Exit(0))
				})

				It("should fail log in and display an error informing the user they need to log out", func() {
					username, password := helpers.GetCredentials()
					session := helpers.CF("login", "-p", password, "-u", username)
					Eventually(session).Should(Say("API endpoint:\\s+" + helpers.GetAPI()))
					Eventually(session).Should(Say(`API endpoint:\s+` + helpers.GetAPI() + `\s+\(API version: \d\.\d{1,3}\.\d{1,3}\)`))
					// The following message is a bit strange in the output. Consider removing?
					Eventually(session).Should(Say("Not logged in. Use 'cf login' to log in."))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Say("Service account currently logged in. Use 'cf logout' to log out service account and try again."))
					Eventually(session).Should(Exit(1))
				})
			})
		})
	})
})
