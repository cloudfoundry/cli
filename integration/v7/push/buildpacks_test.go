package push

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
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

			When("resetting the buildpack to autodetection", func() {
				It("successfully pushes the app with -b default", func() {
					helpers.WithHelloWorldApp(func(appDir string) {
						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir},
							PushCommandName, appName,
							"-b", "default",
						)

						Eventually(session).Should(Exit(0))

						Expect(session).To(Say(`name:\s+%s`, appName))
						Expect(session).To(Say(`requested state:\s+started`))
						Expect(session).To(Say("buildpacks:"))
						Expect(session).To(Say(`staticfile_buildpack\s+\d+.\d+.\d+`))
					})
				})

				It("successfully pushes the app with -b null", func() {
					helpers.WithHelloWorldApp(func(appDir string) {
						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir},
							PushCommandName, appName,
							"-b", "null",
						)

						Eventually(session).Should(Exit(0))

						Expect(session).To(Say(`name:\s+%s`, appName))
						Expect(session).To(Say(`requested state:\s+started`))
						Expect(session).To(Say("buildpacks:"))
						Expect(session).To(Say(`staticfile_buildpack\s+\d+.\d+.\d+`))
					})
				})
			})

			When("omitting the buildpack", func() {
				It("continues using previously set buildpack", func() {
					helpers.WithHelloWorldApp(func(appDir string) {
						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, PushCommandName, appName)

						Eventually(session).Should(Exit(1))

						Expect(session).To(Say("FAILED"))
					})
				})
			})
		})

		When("the buildpack is invalid", func() {
			It("errors and does not push the app", func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, PushCommandName, appName, "-b", "wut")

					Eventually(session).Should(Exit(1))

					Expect(session.Err).To(Say(`For application '%s': Specified unknown buildpack name: "wut"`, appName))
					Expect(session).To(Say("FAILED"))
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

							Eventually(session).Should(Exit(0))

							Expect(session).To(Say(`name:\s+%s`, appName))
							Expect(session).To(Say(`requested state:\s+started`))
							Expect(session).To(Say("buildpacks:"))
							Expect(session).To(Say(`staticfile_buildpack\s+\d+.\d+.\d+`))
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

							Eventually(session).Should(Exit(0))

							Expect(session).To(Say("Ruby Buildpack"))
							Expect(session).To(Say("Go Buildpack"))
							Expect(session).To(Say(`name:\s+%s`, appName))
							Expect(session).To(Say(`requested state:\s+started`))
							Expect(session).To(Say("buildpacks:"))
							Expect(session).To(Say(`ruby_buildpack\s+\d+.\d+.\d+`))
							Expect(session).To(Say(`go_buildpack\s+\d+.\d+.\d+`))
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

						Eventually(session).Should(Exit(0))

						Expect(session).To(Say(`name:\s+%s`, appName))
						Expect(session).To(Say(`requested state:\s+started`))
						Expect(session).To(Say("buildpacks:"))
						Expect(session).To(Say(`https://github.com/cloudfoundry/staticfile-buildpack\s+\d+.\d+.\d+`))
					})
				})
			})
		})
	})
})
