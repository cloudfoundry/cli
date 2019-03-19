package push

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	"io/ioutil"
	"path/filepath"
)

var _ = Describe("pushing docker images", func() {
	var (
		appName string
		tempDir string
	)

	BeforeEach(func() {
		appName = helpers.PrefixedRandomName("app")
		var err error
		tempDir, err = ioutil.TempDir("", "simple-manifest-test")
		Expect(err).ToNot(HaveOccurred())
	})

	var AssertThatItPrintsSuccessfulDockerPushOutput = func(session *Session, dockerImage string) {
		Eventually(session).Should(Say(`name:\s+%s`, appName))
		Eventually(session).Should(Say(`requested state:\s+started`))
		Eventually(session).Should(Say("stack:"))
		Consistently(session).ShouldNot(Say("buildpacks:"))
		Eventually(session).Should(Say(`docker image:\s+%s`, dockerImage))
		Eventually(session).Should(Exit(0))
	}

	When("a public docker image is specified in command line", func() {
		When("the docker image is valid", func() {
			It("uses the specified docker image", func() {
				session := helpers.CF(PushCommandName, appName, "-o", PublicDockerImage)

				AssertThatItPrintsSuccessfulDockerPushOutput(session, PublicDockerImage)
			})
		})

		When("the docker image is invalid", func() {
			It("displays an error and exits 1", func() {
				session := helpers.CF(PushCommandName, appName, "-o", "some-invalid-docker-image")
				Eventually(session.Err).Should(Say("StagingError - Staging error: staging failed"))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})
	})

	When("a public docker image is only provided in the manifest", func() {
		When("the docker image is valid", func() {
			It("uses the specified docker image", func() {
				manifestPath := filepath.Join(tempDir, "manifest.yml")

				helpers.WriteManifest(manifestPath, map[string]interface{}{
					"applications": []map[string]interface{}{
						{
							"name": appName,
							"docker": map[string]interface{}{
								"image": PublicDockerImage,
							},
						},
					},
				})
				session := helpers.CF(
					PushCommandName, appName,
					"-f", manifestPath,
				)

				AssertThatItPrintsSuccessfulDockerPushOutput(session, PublicDockerImage)
			})
		})

		When("the docker image is invalid", func() {
			It("displays an error and exits 1", func() {
				manifestPath := filepath.Join(tempDir, "manifest.yml")

				helpers.WriteManifest(manifestPath, map[string]interface{}{
					"applications": []map[string]interface{}{
						{
							"name": appName,
							"docker": map[string]interface{}{
								"image": "some-invalid-docker-image",
							},
						},
					},
				})
				session := helpers.CF(
					PushCommandName, appName,
					"-f", manifestPath,
				)

				Eventually(session.Err).Should(Say("StagingError - Staging error: staging failed"))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})
	})
	When("a docker username and password are provided with a private image", func() {
		var (
			privateDockerImage    string
			privateDockerUsername string
			privateDockerPassword string
		)

		BeforeEach(func() {
			privateDockerImage, privateDockerUsername, privateDockerPassword = helpers.SkipIfPrivateDockerInfoNotSet()
		})

		When("the docker password is provided via environment variable", func() {
			It("uses the specified private docker image", func() {
				session := helpers.CustomCF(
					helpers.CFEnv{
						EnvVars: map[string]string{"CF_DOCKER_PASSWORD": privateDockerPassword},
					},
					PushCommandName, appName,
					"--docker-username", privateDockerUsername,
					"--docker-image", privateDockerImage,
				)

				Eventually(session).Should(Say("Using docker repository password from environment variable CF_DOCKER_PASSWORD."))
				Consistently(session).ShouldNot(Say("Docker password"))

				AssertThatItPrintsSuccessfulDockerPushOutput(session, privateDockerImage)
			})
		})

		When("the docker password is not provided", func() {
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

				Eventually(session).Should(Say("Environment variable CF_DOCKER_PASSWORD not set."))
				Eventually(session).Should(Say("Docker password"))

				AssertThatItPrintsSuccessfulDockerPushOutput(session, privateDockerImage)
			})
		})
	})

	When("a docker username and private image are only provided in the manifest", func() {
		var (
			privateDockerImage    string
			privateDockerUsername string
			privateDockerPassword string
		)

		BeforeEach(func() {
			privateDockerImage, privateDockerUsername, privateDockerPassword = helpers.SkipIfPrivateDockerInfoNotSet()
		})

		When("the docker password is provided via environment variable", func() {
			It("uses the specified private docker image", func() {
				manifestPath := filepath.Join(tempDir, "manifest.yml")

				helpers.WriteManifest(manifestPath, map[string]interface{}{
					"applications": []map[string]interface{}{
						{
							"name": appName,
							"docker": map[string]interface{}{
								"image":    privateDockerImage,
								"username": privateDockerUsername,
							},
						},
					},
				})
				session := helpers.CustomCF(
					helpers.CFEnv{
						EnvVars: map[string]string{"CF_DOCKER_PASSWORD": privateDockerPassword},
					},
					PushCommandName, appName,
					"-f", manifestPath,
				)

				Eventually(session).Should(Say("Using docker repository password from environment variable CF_DOCKER_PASSWORD."))
				Consistently(session).ShouldNot(Say("Docker password"))

				AssertThatItPrintsSuccessfulDockerPushOutput(session, privateDockerImage)
			})
		})

		When("the docker password is not provided", func() {
			var buffer *Buffer

			BeforeEach(func() {
				buffer = NewBuffer()
				_, err := buffer.Write([]byte(privateDockerPassword + "\n"))
				Expect(err).NotTo(HaveOccurred())
			})

			It("prompts for the docker password", func() {
				manifestPath := filepath.Join(tempDir, "manifest.yml")
				helpers.WriteManifest(manifestPath, map[string]interface{}{
					"applications": []map[string]interface{}{
						{
							"name": appName,
							"docker": map[string]interface{}{
								"image":    privateDockerImage,
								"username": privateDockerUsername,
							},
						},
					},
				})
				session := helpers.CFWithStdin(buffer,
					PushCommandName,
					appName,
					"-f", manifestPath,
				)

				Eventually(session).Should(Say("Environment variable CF_DOCKER_PASSWORD not set."))
				Eventually(session).Should(Say("Docker password"))

				AssertThatItPrintsSuccessfulDockerPushOutput(session, privateDockerImage)
			})
		})
	})

})
