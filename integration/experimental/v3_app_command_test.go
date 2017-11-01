package experimental

import (
	"encoding/json"
	"fmt"
	"strings"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("v3-app command", func() {
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

	Describe("help", func() {
		Context("when --help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("v3-app", "--help")

				Eventually(session.Out).Should(Say("NAME:"))
				Eventually(session.Out).Should(Say("v3-app - Display health and status for an app"))
				Eventually(session.Out).Should(Say("USAGE:"))
				Eventually(session.Out).Should(Say("cf v3-app APP_NAME [--guid]"))
				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say("--guid\\s+Retrieve and display the given app's guid.  All other health and status output for the app is suppressed."))

				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("when the app name is not provided", func() {
		It("tells the user that the app name is required, prints help text, and exits 1", func() {
			session := helpers.CF("v3-app")

			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `APP_NAME` was not provided"))
			Eventually(session.Out).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	It("displays the experimental warning", func() {
		session := helpers.CF("v3-app", appName)
		Eventually(session.Out).Should(Say("This command is in EXPERIMENTAL stage and may change without notice"))
		Eventually(session).Should(Exit())
	})

	Context("when the environment is not setup correctly", func() {
		Context("when no API endpoint is set", func() {
			BeforeEach(func() {
				helpers.UnsetAPI()
			})

			It("fails with no API endpoint set message", func() {
				session := helpers.CF("v3-app", appName)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("No API endpoint set\\. Use 'cf login' or 'cf api' to target an endpoint\\."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the v3 api does not exist", func() {
			var server *Server

			BeforeEach(func() {
				server = helpers.StartAndTargetServerWithoutV3API()
			})

			AfterEach(func() {
				server.Close()
			})

			It("fails with error message that the minimum version is not met", func() {
				session := helpers.CF("v3-app", appName)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("This command requires CF API version 3\\.27\\.0 or higher\\."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the v3 api version is lower than the minimum version", func() {
			var server *Server

			BeforeEach(func() {
				server = helpers.StartAndTargetServerWithV3Version("3.0.0")
			})

			AfterEach(func() {
				server.Close()
			})

			It("fails with error message that the minimum version is not met", func() {
				session := helpers.CF("v3-app", appName)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("This command requires CF API version 3\\.27\\.0 or higher\\."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when not logged in", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
			})

			It("fails with not logged in message", func() {
				session := helpers.CF("v3-app", appName)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("Not logged in\\. Use 'cf login' to log in\\."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when there is no org set", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
				helpers.LoginCF()
			})

			It("fails with no org targeted error message", func() {
				session := helpers.CF("v3-app", appName)
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("No org targeted, use 'cf target -o ORG' to target an org\\."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when there is no space set", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
				helpers.LoginCF()
				helpers.TargetOrg(ReadOnlyOrg)
			})

			It("fails with no space targeted error message", func() {
				session := helpers.CF("v3-app", appName)
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("No space targeted, use 'cf target -s SPACE' to target a space\\."))
				Eventually(session).Should(Exit(1))
			})
		})
	})

	Context("when the environment is set up correctly", func() {
		BeforeEach(func() {
			setupCF(orgName, spaceName)
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		Context("when the app exists", func() {
			var domainName string

			BeforeEach(func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "v3-push", appName)).Should(Exit(0))
				})

				domainName = defaultSharedDomain()
			})

			It("displays the app summary", func() {
				userName, _ := helpers.GetCredentials()

				session := helpers.CF("v3-app", appName)
				Eventually(session.Out).Should(Say("Showing health and status for app %s in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, userName))

				Eventually(session.Out).Should(Say("name:\\s+%s", appName))
				Eventually(session.Out).Should(Say("requested state:\\s+started"))
				Eventually(session.Out).Should(Say("processes:\\s+web:1/1"))
				Eventually(session.Out).Should(Say("memory usage:\\s+\\d+[KMG] x 1"))
				Eventually(session.Out).Should(Say("routes:\\s+%s\\.%s", appName, domainName))
				Eventually(session.Out).Should(Say("stack:\\s+cflinuxfs2"))
				Eventually(session.Out).Should(Say("buildpacks:\\s+staticfile"))
				Eventually(session.Out).Should(Say("web:1/1"))
				Eventually(session.Out).Should(Say("#0\\s+running\\s+\\d{4}-\\d{2}-\\d{2} \\d{2}:\\d{2}:\\d{2} [AP]M"))

				Eventually(session).Should(Exit(0))
			})

			Context("when the app is stopped", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("stop", appName)).Should(Exit(0))
				})

				It("displays that there are no running instances of the app", func() {
					userName, _ := helpers.GetCredentials()

					session := helpers.CF("v3-app", appName)

					Eventually(session.Out).Should(Say(`Showing health and status for app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
					Consistently(session.Out).ShouldNot(Say(`state\s+since\s+cpu\s+memory\s+disk`))
					Eventually(session.Out).Should(Say("There are no running instances of this app"))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when the --guid flag is given", func() {
				var appGUID string

				BeforeEach(func() {
					session := helpers.CF("curl", fmt.Sprintf("/v3/apps?names=%s", appName))
					Eventually(session).Should(Exit(0))
					rawJSON := strings.TrimSpace(string(session.Out.Contents()))
					var AppInfo struct {
						Resources []struct {
							GUID string `json:"guid"`
						} `json:"resources"`
					}

					err := json.Unmarshal([]byte(rawJSON), &AppInfo)
					Expect(err).NotTo(HaveOccurred())

					appGUID = AppInfo.Resources[0].GUID
				})

				It("displays the app guid", func() {
					session := helpers.CF("v3-app", "--guid", appName)
					Eventually(session).Should(Say(appGUID))
					Eventually(session).Should(Exit(0))
				})
			})
		})

		Context("when the app is a docker app", func() {
			var domainName string

			BeforeEach(func() {
				Eventually(helpers.CF("v3-push", appName, "-o", PublicDockerImage)).Should(Exit(0))
				domainName = defaultSharedDomain()
			})

			It("displays the app summary", func() {
				userName, _ := helpers.GetCredentials()

				session := helpers.CF("v3-app", appName)
				Eventually(session.Out).Should(Say("Showing health and status for app %s in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, userName))

				Eventually(session.Out).Should(Say("name:\\s+%s", appName))
				Eventually(session.Out).Should(Say("requested state:\\s+started"))
				Eventually(session.Out).Should(Say("processes:\\s+web:1/1"))
				Eventually(session.Out).Should(Say("memory usage:\\s+\\d+[KMG] x 1"))
				Eventually(session.Out).Should(Say("routes:\\s+%s\\.%s", appName, domainName))
				Eventually(session.Out).Should(Say("stack:\\s+"))
				Eventually(session.Out).Should(Say("docker image:\\s+cloudfoundry/diego-docker-app-custom"))
				Eventually(session.Out).Should(Say("web:1/1"))
				Eventually(session.Out).Should(Say("#0\\s+running\\s+\\d{4}-\\d{2}-\\d{2} \\d{2}:\\d{2}:\\d{2} [AP]M"))

				Eventually(session).Should(Exit(0))
			})
		})

		Context("when the app does not exist", func() {
			It("displays app not found and exits 1", func() {
				invalidAppName := "invalid-app-name"
				session := helpers.CF("v3-app", invalidAppName)
				userName, _ := helpers.GetCredentials()

				Eventually(session.Out).Should(Say("Showing health and status for app %s in org %s / space %s as %s\\.\\.\\.", invalidAppName, orgName, spaceName, userName))
				Eventually(session.Err).Should(Say("App %s not found", invalidAppName))
				Eventually(session.Out).Should(Say("FAILED"))

				Eventually(session).Should(Exit(1))
			})

			Context("when the --guid flag is given", func() {
				It("tells the user that the app is not found and exits 1", func() {
					appName := helpers.PrefixedRandomName("invalid-app")
					session := helpers.CF("v3-app", "--guid", appName)

					Eventually(session.Out).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("App %s not found", appName))
					Eventually(session).Should(Exit(1))
				})
			})
		})
	})
})
