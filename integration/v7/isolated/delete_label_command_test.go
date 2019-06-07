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

var _ = Describe("delete-label command", func() {
	When("--help flag is given", func() {
		It("Displays command usage to output", func() {
			session := helpers.CF("delete-label", "--help")

			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Say(`\s+delete-label - Delete a label \(key-value pairs\) for an API resource`))
			Eventually(session).Should(Say("USAGE:"))
			Eventually(session).Should(Say(`\s+cf delete-label RESOURCE RESOURCE_NAME KEY`))
			Eventually(session).Should(Say("EXAMPLES:"))
			Eventually(session).Should(Say(`\s+cf delete-label app dora ci_signature_sha2`))
			Eventually(session).Should(Say("RESOURCES:"))
			Eventually(session).Should(Say(`\s+app`))
			Eventually(session).Should(Say("SEE ALSO:"))
			Eventually(session).Should(Say(`\s+set-label, labels`))
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

		type commonResource struct {
			Metadata struct {
				Labels map[string]string
			}
		}

		BeforeEach(func() {
			username, _ = helpers.GetCredentials()
			helpers.LoginCF()
			orgName = helpers.NewOrgName()
			helpers.CreateOrg(orgName)
		})

		When("deleting labels from an app", func() {
			BeforeEach(func() {
				spaceName = helpers.NewSpaceName()
				appName = helpers.PrefixedRandomName("app")

				helpers.SetupCF(orgName, spaceName)
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CF("push", appName, "-p", appDir, "--no-start")).Should(Exit(0))
				})

				session := helpers.CF("set-label", "app", appName, "some-key=some-value", "some-other-key=some-other-value", "some-third-key=other")
				Eventually(session).Should(Exit(0))
			})

			It("deletes the specified labels on the app", func() {
				session := helpers.CF("delete-label", "app", appName, "some-other-key", "some-third-key")
				Eventually(session).Should(Say(regexp.QuoteMeta(`Deleting label(s) for app %s in org %s / space %s as %s...`), appName, orgName, spaceName, username))
				Consistently(session).ShouldNot(Say("\n\nOK"))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))

				appGuid := helpers.AppGUID(appName)
				session = helpers.CF("curl", fmt.Sprintf("/v3/apps/%s", appGuid))
				Eventually(session).Should(Exit(0))
				appJSON := session.Out.Contents()

				var app commonResource
				Expect(json.Unmarshal(appJSON, &app)).To(Succeed())
				Expect(len(app.Metadata.Labels)).To(Equal(1))
				Expect(app.Metadata.Labels["some-key"]).To(Equal("some-value"))
			})
		})

		When("Deleting labels from an org", func() {
			BeforeEach(func() {
				session := helpers.CF("set-label", "org", orgName, "pci=true", "public-facing=false")
				Eventually(session).Should(Exit(0))
			})

			It("deletes the specified labels on the org", func() {
				session := helpers.CF("delete-label", "org", orgName, "public-facing")
				Eventually(session).Should(Say(regexp.QuoteMeta(`Deleting label(s) for org %s as %s...`), orgName, username))
				Consistently(session).ShouldNot(Say("\n\nOK"))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))

				orgGUID := helpers.GetOrgGUID(orgName)
				session = helpers.CF("curl", fmt.Sprintf("/v3/organizations/%s", orgGUID))
				Eventually(session).Should(Exit(0))
				orgJSON := session.Out.Contents()

				var org commonResource
				Expect(json.Unmarshal(orgJSON, &org)).To(Succeed())
				Expect(len(org.Metadata.Labels)).To(Equal(1))
				Expect(org.Metadata.Labels["pci"]).To(Equal("true"))
			})

		})
	})
})
