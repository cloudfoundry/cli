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
				Eventually(session).Should(Say(`\s+cf set-label org business pci=true public-facing=false`))
				Eventually(session).Should(Say(`\s+cf set-label space business_space public-facing=false owner=jane_doe`))
				Eventually(session).Should(Say("RESOURCES:"))
				Eventually(session).Should(Say(`\s+app`))
				Eventually(session).Should(Say(`\s+org`))
				Eventually(session).Should(Say(`\s+space`))
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

		When("assigning label to app", func() {
			BeforeEach(func() {
				spaceName = helpers.NewSpaceName()
				appName = helpers.PrefixedRandomName("app")

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
				appGUID := helpers.AppGUID(appName)
				session = helpers.CF("curl", fmt.Sprintf("/v3/apps/%s", appGUID))
				Eventually(session).Should(Exit(0))
				appJSON := session.Out.Contents()
				var app commonResource
				Expect(json.Unmarshal(appJSON, &app)).To(Succeed())
				Expect(len(app.Metadata.Labels)).To(Equal(2))
				Expect(app.Metadata.Labels["some-key"]).To(Equal("some-value"))
				Expect(app.Metadata.Labels["some-other-key"]).To(Equal("some-other-value"))
			})

			When("the app is unknown", func() {
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

			When("more than one value is provided for the same key", func() {
				It("uses the last value", func() {
					session := helpers.CF("set-label", "app", appName, "owner=sue", "owner=beth")
					Eventually(session).Should(Exit(0))
					appGUID := helpers.AppGUID(appName)
					session = helpers.CF("curl", fmt.Sprintf("/v3/apps/%s", appGUID))
					Eventually(session).Should(Exit(0))
					appJSON := session.Out.Contents()
					var app commonResource
					Expect(json.Unmarshal(appJSON, &app)).To(Succeed())
					Expect(len(app.Metadata.Labels)).To(Equal(1))
					Expect(app.Metadata.Labels["owner"]).To(Equal("beth"))
				})
			})
		})

		When("assigning label to org", func() {
			It("sets the specified labels on the org", func() {
				session := helpers.CF("set-label", "org", orgName, "pci=true", "public-facing=false")
				Eventually(session).Should(Say(regexp.QuoteMeta(`Setting label(s) for org %s as %s...`), orgName, username))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))

				orgGUID := helpers.GetOrgGUID(orgName)
				session = helpers.CF("curl", fmt.Sprintf("/v3/organizations/%s", orgGUID))
				Eventually(session).Should(Exit(0))
				orgJSON := session.Out.Contents()
				var org commonResource
				Expect(json.Unmarshal(orgJSON, &org)).To(Succeed())
				Expect(len(org.Metadata.Labels)).To(Equal(2))
				Expect(org.Metadata.Labels["pci"]).To(Equal("true"))
				Expect(org.Metadata.Labels["public-facing"]).To(Equal("false"))
			})

			When("the org is unknown", func() {
				It("displays an error", func() {
					session := helpers.CF("set-label", "org", "non-existent-org", "some-key=some-value")
					Eventually(session.Err).Should(Say("Organization 'non-existent-org' not found"))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("the label has an empty key and an invalid value", func() {
				It("displays an error", func() {
					session := helpers.CF("set-label", "org", orgName, "=test", "sha2=108&eb90d734")
					Eventually(session.Err).Should(Say("Metadata key error: label key cannot be empty string, Metadata value error: label '108&eb90d734' contains invalid characters"))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("the label does not include a '=' to separate the key and value", func() {
				It("displays an error", func() {
					session := helpers.CF("set-label", "org", orgName, "test-label")
					Eventually(session.Err).Should(Say("Metadata error: no value provided for label 'test-label'"))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("more than one value is provided for the same key", func() {
				It("uses the last value", func() {
					session := helpers.CF("set-label", "org", orgName, "owner=sue", "owner=beth")
					Eventually(session).Should(Exit(0))
					orgGUID := helpers.GetOrgGUID(orgName)
					session = helpers.CF("curl", fmt.Sprintf("/v3/organizations/%s", orgGUID))
					Eventually(session).Should(Exit(0))
					orgJSON := session.Out.Contents()
					var org commonResource
					Expect(json.Unmarshal(orgJSON, &org)).To(Succeed())
					Expect(len(org.Metadata.Labels)).To(Equal(1))
					Expect(org.Metadata.Labels["owner"]).To(Equal("beth"))
				})
			})
		})
	})
})
