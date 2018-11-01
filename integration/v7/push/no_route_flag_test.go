// +build !partialPush

package push

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = When("the --no-route flag is set", func() {
	var (
		appName string
	)

	BeforeEach(func() {
		appName = helpers.PrefixedRandomName("app")
	})

	It("does not map any routes to the app", func() {
		helpers.WithHelloWorldApp(func(appDir string) {
			session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, PushCommandName, appName, "--no-route")
			Consistently(session).ShouldNot(Say(`Mapping routes\.\.\.`))
			Eventually(session).Should(Say(`name:\s+%s`, appName))
			Eventually(session).Should(Say(`requested state:\s+started`))
			Eventually(session).Should(Say(`routes:\s+\n`))
			Eventually(session).Should(Exit(0))
		})
	})
})
