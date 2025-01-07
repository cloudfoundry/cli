package push

import (
	"time"

	"code.cloudfoundry.org/cli/v9/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("push with memory flag", func() {
	var (
		appName string
	)

	BeforeEach(func() {
		appName = helpers.NewAppName()
	})

	Context("when the -m flag is provided with application memory", func() {
		It("creates the app with the specified memory", func() {
			helpers.WithHelloWorldApp(func(dir string) {
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir},
					PushCommandName, appName,
					"-m", "70M",
				)
				Eventually(session).Should(Exit(0))
			})

			time.Sleep(5 * time.Second)
			session := helpers.CF("app", appName)
			Eventually(session).Should(Say(`name:\s+%s`, appName))
			Eventually(session).Should(Say(`last uploaded:\s+%s`, helpers.ReadableDateTimeRegex))
			Eventually(session).Should(Say(`memory usage:\s+70M`))
			Eventually(session).Should(Exit(0))
		})
	})
})
