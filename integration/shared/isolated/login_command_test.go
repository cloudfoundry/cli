package isolated

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"time"

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

	Describe("Invalid Command Usage", func() {
		When("a random flag is passed in", func() {
			It("should exit 1 and display an unknown flag error message", func() {
				session := helpers.CF("login", "--test")
				Eventually(session).Should(Exit(1))

				Expect(session.Err).Should(Say("Incorrect Usage: unknown flag `test'"))
			})
		})
	})

	Describe("Minimum Version Check", func() {
		When("the api version is less than the minimum supported version", func() {
			var server *ghttp.Server

			BeforeEach(func() {
				server = helpers.StartServerWithAPIVersions("2.99.9", "3.34.9")

				fakeTokenResponse := map[string]string{
					"access_token":  helpers.BuildTokenString(time.Now()),
					"token_type":    "bearer",
					"refresh_token": "refresh-token",
				}
				server.RouteToHandler(http.MethodPost, "/oauth/token",
					ghttp.RespondWithJSONEncoded(http.StatusOK, fakeTokenResponse))
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
				helpers.TurnOnExperimentalLogin()
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
				helpers.TurnOffExperimentalLogin()
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
		BeforeEach(func() {
			helpers.TurnOnExperimentalLogin()
		})

		AfterEach(func() {
			helpers.TurnOffExperimentalLogin()
		})

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
		BeforeEach(func() {
			helpers.TurnOnExperimentalLogin()
		})

		AfterEach(func() {
			helpers.TurnOffExperimentalLogin()
		})

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

	Describe("Target Organization", func() {
		var (
			orgName  string
			username string
			password string
		)

		BeforeEach(func() {
			helpers.TurnOnExperimentalLogin()
			helpers.LoginCF()
			orgName = helpers.NewOrgName()
			session := helpers.CF("create-org", orgName)
			Eventually(session).Should(Exit(0))
			username, password = helpers.CreateUserInOrgRole(orgName, "OrgManager")
		})

		AfterEach(func() {
			helpers.TurnOffExperimentalLogin()
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
						server = helpers.StartAndTargetServerWithAPIVersions(helpers.DefaultV2Version, helpers.DefaultV3Version)
						helpers.AddLoginRoutes(server)
						helpers.AddFiftyOneOrgs(server)
					})

					It("displays a message and prompt the user for the org name", func() {
						input := NewBuffer()
						_, wErr := input.Write([]byte(fmt.Sprintf("%s\n", "org20"))) // "org20" is one of the orgs in the text fixture
						Expect(wErr).ToNot(HaveOccurred())

						session := helpers.CFWithStdin(input, "login", "-u", username, "-p", password, "--skip-ssl-validation")

						Eventually(session).Should(Say("There are too many options to display; please type in the name."))
						Eventually(session).Should(Say("Org:\\s+org20"))
						Eventually(session).Should(Exit(0))
					})
				})

				When("user selects an organization by using numbered list", func() {
					// required
					It("prompts the user for org and target the selected org", func() {
						input := NewBuffer()
						_, err := input.Write([]byte("1\n"))
						Expect(err).ToNot(HaveOccurred())
						var session *Session
						// TODO: do we still need this?
						if skipSSLValidation {
							session = helpers.CFWithStdin(input, "login", "-u", username, "-p", password, "-a", apiURL, "--skip-ssl-validation")
						} else {
							session = helpers.CFWithStdin(input, "login", "-u", username, "-p", password, "-a", apiURL)
						}

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
						// allowed to change
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
					// required
					It("prompts the user for an org and then targets the selected org", func() {
						input := NewBuffer()
						_, err := input.Write([]byte(fmt.Sprintf("%s\n", orgName)))
						Expect(err).ToNot(HaveOccurred())

						var session *Session
						if skipSSLValidation {
							session = helpers.CFWithStdin(input, "login", "-u", username, "-p", password, "-a", apiURL, "--skip-ssl-validation")
						} else {
							session = helpers.CFWithStdin(input, "login", "-u", username, "-p", password, "-a", apiURL)
						}
						Eventually(session).Should(Say(`\d\. %s`, orgName))
						Eventually(session).Should(Exit(0))

						targetSession := helpers.CF("target")
						Eventually(targetSession).Should(Exit(0))
						Eventually(targetSession).Should(Say(`org:\s+%s`, orgName))
					})
				})

				When("user does not select an organization", func() {
					// allowed to change
					It("succesfully logs in but does not target any org", func() {
						input := NewBuffer()
						_, err := input.Write([]byte("\n"))
						Expect(err).ToNot(HaveOccurred())

						var session *Session
						if skipSSLValidation {
							session = helpers.CFWithStdin(input, "login", "-u", username, "-p", password, "-a", apiURL, "--skip-ssl-validation")
						} else {
							session = helpers.CFWithStdin(input, "login", "-u", username, "-p", password, "-a", apiURL)
						}
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

		When("the -o flag is passed", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
			})

			When("the organization is valid", func() {
				It("targets the organization that was passed as an argument", func() {
					session := helpers.CF("login", "-u", username, "-p", password, "-o", orgName)

					Eventually(session).Should(Exit(0))
					Eventually(session).Should(Say(`Org:\s+%s`, orgName))

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

			When("the -s flag is passed", func() {
				BeforeEach(func() {
					helpers.TurnOnExperimentalLogin()
				})

				AfterEach(func() {
					helpers.TurnOffExperimentalLogin()
				})

				It("targets the org and the space", func() {
					session := helpers.CF("login", "-u", username, "-p", password, "-a", apiURL, "-s", spaceName, "--skip-ssl-validation")

					Eventually(session).Should(Say(`Targeted org:\s+%s`, orgName))
					Eventually(session).Should(Say(`\n\nTargeted space:\s+%s`, spaceName))

					Eventually(session).Should(Say(`Org:\s+%s`, orgName))
					Eventually(session).Should(Say(`Space:\s+%s`, spaceName))
					Eventually(session).Should(Exit(0))

					sessionOutput := string(session.Out.Contents())
					Expect(sessionOutput).To(MatchRegexp(`\S\n\n\n\nAPI`))

					targetSession := helpers.CF("target")
					Eventually(targetSession).Should(Exit(0))
					Eventually(targetSession).Should(Say(`org:\s+%s`, orgName))
					Eventually(targetSession).Should(Say(`space:\s+%s`, spaceName))
				})

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
					helpers.TurnOnExperimentalLogin()
					orgName2 := helpers.NewOrgName()
					session := helpers.CF("create-org", orgName2)
					Eventually(session).Should(Exit(0))
					session = helpers.CF("set-org-role", username, orgName2, "OrgManager")
					Eventually(session).Should(Exit(0))
				})

				AfterEach(func() {
					helpers.TurnOffExperimentalLogin()
				})

				It("targets the org and the space", func() {
					stdin := NewBuffer()
					session := helpers.CFWithStdin(stdin, "login", "-u", username, "-p", password, "-a", apiURL, "-s", spaceName, "--skip-ssl-validation")
					_, writeErr := stdin.Write([]byte(orgName + "\n"))
					Expect(writeErr).ToNot(HaveOccurred())

					Eventually(session).Should(Say(`Targeted org:\s+%s`, orgName))
					Eventually(session).Should(Say(`\n\nTargeted space:\s+%s`, spaceName))

					Eventually(session).Should(Say(`Org:\s+%s`, orgName))
					Eventually(session).Should(Say(`Space:\s+%s`, spaceName))
					Eventually(session).Should(Exit(0))

					sessionOutput := string(session.Out.Contents())
					Expect(sessionOutput).To(MatchRegexp(`\S\n\n\n\nAPI`))

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
						session := helpers.CFWithStdin(stdin, "login", "-u", username, "-p", password, "-a", apiURL, "-s", spaceName, "--skip-ssl-validation")
						_, writeErr := stdin.Write([]byte(orgName + "\n"))
						Expect(writeErr).ToNot(HaveOccurred())
						Eventually(session).Should(Exit(1))
						Eventually(session).Should(Say("FAILED"))
						Eventually(session.Err).Should(Say("Space '%s' not found", spaceName))

						targetSession := helpers.CF("target")
						Eventually(targetSession).Should(Exit(0))
						Eventually(targetSession).Should(Say(`org:\s+%s`, orgName))
						Eventually(targetSession).ShouldNot(Say(`space:\s+%s`, spaceName))
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
					expectedSpaceName := re.FindStringSubmatch(string(session.Out.Contents()))[1]

					targetSession := helpers.CF("target")
					Eventually(targetSession).Should(Exit(0))
					Eventually(targetSession).Should(Say(`space:\s+%s`, expectedSpaceName))
				})

				It("reprompts the user if an invalid number is entered", func() {
					input := NewBuffer()
					_, err := input.Write([]byte("4\n"))
					Expect(err).ToNot(HaveOccurred())

					session := helpers.CFWithStdin(input, "login", "-u", username, "-p", password, "-a", apiURL, "--skip-ssl-validation")
					Eventually(session).Should(Say(regexp.QuoteMeta("Select a space (or press enter to skip):")))
					Eventually(session).Should(Say(regexp.QuoteMeta("Select a space (or press enter to skip):")))
					session.Interrupt()
					Eventually(session).Should(Exit())
				})

				It("prompts the user to pick their space by name", func() {
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
					_, err := input.Write([]byte(" \n"))
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
						Eventually(session).Should(Say("Space %s not found", spaceName))

						targetSession := helpers.CF("target")
						Eventually(targetSession).Should(Exit(0))
						Eventually(targetSession).Should(Say(`org:\s+%s`, orgName))
						Eventually(targetSession).ShouldNot(Say(`space:\s+%s`, spaceName))
						Eventually(targetSession).Should(Say("No space targeted, use 'cf target -s SPACE'"))
					})
				})
			})
		})
	})

	Describe("User Credentials", func() {
		BeforeEach(func() {
			helpers.TurnOnExperimentalLogin()
		})

		AfterEach(func() {
			helpers.TurnOffExperimentalLogin()
		})

		It("prompts the user for email and password", func() {
			username, password := helpers.GetCredentials()
			buffer := NewBuffer()
			_, err := buffer.Write([]byte(fmt.Sprintf("%s\n%s\n", username, password)))
			Expect(err).ToNot(HaveOccurred())
			session := helpers.CFWithStdin(buffer, "login")
			Eventually(session).Should(Say("Email:"))
			Eventually(session).Should(Say("\n\n"))
			Eventually(session).Should(Say("Password:"))
			Eventually(session).Should(Say("\n\n"))
			Eventually(session).Should(Exit(0))
		})

		When("the user's account has been locked due to too many failed attempts", func() {
			var username string

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
				username, password := helpers.GetCredentials()
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
				username, password := helpers.GetCredentials()
				input := NewBuffer()
				_, err := input.Write([]byte(username + "\n"))
				Expect(err).ToNot(HaveOccurred())
				session := helpers.CFWithStdin(input, "login", "-p", password)
				Eventually(session).Should(Say("Email: "))
				Eventually(session).Should(Exit(0))
			})

			When("the password flag is given incorrectly", func() {
				It("Prompts the user two more times before exiting with an error", func() {
					username, _ := helpers.GetCredentials()
					input := NewBuffer()
					_, err := input.Write([]byte(username + "\n" + "bad-password\n" + "bad-password\n"))
					Expect(err).ToNot(HaveOccurred())
					session := helpers.CFWithStdin(input, "login", "-p", "bad-password")
					Eventually(session).Should(Say("Email: "))
					Eventually(session.Err).Should(Say("Credentials were rejected, please try again."))
					Eventually(session).Should(Say("Password: "))
					Eventually(session.Err).Should(Say("Credentials were rejected, please try again."))
					Eventually(session).Should(Say("Not logged in. Use 'cf login' to log in."))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("Unable to authenticate."))
					Eventually(session).Should(Exit(1))
				})
			})
		})

		When("multiple interactive prompts are used", func() {
			var (
				orgName  string
				orgName2 string
				username string
				password string
			)

			BeforeEach(func() {
				helpers.LoginCF()
				orgName = helpers.NewOrgName()
				session := helpers.CF("create-org", orgName)
				Eventually(session).Should(Exit(0))
				username, password = helpers.CreateUserInOrgRole(orgName, "OrgManager")

				orgName2 = helpers.NewOrgName()
				Eventually(helpers.CF("create-org", orgName2)).Should(Exit(0))
				setOrgRoleSession := helpers.CF("set-org-role", username, orgName2, "OrgManager")
				Eventually(setOrgRoleSession).Should(Exit(0))
			})

			It("should accept each value", func() {
				input := NewBuffer()
				_, err := input.Write([]byte(username + "\n" + password + "\n" + orgName + "\n"))
				Expect(err).ToNot(HaveOccurred())
				session := helpers.CFWithStdin(input, "login")
				Eventually(session).Should(Exit(0))
			})

			When("MFA is enabled", func() {
				var (
					password string
					mfaCode  string
					server   *ghttp.Server
				)

				BeforeEach(func() {
					password = "some-password"
					mfaCode = "123456"
					server = helpers.StartAndTargetServerWithAPIVersions(helpers.DefaultV2Version, helpers.DefaultV3Version)
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
						Eventually(session).Should(Say("\n\n"))
						Eventually(session).Should(Say("Password:"))
						Eventually(session).Should(Say("\n\n"))
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
						Eventually(session).Should(Say("Not logged in. Use 'cf login' to log in."))
						Eventually(session).Should(Say("FAILED"))
						Eventually(session.Err).Should(Say("Unable to authenticate."))

						Eventually(session).Should(Exit(1))
					})
				})
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
					Eventually(session).Should(Say(`API endpoint:\s+` + helpers.GetAPI() + `\s+\(API version: \d\.\d{1,3}\.\d{1,3}\)`))
					Eventually(session).Should(Say(`Not logged in. Use 'cf login' to log in.`))
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
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("Service account currently logged in. Use 'cf logout' to log out service account and try again."))
					Eventually(session).Should(Exit(1))
				})
			})
		})
	})

	Describe("Authenticating as a user, through a custom client", func() {
		var (
			accessTokenExpiration time.Duration
			username              string
			password              string
			customClientID        string
			customClientSecret    string
		)

		BeforeEach(func() {
			helpers.TurnOnExperimentalLogin()
			customClientID, customClientSecret = helpers.SkipIfCustomClientCredentialsNotSet()

			helpers.LoginCF()
			username, password = helpers.CreateUser()

			helpers.SetConfig(func(config *configv3.Config) {
				config.ConfigFile.UAAOAuthClient = customClientID
				config.ConfigFile.UAAOAuthClientSecret = customClientSecret
				config.ConfigFile.UAAGrantType = ""
			})

			session := helpers.CF("login", "-u", username, "-p", password)
			Eventually(session).Should(Exit(0))
		})

		AfterEach(func() {
			helpers.TurnOffExperimentalLogin()
		})

		It("gets a token whose settings match those of the custom client", func() {
			accessTokenExpiration = 120 // this was configured in the pipeline

			config := helpers.GetConfig()

			jwt := helpers.ParseTokenString(config.ConfigFile.AccessToken)
			expires, expIsSet := jwt.Claims().Expiration()
			Expect(expIsSet).To(BeTrue())

			iat, iatIsSet := jwt.Claims().IssuedAt()

			Expect(iatIsSet).To(BeTrue())
			Expect(expires.Sub(iat)).To(Equal(accessTokenExpiration * time.Second))
		})

		It("warns the user that this configuration is deprecated", func() {
			deprecationMessage := "Deprecation warning: Manually writing your client credentials to the config.json is deprecated and will be removed in the future. For similar functionality, please use the `cf auth --client-credentials` command instead."

			session := helpers.CF("login", "-u", username, "-p", password)
			Eventually(session.Err).Should(Say(deprecationMessage))
			Eventually(session).Should(Exit(0))
		})

		When("the token has expired", func() {
			BeforeEach(func() {
				helpers.SetConfig(func(config *configv3.Config) {
					config.ConfigFile.AccessToken = helpers.ExpiredAccessToken()
				})
			})

			It("re-authenticates using the custom client", func() {
				session := helpers.CF("orgs")
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
