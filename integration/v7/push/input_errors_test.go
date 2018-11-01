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

var _ = Describe("input errors", func() {
	var (
		appName string
	)

	BeforeEach(func() {
		appName = helpers.PrefixedRandomName("app")
	})

	When("the app name is not provided", func() {
		It("tells the user that the app name is required, prints help text, and exits 1", func() {
			session := helpers.CF(PushCommandName)

			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `APP_NAME` was not provided"))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	When("the -p flag path does not exist", func() {
		It("tells the user that the flag requires an arg, prints help text, and exits 1", func() {
			session := helpers.CF(PushCommandName, appName, "-p", "path/that/does/not/exist")

			Eventually(session.Err).Should(Say("Incorrect Usage: The specified path 'path/that/does/not/exist' does not exist."))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	Describe("argument combination errors", func() {
		When("the --docker-username is provided without the -o flag", func() {
			It("displays an error and exits 1", func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					session := helpers.CF(PushCommandName, appName, "--docker-username", "some-username")
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("Incorrect Usage: '--docker-image, -o' and '--docker-username' must be used together."))
					Eventually(session).Should(Say("NAME:"))
					Eventually(session).Should(Exit(1))
				})
			})
		})

		When("the --docker-username and -p flags are provided together", func() {
			It("displays an error and exits 1", func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					session := helpers.CF(PushCommandName, appName, "--docker-username", "some-username", "-p", appDir)
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("Incorrect Usage: '--docker-image, -o' and '--docker-username' must be used together."))
					Eventually(session).Should(Say("NAME:"))
					Eventually(session).Should(Exit(1))
				})
			})
		})

		When("the --docker-username is provided without a password", func() {
			var oldPassword string

			BeforeEach(func() {
				oldPassword = os.Getenv("CF_DOCKER_PASSWORD")
				err := os.Unsetenv("CF_DOCKER_PASSWORD")
				Expect(err).ToNot(HaveOccurred())
			})

			AfterEach(func() {
				err := os.Setenv("CF_DOCKER_PASSWORD", oldPassword)
				Expect(err).ToNot(HaveOccurred())
			})

			It("displays an error and exits 1", func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					session := helpers.CF(PushCommandName, appName, "--docker-username", "some-username", "--docker-image", "some-image")
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say(`Environment variable CF_DOCKER_PASSWORD not set\.`))
					Eventually(session).Should(Exit(1))
				})
			})
		})

		When("the -o and -p flags are provided together", func() {
			It("displays an error and exits 1", func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					session := helpers.CF(PushCommandName, appName, "-o", PublicDockerImage, "-p", appDir)
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("Incorrect Usage: The following arguments cannot be used together: --docker-image, -o, -p"))
					Eventually(session).Should(Say("NAME:"))
					Eventually(session).Should(Exit(1))
				})
			})
		})

		When("the -o and -b flags are provided together", func() {
			It("displays an error and exits 1", func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					session := helpers.CF(PushCommandName, appName, "-o", PublicDockerImage, "-b", "some-buildpack")
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("Incorrect Usage: The following arguments cannot be used together: -b, --docker-image, -o"))
					Eventually(session).Should(Say("NAME:"))
					Eventually(session).Should(Exit(1))
				})
			})
		})
	})
})
