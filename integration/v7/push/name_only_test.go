// +build !partialPush

package push

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = When("only the name is provided", func() {
	var (
		appName    string
		userName   string
		domainName string
	)

	BeforeEach(func() {
		appName = helpers.PrefixedRandomName("app")
		userName, _ = helpers.GetCredentials()
		domainName = helpers.DefaultSharedDomain()
	})

	When("the app does not already exist", func() {
		It("creates a new app", func() {
			helpers.WithHelloWorldApp(func(appDir string) {
				session := helpers.CustomCF(
					helpers.CFEnv{WorkingDirectory: appDir},
					PushCommandName, appName,
				)

				Eventually(session).Should(Say(`Creating app %s in org %s / space %s as %s\.\.\.`, appName, organization, space, userName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Say(`Uploading and creating bits package for app %s in org %s / space %s as %s\.\.\.`, appName, organization, space, userName))
				Eventually(session).Should(Say("OK"))
				Consistently(session).ShouldNot(Say("Stopping app %s", appName))
				Eventually(session).Should(Say(`Staging package for app %s in org %s / space %s as %s\.\.\.`, appName, organization, space, userName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Say(`Setting app %s to droplet .+ in org %s / space %s as %s\.\.\.`, appName, organization, space, userName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Say(`Mapping routes\.\.\.`))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Say(`Starting app %s in org %s / space %s as %s\.\.\.`, appName, organization, space, userName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Say(`Waiting for app to start\.\.\.`))
				Eventually(session).Should(Say(`Showing health and status for app %s in org %s / space %s as %s\.\.\.`, appName, organization, space, userName))
				Eventually(session).Should(Say(`name:\s+%s`, appName))
				Eventually(session).Should(Say(`requested state:\s+started`))
				Eventually(session).Should(Say(`routes:\s+%s\.%s`, appName, domainName))
				Eventually(session).Should(Say(`stack:\s+cflinuxfs2`))
				Eventually(session).Should(Say(`buildpacks:\s+staticfile`))
				Eventually(session).Should(Say(`type:\s+web`))
				Eventually(session).Should(Say(`instances:\s+1/1`))
				Eventually(session).Should(Say(`memory usage:\s+\d+(M|G)`))
				Eventually(session).Should(Say(`state\s+since\s+cpu\s+memory\s+disk`))
				Eventually(session).Should(Say(`#0\s+running\s+\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2} [AP]M`))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the app exists", func() {
		BeforeEach(func() {
			helpers.WithHelloWorldApp(func(appDir string) {
				Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir},
					PushCommandName, appName,
				)).Should(Exit(0))
			})
		})

		It("pushes the app", func() {
			helpers.WithHelloWorldApp(func(appDir string) {
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir},
					PushCommandName, appName,
				)

				Eventually(session).Should(Say(`Updating app %s in org %s / space %s as %s\.\.\.`, appName, organization, space, userName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Say(`Uploading and creating bits package for app %s in org %s / space %s as %s\.\.\.`, appName, organization, space, userName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Say(`Stopping app %s in org %s / space %s as %s\.\.\.`, appName, organization, space, userName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Say(`Staging package for app %s in org %s / space %s as %s\.\.\.`, appName, organization, space, userName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Say(`Setting app %s to droplet .+ in org %s / space %s as %s\.\.\.`, appName, organization, space, userName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Say(`Mapping routes\.\.\.`))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Say(`Starting app %s in org %s / space %s as %s\.\.\.`, appName, organization, space, userName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Say(`Waiting for app to start\.\.\.`))
				Eventually(session).Should(Say(`Showing health and status for app %s in org %s / space %s as %s\.\.\.`, appName, organization, space, userName))
				Eventually(session).Should(Say(`name:\s+%s`, appName))
				Eventually(session).Should(Say(`requested state:\s+started`))
				Eventually(session).Should(Say(`routes:\s+%s\.%s`, appName, domainName))
				Eventually(session).Should(Say(`stack:\s+cflinuxfs2`))

				Eventually(session).Should(Say(`buildpacks:\s+staticfile`))
				Eventually(session).Should(Say(`type:\s+web`))
				Eventually(session).Should(Say(`instances:\s+1/1`))
				Eventually(session).Should(Say(`memory usage:\s+\d+(M|G)`))
				Eventually(session).Should(Say(`state\s+since\s+cpu\s+memory\s+disk`))
				Eventually(session).Should(Say(`#0\s+running\s+\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2} [AP]M`))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the app crashes", func() {
		It("pushes the app", func() {
			helpers.WithCrashingApp(func(appDir string) {
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, PushCommandName, appName)
				Eventually(session).Should(Say(`Creating app %s in org %s / space %s as %s\.\.\.`, appName, organization, space, userName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Say(`Uploading and creating bits package for app %s in org %s / space %s as %s\.\.\.`, appName, organization, space, userName))
				Eventually(session).Should(Say("OK"))
				Consistently(session).ShouldNot(Say("Stopping app %s", appName))
				Eventually(session).Should(Say(`Staging package for app %s in org %s / space %s as %s\.\.\.`, appName, organization, space, userName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Say(`Setting app %s to droplet .+ in org %s / space %s as %s\.\.\.`, appName, organization, space, userName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Say(`Mapping routes\.\.\.`))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Say(`Starting app %s in org %s / space %s as %s\.\.\.`, appName, organization, space, userName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Say(`Waiting for app to start\.\.\.`))
				Eventually(session).Should(Say(`Showing health and status for app %s in org %s / space %s as %s\.\.\.`, appName, organization, space, userName))
				Eventually(session).Should(Say(`name:\s+%s`, appName))
				Eventually(session).Should(Say(`requested state:\s+started`))
				Eventually(session).Should(Say(`routes:\s+%s\.%s`, appName, domainName))
				Eventually(session).Should(Say(`stack:\s+cflinuxfs2`))
				Eventually(session).Should(Say(`buildpacks:\s+ruby`))
				Eventually(session).Should(Say(`type:\s+web`))
				Eventually(session).Should(Say(`instances:\s+0/1`))
				Eventually(session).Should(Say(`memory usage:\s+\d+(M|G)`))
				Eventually(session).Should(Say(`state\s+since\s+cpu\s+memory\s+disk`))
				Eventually(session).Should(Say(`#0\s+crashed\s+\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2} [AP]M`))
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
