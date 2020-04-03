package isolated

import (
	"fmt"
	"io/ioutil"
	"net/http"

	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("copy-source command", func() {
	var (
		appName1  string
		appName2  string
		orgName   string
		spaceName string
	)

	Describe("help", func() {
		When("--help flag is set", func() {
			It("appears in cf help -a", func() {
				session := helpers.CF("help", "-a")
				Eventually(session).Should(Exit(0))
				Expect(session).To(HaveCommandInCategoryWithDescription("copy-source", "APPS", "Copies the source code of an application to another existing application and restages that application"))
			})

			It("Displays command usage to output", func() {
				session := helpers.CF("copy-source", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("copy-source - Copies the source code of an application to another existing application and restages that application"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say(`cf copy-source SOURCE_APP DESTINATION_APP`))
				Eventually(session).Should(Say("ENVIRONMENT:"))
				Eventually(session).Should(Say(`CF_STAGING_TIMEOUT=15\s+Max wait time for staging, in minutes`))
				Eventually(session).Should(Say(`CF_STARTUP_TIMEOUT=5\s+Max wait time for app instance startup, in minutes`))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("apps, push, restage, restart, target"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Describe("command behavior", func() {
		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()

			helpers.SetupCF(orgName, spaceName)

			appName1 = helpers.PrefixedRandomName("hello")
			appName2 = helpers.PrefixedRandomName("banana")

			helpers.WithHelloWorldApp(func(appDir string) {
				Eventually(helpers.CF("push", appName1, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
			})

			helpers.WithBananaPantsApp(func(appDir string) {
				Eventually(helpers.CF("push", appName2, "--no-start", "-p", appDir, "-b", "staticfile_buildpack")).Should(Exit(0))
			})
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		It("copies the app", func() {
			session := helpers.CF("copy-source", appName1, appName2)
			Eventually(session).Should(Say("Copying source from app %s to target app %s", appName1, appName2))
			Eventually(session).Should(Exit(0))

			resp, err := http.Get(fmt.Sprintf("http://%s.%s", appName2, helpers.DefaultSharedDomain()))
			Expect(err).ToNot(HaveOccurred())
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(body)).To(MatchRegexp("hello world"))
		})
	})
})
