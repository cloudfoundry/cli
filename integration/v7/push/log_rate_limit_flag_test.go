package push

import (
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("push with log rate limit flag", func() {
	var (
		appName string
	)

	BeforeEach(func() {
		helpers.SkipIfVersionLessThan(ccversion.MinVersionLogRateLimitingV3)

		appName = helpers.NewAppName()
	})

	Context("when the -l flag is provided with application log rate limit", func() {
		It("creates the app with the specified log rate limit", func() {
			helpers.WithHelloWorldApp(func(dir string) {
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir},
					PushCommandName, appName,
					"-l", "5K",
				)
				Eventually(session).Should(Exit(0))
			})

			time.Sleep(5 * time.Second)
			session := helpers.CF("app", appName)
			Eventually(session).Should(Say(`name:\s+%s`, appName))
			Eventually(session).Should(Say(`last uploaded:\s+%s`, helpers.ReadableDateTimeRegex))
			//TODO: check output of push command
			// Eventually(session).Should(Say(`log rate usage per second:\s+5K`))
			// Eventually(session).Should(Say(`\s+state\s+since\s+cpu\s+memory\s+disk\s+log rate per second`))
			Eventually(session).Should(Say(`#0\s+running\s+\d{4}-[01]\d-[0-3]\dT[0-2][0-9]:[0-5]\d:[0-5]\dZ`))
			Eventually(session).Should(Exit(0))
		})
	})
})
