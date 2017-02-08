package isolated

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Push with health check", func() {
	Context("help", func() {
		Context("when displaying help in the refactor", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("push", "--help")
				Eventually(session).Should(Say("--health-check-type, -u\\s+Application health check type \\(Default: 'port', 'none' accepted for 'process', 'http' implies endpoint '/'\\)"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("when the environment is set up correctly", func() {
		var (
			appName   string
			orgName   string
			spaceName string
		)

		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			spaceName = helpers.PrefixedRandomName("SPACE")

			setupCF(orgName, spaceName)

			appName = helpers.PrefixedRandomName("app")
		})

		Context("when displaying help in the old code", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("push")
				Eventually(session).Should(Say("--health-check-type, -u\\s+Application health check type \\(Default: 'port', 'none' accepted for 'process', 'http' implies endpoint '/'\\)"))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when pushing with flags", func() {
			Context("when setting the health check", func() {
				DescribeTable("displays the correct health check type",
					func(healthCheckType string, endpoint string) {
						helpers.WithHelloWorldApp(func(appDir string) {
							Eventually(helpers.CF("push", appName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "-u", healthCheckType)).Should(Exit(0))
						})

						session := helpers.CF("get-health-check", appName)
						Eventually(session.Out).Should(Say("Health check type:\\s+%s", healthCheckType))
						Eventually(session.Out).Should(Say("Endpoint \\(for http type\\):\\s+%s\n", endpoint))
						Eventually(session).Should(Exit(0))
					},

					Entry("when the health check type is none", "none", ""),
					Entry("when the health check type is process", "process", ""),
					Entry("when the health check type is port", "port", ""),
					Entry("when the health check type is http", "http", "/"),
				)
			})

			Context("when the health check type is not 'http'", func() {
				BeforeEach(func() {
					helpers.WithHelloWorldApp(func(appDir string) {
						Eventually(helpers.CF("push", appName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "-u", "port")).Should(Exit(0))
					})
				})

				Context("when the health check type is set to 'http'", func() {
					BeforeEach(func() {
						helpers.WithHelloWorldApp(func(appDir string) {
							Eventually(helpers.CF("push", appName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "-u", "http")).Should(Exit(0))
						})
					})

					It("sets the endpoint to /", func() {
						session := helpers.CF("get-health-check", appName)
						Eventually(session.Out).Should(Say("Endpoint \\(for http type\\):\\s+\\/\n"))
						Eventually(session).Should(Exit(0))
					})
				})
			})

			Context("when the app already has a health check 'http' endpoint set", func() {
				BeforeEach(func() {
					helpers.WithHelloWorldApp(func(appDir string) {
						Eventually(helpers.CF("push", appName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "-u", "http")).Should(Exit(0))
					})

					Eventually(helpers.CF("set-health-check", appName, "http", "--endpoint", "/some-endpoint")).Should(Exit(0))

					appGUID := helpers.AppGUID(appName)
					session := helpers.CF("curl", fmt.Sprintf("/v2/apps/%s", appGUID))
					Eventually(session.Out).Should(Say(`"health_check_http_endpoint": "/some-endpoint"`))
					Eventually(session).Should(Exit(0))
				})

				Context("when updating the health check type to 'http'", func() {
					BeforeEach(func() {
						helpers.WithHelloWorldApp(func(appDir string) {
							Eventually(helpers.CF("push", appName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "-u", "http")).Should(Exit(0))
						})
					})

					It("sets the endpoint to the default", func() {
						session := helpers.CF("get-health-check", appName)
						Eventually(session.Out).Should(Say("Endpoint \\(for http type\\):\\s+\\/\n"))
						Eventually(session).Should(Exit(0))
					})
				})

				Context("when updating the health check type to something other than 'http'", func() {
					BeforeEach(func() {
						helpers.WithHelloWorldApp(func(appDir string) {
							Eventually(helpers.CF("push", appName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "-u", "port")).Should(Exit(0))
						})
					})

					It("preserves the existing endpoint", func() {
						appGUID := helpers.AppGUID(appName)
						session := helpers.CF("curl", fmt.Sprintf("/v2/apps/%s", appGUID))
						Eventually(session.Out).Should(Say(`"health_check_http_endpoint": "/some-endpoint"`))
						Eventually(session).Should(Exit(0))
					})
				})
			})
		})

		Context("when pushing with manifest", func() {
			Context("when the health type is http and an endpoint is provided", func() {
				It("sets the health check type and endpoint", func() {
					helpers.WithHelloWorldApp(func(appDir string) {
						manifestContents := []byte(fmt.Sprintf(`
---
applications:
- name: %s
  memory: 128M
  health-check-type: http
  health-check-http-endpoint: /some-endpoint
`, appName))
						manifestPath := filepath.Join(appDir, "manifest.yml")
						err := ioutil.WriteFile(manifestPath, manifestContents, 0666)
						Expect(err).ToNot(HaveOccurred())

						Eventually(helpers.CF("push", "--no-start", "-p", appDir, "-f", manifestPath, "-b", "staticfile_buildpack")).Should(Exit(0))
					})

					session := helpers.CF("get-health-check", appName)
					Eventually(session.Out).Should(Say("Health check type:\\s+http"))
					Eventually(session.Out).Should(Say("Endpoint \\(for http type\\):\\s+/some-endpoint\n"))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when pushing an existing app", func() {
				It("updates the health check type and endpoint", func() {
					helpers.WithHelloWorldApp(func(appDir string) {
						manifestContents := []byte(fmt.Sprintf(`
---
applications:
- name: %s
  memory: 128M
  health-check-type: http
  health-check-http-endpoint: /some-endpoint
`, appName))
						manifestPath := filepath.Join(appDir, "manifest.yml")
						err := ioutil.WriteFile(manifestPath, manifestContents, 0666)
						Expect(err).ToNot(HaveOccurred())
						Eventually(helpers.CF("push", "--no-start", "-p", appDir, "-f", manifestPath, "-b", "staticfile_buildpack")).Should(Exit(0))
					})

					helpers.WithHelloWorldApp(func(appDir string) {
						manifestContents := []byte(fmt.Sprintf(`
---
applications:
- name: %s
  memory: 128M
  health-check-type: http
  health-check-http-endpoint: /new-endpoint
`, appName))
						manifestPath := filepath.Join(appDir, "manifest.yml")
						err := ioutil.WriteFile(manifestPath, manifestContents, 0666)
						Expect(err).ToNot(HaveOccurred())
						Eventually(helpers.CF("push", "--no-start", "-p", appDir, "-f", manifestPath, "-b", "staticfile_buildpack")).Should(Exit(0))
					})

					session := helpers.CF("get-health-check", appName)
					Eventually(session.Out).Should(Say("Health check type:\\s+http"))
					Eventually(session.Out).Should(Say("Endpoint \\(for http type\\):\\s+/new-endpoint\n"))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when the health type is not http and an endpoint is provided", func() {
				It("displays an error", func() {
					helpers.WithHelloWorldApp(func(appDir string) {
						manifestContents := []byte(fmt.Sprintf(`
---
applications:
- name: %s
  memory: 128M
  health-check-type: port
  health-check-http-endpoint: /some-endpoint
`, appName))
						manifestPath := filepath.Join(appDir, "manifest.yml")
						err := ioutil.WriteFile(manifestPath, manifestContents, 0666)
						Expect(err).ToNot(HaveOccurred())

						session := helpers.CF("push", "--no-start", "-p", appDir, "-f", manifestPath, "-b", "staticfile_buildpack")
						Eventually(session.Out).Should(Say("Health check type must be 'http' to set a health check HTTP endpoint."))
						Eventually(session).Should(Exit(1))
					})
				})
			})

			Context("when passing an 'http' health check type with the -u option", func() {
				It("resets the endpoint to the default", func() {
					helpers.WithHelloWorldApp(func(appDir string) {
						manifestContents := []byte(fmt.Sprintf(`
---
applications:
- name: %s
  memory: 128M
  health-check-type: http
  health-check-http-endpoint: /some-endpoint
`, appName))
						manifestPath := filepath.Join(appDir, "manifest.yml")
						err := ioutil.WriteFile(manifestPath, manifestContents, 0666)
						Expect(err).ToNot(HaveOccurred())

						Eventually(helpers.CF("push", "--no-start", "-p", appDir, "-f", manifestPath, "-b", "staticfile_buildpack", "-u", "http")).Should(Exit(0))
					})

					session := helpers.CF("get-health-check", appName)
					Eventually(session.Out).Should(Say("Health check type:\\s+http"))
					Eventually(session.Out).Should(Say("Endpoint \\(for http type\\):\\s+/\n"))
					Eventually(session).Should(Exit(0))
				})
			})
		})

		Context("when the manifest contains the health check type attribute", func() {
			DescribeTable("displays the correct health check type",
				func(healthCheckType string, endpoint string) {
					helpers.WithHelloWorldApp(func(appDir string) {
						manifestContents := []byte(fmt.Sprintf(`
---
applications:
- name: %s
  memory: 128M
  health-check-type: %s
`, appName, healthCheckType))
						manifestPath := filepath.Join(appDir, "manifest.yml")
						err := ioutil.WriteFile(manifestPath, manifestContents, 0666)
						Expect(err).ToNot(HaveOccurred())

						Eventually(helpers.CF("push", "--no-start", "-p", appDir, "-f", manifestPath, "-b", "staticfile_buildpack")).Should(Exit(0))
					})

					session := helpers.CF("get-health-check", appName)
					Eventually(session.Out).Should(Say("Health check type:\\s+%s", healthCheckType))
					Eventually(session.Out).Should(Say("Endpoint \\(for http type\\):\\s+%s\n", endpoint))
					Eventually(session).Should(Exit(0))
				},

				Entry("when the health check type is none", "none", ""),
				Entry("when the health check type is process", "process", ""),
				Entry("when the health check type is port", "port", ""),
				Entry("when the health check type is http", "http", "/"),
			)

			Context("when passing a health check type with the -u option", func() {
				It("overrides any health check types in the manifest", func() {
					helpers.WithHelloWorldApp(func(appDir string) {
						manifestContents := []byte(fmt.Sprintf(`
---
applications:
- name: %s
  memory: 128M
  health-check-type: port
`, appName))
						manifestPath := filepath.Join(appDir, "manifest.yml")
						err := ioutil.WriteFile(manifestPath, manifestContents, 0666)
						Expect(err).ToNot(HaveOccurred())

						Eventually(helpers.CF("push", "--no-start", "-p", appDir, "-f", manifestPath, "-b", "staticfile_buildpack", "-u", "http")).Should(Exit(0))
					})

					session := helpers.CF("get-health-check", appName)
					Eventually(session.Out).Should(Say("Health check type:\\s+http"))
					Eventually(session).Should(Exit(0))
				})
			})
		})
	})
})
