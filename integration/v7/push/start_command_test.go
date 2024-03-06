package push

import (
	"fmt"

	"code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo/v2"
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

	It("sets the start command correctly for the pushed app", func() {
		helpers.WithHelloWorldApp(func(dir string) {
			initialPush := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName)
			Eventually(initialPush).Should(Exit(0))

			session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir},
				PushCommandName, appName,
				"--start-command", fmt.Sprintf("%s && echo hello", helpers.ModernStaticfileBuildpackStartCommand))
			Eventually(session).Should(Say(`type:\s+web`))
			Eventually(session).Should(Say(`start command:\s+%s && echo hello`, helpers.ModernStaticfileBuildpackStartCommandRegex))
			Eventually(session).Should(Exit(0))

			By("not providing a custom start command again, it reuses the previous custom start command")
			session = helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName)
			Eventually(session).Should(Say(`type:\s+web`))
			Eventually(session).Should(Say(`start command:\s+%s && echo hello`, helpers.ModernStaticfileBuildpackStartCommandRegex))
			Eventually(session).Should(Exit(0))
		})
	})

	DescribeTable("resetting the start command",
		func(startCommand string) {
			helpers.WithHelloWorldApp(func(dir string) {
				initialPush := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir},
					PushCommandName, appName,
					"--start-command", fmt.Sprintf("%s && echo hello", helpers.ModernStaticfileBuildpackStartCommand))
				Eventually(initialPush).Should(Exit(0))

				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir},
					PushCommandName, appName,
					"--start-command", startCommand)
				Eventually(session).Should(Say(`type:\s+web`))
				Consistently(session).ShouldNot(Say(`start command:\s+%s && echo hello`, helpers.ModernStaticfileBuildpackStartCommandRegex))
				Eventually(session).Should(Say(`start command:\s+%s`, helpers.ModernStaticfileBuildpackStartCommandRegex))
				Eventually(session).Should(Exit(0))
			})
		},

		Entry("the start command is set to 'default'", "default"),
		Entry("the start command is set to 'null'", "null"),
	)
})
