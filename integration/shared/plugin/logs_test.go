package plugin

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("logs", func() {
	BeforeEach(func() {
		installTestPlugin()
	})

	AfterEach(func() {
		uninstallTestPlugin()
	})

	var (
		organization string
		space        string
		appName      string
	)

	BeforeEach(func() {
		organization, space = createTargetedOrgAndSpace()
		appName = helpers.PrefixedRandomName("APP")
	})

	AfterEach(func() {
		helpers.QuickDeleteSpace(space)
		helpers.QuickDeleteOrg(organization)
	})

	When("pushing an application from a plugin", func() {
		It("outputs logs from the staging process", func() {
			helpers.WithHelloWorldApp(func(appDir string) {
				session := helpers.CF("CliCommand", "push",
					appName, "-p", appDir, "-b", "staticfile_buildpack")
				Eventually(session).Should(Exit(0))
				Expect(session).To(Say("buildpack: staticfile_buildpack"))
				Expect(session).To(Say("Creating app"))
			})
		})
	})

	XWhen("tailing logs for an app from a plugin", func() {
		BeforeEach(func() {
			helpers.WithHelloWorldApp(func(appDir string) {
				session := helpers.CF("push", appName, "-p", appDir, "-b", "staticfile_buildpack")
				Eventually(session).Should(Exit(0))
			})
		})

		It("outputs the application logs", func() {
			logSession := helpers.CF("CliCommand", "logs", appName)

			restageSession := helpers.CF("restage", appName)
			Eventually(restageSession).Should(Exit(0))

			Eventually(logSession).Should(Say("Staticfile Buildpack version"))
			logSession.Kill()

			Eventually(logSession).Should(Exit())
		})

	})

	XWhen("viewing recent logs for an app from a plugin", func() {

		BeforeEach(func() {
			helpers.WithHelloWorldApp(func(appDir string) {
				session := helpers.CF("push", appName, "-p", appDir, "-b", "staticfile_buildpack")
				Eventually(session).Should(Exit(0))
			})
		})

		It("outputs the recent application logs", func() {
			session := helpers.CF("CliCommand", "logs", appName, "--recent")
			Eventually(session).Should(Exit(0))
			Expect(session).To(Say("Staticfile Buildpack version"))
		})

	})
})
