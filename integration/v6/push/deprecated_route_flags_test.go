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
		localArgs     []string
		session       *Session
	)

	BeforeEach(func() {
		appName = helpers.NewAppName()
		host = helpers.NewAppName()

		privateDomain = helpers.NewDomainName()
		domain := helpers.NewDomain(organization, privateDomain)
		domain.Create()
	})
	JustBeforeEach(func() {
		helpers.WithHelloWorldApp(func(dir string) {
			allArgs := []string{PushCommandName, appName, "--no-start"}
			allArgs = append(allArgs, localArgs...)
			session = helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, allArgs...)
			Eventually(session).Should(Exit(0))
		})
	})

	When("no deprecated flags are provided", func() {
		It("does not output a deprecation warning", func() {
			Expect(string(session.Err.Contents())).ToNot(ContainSubstring("deprecated"))
		})
	})

	When("the -d (domains) flag is provided", func() {
		BeforeEach(func() {
			localArgs = []string{"-d", privateDomain}
		})
		It("outputs a deprecation warning", func() {
			Eventually(session.Err).Should(Say(deprecationTemplate, "-d"))
		})
	})

	When("the --hostname flag is provided", func() {
		BeforeEach(func() {
			localArgs = []string{"--hostname", host}
		})
		It("outputs a deprecation warning", func() {
			Eventually(session.Err).Should(Say(deprecationTemplate, "--hostname"))
		})
	})

	When("the --no-hostname flag is provided", func() {
		BeforeEach(func() {
			localArgs = []string{"--no-hostname", "-d", privateDomain}
		})
		It("outputs a deprecation warning", func() {
			Eventually(session.Err).Should(Say(deprecationTemplate, "--no-hostname"))
		})
	})

	When("the --route-path flag is provided", func() {
		BeforeEach(func() {
			localArgs = []string{"--route-path", "some-path"}
		})
		It("outputs a deprecation warning", func() {
			Eventually(session.Err).Should(Say(deprecationTemplate, "--route-path"))
		})
	})

})
