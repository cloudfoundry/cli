package helpers_test

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

func ExampleWithHelloWorldApp() {
	var appName string

	When("the app exists", func() {
		BeforeEach(func() {
			appName = helpers.PrefixedRandomName("app")

			helpers.WithHelloWorldApp(func(appDir string) {
				Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "push", appName)).Should(Exit(0))
			})
		})
	})
}
