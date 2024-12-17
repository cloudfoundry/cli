package isolated

import (
	"fmt"

	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("revision command", func() {
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
				Expect(session).To(HaveCommandInCategoryWithDescription("revision", "EXPERIMENTAL COMMANDS", "Show details for a specific app revision"))
			})

			It("Displays revision command usage to output", func() {
				session := helpers.CF("revision", "--help")

				Eventually(session).Should(Exit(0))

				Expect(session).To(Say("NAME:"))
				Expect(session).To(Say("revision - Show details for a specific app revision"))
				Expect(session).To(Say("USAGE:"))
				Expect(session).To(Say(`cf revision APP_NAME [--version VERSION]`))
				Expect(session).To(Say("OPTIONS:"))
				Expect(session).To(Say("--version      The integer representing the specific revision to show"))
				Expect(session).To(Say("SEE ALSO:"))
				Expect(session).To(Say("revisions, rollback"))
			})
		})
	})

	When("targetting and org and space", func() {
		BeforeEach(func() {
			helpers.SetupCF(orgName, spaceName)
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		// TODO: #173456870 when app not provided, app does not exist, revision doesn't exist cases

		When("the requested app and revision both exist", func() {
			BeforeEach(func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CF("create-app", appName)).Should(Exit(0))
					Eventually(helpers.CF("set-env", appName, "foo", "bar1")).Should(Exit(0))
					Eventually(helpers.CF("push", appName, "-p", appDir)).Should(Exit(0))
					Eventually(helpers.CF("push", appName, "-p", appDir)).Should(Exit(0))
				})
			})

			FIt("shows details about the revision", func() {
				session := helpers.CF("revision", appName, "--version", "1")
				Eventually(session).Should(Exit(0))

				Expect(session).Should(Say(
					fmt.Sprintf("Showing revision 1 for app %s in org %s / space %s as %s...", appName, orgName, spaceName, username),
				))
				Expect(session).Should(Say(`revision:        1`))
				Expect(session).Should(Say(`deployed:        false`))
				Expect(session).Should(Say(`description:     Initial revision`))
				Expect(session).Should(Say(`deployable:      true`))
				Expect(session).Should(Say(`revision GUID:   \S+\n`))
				Expect(session).Should(Say(`droplet GUID:    \S+\n`))
				Expect(session).Should(Say(`created on:      \S+\n`))

				// Expect(session).Should(Say(`labels:`))
				// Expect(session).Should(Say(`label: foo1`))

				// Expect(session).Should(Say(`annotations:`))
				// Expect(session).Should(Say(`annotation: foo1`))

				Expect(session).Should(Say(`application environment variables:`))
				Expect(session).Should(Say(`foo:   bar1`))

				session = helpers.CF("revision", appName, "--version", "2")
				Eventually(session).Should(Exit(0))
				Expect(session).Should(Say(
					fmt.Sprintf("Showing revision 2 for app %s in org %s / space %s as %s...", appName, orgName, spaceName, username),
				))
				Expect(session).Should(Say(`revision:        2`))
				Expect(session).Should(Say(`deployed:        true`))
				Expect(session).Should(Say(`description:     New droplet deployed`))
				Expect(session).Should(Say(`deployable:      true`))
				Expect(session).Should(Say(`revision GUID:   \S+\n`))
				Expect(session).Should(Say(`droplet GUID:    \S+\n`))
				Expect(session).Should(Say(`created on:      \S+\n`))

				Expect(session).Should(Say(`application environment variables:`))
				Expect(session).Should(Say(`foo:   bar1`))

				// Expect(session).Should(Say(`labels:`))
				// Expect(session).Should(Say(`label: foo2`))

				// Expect(session).Should(Say(`annotations:`))
				// Expect(session).Should(Say(`annotation: foo2`))
			})
		})
	})
})
