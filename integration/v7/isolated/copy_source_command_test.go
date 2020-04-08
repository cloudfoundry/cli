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
		sourceAppName   string
		targetAppName   string
		orgName         string
		secondOrgName   string
		spaceName       string
		secondSpaceName string
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
				helpText(session)
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Describe("command behavior without flags", func() {
		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()

			helpers.SetupCF(orgName, spaceName)

			sourceAppName = helpers.PrefixedRandomName("hello")
			targetAppName = helpers.PrefixedRandomName("banana")

			helpers.WithHelloWorldApp(func(appDir string) {
				Eventually(helpers.CF("push", sourceAppName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
			})

			helpers.WithBananaPantsApp(func(appDir string) {
				Eventually(helpers.CF("push", targetAppName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack")).Should(Exit(0))
			})
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		It("copies the app", func() {
			session := helpers.CF("copy-source", sourceAppName, targetAppName)
			Eventually(session).Should(Say("Copying source from app %s to target app %s", sourceAppName, targetAppName))
			Eventually(session).Should(Exit(0))

			resp, err := http.Get(fmt.Sprintf("http://%s.%s", targetAppName, helpers.DefaultSharedDomain()))
			Expect(err).ToNot(HaveOccurred())
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(body)).To(MatchRegexp("hello world"))
		})
	})

	Describe("command behavior with a space flag", func() {
		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()
			secondSpaceName = helpers.NewSpaceName()

			helpers.SetupCF(orgName, secondSpaceName)
			helpers.CreateSpace(spaceName)

			sourceAppName = helpers.PrefixedRandomName("hello")
			targetAppName = helpers.PrefixedRandomName("banana")

			helpers.WithHelloWorldApp(func(appDir string) {
				Eventually(helpers.CF("push", sourceAppName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
			})

			helpers.TargetOrgAndSpace(orgName, spaceName)

			helpers.WithBananaPantsApp(func(appDir string) {
				Eventually(helpers.CF("push", targetAppName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack")).Should(Exit(0))
			})
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		It("copies the app to the provided space", func() {
			username, _ := helpers.GetCredentials()
			session := helpers.CF("copy-source", sourceAppName, targetAppName, "--space", secondSpaceName)
			Eventually(session).Should(Say("Copying source from app %s to target app %s in org %s / space %s as %s...", sourceAppName, targetAppName, orgName, secondSpaceName, username))
			Eventually(session).Should(Exit(0))

			resp, err := http.Get(fmt.Sprintf("http://%s.%s", targetAppName, helpers.DefaultSharedDomain()))
			Expect(err).ToNot(HaveOccurred())
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(body)).To(MatchRegexp("hello world"))
		})
	})

	FDescribe("command behavior with a space flag and an org flag", func() {
		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			secondOrgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()
			secondSpaceName = helpers.NewSpaceName()

			sourceAppName = helpers.PrefixedRandomName("hello")
			targetAppName = helpers.PrefixedRandomName("banana")

			helpers.SetupCF(orgName, spaceName)

			helpers.WithHelloWorldApp(func(appDir string) {
				Eventually(helpers.CF("push", sourceAppName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
			})

			helpers.SetupCF(secondOrgName, secondSpaceName)

			helpers.WithBananaPantsApp(func(appDir string) {
				Eventually(helpers.CF("push", targetAppName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack")).Should(Exit(0))
			})

			helpers.TargetOrgAndSpace(orgName, spaceName)
		})

		AfterEach(func() {
			//helpers.QuickDeleteOrg(orgName)
			//helpers.QuickDeleteOrg(secondOrgName)
		})

		It("copies the app to the provided space", func() {
			username, _ := helpers.GetCredentials()
			session := helpers.CF("copy-source", sourceAppName, targetAppName, "--organization", secondOrgName, "--space", secondSpaceName)
			Eventually(session).Should(Say("Copying source from app %s to target app %s in org %s / space %s as %s...", sourceAppName, targetAppName, secondOrgName, secondSpaceName, username))
			Eventually(session).Should(Exit(0))

			resp, err := http.Get(fmt.Sprintf("http://%s.%s", targetAppName, helpers.DefaultSharedDomain()))
			Expect(err).ToNot(HaveOccurred())
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(body)).To(MatchRegexp("hello world"))
		})
	})

	Describe("command behavior with an invalid org name", func() {
		var invalidOrgName string
		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			invalidOrgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()

			sourceAppName = helpers.PrefixedRandomName("hello")
			targetAppName = helpers.PrefixedRandomName("banana")

			helpers.SetupCF(orgName, spaceName)

			helpers.WithHelloWorldApp(func(appDir string) {
				Eventually(helpers.CF("push", sourceAppName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
			})

			helpers.WithBananaPantsApp(func(appDir string) {
				Eventually(helpers.CF("push", targetAppName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack")).Should(Exit(0))
			})
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		It("copies the app to the provided space", func() {
			username, _ := helpers.GetCredentials()
			session := helpers.CF("copy-source", sourceAppName, targetAppName, "--organization", invalidOrgName, "--space", spaceName)
			Eventually(session).Should(Say("Copying source from app %s to target app %s in org %s / space %s as %s...", sourceAppName, targetAppName, invalidOrgName, secondSpaceName, username))
			Eventually(session).Should(Say("Organization '%s' not found.", invalidOrgName))
			Eventually(session).Should(Say("FAILED"))
			helpText(session)
			Eventually(session).Should(Exit(1))
		})
	})

	Describe("command behavior with an org name only (no space)", func() {
		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			secondOrgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()
			secondSpaceName = helpers.NewSpaceName()

			sourceAppName = helpers.PrefixedRandomName("hello")
			targetAppName = helpers.PrefixedRandomName("banana")

			helpers.SetupCF(orgName, spaceName)

			helpers.WithHelloWorldApp(func(appDir string) {
				Eventually(helpers.CF("push", sourceAppName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
			})

			helpers.SetupCF(secondOrgName, secondSpaceName)

			helpers.WithBananaPantsApp(func(appDir string) {
				Eventually(helpers.CF("push", targetAppName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack")).Should(Exit(0))
			})

			helpers.TargetOrgAndSpace(orgName, spaceName)
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		It("copies the app to the provided space", func() {
			username, _ := helpers.GetCredentials()
			session := helpers.CF("copy-source", sourceAppName, targetAppName, "--organization", secondOrgName)
			Eventually(session).Should(Say("Copying source from app %s to target app %s in org %s / space %s as %s...", sourceAppName, targetAppName, secondOrgName, secondSpaceName, username))
			Eventually(session).Should(Say("Incorrect Usage: '--organization, -o' requires '--space, -s' to be specified"))
			Eventually(session).Should(Say("FAILED"))
			helpText(session)
			Eventually(session).Should(Exit(1))
		})
	})
})

func helpText(session *Session) {
	Eventually(session).Should(Say("NAME:"))
	Eventually(session).Should(Say("copy-source - Copies the source code of an application to another existing application and restages that application"))
	Eventually(session).Should(Say("USAGE:"))
	Eventually(session).Should(Say(`cf copy-source SOURCE_APP DESTINATION_APP`))
	Eventually(session).Should(Say("OPTIONS:"))
	Eventually(session).Should(Say(`--organization, -o\s+Org that contains the destination application`))
	Eventually(session).Should(Say(`--space, -s\s+Space that contains the destination application`))
	Eventually(session).Should(Say("ENVIRONMENT:"))
	Eventually(session).Should(Say(`CF_STAGING_TIMEOUT=15\s+Max wait time for staging, in minutes`))
	Eventually(session).Should(Say(`CF_STARTUP_TIMEOUT=5\s+Max wait time for app instance startup, in minutes`))
	Eventually(session).Should(Say("SEE ALSO:"))
	Eventually(session).Should(Say("apps, push, restage, restart, target"))
}
