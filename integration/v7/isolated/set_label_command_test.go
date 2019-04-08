package isolated

import (
	"encoding/json"
	"fmt"
	"regexp"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("set-label command", func() {
	Describe("help", func() {
		When("--help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("set-label", "--help")

				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say(`\s+set-label - Set a label \(key-value pairs\) for an API resource`))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say(`\s+cf set-label RESOURCE RESOURCE_NAME KEY=VALUE\.\.\.`))
				Eventually(session).Should(Say("EXAMPLES:"))
				Eventually(session).Should(Say(`\s+cf set-label app dora env=production`))
				Eventually(session).Should(Say("RESOURCES:"))
				Eventually(session).Should(Say(`\s+APP`))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say(`\s+delete-label, labels`))

				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the environment is set up correctly", func() {
		var (
			orgName   string
			spaceName string
			appName   string
			username  string
		)

		type AppResource struct {
			Metadata struct {
				Labels map[string]string
			}
		}

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

		It("sets the specified labels on the app", func() {
			session := helpers.CF("set-label", "app", appName, "some-key=some-value", "some-other-key=some-other-value")
			Eventually(session).Should(Say(regexp.QuoteMeta(`Setting label(s) for app %s in org %s / space %s as %s...`), appName, orgName, spaceName, username))
			Eventually(session).Should(Say("OK"))
			Eventually(session).Should(Exit(0))
			appGuid := helpers.AppGUID(appName)
			session = helpers.CF("curl", fmt.Sprintf("/v3/apps/%s", appGuid))
			Eventually(session).Should(Exit(0))
			appJSON := session.Out.Contents()
			var app AppResource
			Expect(json.Unmarshal(appJSON, &app)).To(Succeed())
			Expect(len(app.Metadata.Labels)).To(Equal(2))
			Expect(app.Metadata.Labels["some-key"]).To(Equal("some-value"))
			Expect(app.Metadata.Labels["some-other-key"]).To(Equal("some-other-value"))
		})

		When("there the app is unknown", func() {
			It("displays an error", func() {
				session := helpers.CF("set-label", "app", "non-existent-app", "some-key=some-value")
				Eventually(session.Err).Should(Say("App 'non-existent-app' not found"))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the label has an empty key and an invalid value", func() {
			It("displays an error", func() {
				session := helpers.CF("set-label", "app", appName, "=test", "sha2=108&eb90d734")
				Eventually(session.Err).Should(Say("Metadata key error: label key cannot be empty string, Metadata value error: label '108&eb90d734' contains invalid characters"))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the label does not include a '=' to separate the key and value", func() {
			It("displays an error", func() {
				session := helpers.CF("set-label", "app", appName, "test-label")
				Eventually(session.Err).Should(Say("Metadata error: no value provided for label 'test-label'"))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})
	})
})
