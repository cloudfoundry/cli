package isolated

import (
	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("create-app command", func() {
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
				Expect(session).To(HaveCommandInCategoryWithDescription("create-app", "APPS", "Create an Application in the target space"))
			})

			It("Displays command usage to output", func() {
				session := helpers.CF("create-app", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("create-app - Create an Application in the target space"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say(`cf create-app APP_NAME \[--app-type \(buildpack | docker\)\]`))
				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say(`--app-type\s+App lifecycle type to stage and run the app \(Default: buildpack\)`))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("app, apps, push"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the app name is not provided", func() {
		It("tells the user that the app name is required, prints help text, and exits 1", func() {
			session := helpers.CF("create-app")

			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `APP_NAME` was not provided"))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	When("app type is not supported", func() {
		It("tells the user that the app type is incorrect, prints help text, and exits 1", func() {
			session := helpers.CF("create-app", appName, "--app-type", "unknown-app-type")

			Eventually(session.Err).Should(Say("Incorrect Usage: Invalid value `unknown-app-type' for option `--app-type'. Allowed values are: buildpack or docker"))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "create-app", appName)
		})
	})

	When("the environment is set up correctly", func() {
		BeforeEach(func() {
			helpers.SetupCF(orgName, spaceName)
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		When("the app does not exist", func() {
			It("creates the app with default app type buildpack", func() {
				session := helpers.CF("create-app", appName)
				userName, _ := helpers.GetCredentials()
				Eventually(session).Should(Say("Creating app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))

				session = helpers.CF("app", appName)
				Eventually(session).Should(Say("buildpacks:"))
				Eventually(session).Should(Exit(0))
			})

			When("app type is specified", func() {
				It("creates the app with the specified app type", func() {
					session := helpers.CF("create-app", appName, "--app-type", "docker")
					userName, _ := helpers.GetCredentials()
					Eventually(session).Should(Say("Creating app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))

					session = helpers.CF("app", appName)
					Eventually(session).Should(Say("docker image:"))
					Eventually(session).Should(Exit(0))
				})
			})
		})

		When("the app already exists", func() {
			BeforeEach(func() {
				Eventually(helpers.CF("create-app", appName)).Should(Exit(0))
			})

			It("fails to create the app", func() {
				session := helpers.CF("create-app", appName)
				userName, _ := helpers.GetCredentials()
				Eventually(session).Should(Say("Creating app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName))
				Eventually(session).Should(Say(`App with the name '%s' already exists\.`, appName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
