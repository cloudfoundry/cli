package push

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("deprecated route command-line flags", func() {

	const deprecationTemplate = "Deprecation warning: Use of the '%[1]s' command-line flag option is deprecated in favor of the 'routes' property in the manifest. Please see https://docs.cloudfoundry.org/devguide/deploy-apps/manifest-attributes.html#routes for usage information. The '%[1]s' command-line flag option will be removed in the future."

	var (
		appName       string
		host          string
		privateDomain string
	)

	BeforeEach(func() {
		appName = helpers.NewAppName()
		host = helpers.NewAppName()

		privateDomain = helpers.NewDomainName()
		domain := helpers.NewDomain(organization, privateDomain)
		domain.Create()
	})

	When("no deprecated flags are provided", func() {
		It("does not output a deprecation warning", func() {
			helpers.WithHelloWorldApp(func(dir string) {
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--no-start")
				Eventually(session).Should(Exit(0))
				Expect(string(session.Err.Contents())).ToNot(ContainSubstring("deprecated"))
			})
		})
	})

	When("the -d (domains) flag is provided", func() {
		It("outputs a deprecation warning", func() {
			helpers.WithHelloWorldApp(func(dir string) {
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--no-start", "-d", privateDomain)
				Eventually(session.Err).Should(Say(deprecationTemplate, "-d"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the --hostname flag is provided", func() {
		It("outputs a deprecation warning", func() {
			helpers.WithHelloWorldApp(func(dir string) {
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--no-start", "--hostname", host)
				Eventually(session.Err).Should(Say(deprecationTemplate, "--hostname"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the --no-hostname flag is provided", func() {
		It("outputs a deprecation warning", func() {
			helpers.WithHelloWorldApp(func(dir string) {
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--no-start", "--no-hostname", "-d", privateDomain)
				Eventually(session.Err).Should(Say(deprecationTemplate, "--no-hostname"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the --route-path flag is provided", func() {
		It("outputs a deprecation warning", func() {
			helpers.WithHelloWorldApp(func(dir string) {
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--no-start", "--route-path", "some-path")
				Eventually(session.Err).Should(Say(deprecationTemplate, "--route-path"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

})
