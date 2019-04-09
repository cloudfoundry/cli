package isolated

import (
	"regexp"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("labels command", func() {
	When("--help flag is given", func() {
		It("Displays command usage", func() {
			session := helpers.CF("labels", "--help")

			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Say(`\s+labels - List all labels \(key-value pairs\) for an API resource`))
			Eventually(session).Should(Say("USAGE:"))
			Eventually(session).Should(Say(`\s+cf labels RESOURCE RESOURCE_NAME`))
			Eventually(session).Should(Say("EXAMPLES:"))
			Eventually(session).Should(Say(`\s+cf labels app dora`))
			Eventually(session).Should(Say("RESOURCES:"))
			Eventually(session).Should(Say(`\s+APP`))
			Eventually(session).Should(Say("SEE ALSO:"))
			Eventually(session).Should(Say(`\s+set-label, delete-label`))
			Eventually(session).Should(Exit(0))
		})
	})

	When("the environment is set up correctly", func() {
		var (
			orgName   string
			spaceName string
			appName   string
			username  string
		)

		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()
			appName = helpers.PrefixedRandomName("app")

			username, _ = helpers.GetCredentials()
			helpers.SetupCF(orgName, spaceName)
			helpers.WithHelloWorldApp(func(appDir string) {
				Eventually(helpers.CF("push", appName, "-p", appDir)).Should(Exit(0))
			})
		})

		When("there are labels set on the application", func() {
			BeforeEach(func() {
				session := helpers.CF("set-label", "app", appName, "some-other-key=some-other-value", "some-key=some-value")
				Eventually(session).Should(Exit(0))
			})
			It("lists the labels", func() {
				session := helpers.CF("labels", "app", appName)
				Eventually(session).Should(Say(regexp.QuoteMeta("Getting labels for app %s in org %s / space %s as %s...\n\n"), appName, orgName, spaceName, username))
				Eventually(session).Should(Say(`Key\s+Value`))
				Eventually(session).Should(Say(`some-key\s+some-value`))
				Eventually(session).Should(Say(`some-other-key\s+some-other-value`))
				Eventually(session).Should(Exit(0))
			})
		})

		When("there are no labels set on the application", func() {
			It("indicates that there are no labels", func() {
				session := helpers.CF("labels", "app", appName)
				Eventually(session).Should(Say(regexp.QuoteMeta("Getting labels for app %s in org %s / space %s as %s...\n\n"), appName, orgName, spaceName, username))
				Expect(session).ToNot(Say(`Key\s+Value`))
				Eventually(session).Should(Say("No labels found."))
				Eventually(session).Should(Exit(0))
			})
		})

		When("the app does not exist", func() {
			It("displays an error", func() {
				session := helpers.CF("labels", "app", "non-existent-app")
				Eventually(session).Should(Say(regexp.QuoteMeta("Getting labels for app non-existent-app in org %s / space %s as %s...\n\n"), orgName, spaceName, username))
				Eventually(session.Err).Should(Say("App 'non-existent-app' not found"))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})
	})
})
