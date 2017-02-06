package isolated

import (
	"fmt"

	"code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Push with health check", func() {
	Context("help", func() {
		Context("when --help flag is set", func() {
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

		Context("setting the health check", func() {
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

		Context("when the health check is changed from another type to http", func() {
			BeforeEach(func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CF("push", appName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "-u", "port")).Should(Exit(0))
				})
			})

			It("sets the endpoint to /", func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CF("push", appName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "-u", "http")).Should(Exit(0))
				})

				session := helpers.CF("get-health-check", appName)
				Eventually(session.Out).Should(Say("Endpoint (for http type):\\s+/\n"))
				Eventually(session).Should(Exit(0))
			})
		})

		Context("when the health check is changed from http to another type", func() {
			BeforeEach(func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CF("push", appName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "-u", "http")).Should(Exit(0))
				})
				Eventually(helpers.CF("set-health-check", appName, "http", "--endpoint", "/some-endpoint")).Should(Exit(0))
			})

			It("preserves the current endpoint", func() {
				appGUID := helpers.AppGUID(appName)

				session := helpers.CF("curl", fmt.Sprintf("/v2/apps/%s", appGUID))
				Eventually(session).Should(Exit(0))
				Eventually(session.Out).Should(Say(`"health_check_http_endpoint": "/some-endpoint"`))

				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CF("push", appName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "-u", "port")).Should(Exit(0))
				})

				session = helpers.CF("curl", fmt.Sprintf("/v2/apps/%s", appGUID))
				Eventually(session).Should(Exit(0))
				Eventually(session.Out).Should(Say(`"health_check_http_endpoint": "/some-endpoint"`))
			})
		})
	})
})
