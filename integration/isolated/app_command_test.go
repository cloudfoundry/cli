package isolated

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("app command", func() {
	Describe("help", func() {
		Context("when --help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("app", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("app - Display health and status for an app"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf app APP_NAME"))
				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say("--guid      Retrieve and display the given app's guid.  All other health and status output for the app is suppressed."))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("apps, events, logs, map-route, push, unmap-route"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("when the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "app", "some-app")
		})
	})

	Context("when the environment is set up correctly", func() {
		var (
			orgName   string
			spaceName string
		)

		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()

			setupCF(orgName, spaceName)
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		Context("when the app name is not provided", func() {
			It("tells the user that the app name is required, prints help text, and exits 1", func() {
				session := helpers.CF("app")

				Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `APP_NAME` was not provided"))
				Eventually(session.Out).Should(Say("NAME:"))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the app does not exist", func() {
			Context("when no flags are given", func() {
				It("tells the user that the app is not found and exits 1", func() {
					appName := helpers.PrefixedRandomName("app")
					session := helpers.CF("app", appName)

					Eventually(session.Out).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("App %s not found", appName))
					Eventually(session).Should(Exit(1))
				})
			})

			Context("when the --guid flag is given", func() {
				It("tells the user that the app is not found and exits 1", func() {
					appName := helpers.PrefixedRandomName("app")
					session := helpers.CF("app", "--guid", appName)

					Eventually(session.Out).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("App %s not found", appName))
					Eventually(session).Should(Exit(1))
				})
			})
		})

		Context("when the app does exist", func() {
			Context("when the app is a buildpack app", func() {
				var (
					domainName string
					tcpDomain  helpers.Domain
					appName    string
				)

				BeforeEach(func() {
					Eventually(helpers.CF("create-isolation-segment", RealIsolationSegment)).Should(Exit(0))
					Eventually(helpers.CF("enable-org-isolation", orgName, RealIsolationSegment)).Should(Exit(0))
					Eventually(helpers.CF("set-space-isolation-segment", spaceName, RealIsolationSegment)).Should(Exit(0))

					appName = helpers.PrefixedRandomName("app")
					domainName = defaultSharedDomain()
					tcpDomain = helpers.NewDomain(orgName, helpers.DomainName("tcp"))
					tcpDomain.CreateWithRouterGroup(helpers.FindOrCreateTCPRouterGroup(GinkgoParallelNode()))
					helpers.WithHelloWorldApp(func(appDir string) {
						manifestContents := []byte(fmt.Sprintf(`
---
applications:
- name: %s
  memory: 128M
  instances: 2
  disk_quota: 128M
  routes:
  - route: %s.%s
  - route: %s:1024
`, appName, appName, domainName, tcpDomain.Name))
						manifestPath := filepath.Join(appDir, "manifest.yml")
						err := ioutil.WriteFile(manifestPath, manifestContents, 0666)
						Expect(err).ToNot(HaveOccurred())

						// Create manifest
						Eventually(helpers.CF("push", appName, "-p", appDir, "-f", manifestPath, "-b", "staticfile_buildpack")).Should(Exit(0))
					})
				})

				Context("when the app is started and has 2 instances", func() {
					It("displays the app information with instances table", func() {
						session := helpers.CF("app", appName)
						Eventually(session).Should(Say("name:\\s+%s", appName))
						Eventually(session).Should(Say("requested state:\\s+started"))
						Eventually(session).Should(Say("instances:\\s+2/2"))
						Eventually(session).Should(Say("isolation segment:\\s+%s", RealIsolationSegment))
						Eventually(session).Should(Say("usage:\\s+128M x 2 instances"))
						Eventually(session).Should(Say("routes:\\s+[\\w\\d-]+\\.%s, %s:1024", domainName, tcpDomain.Name))
						Eventually(session).Should(Say("last uploaded:\\s+\\w{3} [0-3]\\d \\w{3} [0-2]\\d:[0-5]\\d:[0-5]\\d \\w+ \\d{4}"))
						Eventually(session).Should(Say("stack:\\s+cflinuxfs2"))
						Eventually(session).Should(Say("buildpack:\\s+staticfile_buildpack"))
						Eventually(session).Should(Say(""))
						Eventually(session).Should(Say("state\\s+since\\s+cpu\\s+memory\\s+disk\\s+details"))
						Eventually(session).Should(Say("#0\\s+running\\s+\\d{4}-[01]\\d-[0-3]\\dT[0-2][0-9]:[0-5]\\d:[0-5]\\dZ\\s+\\d+\\.\\d+%.*of 128M.*of 128M"))
						Eventually(session).Should(Say("#1\\s+running\\s+\\d{4}-[01]\\d-[0-3]\\dT[0-2][0-9]:[0-5]\\d:[0-5]\\dZ\\s+\\d+\\.\\d+%.*of 128M.*of 128M"))
						Eventually(session).Should(Exit(0))
					})
				})

				Context("when the app is stopped", func() {
					BeforeEach(func() {
						Eventually(helpers.CF("stop", appName)).Should(Exit(0))
					})

					It("displays the app information", func() {
						session := helpers.CF("app", appName)
						Eventually(session).Should(Say("name:\\s+%s", appName))
						Eventually(session).Should(Say("requested state:\\s+stopped"))
						Eventually(session).Should(Say("instances:\\s+0/2"))
						Eventually(session).Should(Say("usage:\\s+128M x 2 instances"))
						Eventually(session).Should(Say("routes:\\s+[\\w\\d-]+.%s, %s:1024", domainName, tcpDomain.Name))
						Eventually(session).Should(Say("last uploaded:"))
						Eventually(session).Should(Say("stack:\\s+cflinuxfs2"))
						Eventually(session).Should(Say("buildpack:\\s+staticfile_buildpack"))

						Eventually(session).Should(Say("There are no running instances of this app."))
						Eventually(session).Should(Exit(0))
					})
				})

				Context("when the app has 0 instances", func() {
					BeforeEach(func() {
						Eventually(helpers.CF("scale", appName, "-i", "0")).Should(Exit(0))
					})

					It("displays the app information", func() {
						session := helpers.CF("app", appName)
						Eventually(session).Should(Say("name:\\s+%s", appName))
						Eventually(session).Should(Say("requested state:\\s+started"))
						Eventually(session).Should(Say("instances:\\s+0/0"))
						Eventually(session).Should(Say("usage:\\s+128M x 0 instances"))
						Eventually(session).Should(Say("routes:\\s+[\\w\\d-]+\\.%s, %s:1024", domainName, tcpDomain.Name))
						Eventually(session).Should(Say("last uploaded:"))
						Eventually(session).Should(Say("stack:\\s+cflinuxfs2"))
						Eventually(session).Should(Say("buildpack:\\s+staticfile_buildpack"))

						Eventually(session).Should(Say("There are no running instances of this app."))
						Eventually(session).Should(Exit(0))
					})
				})

				Context("when the --guid flag is given", func() {
					var appGUID string

					BeforeEach(func() {
						session := helpers.CF("curl", fmt.Sprintf("/v2/apps?q=name:%s", appName))
						Eventually(session).Should(Exit(0))
						rawJSON := strings.TrimSpace(string(session.Out.Contents()))
						var AppInfo struct {
							Resources []struct {
								Metadata struct {
									GUID string `json:"guid"`
								} `json:"metadata"`
							} `json:"resources"`
						}

						err := json.Unmarshal([]byte(rawJSON), &AppInfo)
						Expect(err).NotTo(HaveOccurred())

						appGUID = AppInfo.Resources[0].Metadata.GUID
					})

					It("displays the app guid", func() {
						session := helpers.CF("app", "--guid", appName)
						Eventually(session).Should(Say(appGUID))
						Eventually(session).Should(Exit(0))
					})
				})
			})

			Context("when the app is a Docker app", func() {
				var (
					tcpDomain helpers.Domain
					appName   string
				)

				BeforeEach(func() {
					appName = helpers.PrefixedRandomName("app")
					tcpDomain = helpers.NewDomain(orgName, helpers.DomainName("tcp"))
					tcpDomain.CreateWithRouterGroup(helpers.FindOrCreateTCPRouterGroup(GinkgoParallelNode()))
					Eventually(helpers.CF("push", appName, "-o", DockerImage)).Should(Exit())
				})

				It("displays the docker image and does not display buildpack", func() {
					session := helpers.CF("app", appName)
					Eventually(session).Should(Say("name:\\s+%s", appName))
					Consistently(session).ShouldNot(Say("buildpack:"))
					Eventually(session).Should(Say("docker image:\\s+%s", DockerImage))
					Consistently(session).ShouldNot(Say("buildpack:"))
					Eventually(session).Should(Exit(0))
				})
			})
		})
	})
})
