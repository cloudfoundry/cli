package push

import (
	"fmt"
	"regexp"
	"strings"

	"code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("push with only an app name", func() {
	var (
		appName  string
		username string
	)

	BeforeEach(func() {
		appName = helpers.NewAppName()
		username, _ = helpers.GetCredentials()
	})

	Describe("app existence", func() {
		Context("when the app does not exist", func() {
			It("creates the app", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName)
					Eventually(session).Should(Say("Pushing app %s to org %s / space %s as %s\\.\\.\\.", appName, organization, space, username))
					Eventually(session).Should(Say("Getting app info\\.\\.\\."))
					Eventually(session).Should(Say("Creating app with these attributes\\.\\.\\."))
					Eventually(session).Should(Say("\\+\\s+name:\\s+%s", appName))
					Eventually(session).Should(Say("\\s+routes:"))
					Eventually(session).Should(Say("(?i)\\+\\s+%s.%s", appName, defaultSharedDomain()))
					Eventually(session).Should(Say("Mapping routes\\.\\.\\."))
					Eventually(session).Should(Say("Comparing local files to remote cache\\.\\.\\."))
					Eventually(session).Should(Say("Packaging files to upload\\.\\.\\."))
					Eventually(session).Should(Say("Uploading files\\.\\.\\."))
					Eventually(session).Should(Say("100.00%"))
					Eventually(session).Should(Say("Waiting for API to complete processing files\\.\\.\\."))
					helpers.ConfirmStagingLogs(session)
					Eventually(session).Should(Say("Waiting for app to start\\.\\.\\."))
					Eventually(session).Should(Say("requested state:\\s+started"))
					Eventually(session).Should(Exit(0))
				})

				session := helpers.CF("app", appName)
				Eventually(session).Should(Say("name:\\s+%s", appName))
				Eventually(session).Should(Exit(0))
			})

			Context("when the app has a non-standard name", func() {
				BeforeEach(func() {
					appName = helpers.PrefixedRandomName("App Name One$")
				})

				It("sanitizes the name and creates the app", func() {
					sanitizedName := fmt.Sprintf("app-name-one-%s", strings.SplitN(appName, "-", 2)[1])

					helpers.WithHelloWorldApp(func(dir string) {
						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName)
						Eventually(session).Should(Say("Pushing app %s to org %s / space %s as %s\\.\\.\\.", regexp.QuoteMeta(appName), organization, space, username))
						Eventually(session).Should(Say("Getting app info\\.\\.\\."))
						Eventually(session).Should(Say("Creating app with these attributes\\.\\.\\."))
						Eventually(session).Should(Say("\\+\\s+name:\\s+%s", regexp.QuoteMeta(appName)))
						Eventually(session).Should(Say("\\s+routes:"))
						Eventually(session).Should(Say("(?i)\\+\\s+%s.%s", regexp.QuoteMeta(sanitizedName), defaultSharedDomain()))
						Eventually(session).Should(Say("Mapping routes\\.\\.\\."))
						Eventually(session).Should(Say("Comparing local files to remote cache\\.\\.\\."))
						Eventually(session).Should(Say("Packaging files to upload\\.\\.\\."))
						Eventually(session).Should(Say("Uploading files\\.\\.\\."))
						Eventually(session).Should(Say("100.00%"))
						Eventually(session).Should(Say("Waiting for API to complete processing files\\.\\.\\."))
						helpers.ConfirmStagingLogs(session)
						Eventually(session).Should(Say("Waiting for app to start\\.\\.\\."))
						Eventually(session).Should(Say("requested state:\\s+started"))
						Eventually(session).Should(Exit(0))
					})

					session := helpers.CF("app", appName)
					Eventually(session).Should(Say("name:\\s+%s", regexp.QuoteMeta(appName)))
					Eventually(session).Should(Exit(0))
				})
			})
		})

		Context("when the app exists", func() {
			BeforeEach(func() {
				helpers.WithHelloWorldApp(func(dir string) {
					Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, "push", appName)).Should(Exit(0))
				})
			})

			It("updates the app", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName)
					Eventually(session).Should(Say("Getting app info\\.\\.\\."))
					Eventually(session).Should(Say("Updating app with these attributes\\.\\.\\."))
					Eventually(session).Should(Say("(?m)^\\s+name:\\s+%s$", appName))
					Eventually(session).Should(Say("\\s+routes:"))
					Eventually(session).Should(Say("(?mi)^\\s+%s.%s$", strings.ToLower(appName), defaultSharedDomain()))
					Eventually(session).Should(Say("Mapping routes\\.\\.\\."))
					Eventually(session).Should(Say("Uploading files\\.\\.\\."))
					Eventually(session).Should(Say("100.00%"))
					Eventually(session).Should(Say("Waiting for API to complete processing files\\.\\.\\."))
					Eventually(session).Should(Say("Stopping app\\.\\.\\."))
					helpers.ConfirmStagingLogs(session)
					Eventually(session).Should(Say("Waiting for app to start\\.\\.\\."))
					Eventually(session).Should(Say("requested state:\\s+started"))
					Eventually(session).Should(Exit(0))
				})

				session := helpers.CF("app", appName)
				Eventually(session).Should(Say("name:\\s+%s", appName))
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
