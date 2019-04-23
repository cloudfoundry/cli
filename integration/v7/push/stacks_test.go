package push

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("stack", func() {
	var (
		appName string
	)

	BeforeEach(func() {
		appName = helpers.PrefixedRandomName("app")
	})

	When("the --stack flag is provided", func() {
		It("successfully pushes the app with the provided stack", func() {
			helpers.WithHelloWorldApp(func(appDir string) {
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir},
					PushCommandName, appName,
					"--stack", WindowsStack,
					"--no-start",
				)
				Eventually(session).Should(Exit(0))
				Expect(helpers.AppJSON(appName)).To(MatchRegexp(`"stack":\s*"%s"`, WindowsStack))
			})
		})

		It("fails to push the app with an invalid stack", func() {
			helpers.WithHelloWorldApp(func(appDir string) {
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir},
					PushCommandName, appName,
					"--stack", "invalidStack",
				)

				Eventually(session.Err).Should(Say(`Stack must be an existing stack`))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})
	})
})
