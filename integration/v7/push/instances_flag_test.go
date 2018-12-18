package push

import (
	"time"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("push with instances flag", func() {
	var (
		appName string
	)

	BeforeEach(func() {
		appName = helpers.NewAppName()
	})

	Context("when the -i flag is provided with an instance count", func() {
		It("creates the app with the specified number of web instances", func() {
			helpers.WithHelloWorldApp(func(dir string) {
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir},
					PushCommandName, appName,
					"-i", "3",
				)
				Eventually(session).Should(Exit(0))
			})

			time.Sleep(5 * time.Second)
			session := helpers.CF("app", appName)
			Eventually(session).Should(Say(`name:\s+%s`, appName))
			Eventually(session).Should(Say(`last uploaded:\s+\w{3} \d{1,2} \w{3} \d{2}:\d{2}:\d{2} \w{3} \d{4}`))
			Eventually(session).Should(Say(`\s+state\s+since\s+cpu\s+memory\s+disk`))
			Eventually(session).Should(Say(`#0\s+running\s+\d{4}-[01]\d-[0-3]\dT[0-2][0-9]:[0-5]\d:[0-5]\dZ`))
			Eventually(session).Should(Say(`#1\s+running\s+\d{4}-[01]\d-[0-3]\dT[0-2][0-9]:[0-5]\d:[0-5]\dZ`))
			Eventually(session).Should(Say(`#2\s+running\s+\d{4}-[01]\d-[0-3]\dT[0-2][0-9]:[0-5]\d:[0-5]\dZ`))
			Eventually(session).Should(Exit(0))
		})
	})
})
