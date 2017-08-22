package push

import (
	"code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("push with different instances values", func() {
	var (
		appName string
	)

	BeforeEach(func() {
		appName = helpers.NewAppName()
	})

	Context("when instances flag is provided", func() {
		Context("when instances flag is greater than 0", func() {
			It("pushes an app with specified number of instances", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					By("pushing an app with 2 instances")
					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir},
						PushCommandName, appName,
						"-i", "2",
					)
					Eventually(session).Should(Say("\\s+instances:\\s+2"))
					Eventually(session).Should(Exit(0))

					session = helpers.CF("app", appName)
					Eventually(session).Should(Say("instances:\\s+\\d/2"))
					Eventually(session).Should(Exit(0))

					By("updating an app with 1 instance")
					session = helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir},
						PushCommandName, appName,
						"-i", "1",
					)
					Eventually(session).Should(Say("\\-\\s+instances:\\s+2"))
					Eventually(session).Should(Say("\\+\\s+instances:\\s+1"))
					Eventually(session).Should(Exit(0))

					session = helpers.CF("app", appName)
					Eventually(session).Should(Say("instances:\\s+\\d/1"))
					Eventually(session).Should(Exit(0))
				})
			})
		})

		Context("when instances flag is set to 0", func() {
			It("pushes an app with 0 instances", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					By("pushing an app with 0 instances")
					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir},
						PushCommandName, appName,
						"-i", "0",
					)
					Eventually(session).Should(Say("\\s+instances:\\s+0"))
					Eventually(session).Should(Exit(0))

					session = helpers.CF("app", appName)
					Eventually(session).Should(Say("instances:\\s+\\d/0"))
					Eventually(session).Should(Exit(0))

					By("updating an app to 1 instance")
					session = helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir},
						PushCommandName, appName,
						"-i", "1",
					)
					Eventually(session).Should(Say("\\-\\s+instances:\\s+0"))
					Eventually(session).Should(Say("\\+\\s+instances:\\s+1"))
					Eventually(session).Should(Exit(0))

					session = helpers.CF("app", appName)
					Eventually(session).Should(Say("instances:\\s+\\d/1"))
					Eventually(session).Should(Exit(0))

					By("updating an app back to 0 instances")
					session = helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir},
						PushCommandName, appName,
						"-i", "0",
					)
					Eventually(session).Should(Say("\\-\\s+instances:\\s+1"))
					Eventually(session).Should(Say("\\+\\s+instances:\\s+0"))
					Eventually(session).Should(Exit(0))

					session = helpers.CF("app", appName)
					Eventually(session).Should(Say("instances:\\s+\\d/0"))
					Eventually(session).Should(Exit(0))
				})
			})
		})
	})

	Context("when instances flag is not provided", func() {
		Context("when app does not exist", func() {
			It("pushes an app with default number of instances", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir},
						PushCommandName, appName,
					)
					Eventually(session).Should(Say("\\s+instances:\\s+1"))
					Eventually(session).Should(Exit(0))
				})

				session := helpers.CF("app", appName)
				Eventually(session).Should(Say("instances:\\s+\\d/1"))
				Eventually(session).Should(Exit(0))
			})
		})

		Context("when app exists with some instances", func() {
			It("does not update the number of instances", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					By("pushing an app with 2 instances")
					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir},
						PushCommandName, appName,
						"-i", "2",
					)
					Eventually(session).Should(Say("\\s+instances:\\s+2"))
					Eventually(session).Should(Exit(0))

					session = helpers.CF("app", appName)
					Eventually(session).Should(Say("instances:\\s+\\d/2"))
					Eventually(session).Should(Exit(0))

					By("pushing an app with no instances specified")
					session = helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir},
						PushCommandName, appName,
					)
					Eventually(session).Should(Say("\\s+instances:\\s+2"))
					Eventually(session).Should(Exit(0))
				})
			})
		})
	})
})
