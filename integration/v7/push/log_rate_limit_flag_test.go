package push

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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

			helpers.WaitForLogRateLimitToTakeEffect(appName, 0, 0, false, "5K")
		})
	})
})
