package push

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("push with different buildpack values", func() {
	var (
		appName string
	)

	staticFileGitRepo := "https://github.com/cloudfoundry/staticfile-buildpack#v1.4.44"
	rubyBuildpackGitRepo := "https://github.com/cloudfoundry/ruby-buildpack#v1.7.44"

	BeforeEach(func() {
		appName = helpers.NewAppName()
	})

	When("the buildpack flag is provided", func() {
		When("only one buildpack is provided", func() {
			BeforeEach(func() {
				helpers.WithHelloWorldApp(func(dir string) {
					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir},
						PushCommandName, appName,
						"-b", "binary_buildpack",
						"--no-start",
					)

					Eventually(session).Should(Say(`(?m)\s+buildpacks:\s+\+\s+binary_buildpack`))
					Eventually(session).Should(Exit(0))
				})
			})

			It("pushing a staticfile app with a null buildpack sets buildpack to auto-detected (staticfile)", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir},
						PushCommandName, appName,
						"-b", "null",
					)
					Eventually(session).Should(Say(`(?m)\s+buildpacks:\s+-\s+binary_buildpack`))
					Eventually(session).Should(Say(`buildpacks?:\s+staticfile`))
					Eventually(session).Should(Exit(0))
				})
			})

			It("pushing a staticfile app with a default buildpack sets buildpack to auto-detected (staticfile)", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir},
						PushCommandName, appName,
						"-b", "default",
					)
					Eventually(session).Should(Say(`(?m)\s+buildpacks:\s+-\s+binary_buildpack`))
					Eventually(session).Should(Say(`buildpacks?:\s+staticfile`))
					Eventually(session).Should(Exit(0))
				})
			})
		})

		When("multiple instances of buildpack are provided", func() {
			When("the buildpacks do not use the default stack", func() {
				var (
					buildpacks      []string
					nonDefaultStack string
				)

				BeforeEach(func() {
					nonDefaultStack = helpers.CreateStack()
					buildpacks = []string{helpers.NewBuildpackName(), helpers.NewBuildpackName()}
					for _, buildpack := range buildpacks {
						helpers.SetupBuildpackWithStack(buildpack, nonDefaultStack)
					}
				})

				When("a stack is provided", func() {
					It("pushes the app successfully with multiple buildpacks using the stack specified", func() {
						helpers.WithProcfileApp(func(dir string) {
							tempfile := filepath.Join(dir, "index.html")
							err := ioutil.WriteFile(tempfile, []byte(fmt.Sprintf("hello world %d", rand.Int())), 0666)
							Expect(err).ToNot(HaveOccurred())

							session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir},
								PushCommandName, appName,
								"-b", buildpacks[0], "-b", buildpacks[1], "-s", nonDefaultStack, "--no-start",
							)
							Eventually(session).Should(Exit(0))
						})

						session := helpers.CF("curl", fmt.Sprintf("v3/apps/%s", helpers.AppGUID(appName)))

						Eventually(session).Should(Say(`\s+"buildpacks":\s+`))
						Eventually(session).Should(Say(`\s+"%s"`, buildpacks[0]))
						Eventually(session).Should(Say(`\s+"%s"`, buildpacks[1]))
						Eventually(session).Should(Say(`"stack":\s+"%s"`, nonDefaultStack))
						Eventually(session).Should(Exit(0))
					})
				})
			})

			When("the app does NOT have existing buildpack configurations", func() {
				It("pushes the app successfully with multiple buildpacks", func() {
					helpers.WithProcfileApp(func(dir string) {
						tempfile := filepath.Join(dir, "index.html")
						err := ioutil.WriteFile(tempfile, []byte(fmt.Sprintf("hello world %d", rand.Int())), 0666)
						Expect(err).ToNot(HaveOccurred())

						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir},
							PushCommandName, appName,
							"-b", "staticfile_buildpack", "-b", "ruby_buildpack", "--no-start",
						)
						Eventually(session).Should(Exit(0))
					})

					session := helpers.CF("curl", fmt.Sprintf("v3/apps/%s", helpers.AppGUID(appName)))

					Eventually(session).Should(Say(`\s+"buildpacks":\s+`))
					Eventually(session).Should(Say(`\s+"staticfile_buildpack"`))
					Eventually(session).Should(Say(`\s+"ruby_buildpack"`))
					Eventually(session).Should(Exit(0))
				})
			})

			When("the app has existing buildpacks", func() {
				It("pushes the app successfully and overrides the existing buildpacks", func() {
					helpers.WithHelloWorldApp(func(dir string) {
						helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
							"applications": []map[string]interface{}{
								{
									"name": appName,
									"buildpacks": []string{
										"ruby_buildpack",
										"staticfile_buildpack",
									},
								},
							},
						})
						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir},
							PushCommandName, appName,
							"-b", "php_buildpack", "-b", "go_buildpack", "--no-start",
						)
						Eventually(session).Should(Exit(0))
					})

					session := helpers.CF("curl", fmt.Sprintf("v3/apps/%s", helpers.AppGUID(appName)))

					Eventually(session).Should(Say(`\s+"buildpacks":\s+`))
					Eventually(session).Should(Say(`php_buildpack`))
					Eventually(session).Should(Say(`go_buildpack`))
					Eventually(session).Should(Exit(0))
				})
			})

			When("the app has existing `buildpack`", func() {
				It("pushes the app successfully and overrides the existing buildpacks", func() {
					helpers.WithHelloWorldApp(func(dir string) {
						helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
							"applications": []map[string]interface{}{
								{
									"name":      appName,
									"buildpack": "staticfile_buildpack",
								},
							},
						})
						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir},
							PushCommandName, appName,
							"-b", "php_buildpack", "-b", "go_buildpack", "--no-start",
						)
						Eventually(session).Should(Exit(0))
					})

					session := helpers.CF("curl", fmt.Sprintf("v3/apps/%s", helpers.AppGUID(appName)))
					Eventually(session).Should(Exit(0))

					Expect(session).Should(Say(`\s+"buildpacks":\s+`))
					Expect(session).ShouldNot(Say(`staticfile_buildpack`))
					Expect(session).Should(Say(`php_buildpack`))
					Expect(session).Should(Say(`go_buildpack`))
				})
			})

			When("one of the buildpacks provided is null or default", func() {
				It("fails and prints an error", func() {
					helpers.WithProcfileApp(func(dir string) {
						tempfile := filepath.Join(dir, "index.html")
						err := ioutil.WriteFile(tempfile, []byte(fmt.Sprintf("hello world %d", rand.Int())), 0666)
						Expect(err).ToNot(HaveOccurred())

						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir},
							PushCommandName, appName,
							"-b", "staticfile_buildpack", "-b", "null", "--no-start",
						)
						Eventually(session).Should(Exit(1))
						Eventually(session.Err).Should(Say("Multiple buildpacks flags cannot have null/default option."))
					})
				})
			})
		})
	})

	When("buildpack is provided via manifest", func() {
		It("sets buildpack and returns a warning", func() {
			helpers.WithHelloWorldApp(func(dir string) {
				helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
					"applications": []map[string]interface{}{
						{
							"name":      appName,
							"buildpack": "staticfile_buildpack",
						},
					},
				})
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "no-start")
				Eventually(session).Should(Say(`(?m)\s+buildpacks:\s+\+\s+staticfile_buildpack`))
				Eventually(session.Err).Should(Say(`Deprecation warning: Use of 'buildpack'`))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("buildpacks (plural) is provided via manifest", func() {
		When("mutiple buildpacks are specified", func() {
			It("sets all buildpacks correctly for the pushed app", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
						"applications": []map[string]interface{}{
							{
								"name": appName,
								"buildpacks": []string{
									rubyBuildpackGitRepo,
									staticFileGitRepo,
								},
							},
						},
					})
					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName)
					Eventually(session).Should(Exit(0))
				})

				session := helpers.CF("curl", fmt.Sprintf("v3/apps/%s", helpers.AppGUID(appName)))

				Eventually(session).Should(Say(`https://github.com/cloudfoundry/ruby-buildpack#v1.7.44`))
				Eventually(session).Should(Say(`https://github.com/cloudfoundry/staticfile-buildpack#v1.4.44`))
				Eventually(session).Should(Exit(0))
			})
		})

		When("only one buildpack is specified", func() {
			It("sets only one buildpack for the pushed app", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
						"applications": []map[string]interface{}{
							{
								"name": appName,
								"buildpacks": []string{
									staticFileGitRepo,
								},
							},
						},
					})
					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName)
					Eventually(session).Should(Exit(0))
				})

				session := helpers.CF("curl", fmt.Sprintf("v3/apps/%s", helpers.AppGUID(appName)))

				// TODO: fix during app command rework to actually test that the second buildpack does not exist
				Eventually(session).Should(Say(`https://github.com/cloudfoundry/staticfile-buildpack#v1.4.44`))
				Eventually(session).Should(Exit(0))
			})
		})

		When("empty list of buildpacks is specified", func() {
			It("autodetects the buildpack", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "-b", "staticfile_buildpack", "--no-start")
					Eventually(session).Should(Exit(0))

					helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
						"applications": []map[string]interface{}{
							{
								"name":       appName,
								"buildpacks": []string{},
							},
						},
					})
					session = helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName)
					Eventually(session).Should(Exit(0))
				})

				By("displaying an empty buildpacks field")
				session := helpers.CF("curl", fmt.Sprintf("v3/apps/%s", helpers.AppGUID(appName)))

				Eventually(session).Should(Say(`"buildpacks": \[\]`))
				Eventually(session).Should(Exit(0))
			})
		})

		When("an empty string is specified", func() {
			It("rasises an error", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
						"applications": []map[string]interface{}{
							{
								"name":       appName,
								"buildpacks": nil,
							},
						},
					})
					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName)
					Eventually(session).Should(Exit(1))
					Eventually(session.Err).Should(Say("Buildpacks property cannot be an empty string."))
				})
			})
		})
	})

	When("both buildpack and buildpacks are provided via manifest", func() {
		It("returns an error", func() {
			helpers.WithHelloWorldApp(func(dir string) {
				helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
					"applications": []map[string]interface{}{
						{
							"name":      appName,
							"buildpack": "ruby_buildpack",
							"buildpacks": []string{
								staticFileGitRepo,
							},
						},
					},
				})
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName)

				Eventually(session).Should(Exit(1))
				Eventually(session.Err).Should(Say("Application %s cannot use the combination of properties: buildpack, buildpacks", appName))
			})
		})
	})

	When("both buildpacks and docker are provided via manfest", func() {
		It("returns an error", func() {
			helpers.WithHelloWorldApp(func(dir string) {
				helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
					"applications": []map[string]interface{}{
						{
							"name": appName,
							"docker": map[string]interface{}{
								"image": PublicDockerImage,
							},
							"buildpacks": []string{
								staticFileGitRepo,
							},
						},
					},
				})
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName)

				Eventually(session).Should(Exit(1))
				Eventually(session.Err).Should(Say("Application %s cannot use the combination of properties: docker, buildpacks", appName))
			})
		})
	})

	When("both buildpacks and docker are provided via flags", func() {
		It("returns an error", func() {
			helpers.WithHelloWorldApp(func(dir string) {
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir},
					PushCommandName, appName, "-o", PublicDockerImage, "-b", "ruby_buildpack", "-b", "staticfile_buildpack",
				)

				Eventually(session).Should(Exit(1))
				Eventually(session.Err).Should(Say("Incorrect Usage: The following arguments cannot be used together: -b, --docker-image, -o"))
			})
		})
	})

	When("buildpack is provided via manifest and droplet is provided via flags", func() {
		var tempDroplet string

		BeforeEach(func() {
			f, err := ioutil.TempFile("", "INT-push-buildpack-droplet-")
			Expect(err).ToNot(HaveOccurred())
			Expect(f.Close()).ToNot(HaveOccurred())

			tempDroplet = f.Name()
		})

		AfterEach(func() {
			Expect(os.RemoveAll(tempDroplet)).ToNot(HaveOccurred())
		})

		It("returns an error", func() {
			helpers.WithHelloWorldApp(func(dir string) {
				helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
					"applications": []map[string]interface{}{
						{
							"name":      appName,
							"buildpack": staticFileGitRepo,
						},
					},
				})
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--droplet", tempDroplet)

				Eventually(session).Should(Exit(1))
				Eventually(session.Err).Should(Say("Application %s cannot use the combination of properties: droplet, buildpack", appName))
			})
		})
	})

	When("buildpacks is provided via manifest and droplet is provided via flags", func() {
		var tempDroplet string

		BeforeEach(func() {
			f, err := ioutil.TempFile("", "INT-push-buildpack-droplet-")
			Expect(err).ToNot(HaveOccurred())
			Expect(f.Close()).ToNot(HaveOccurred())

			tempDroplet = f.Name()
		})

		AfterEach(func() {
			Expect(os.RemoveAll(tempDroplet)).ToNot(HaveOccurred())
		})

		It("returns an error", func() {
			helpers.WithHelloWorldApp(func(dir string) {
				helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
					"applications": []map[string]interface{}{
						{
							"name": appName,
							"buildpacks": []string{
								staticFileGitRepo,
							},
						},
					},
				})
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--droplet", tempDroplet)

				Eventually(session).Should(Exit(1))
				Eventually(session.Err).Should(Say("Application %s cannot use the combination of properties: droplet, buildpacks", appName))
			})
		})
	})

	When("both buildpack and droplet are provided via flags", func() {
		var tempDroplet string

		BeforeEach(func() {
			f, err := ioutil.TempFile("", "INT-push-buildpack-droplet-")
			Expect(err).ToNot(HaveOccurred())
			Expect(f.Close()).ToNot(HaveOccurred())

			tempDroplet = f.Name()
		})

		AfterEach(func() {
			Expect(os.RemoveAll(tempDroplet)).ToNot(HaveOccurred())
		})

		It("returns an error", func() {
			helpers.WithHelloWorldApp(func(dir string) {
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir},
					PushCommandName, appName, "--droplet", tempDroplet, "-b", "staticfile_buildpack",
				)

				Eventually(session).Should(Exit(1))
				Eventually(session.Err).Should(Say("Application %s cannot use the combination of properties: droplet, buildpack", appName))
			})
		})
	})

	When("both buildpacks and droplet are provided via flags", func() {
		var tempDroplet string

		BeforeEach(func() {
			f, err := ioutil.TempFile("", "INT-push-buildpack-droplet-")
			Expect(err).ToNot(HaveOccurred())
			Expect(f.Close()).ToNot(HaveOccurred())

			tempDroplet = f.Name()
		})

		AfterEach(func() {
			Expect(os.RemoveAll(tempDroplet)).ToNot(HaveOccurred())
		})

		It("returns an error", func() {
			helpers.WithHelloWorldApp(func(dir string) {
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir},
					PushCommandName, appName, "--droplet", tempDroplet, "-b", "ruby_buildpack", "-b", "staticfile_buildpack",
				)

				Eventually(session).Should(Exit(1))
				Eventually(session.Err).Should(Say("Application %s cannot use the combination of properties: droplet, buildpacks", appName))
			})
		})
	})
})
