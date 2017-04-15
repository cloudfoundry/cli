package push

import (
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
					Eventually(session).Should(Say("Creating app %s in org %s / space %s as %s...", appName, organization, space, username))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Say("Uploading %s...", appName))
					Eventually(session).Should(Say("Uploading app files from: %s", dir))
					Eventually(session).Should(Say("Uploading .*, 2 files"))
					Eventually(session).Should(Say("Done uploading"))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Say("Starting app %s in org %s / space %s as %s...", appName, organization, space, username))
					Eventually(session).Should(Say("Downloaded staticfile_buildpack"))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Say("requested state:\\s+started"))
					Eventually(session).Should(Exit(0))
				})

				session := helpers.CF("app", appName)
				Eventually(session).Should(Say("name:\\s+%s", appName))
				Eventually(session).Should(Exit(0))
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
					Eventually(session).Should(Say("Updating app %s in org %s / space %s as %s...", appName, organization, space, username))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Say("Uploading %s...", appName))
					Eventually(session).Should(Say("Uploading app files from: %s", dir))
					Eventually(session).Should(Say("Uploading .*, 2 files"))
					Eventually(session).Should(Say("Done uploading"))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Say("Stopping app %s in org %s / space %s as %s...", appName, organization, space, username))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Say("Starting app %s in org %s / space %s as %s...", appName, organization, space, username))
					Eventually(session).Should(Say("Downloaded staticfile_buildpack"))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Say("requested state:\\s+started"))
					Eventually(session).Should(Exit(0))
				})

				session := helpers.CF("app", appName)
				Eventually(session).Should(Say("name:\\s+%s", appName))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Describe("route existence", func() {
		Context("when the route does not exist", func() {
			It("creates and binds the route", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName)
					Eventually(session).Should(Say("Creating route %s.%s...", strings.ToLower(appName), defaultSharedDomain()))
					Eventually(session).Should(Say("OK"))

					Eventually(session).Should(Say("Binding %s.%s to %s...", strings.ToLower(appName), defaultSharedDomain(), appName))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))
				})
			})
		})

		Context("when the route exists in the current space", func() {
			BeforeEach(func() {
				session := helpers.CF("create-route", space, defaultSharedDomain(), "-n", appName)
				Eventually(session).Should(Exit(0))
			})

			It("should not create the route", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName)
					Consistently(session).ShouldNot(Say("Creating route %s.%s...", strings.ToLower(appName), defaultSharedDomain()))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when the route is not bound to an app of the same name", func() {
				It("binds the route to the app", func() {
					helpers.WithHelloWorldApp(func(dir string) {
						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName)
						Eventually(session).Should(Say("Using route %s.%s", strings.ToLower(appName), defaultSharedDomain()))
						Eventually(session).Should(Say("Binding %s.%s to %s...", strings.ToLower(appName), defaultSharedDomain(), appName))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Exit(0))
					})
				})
			})

			Context("when the route is already bound to the application", func() {
				BeforeEach(func() {
					helpers.WithHelloWorldApp(func(dir string) {
						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName)
						Eventually(session).Should(Exit(0))
					})
				})

				It("does not rebind the route", func() {
					helpers.WithHelloWorldApp(func(dir string) {
						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName)
						Consistently(session).ShouldNot(Say("Binding %s.%s to %s...", strings.ToLower(appName), defaultSharedDomain(), appName))
						Eventually(session).Should(Exit(0))
					})
				})
			})
		})

		Context("when the route exists in a different space", func() {
			BeforeEach(func() {
				otherSpace := helpers.NewSpaceName()
				Eventually(helpers.CF("create-space", otherSpace)).Should(Exit(0))
				Eventually(helpers.CF("create-route", otherSpace, defaultSharedDomain(), "-n", appName)).Should(Exit(0))
			})

			It("errors", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName)
					Eventually(session).Should(Say("Using route %s.%s", strings.ToLower(appName), defaultSharedDomain()))
					Eventually(session).Should(Say("Binding %s.%s to %s...", strings.ToLower(appName), defaultSharedDomain(), appName))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("The route %s.%s is already in use.", appName, defaultSharedDomain()))
					Eventually(session.Err).Should(Say("TIP: Change the hostname with -n HOSTNAME or use --random-route to generate a new route and then push again."))
					Eventually(session).Should(Exit(1))
				})
			})
		})
	})
})
