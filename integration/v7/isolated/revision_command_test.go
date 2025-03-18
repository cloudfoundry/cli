package isolated

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	. "code.cloudfoundry.org/cli/v8/cf/util/testhelpers/matchers"
	"code.cloudfoundry.org/cli/v8/integration/helpers"
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
				Expect(session).To(HaveCommandInCategoryWithDescription("revision", "APPS", "Show details for a specific app revision"))
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

		When("the requested revision version does not exist", func() {
			BeforeEach(func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CF("create-app", appName)).Should(Exit(0))
					Eventually(helpers.CF("set-env", appName, "foo", "bar1")).Should(Exit(0))
					Eventually(helpers.CF("push", appName, "-p", appDir)).Should(Exit(0))
				})
			})
			It("displays revision not found", func() {
				session := helpers.CF("revision", appName, "--version", "125")
				Eventually(session).Should(Exit(1))

				Expect(session).Should(Say(
					fmt.Sprintf("Showing revision 125 for app %s in org %s / space %s as %s...", appName, orgName, spaceName, username),
				))
				Expect(session.Err).Should(Say("Revision '125' not found"))
			})
		})

		When("the requested app and revision both exist", func() {
			BeforeEach(func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CF("create-app", appName)).Should(Exit(0))
					Eventually(helpers.CF("set-env", appName, "foo", "bar1")).Should(Exit(0))
					Eventually(helpers.CF("push", appName, "-p", appDir)).Should(Exit(0))
					Eventually(helpers.CF("push", appName, "-p", appDir)).Should(Exit(0))
				})
			})

			It("shows details about the revision", func() {
				cmd := exec.Command("bash", "-c", "cf revision "+appName+" --version 1 | grep \"revision GUID\" | sed -e 's/.*:\\s*//' -e 's/^[ \\t]*//'")
				var stdout bytes.Buffer
				cmd.Stdout = &stdout
				err := cmd.Run()
				if err != nil {
					return
				}
				revisionGUID := strings.TrimSpace(stdout.String())
				data := map[string]interface{}{
					"metadata": map[string]interface{}{
						"labels": map[string]string{
							"label": "foo3",
						},
						"annotations": map[string]string{
							"annotation": "foo3",
						},
					},
				}
				metadata, err := json.Marshal(data)
				Expect(err).NotTo(HaveOccurred())

				url := "/v3/revisions/" + string(revisionGUID)
				Eventually(helpers.CF("curl", "-X", "PATCH", url, "-d", string(metadata))).Should(Exit(0))

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

				Expect(session).Should(Say(`labels:`))
				Expect(session).Should(Say(`label:   foo3`))

				Expect(session).Should(Say(`annotations:`))
				Expect(session).Should(Say(`annotation:   foo3`))

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

				Expect(session).Should(Say(`labels:`))
				Expect(session).Should(Say(`annotations:`))

				Expect(session).Should(Say(`application environment variables:`))
				Expect(session).Should(Say(`foo:   bar1`))
			})
		})

		When("the revision version is not mentioned", func() {
			BeforeEach(func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CF("create-app", appName)).Should(Exit(0))
					Eventually(helpers.CF("push", appName, "-p", appDir)).Should(Exit(0))
					Eventually(helpers.CF("push", appName, "-p", appDir)).Should(Exit(0))
					Eventually(helpers.CF("push", appName, "-p", appDir, "--strategy", "canary")).Should(Exit(0))
				})
			})

			It("shows all the deployed revisions", func() {
				session := helpers.CF("revision", appName)
				Eventually(session).Should(Exit(0))

				Expect(session).Should(Say(
					fmt.Sprintf("Showing revisions for app %s in org %s / space %s as %s...", appName, orgName, spaceName, username),
				))
				Expect(session).Should(Say(`revision:        2`))
				Expect(session).Should(Say(`deployed:        true`))
				Expect(session).Should(Say(`description:     New droplet deployed`))
				Expect(session).Should(Say(`deployable:      true`))
				Expect(session).Should(Say(`revision GUID:   \S+\n`))
				Expect(session).Should(Say(`droplet GUID:    \S+\n`))
				Expect(session).Should(Say(`created on:      \S+\n`))

				Expect(session).Should(Say(`labels:`))
				Expect(session).Should(Say(`annotations:`))
				Expect(session).Should(Say(`application environment variables:`))

				Expect(session).Should(Say(`revision:        3`))
				Expect(session).Should(Say(`deployed:        true`))
				Expect(session).Should(Say(`description:     New droplet deployed`))
				Expect(session).Should(Say(`deployable:      true`))
				Expect(session).Should(Say(`revision GUID:   \S+\n`))
				Expect(session).Should(Say(`droplet GUID:    \S+\n`))
				Expect(session).Should(Say(`created on:      \S+\n`))

				Expect(session).Should(Say(`labels:`))
				Expect(session).Should(Say(`annotations:`))
				Expect(session).Should(Say(`application environment variables:`))

				Expect(session).ShouldNot(Say(`revision:        1`))
			})
		})
	})
})
