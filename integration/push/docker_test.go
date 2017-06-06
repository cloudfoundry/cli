package push

import (
	"code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("pushing a docker image", func() {
	var (
		appName string
	)

	BeforeEach(func() {
		appName = helpers.NewAppName()
	})

	Describe("app existence", func() {
		Context("when the app does not exist", func() {
			It("creates the app", func() {
				session := helpers.CF(PushCommandName, appName, "-o", DockerImage)
				Eventually(session).Should(Say("Getting app info\\.\\.\\."))
				Eventually(session).Should(Say("Creating app with these attributes\\.\\.\\."))
				Eventually(session).Should(Say("\\+\\s+name:\\s+%s", appName))
				Eventually(session).Should(Say("\\s+docker image:\\s+%s", DockerImage))
				Eventually(session).Should(Say("\\s+routes:"))
				Eventually(session).Should(Say("(?i)\\+\\s+%s.%s", appName, defaultSharedDomain()))
				Eventually(session).Should(Say("Mapping routes\\.\\.\\."))
				Eventually(session).Should(Say("Staging Complete"))
				Eventually(session).Should(Say("Waiting for app to start\\.\\.\\."))
				Eventually(session).Should(Say("requested state:\\s+started"))
				Eventually(session).Should(Exit(0))

				session = helpers.CF("app", appName)
				Eventually(session).Should(Say("name:\\s+%s", appName))
				Eventually(session).Should(Exit(0))
			})
		})

		Context("when the app exists", func() {
			BeforeEach(func() {
				Eventually(helpers.CF(PushCommandName, appName, "-o", DockerImage)).Should(Exit(0))
			})

			It("updates the app", func() {
				session := helpers.CF(PushCommandName, appName, "-o", DockerImage)
				Eventually(session).Should(Say("Getting app info\\.\\.\\."))
				Eventually(session).Should(Say("Updating app with these attributes\\.\\.\\."))
				Eventually(session).Should(Say("\\s+name:\\s+%s", appName))
				Eventually(session).Should(Say("\\s+docker image:\\s+%s", DockerImage))
				Eventually(session).Should(Say("\\s+routes:"))
				Eventually(session).Should(Say("(?i)\\s+%s.%s", appName, defaultSharedDomain()))
				Eventually(session).Should(Say("Mapping routes\\.\\.\\."))
				Eventually(session).Should(Say("Waiting for app to start\\.\\.\\."))
				Eventually(session).Should(Say("requested state:\\s+started"))
				Eventually(session).Should(Exit(0))

				session = helpers.CF("app", appName)
				Eventually(session).Should(Say("name:\\s+%s", appName))
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
