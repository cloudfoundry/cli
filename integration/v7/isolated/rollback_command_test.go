package isolated

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("rollback command", func() {
	var (
		orgName   string
		spaceName string
		appName   string
	)

	BeforeEach(func() {
		orgName = helpers.NewOrgName()
		spaceName = helpers.NewSpaceName()
		appName = helpers.PrefixedRandomName("app")
	})

	Describe("help", func() {
		When("--help flag is set", func() {
			It("appears in cf help -a", func() {
				session := helpers.CF("help", "-a")
				Eventually(session).Should(Exit(0))
				Expect(session).To(HaveCommandInCategoryWithDescription("rollback", "EXPERIMENTAL COMMANDS", "Rollback to the specified revision of an app"))
			})

			It("Displays rollback command usage to output", func() {
				session := helpers.CF("rollback", "--help")

				Eventually(session).Should(Exit(0))

				Expect(session).To(Say("NAME:"))
				Expect(session).To(Say("rollback - Rollback to the specified revision of an app"))
				Expect(session).To(Say("USAGE:"))
				Expect(session).To(Say(`cf rollback APP_NAME \[--revision REVISION_NUMBER\]`))
				Expect(session).To(Say("OPTIONS:"))
				Expect(session).To(Say("--revision      Roll back to the given app revision"))
				Expect(session).To(Say("SEE ALSO:"))
				Expect(session).To(Say("revisions"))
			})
		})
	})

	When("the environment is set up correctly", func() {
		BeforeEach(func() {
			helpers.SetupCF(orgName, spaceName)
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		Describe("version dependent display", func() {

			var domainName string

			BeforeEach(func() {
				domainName = helpers.DefaultSharedDomain()
			})

			When("the app is started and has 2 instances", func() {
				BeforeEach(func() {
					helpers.WithHelloWorldApp(func(appDir string) {
						manifestContents := []byte(fmt.Sprintf(`
---
applications:
- name: %s
  memory: 128M
  instances: 2
  disk_quota: 128M
  routes:
  - route: %s.%s
`, appName, appName, domainName))
						manifestPath := filepath.Join(appDir, "manifest.yml")
						err := ioutil.WriteFile(manifestPath, manifestContents, 0666)
						Expect(err).ToNot(HaveOccurred())

						// Create manifest
						Eventually(helpers.CF("push", appName, "-p", appDir, "-f", manifestPath, "-b", "staticfile_buildpack")).Should(Exit(0))
						Eventually(helpers.CF("push", appName, "-p", appDir, "-f", manifestPath, "-b", "staticfile_buildpack")).Should(Exit(0))
					})
				})

				It("succeeds", func() {
					userName, _ := helpers.GetCredentials()

					session := helpers.CF("rollback", appName, "--revision", "1")

					Eventually(session).Should(Say(`Rolling back to revision 1 for app %s in org %s / space %s as %s...`, appName, orgName, spaceName, userName))
					Eventually(session).Should(Say(`OK`))
					Eventually(session).Should(Exit(0))
					session = helpers.CF("revisions", appName)
					Eventually(session).Should(Say(`3\s+[\w\-]+\s+New droplet deployed.`))
					Eventually(session).Should(Exit(0))
				})
			})
		})
	})
})
