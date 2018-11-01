// +build !partialPush

package global

import (
	"fmt"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("buildpacks command", func() {

	When("--help is passed", func() {
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
			Skip("Unrefactored command is writing login errors to STDOUT; remove skip when refactored")
			helpers.CheckEnvironmentTargetedCorrectly(false, false, ReadOnlyOrg, "buildpacks")
		})

		It("displays an error and exits 1", func() {
			helpers.UnrefactoredCheckEnvironmentTargetedCorrectly(false, false, ReadOnlyOrg, "buildpacks")
		})
	})

	When("too many args are passed", func() {
		It("displays FAILED, then usage, then exits 1", func() {
			session := helpers.CF("buildpacks", "no-further-args-allowed")
			Eventually(session).Should(Say("FAILED"))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Say("buildpacks - List all buildpacks"))
			Eventually(session).Should(Say("USAGE:"))
			Eventually(session).Should(Say("cf buildpacks"))
			Eventually(session).Should(Exit(1))
		})
	})

	When("the targeted API supports stack association", func() {
		BeforeEach(func() {
			helpers.SkipIfVersionLessThan(ccversion.MinVersionBuildpackStackAssociationV2)
		})

		It("lists the buildpacks with the stack column", func() {
			helpers.LoginCF()
			session := helpers.CF("buildpacks")
			Eventually(session).Should(Say("Getting buildpacks..."))
			Eventually(session).Should(Say(`buildpack\s+position\s+enabled\s+locked\s+filename\s+stack`))

			buildpackNameRegex := `staticfile_buildpack`
			positionRegex := `\d+`
			boolRegex := `(true|false)`
			buildpackFileRegex := `staticfile-buildpack-\S+`
			stackRegex := `(cflinuxfs[23]|windows.+)`

			Eventually(session).Should(Say(fmt.Sprintf(`%s\s+%s\s+%s\s+%s\s+%s\s+%s`, buildpackNameRegex,
				positionRegex, boolRegex, boolRegex, buildpackFileRegex, stackRegex)))
			Eventually(session).Should(Exit(0))
		})
	})

	When("the targeted API does not support stack association", func() {
		BeforeEach(func() {
			helpers.SkipIfVersionAtLeast(ccversion.MinVersionBuildpackStackAssociationV2)
		})

		It("includes the stack column but does not include the stack values", func() {
			helpers.LoginCF()
			session := helpers.CF("buildpacks")
			Eventually(session).Should(Say("Getting buildpacks..."))
			Eventually(session).Should(Say(`buildpack\s+position\s+enabled\s+locked\s+filename\s+stack\n`))

			buildpackNameRegex := `staticfile_buildpack`
			positionRegex := `\d+`
			boolRegex := `(true|false)`
			buildpackFileRegex := `staticfile_buildpack-\S+`

			Eventually(session).Should(Say(fmt.Sprintf(`%s\s+%s\s+%s\s+%s\s+%s\n`, buildpackNameRegex,
				positionRegex, boolRegex, boolRegex, buildpackFileRegex)))
			Eventually(session).Should(Exit(0))
		})
	})
})
