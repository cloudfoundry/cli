package push

import (
	"code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("HTTP random route", func() {
	var (
		appName string
	)

	BeforeEach(func() {
		appName = "short-app-name" // used on purpose to fit route length requirement
	})

	It("generates a random route for the app", func() {
		helpers.WithHelloWorldApp(func(dir string) {
			session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--random-route", "--no-start")
			Eventually(session).Should(Say("routes:"))
			Eventually(session).Should(Say("(?i)\\+\\s+%s-[\\w\\d]+-[\\w\\d]+.%s", appName, defaultSharedDomain()))
			Eventually(session).Should(Exit(0))
		})

		session := helpers.CF("app", appName)
		Eventually(session).Should(Say("name:\\s+%s", appName))
		Eventually(session).Should(Say("(?i)routes:\\s+%s-[\\w\\d]+-[\\w\\d]+.%s", appName, defaultSharedDomain()))
		Eventually(session).Should(Exit(0))
	})
})
