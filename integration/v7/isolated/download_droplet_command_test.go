package isolated

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("download-droplet command", func() {
	var (
		helpText func(*Session)
		appName  string
	)

	BeforeEach(func() {
		appName = helpers.NewAppName()

		helpText = func(session *Session) {
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Say("download-droplet - Download an application droplet"))
			Eventually(session).Should(Say("USAGE:"))
			Eventually(session).Should(Say(`cf download-droplet APP_NAME \[--droplet DROPLET_GUID\] \[--path /path/to/droplet.tgz\]`))
			Eventually(session).Should(Say("OPTIONS:"))
			Eventually(session).Should(Say(`--droplet\s+The guid of the droplet to download \(default: app's current droplet\).`))
			Eventually(session).Should(Say(`--path, -p\s+File path to download droplet to \(default: current working directory\).`))
			Eventually(session).Should(Say("SEE ALSO:"))
			Eventually(session).Should(Say("apps, droplets, push, set-droplet"))
		}
	})

	Describe("help", func() {
		When("--help flag is set", func() {
			It("appears in cf help -a", func() {
				session := helpers.CF("help", "-a")
				Eventually(session).Should(Exit(0))
				Expect(session).To(HaveCommandInCategoryWithDescription("download-droplet", "APPS", "Download an application droplet"))
			})

			It("Displays command usage to output", func() {
				session := helpers.CF("download-droplet", "--help")
				helpText(session)
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the app name is not provided", func() {
		It("tells the user that the app name is required, prints help text, and exits 1", func() {
			session := helpers.CF("download-droplet")

			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `APP_NAME` was not provided"))
			helpText(session)
			Eventually(session).Should(Exit(1))
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "download-droplet", appName)
		})
	})

	When("the environment is setup correctly", func() {
		var (
			spaceName string
			orgName   string
			userName  string
		)

		BeforeEach(func() {
			spaceName = helpers.NewSpaceName()
			orgName = helpers.NewOrgName()

			helpers.SetupCF(orgName, spaceName)
			userName, _ = helpers.GetCredentials()
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		When("the app has a current droplet", func() {
			var (
				dropletPath string
				dropletGUID string
			)

			BeforeEach(func() {
				helpers.CreateApp(appName)

				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "push", appName)).Should(Exit(0))
				})
				dropletSession := helpers.CF("droplets", appName)
				Eventually(dropletSession).Should(Exit(0))
				regex := regexp.MustCompile(`(.+)\s+\(current\)`)
				matches := regex.FindStringSubmatch(string(dropletSession.Out.Contents()))
				Expect(matches).To(HaveLen(2))
				dropletGUID = matches[1]

				dir, err := os.Getwd()
				Expect(err).ToNot(HaveOccurred())
				dropletPath = filepath.Join(dir, "droplet_"+dropletGUID+".tgz")
			})

			AfterEach(func() {
				os.RemoveAll(dropletPath)
			})

			It("downloads the droplet successfully", func() {
				session := helpers.CF("download-droplet", appName)
				Eventually(session).Should(Say(`Downloading current droplet for app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
				Eventually(session).Should(helpers.SayPath(`Droplet downloaded successfully at %s`, dropletPath))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))

				_, err := os.Stat(dropletPath)
				Expect(err).ToNot(HaveOccurred())
			})

			When("a path to a directory is provided", func() {
				BeforeEach(func() {
					tmpDir, err := ioutil.TempDir("", "droplets")
					Expect(err).NotTo(HaveOccurred())
					dropletPath = tmpDir
				})

				It("downloads the droplet to the given path successfully", func() {
					session := helpers.CF("download-droplet", appName, "--path", dropletPath)
					Eventually(session).Should(Say(`Downloading current droplet for app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
					Eventually(session).Should(helpers.SayPath(`Droplet downloaded successfully at %s`, dropletPath))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))

					_, err := os.Stat(filepath.Join(dropletPath, "droplet_"+dropletGUID+".tgz"))
					Expect(err).ToNot(HaveOccurred())
				})
			})

			When("a path to a file is provided", func() {
				BeforeEach(func() {
					tmpDir, err := ioutil.TempDir("", "droplets")
					Expect(err).NotTo(HaveOccurred())
					dropletPath = filepath.Join(tmpDir, "my-droplet.tgz")
				})

				It("downloads the droplet to the given path successfully", func() {
					session := helpers.CF("download-droplet", appName, "--path", dropletPath)
					Eventually(session).Should(Say(`Downloading current droplet for app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
					Eventually(session).Should(helpers.SayPath(`Droplet downloaded successfully at %s`, dropletPath))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))

					_, err := os.Stat(dropletPath)
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})

		When("the app does not exist", func() {
			It("displays app not found and exits 1", func() {
				session := helpers.CF("download-droplet", appName)

				Eventually(session).Should(Say(`Downloading current droplet for app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
				Eventually(session.Err).Should(Say("App '%s' not found", appName))
				Eventually(session).Should(Say("FAILED"))

				Eventually(session).Should(Exit(1))
			})
		})

		When("the app does not have a current droplet", func() {
			BeforeEach(func() {
				helpers.CreateApp(appName)
			})

			It("displays that there is no current droplet and exits 1", func() {
				session := helpers.CF("download-droplet", appName)

				Eventually(session).Should(Say(`Downloading current droplet for app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
				Eventually(session.Err).Should(Say("App '%s' does not have a current droplet.", appName))
				Eventually(session).Should(Say("FAILED"))

				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the droplet flag is passed", func() {
			var (
				dropletPath string
				dropletGUID string
			)

			BeforeEach(func() {
				helpers.CreateApp(appName)

				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "push", appName)).Should(Exit(0))
				})
				dropletSession := helpers.CF("droplets", appName)
				Eventually(dropletSession).Should(Exit(0))
				regex := regexp.MustCompile(`(.+)\s+\(current\)`)
				matches := regex.FindStringSubmatch(string(dropletSession.Out.Contents()))
				Expect(matches).To(HaveLen(2))
				dropletGUID = matches[1]

				dir, err := os.Getwd()
				Expect(err).ToNot(HaveOccurred())
				dropletPath = filepath.Join(dir, "droplet_"+dropletGUID+".tgz")

				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "push", appName)).Should(Exit(0))
				})
			})

			AfterEach(func() {
				os.RemoveAll(dropletPath)
			})

			It("downloads the droplet successfully", func() {
				session := helpers.CF("download-droplet", appName, "--droplet", dropletGUID)
				Eventually(session).Should(Say(`Downloading droplet %s for app %s in org %s / space %s as %s\.\.\.`, dropletGUID, appName, orgName, spaceName, userName))
				Eventually(session).Should(helpers.SayPath(`Droplet downloaded successfully at %s`, dropletPath))
				Eventually(session).Should(Say("OK"))

				_, err := os.Stat(dropletPath)
				Expect(err).ToNot(HaveOccurred())
			})

			When("the app does not contain a droplet with the given guid", func() {
				It("downloads the droplet successfully", func() {
					session := helpers.CF("download-droplet", appName, "--droplet", "bogus")
					Eventually(session).Should(Say(`Downloading droplet bogus for app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
					Eventually(session.Err).Should(Say("Droplet 'bogus' not found for app '%s'", appName))
					Eventually(session).Should(Say("FAILED"))

					Eventually(session).Should(Exit(1))
				})
			})
		})
	})
})
