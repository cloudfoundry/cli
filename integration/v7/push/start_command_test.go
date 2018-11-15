package push

import (
	"fmt"

	"code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("push with different start command values", func() {
	var (
		appName string
	)

	BeforeEach(func() {
		appName = helpers.NewAppName()
	})

	When("the start command flag is provided", func() {
		It("sets the start command correctly for the pushed app", func() {
			helpers.WithHelloWorldApp(func(dir string) {
				By("pushing the app with no provided start command uses detected command")
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName)
				Eventually(session).Should(Say(`type:\s+web`))
				Eventually(session).Should(Say(`start command:\s+%s`, helpers.StaticfileBuildpackStartCommand))
				Eventually(session).Should(Exit(0))

				By("pushing the app with a start command uses provided start command")
				session = helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir},
					PushCommandName, appName,
					"--start-command", fmt.Sprintf("%s && echo hello", helpers.StaticfileBuildpackStartCommand))
				Eventually(session).Should(Say(`type:\s+web`))
				Eventually(session).Should(Say(`start command:\s+%s && echo hello`, helpers.StaticfileBuildpackStartCommand))
				Eventually(session).Should(Exit(0))

				By("pushing the app with no provided start command again uses previously set command")
				session = helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName)
				Eventually(session).Should(Say(`type:\s+web`))
				Eventually(session).Should(Say(`start command:\s+%s && echo hello`, helpers.StaticfileBuildpackStartCommand))
				Eventually(session).Should(Exit(0))

				// By("pushing the app with default uses detected command")
				// session = helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir},
				// 	PushCommandName, appName,
				// 	"--start-command", "default")
				// Eventually(session).Should(Say(`type:\s+web`))
				// Consistently(session).ShouldNot(Say(`start command:\s+%s && echo hello`, helpers.StaticfileBuildpackStartCommand))
				// Eventually(session).Should(Say(`(?m)start command:\s+%s\n`, helpers.StaticfileBuildpackStartCommand))
				// Eventually(session).Should(Exit(0))

				// By("pushing the app with null uses detected command")
				// session = helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir},
				// 	PushCommandName, appName,
				// 	"--start-command", "null")
				// Eventually(session).Should(Say(`type:\s+web`))
				// Consistently(session).ShouldNot(Say(`start command:\s+%s && echo hello`, helpers.StaticfileBuildpackStartCommand))
				// Eventually(session).Should(Say(`(?m)start command:\s+%s\n`, helpers.StaticfileBuildpackStartCommand))
				// Eventually(session).Should(Exit(0))
			})
		})
	})
})
