// +build !partialPush

package push

import (
	"os"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("pushing docker images", func() {
	var (
		appName string
	)

	BeforeEach(func() {
		appName = helpers.PrefixedRandomName("app")
	})

	When("the docker image is valid", func() {
		It("uses the specified docker image", func() {
			session := helpers.CF(PushCommandName, appName, "-o", PublicDockerImage)

			Eventually(session).Should(Say(`name:\s+%s`, appName))
			Eventually(session).Should(Say(`requested state:\s+started`))
			Eventually(session).Should(Say("stack:"))
			Consistently(session).ShouldNot(Say("buildpacks:"))
			Eventually(session).Should(Say(`docker image:\s+%s`, PublicDockerImage))
			Eventually(session).Should(Exit(0))
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

	When("a docker username and password are provided with a private image", func() {
		var (
			privateDockerImage    string
			privateDockerUsername string
			privateDockerPassword string
		)

		BeforeEach(func() {
			privateDockerImage = os.Getenv("CF_INT_DOCKER_IMAGE")
			privateDockerUsername = os.Getenv("CF_INT_DOCKER_USERNAME")
			privateDockerPassword = os.Getenv("CF_INT_DOCKER_PASSWORD")

			if privateDockerImage == "" || privateDockerUsername == "" || privateDockerPassword == "" {
				Skip("CF_INT_DOCKER_IMAGE, CF_INT_DOCKER_USERNAME, or CF_INT_DOCKER_PASSWORD is not set")
			}
		})

		When("the docker passwored is provided via environment variable", func() {
			It("uses the specified private docker image", func() {
				session := helpers.CustomCF(
					helpers.CFEnv{
						EnvVars: map[string]string{"CF_DOCKER_PASSWORD": privateDockerPassword},
					},
					PushCommandName, appName,
					"--docker-username", privateDockerUsername,
					"--docker-image", privateDockerImage,
				)

				Eventually(session).Should(Say(`name:\s+%s`, appName))
				Eventually(session).Should(Say(`requested state:\s+started`))
				Eventually(session).Should(Say("stack:"))
				Consistently(session).ShouldNot(Say("buildpacks:"))
				Eventually(session).Should(Say(`docker image:\s+%s`, privateDockerImage))
				Eventually(session).Should(Exit(0))
			})
		})

		When("the docker passwored is not provided", func() {
			It("returns an error", func() {
				session := helpers.CF(
					PushCommandName, appName,
					"--docker-username", privateDockerUsername,
					"--docker-image", privateDockerImage,
				)

				Eventually(session.Err).Should(Say("Environment variable CF_DOCKER_PASSWORD not set."))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})
	})
})
