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
			It("displays command usage to output", func() {
				session := helpers.CF("push", "--help")
				Eventually(session).Should(Say("--health-check-type, -u\\s+Application health check type \\(Default: 'port', 'none' accepted for 'process', 'http' implies endpoint '/'\\)"))
				Eventually(session).Should(Exit(0))
			})

			It("displays health check timeout (-t) flag description", func() {
				session := helpers.CF("push", "--help")
				Eventually(session).Should(Say("-t\\s+Time \\(in seconds\\) allowed to elapse between starting up an app and the first healthy response from the app"))
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
			spaceName = helpers.NewSpaceName()

			setupCF(orgName, spaceName)

			appName = helpers.PrefixedRandomName("app")
		})

		Context("when displaying help in the old code", func() {
			It("displays command usage to output", func() {
				session := helpers.CF("push")
				Eventually(session).Should(Say("--health-check-type, -u\\s+Application health check type \\(Default: 'port', 'none' accepted for 'process', 'http' implies endpoint '/'\\)"))
				Eventually(session).Should(Exit(1))
			})

			It("displays health check timeout (-t) flag description", func() {
				session := helpers.CF("push")
				Eventually(session).Should(Say("-t\\s+Time \\(in seconds\\) allowed to elapse between starting up an app and the first healthy response from the app"))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when pushing app without a manifest", func() {
			Context("when the app doesn't already exist", func() {
				DescribeTable("displays the correct health check type",
					func(healthCheckType string, endpoint string) {
						helpers.WithHelloWorldApp(func(appDir string) {
							Eventually(helpers.CF("push", appName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "-u", healthCheckType)).Should(Exit(0))
						})

						session := helpers.CF("get-health-check", appName)
						Eventually(session).Should(Say("health check type:\\s+%s", healthCheckType))
						Eventually(session).Should(Say("endpoint \\(for http type\\):\\s+%s\n", endpoint))
						Eventually(session).Should(Exit(0))
					},

					Entry("when the health check type is none", "none", ""),
					Entry("when the health check type is process", "process", ""),
					Entry("when the health check type is port", "port", ""),
					Entry("when the health check type is http", "http", "/"),
				)
			})

			Context("when the app already exists", func() {
				BeforeEach(func() {
					helpers.WithHelloWorldApp(func(appDir string) {
						Eventually(helpers.CF("push", appName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "-u", "port")).Should(Exit(0))
					})
				})

				Context("when the app does not already have a health-check-http-endpoint' configured", func() {
					Context("when setting the health check type to 'http'", func() {
						BeforeEach(func() {
							helpers.WithHelloWorldApp(func(appDir string) {
								Eventually(helpers.CF("push", appName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "-u", "http")).Should(Exit(0))
							})
						})

						It("sets the endpoint to /", func() {
							session := helpers.CF("get-health-check", appName)
							Eventually(session).Should(Say("endpoint \\(for http type\\):\\s+\\/\n"))
							Eventually(session).Should(Exit(0))
						})
					})
				})

				Context("when the app already has a health check 'http' endpoint set", func() {
					BeforeEach(func() {
						Eventually(helpers.CF("set-health-check", appName, "http", "--endpoint", "/some-endpoint")).Should(Exit(0))

						session := helpers.CF("get-health-check", appName)
						Eventually(session).Should(Say("endpoint \\(for http type\\):\\s+/some-endpoint"))
						Eventually(session).Should(Exit(0))
					})

					Context("when the health check type to 'http'", func() {
						BeforeEach(func() {
							helpers.WithHelloWorldApp(func(appDir string) {
								Eventually(helpers.CF("push", appName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "-u", "http")).Should(Exit(0))
							})
						})

						It("preserves the existing endpoint", func() {
							session := helpers.CF("get-health-check", appName)
							Eventually(session).Should(Say("endpoint \\(for http type\\):\\s+\\/some-endpoint\n"))
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
							session := helpers.CF("get-health-check", appName, "-v")
							Eventually(session).Should(Say(`"health_check_http_endpoint": "/some-endpoint"`))
							Eventually(session).Should(Exit(0))
						})
					})
				})
			})
		})

		Context("when pushing with manifest", func() {
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
					Eventually(session).Should(Say("health check type:\\s+%s", healthCheckType))
					Eventually(session).Should(Say("endpoint \\(for http type\\):\\s+%s\n", endpoint))
					Eventually(session).Should(Exit(0))
				},

				Entry("when the health check type is none", "none", ""),
				Entry("when the health check type is process", "process", ""),
				Entry("when the health check type is port", "port", ""),
				Entry("when the health check type is http", "http", "/"),
			)

			Context("when the health check type is not 'http' but an endpoint is provided", func() {
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
						Eventually(session).Should(Say("Health check type must be 'http' to set a health check HTTP endpoint."))
						Eventually(session).Should(Exit(1))
					})
				})
			})

			Context("when the health check type is http and an endpoint is provided", func() {
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
					Eventually(session).Should(Say("health check type:\\s+http"))
					Eventually(session).Should(Say("endpoint \\(for http type\\):\\s+/some-endpoint\n"))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when the app already exists", func() {
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
					Eventually(session).Should(Say("health check type:\\s+http"))
					Eventually(session).Should(Say("endpoint \\(for http type\\):\\s+/new-endpoint\n"))
					Eventually(session).Should(Exit(0))
				})

				It("uses the existing endpoint if one isn't provided", func() {
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
`, appName))
						manifestPath := filepath.Join(appDir, "manifest.yml")
						err := ioutil.WriteFile(manifestPath, manifestContents, 0666)
						Expect(err).ToNot(HaveOccurred())
						Eventually(helpers.CF("push", "--no-start", "-p", appDir, "-f", manifestPath, "-b", "staticfile_buildpack")).Should(Exit(0))
					})

					session := helpers.CF("get-health-check", appName)
					Eventually(session).Should(Say("health check type:\\s+http"))
					Eventually(session).Should(Say("endpoint \\(for http type\\):\\s+/some-endpoint\n"))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when also pushing app with -u option", func() {
				Context("when the -u option is 'port'", func() {
					It("overrides the health check type in the manifest", func() {
						helpers.WithHelloWorldApp(func(appDir string) {
							manifestContents := []byte(fmt.Sprintf(`
---
applications:
- name: %s
  memory: 128M
  health-check-type: http
`, appName))
							manifestPath := filepath.Join(appDir, "manifest.yml")
							err := ioutil.WriteFile(manifestPath, manifestContents, 0666)
							Expect(err).ToNot(HaveOccurred())

							Eventually(helpers.CF("push", "--no-start", "-p", appDir, "-f", manifestPath, "-b", "staticfile_buildpack", "-u", "port")).Should(Exit(0))
						})

						session := helpers.CF("get-health-check", appName)
						Eventually(session).Should(Say("health check type:\\s+port"))
						Eventually(session).Should(Exit(0))
					})
				})

				Context("when the -u option is 'http'", func() {
					It("uses the endpoint in the manifest", func() {
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
						Eventually(session).Should(Say("health check type:\\s+http"))
						Eventually(session).Should(Say("(?m)endpoint \\(for http type\\):\\s+/$"))
						Eventually(session).Should(Exit(0))
					})
				})
			})
		})
	})
})
