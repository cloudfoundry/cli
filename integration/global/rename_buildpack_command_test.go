package global

import (
	"code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("rename buildpack command", func() {
	BeforeEach(func() {
		Skip("until #158613755 is complete")
	})

	Describe("help", func() {
		Context("when --help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("rename-buildpack", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("rename-buildpack - Rename a buildpack"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf rename-buildpack BUILDPACK_NAME NEW_BUILDPACK_NAME"))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("update-buildpack"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("when the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(false, false, ReadOnlyOrg, "rename-buildpack", "fake-buildpack", "some-name")
		})
	})

	Context("when the user is logged in", func() {
		var (
			oldBuildpackName      string
			newBuildpackName      string
			existingBuildpackName string
		)

		BeforeEach(func() {
			helpers.LoginCF()
		})

		Context("when the buildpack exists", func() {
			BeforeEach(func() {
				oldBuildpackName = helpers.NewBuildpack()
				newBuildpackName = helpers.NewBuildpack()

				helpers.BuildpackWithStack(func(buildpackPath string) {
					session := helpers.CF("create-buildpack", oldBuildpackName, buildpackPath, "1")
					Eventually(session).Should(Exit(0))
				}, "")
			})

			Context("when renaming to a unique name", func() {
				It("successfully renames buildpack", func() {
					session := helpers.CF("rename-buildpack", oldBuildpackName, newBuildpackName)
					Eventually(session).Should(Say("Renaming buildpack %s to %s", oldBuildpackName, newBuildpackName))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when renaming to the same name", func() {
				It("successfully renames buildpack", func() {
					session := helpers.CF("rename-buildpack", oldBuildpackName, oldBuildpackName)
					Eventually(session).Should(Say("Renaming buildpack %s to %s", oldBuildpackName, oldBuildpackName))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when renaming to the same name as another buildpack", func() {
				BeforeEach(func() {
					existingBuildpackName = helpers.NewBuildpack()

					helpers.BuildpackWithStack(func(buildpackPath string) {
						session := helpers.CF("create-buildpack", existingBuildpackName, buildpackPath, "1")
						Eventually(session).Should(Exit(0))
					}, "")
				})

				It("returns the appropriate error", func() {
					session := helpers.CF("rename-buildpack", oldBuildpackName, existingBuildpackName)
					Eventually(session).Should(Say("Renaming buildpack %s to %s", oldBuildpackName, existingBuildpackName))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("Buildpack %s already exists without a stack", existingBuildpackName))
					Eventually(session).Should(Exit(1))
				})
			})
		})

		Context("when there are multiple ambiguous buildpacks", func() {
			BeforeEach(func() {
				helpers.SkipIfOneStack()

				stacks := helpers.FetchStacks()

				oldBuildpackName = helpers.NewBuildpack()
				newBuildpackName = helpers.NewBuildpack()

				helpers.BuildpackWithStack(func(buildpackPath string) {
					session := helpers.CF("create-buildpack", oldBuildpackName, buildpackPath, "1")
					Eventually(session).Should(Exit(0))
				}, stacks[0])
				helpers.BuildpackWithStack(func(buildpackPath string) {
					session := helpers.CF("create-buildpack", oldBuildpackName, buildpackPath, "1")
					Eventually(session).Should(Exit(0))
				}, stacks[1])
			})

			It("returns multiple buildpacks found error", func() {
				session := helpers.CF("rename-buildpack", oldBuildpackName, newBuildpackName)
				Eventually(session).Should(Say("Renaming buildpack %s to %s", oldBuildpackName, newBuildpackName))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("Multiple buildpacks named %s found.", oldBuildpackName))
				Eventually(session).Should(Exit(1))
			})
		})
	})
})
