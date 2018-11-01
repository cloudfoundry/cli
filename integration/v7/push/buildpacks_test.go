// +build !partialPush

package push

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("buildpacks", func() {
	var (
		appName string
	)

	BeforeEach(func() {
		appName = helpers.PrefixedRandomName("app")
	})

	When("the -b flag is provided", func() {
		When("an existing application has a buildpack set", func() {
			BeforeEach(func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(
						helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir},
							PushCommandName, appName,
							"-b", "java_buildpack"),
					).Should(Exit(1))
				})
			})

			When("resetting the buildpack to default", func() {
				It("successfully pushes the app", func() {
					helpers.WithHelloWorldApp(func(appDir string) {
						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir},
							PushCommandName, appName,
							"-b", "default",
						)

						Eventually(session).Should(Say(`name:\s+%s`, appName))
						Eventually(session).Should(Say(`requested state:\s+started`))
						Eventually(session).Should(Say(`buildpacks:\s+staticfile`))
						Eventually(session).Should(Exit(0))
					})
				})
			})

			When("omitting the buildpack", func() {
				It("continues using previously set buildpack", func() {
					helpers.WithHelloWorldApp(func(appDir string) {
						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, PushCommandName, appName)
						Eventually(session).Should(Say("FAILED"))
						Eventually(session).Should(Exit(1))
					})
				})
			})
		})

		When("the buildpack is invalid", func() {
			It("errors and does not push the app", func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, PushCommandName, appName, "-b", "wut")
					Eventually(session.Err).Should(Say(`Buildpack "wut" must be an existing admin buildpack or a valid git URI`))
					Consistently(session).ShouldNot(Say("Creating app"))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				})
			})
		})

		When("the buildpack is valid", func() {
			When("the buildpack already exists", func() {
				When("pushing a single buildpack", func() {
					It("uses the specified buildpack", func() {
						helpers.WithHelloWorldApp(func(appDir string) {
							session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir},
								PushCommandName, appName,
								"-b", "staticfile_buildpack",
							)

							Eventually(session).Should(Say(`name:\s+%s`, appName))
							Eventually(session).Should(Say(`requested state:\s+started`))
							Eventually(session).Should(Say(`buildpacks:\s+staticfile`))
							Eventually(session).Should(Exit(0))
						})
					})
				})

				When("pushing a multi-buildpack app", func() {
					It("uses all the provided buildpacks in order", func() {
						helpers.WithMultiBuildpackApp(func(appDir string) {
							session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir},
								PushCommandName, appName,
								"-b", "ruby_buildpack",
								"-b", "go_buildpack",
							)

							Eventually(session).Should(Say("Ruby Buildpack"))
							Eventually(session).Should(Say("Go Buildpack"))

							Eventually(session).Should(Say(`name:\s+%s`, appName))
							Eventually(session).Should(Say(`requested state:\s+started`))
							Eventually(session).Should(Say(`buildpacks:\s+ruby.*go`))
							Eventually(session).Should(Exit(0))
						})
					})
				})
			})

			When("the buildpack is a URL", func() {
				It("uses the specified buildpack", func() {
					helpers.WithHelloWorldApp(func(appDir string) {
						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir},
							PushCommandName, appName,
							"-b", "https://github.com/cloudfoundry/staticfile-buildpack",
						)

						Eventually(session).Should(Say(`name:\s+%s`, appName))
						Eventually(session).Should(Say(`requested state:\s+started`))
						Eventually(session).Should(Say(`buildpacks:\s+staticfile`))
						Eventually(session).Should(Exit(0))
					})
				})
			})
		})
	})
})
