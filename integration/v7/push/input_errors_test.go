package push

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
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

		When("the -o and -p flags are provided together", func() {
			It("displays an error and exits 1", func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					session := helpers.CF(PushCommandName, appName, "-o", PublicDockerImage, "-p", appDir)
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("Incorrect Usage: The following arguments cannot be used together: --docker-image, -o, --path, -p"))
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
					Eventually(session.Err).Should(Say("Incorrect Usage: The following arguments cannot be used together: --buildpack, -b, --docker-image, -o"))
					Eventually(session).Should(Say("NAME:"))
					Eventually(session).Should(Exit(1))
				})
			})
		})
	})
})
