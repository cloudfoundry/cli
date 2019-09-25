package isolated

import (
	"regexp"

	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("buildpacks command", func() {
	When("--help is passed", func() {
		It("appears in cf help -a", func() {
			session := helpers.CF("help", "-a")
			Eventually(session).Should(Exit(0))
			Expect(session).To(HaveCommandInCategoryWithDescription("buildpacks", "BUILDPACKS", "List all buildpacks"))
		})

		It("displays the help message", func() {
			session := helpers.CF("buildpacks", "--help")
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Say("buildpacks - List all buildpacks"))
			Eventually(session).Should(Say("USAGE:"))
			Eventually(session).Should(Say("cf buildpacks"))
			Eventually(session).Should(Say("SEE ALSO:"))
			Eventually(session).Should(Say("push"))
			Eventually(session).Should(Exit(0))
		})
	})

	When("environment is not set up", func() {
		It("displays an error and exits 1", func() {
			helpers.CheckEnvironmentTargetedCorrectly(false, false, ReadOnlyOrg, "buildpacks")
		})
	})

	When("the targeted API supports stack association", func() {
		BeforeEach(func() {
			helpers.LoginCF()
		})

		It("lists the buildpacks with the stack column", func() {
			session := helpers.CF("buildpacks")

			username, _ := helpers.GetCredentials()
			Eventually(session).Should(Say("Getting buildpacks as %s...", username))
			Eventually(session).Should(Say(`position\s+name\s+stack\s+enabled\s+locked\s+filename`))

			positionRegex := `\d+`
			enabledRegex := `true`
			lockedRegex := `false`
			stackRegex := `(cflinuxfs[23]|windows.+)`

			staticfileNameRegex := `staticfile_buildpack`
			// staticfileFileRegex := `staticfile[-_]buildpack-\S+`
			staticfileFileRegex := ""
			Eventually(session).Should(Say(`%s\s+%s\s+%s\s+%s\s+%s\s+%s`,
				positionRegex,
				staticfileNameRegex,
				stackRegex,
				enabledRegex,
				lockedRegex,
				staticfileFileRegex))

			binaryNameRegex := `binary_buildpack`
			// binaryFileRegex := `binary[-_]buildpack-\S+`
			binaryFileRegex := ""
			Eventually(session).Should(Say(`%s\s+%s\s+%s\s+%s\s+%s\s+%s`,
				positionRegex,
				binaryNameRegex,
				stackRegex,
				enabledRegex,
				lockedRegex,
				binaryFileRegex))
			Eventually(session).Should(Exit(0))
		})

		When("the --labels flag is given", func() {
			When("the --labels selector is malformed", func() {
				It("errors", func() {
					session := helpers.CF("buildpacks", "--labels", "malformed in (")
					Eventually(session).Should(Exit(1))
				})
			})

			When("there are buildpacks with labels", func() {
				var (
					buildpack1 string
					buildpack2 string
				)

				BeforeEach(func() {
					buildpack1 = helpers.NewBuildpackName()
					buildpack2 = helpers.NewBuildpackName()
					helpers.SetupBuildpackWithoutStack(buildpack1)
					helpers.SetupBuildpackWithoutStack(buildpack2)
					Eventually(helpers.CF("set-label", "buildpack", buildpack1, "environment=production", "tier=backend")).Should(Exit(0))
					Eventually(helpers.CF("set-label", "buildpack", buildpack2, "environment=staging", "tier=frontend")).Should(Exit(0))
				})

				AfterEach(func() {
					Eventually(helpers.CF("delete-buildpack", buildpack1, "-f")).Should(Exit(0))
					Eventually(helpers.CF("delete-buildpack", buildpack2, "-f")).Should(Exit(0))
				})

				It("lists the filtered buildpacks", func() {
					session := helpers.CF("buildpacks", "--labels", "tier=frontend")
					Eventually(session).Should(Exit(0))

					username, _ := helpers.GetCredentials()
					Expect(session).Should(Say("Getting buildpacks as %s...", username))
					Expect(session).Should(Say(`position\s+name\s+stack\s+enabled\s+locked\s+filename`))

					re := regexp.MustCompile(`(?:\n|\r)\d+\s+\w{8}`)
					buildpackMatches := re.FindAll(session.Out.Contents(), -1)
					Expect(len(buildpackMatches)).To(Equal(1))

					positionRegex := `\d+`
					enabledRegex := `true`
					lockedRegex := `false`
					stackRegex := ``

					Expect(session).Should(Say(`%s\s+%s\s+%s\s+%s\s+%s\s+%s`,
						positionRegex,
						buildpack2,
						stackRegex,
						enabledRegex,
						lockedRegex,
						""))
				})
			})
		})
	})
})
