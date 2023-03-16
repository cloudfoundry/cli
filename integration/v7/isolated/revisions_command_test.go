package isolated

import (
	"bytes"
	"fmt"
	"regexp"

	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("revisions command", func() {
	var (
		orgName   string
		spaceName string
		appName   string
		username  string
	)

	BeforeEach(func() {
		username, _ = helpers.GetCredentials()
		orgName = helpers.NewOrgName()
		spaceName = helpers.NewSpaceName()
		appName = helpers.PrefixedRandomName("app")
	})

	Describe("help", func() {
		When("--help flag is set", func() {
			It("appears in cf help -a", func() {
				session := helpers.CF("help", "-a")
				Eventually(session).Should(Exit(0))
				Expect(session).To(HaveCommandInCategoryWithDescription("revisions", "EXPERIMENTAL COMMANDS", "List revisions of an app"))
			})

			It("Displays command usage to output", func() {
				session := helpers.CF("revisions", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("revisions - List revisions of an app"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf revisions APP_NAME"))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("rollback"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("targeting and org and space", func() {
		BeforeEach(func() {
			helpers.SetupCF(orgName, spaceName)
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		When("An app name is not provided", func() {
			It("Returns the incorrect usage text and help information", func() {
				session := helpers.CF("revisions")
				Eventually(session).Should(Exit(1))
				Expect(session.Err.Contents()).Should(ContainSubstring("Incorrect Usage: the required argument `APP_NAME` was not provided"))
				Expect(session).Should(Say("revisions - List revisions of an app"))
			})
		})

		When("the provided app does not exist", func() {
			It("properly displays app not found error", func() {
				fakeAppName := helpers.PrefixedRandomName("test-fake-app")
				session := helpers.CF("revisions", fakeAppName)
				Eventually(session).Should(Exit(1))
				Expect(session).Should(Say(regexp.QuoteMeta(`Getting revisions for app %s in org %s / space %s as %s...`), fakeAppName, orgName, spaceName, username))
				Expect(session.Err).Should(Say(regexp.QuoteMeta(`App '%s' not found`), fakeAppName))
				Expect(session).To(Say("FAILED"))
			})
		})

		When("an app has been pushed without staging", func() {
			BeforeEach(func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CF("push", appName, "-p", appDir, "--no-start")).Should(Exit(0))
				})
			})

			It("prints a 'not found' message without failing", func() {
				session := helpers.CF("revisions", appName)
				Eventually(session).Should(Exit(0))
				Expect(session).Should(Say(`No \w+ found`))
			})
		})

		When("An app has been pushed several times", func() {
			BeforeEach(func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CF("push", appName, "-p", appDir)).Should(Exit(0))
					Eventually(helpers.CF("push", appName, "-p", appDir)).Should(Exit(0))
				})
			})

			It("Retrieves the revisions", func() {
				session := helpers.CF("revisions", appName)
				Eventually(session).Should(Exit(0))
				Expect(session).Should(Say(regexp.QuoteMeta(`Getting revisions for app %s in org %s / space %s as %s...`), appName, orgName, spaceName, username))

				Expect(session).Should(Say(`2\(deployed\)\s+New droplet deployed.\s+true`))
				Expect(session).Should(Say(`1\s+Initial revision.\s+true`))
			})

			When("revisions are disabled for the app", func() {

				BeforeEach(func() {
					session := helpers.CF("app", appName, "--guid")
					Eventually(session).Should(Exit(0))

					appGuid := bytes.TrimSpace(session.Out.Contents())
					routeToDisableRevisions := fmt.Sprintf(`/v3/apps/%s/features/revisions`, appGuid)
					session = helpers.CF("curl", routeToDisableRevisions, "-X", "PATCH", "-d", `{ "enabled": false }`)
					Eventually(session).Should(Exit(0))
				})
				It("outputs the revisions with a warning", func() {
					session := helpers.CF("revisions", appName)
					Eventually(session).Should(Exit(0))
					Expect(session).Should(Say(regexp.QuoteMeta(`Getting revisions for app %s in org %s / space %s as %s...`), appName, orgName, spaceName, username))
					Expect(session.Err.Contents()).To(ContainSubstring(fmt.Sprintf("Warning: Revisions for app '%s' are disabled. Updates to the app will not create new revisions.", appName)))
					Expect(session).Should(Say("New droplet deployed"))
					Expect(session).Should(Say("Initial revision"))
				})
			})
		})
	})
})
