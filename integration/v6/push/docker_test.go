package push

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("pushing a docker image", func() {
	var (
		appName               string
		privateDockerImage    string
		privateDockerUsername string
		privateDockerPassword string
	)

	BeforeEach(func() {
		helpers.SkipIfClientCredentialsTestMode()
		appName = helpers.NewAppName()
	})

	Describe("when the docker image is provided via command line", func() {
		Describe("a public docker image", func() {
			Describe("app existence", func() {
				When("the app does not exist", func() {
					It("creates the app", func() {
						session := helpers.CF(PushCommandName, appName, "-o", PublicDockerImage)
						validateDockerPush(session, appName, PublicDockerImage)
					})
				})

				When("the app exists", func() {
					BeforeEach(func() {
						Eventually(helpers.CF(PushCommandName, appName, "-o", PublicDockerImage)).Should(Exit(0))
					})

					It("updates the app", func() {
						session := helpers.CF(PushCommandName, appName, "-o", PublicDockerImage)
						Eventually(session).Should(Say(`Updating app with these attributes\.\.\.`))
						Eventually(session).Should(Exit(0))
					})
				})
			})
		})

		Describe("private docker image with a username", func() {
			BeforeEach(func() {
				privateDockerImage, privateDockerUsername, privateDockerPassword = helpers.SkipIfPrivateDockerInfoNotSet()
			})

			When("CF_DOCKER_PASSWORD is set", func() {
				It("push the docker image with those credentials", func() {
					session := helpers.CustomCF(
						helpers.CFEnv{
							EnvVars: map[string]string{"CF_DOCKER_PASSWORD": privateDockerPassword},
						},
						PushCommandName,
						appName,
						"--docker-username", privateDockerUsername,
						"--docker-image", privateDockerImage,
					)
					validateDockerPush(session, appName, privateDockerImage)
				})
			})

			When("the CF_DOCKER_PASSWORD is not set", func() {
				var buffer *Buffer

				BeforeEach(func() {
					buffer = NewBuffer()
					_, err := buffer.Write([]byte(privateDockerPassword + "\n"))
					Expect(err).NotTo(HaveOccurred())
				})

				It("prompts for the docker password", func() {
					session := helpers.CFWithStdin(buffer,
						PushCommandName,
						appName,
						"--docker-username", privateDockerUsername,
						"--docker-image", privateDockerImage,
					)

					validateDockerPassword(session, true)
					validateDockerPush(session, appName, privateDockerImage)
				})
			})
		})
	})

	Describe("docker image in the manifest is provided", func() {
		var appManifest map[string]interface{}

		BeforeEach(func() {
			appManifest = map[string]interface{}{
				"applications": []map[string]interface{}{
					{
						"name": appName,
						"docker": map[string]string{
							"image": PublicDockerImage,
						},
					},
				},
			}
		})

		It("uses the docker image when pushing", func() {
			withManifest(appManifest, func(manifestDir string) {
				session := helpers.CustomCF(
					helpers.CFEnv{WorkingDirectory: manifestDir},
					PushCommandName,
				)

				validateDockerPush(session, appName, PublicDockerImage)
			})
		})

		When("buildpack is set", func() {
			BeforeEach(func() {
				appManifest = map[string]interface{}{
					"applications": []map[string]interface{}{
						{
							"name":      appName,
							"buildpack": "some-buildpack",
							"docker": map[string]string{
								"image": PublicDockerImage,
							},
						},
					},
				}
			})

			It("returns an error", func() {
				withManifest(appManifest, func(manifestDir string) {
					session := helpers.CustomCF(
						helpers.CFEnv{WorkingDirectory: manifestDir},
						PushCommandName,
					)

					Eventually(session.Err).Should(Say("Application %s cannot use the combination of properties: docker, buildpack", appName))
					Eventually(session).Should(Exit(1))
				})
			})
		})

		When("path is set", func() {
			BeforeEach(func() {
				appManifest = map[string]interface{}{
					"applications": []map[string]interface{}{
						{
							"name": appName,
							"docker": map[string]string{
								"image": PublicDockerImage,
							},
							"path": "some-path",
						},
					},
				}
			})

			It("returns an error", func() {
				withManifest(appManifest, func(manifestDir string) {
					session := helpers.CustomCF(
						helpers.CFEnv{WorkingDirectory: manifestDir},
						PushCommandName,
					)

					Eventually(session.Err).Should(Say("Application %s cannot use the combination of properties: docker, path", appName))
					Eventually(session).Should(Exit(1))
				})
			})
		})

		When("user is provided in the manifest", func() {
			BeforeEach(func() {
				privateDockerImage, privateDockerUsername, privateDockerPassword = helpers.SkipIfPrivateDockerInfoNotSet()

				appManifest = map[string]interface{}{
					"applications": []map[string]interface{}{
						{
							"name": appName,
							"docker": map[string]string{
								"image":    privateDockerImage,
								"username": privateDockerUsername,
							},
						},
					},
				}
			})

			When("password is provided in the enviornment", func() {
				It("uses the docker image and credentials when pushing", func() {
					withManifest(appManifest, func(manifestDir string) {
						session := helpers.CustomCF(
							helpers.CFEnv{
								WorkingDirectory: manifestDir,
								EnvVars:          map[string]string{"CF_DOCKER_PASSWORD": privateDockerPassword},
							},
							PushCommandName,
						)

						validateDockerPassword(session, false)
						validateDockerPush(session, appName, privateDockerImage)
					})
				})
			})

			When("password is not provided in the enviornment", func() {
				It("errors out", func() {
					withManifest(appManifest, func(manifestDir string) {
						session := helpers.CustomCF(
							helpers.CFEnv{WorkingDirectory: manifestDir},
							PushCommandName,
						)

						Eventually(session.Err).Should(Say(`Environment variable CF_DOCKER_PASSWORD not set\.`))
						Eventually(session).Should(Exit(1))
					})
				})
			})
		})
	})

	Describe("command line and manifest interaction", func() {
		var appManifest map[string]interface{}

		When("the image and username are provided by both manifest and command line", func() {
			BeforeEach(func() {
				privateDockerImage, privateDockerUsername, privateDockerPassword = helpers.SkipIfPrivateDockerInfoNotSet()

				appManifest = map[string]interface{}{
					"applications": []map[string]interface{}{
						{
							"name": appName,
							"docker": map[string]string{
								"image":    "junk",
								"username": "junk",
							},
						},
					},
				}
			})

			It("command line takes precidence", func() {
				withManifest(appManifest, func(manifestDir string) {
					session := helpers.CustomCF(
						helpers.CFEnv{
							WorkingDirectory: manifestDir,
							EnvVars:          map[string]string{"CF_DOCKER_PASSWORD": privateDockerPassword},
						},
						PushCommandName,
						"--docker-username", privateDockerUsername,
						"--docker-image", privateDockerImage,
					)

					validateDockerPassword(session, false)
					validateDockerPush(session, appName, privateDockerImage)
				})
			})
		})
	})
})

func validateDockerPassword(session *Session, passwordFromPrompt bool) {
	if passwordFromPrompt {
		Eventually(session).Should(Say("Environment variable CF_DOCKER_PASSWORD not set."))
		Eventually(session).Should(Say("Docker password"))
	} else {
		Eventually(session).Should(Say("Using docker repository password from environment variable CF_DOCKER_PASSWORD."))
	}
}

func validateDockerPush(session *Session, appName string, dockerImage string) {
	Eventually(session).Should(Say(`Getting app info\.\.\.`))
	Eventually(session).Should(Say(`Creating app with these attributes\.\.\.`))
	Eventually(session).Should(Say(`\+\s+name:\s+%s`, appName))
	Eventually(session).Should(Say(`\s+docker image:\s+%s`, dockerImage))
	helpers.ConfirmStagingLogs(session)
	Eventually(session).Should(Say(`Waiting for app to start\.\.\.`))
	Eventually(session).Should(Say(`requested state:\s+started`))
	Eventually(session).Should(Exit(0))

	session = helpers.CF("app", appName)
	Eventually(session).Should(Say(`name:\s+%s`, appName))
	Eventually(session).Should(Exit(0))
}

func withManifest(manifest map[string]interface{}, f func(manifestDir string)) {
	dir, err := ioutil.TempDir("", "simple-app")
	Expect(err).ToNot(HaveOccurred())
	defer os.RemoveAll(dir)

	helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), manifest)
	f(dir)
}
