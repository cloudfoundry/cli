package isolated

import (
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
			helpers.CheckEnvironmentTargetedCorrectly(false, false, ReadOnlyOrg, "buildpacks")
		})
	})

	When("too many args are passed", func() {
		It("displays FAILED, then usage, then exits 1", func() {
			Skip("blocked on #162975536")
			session := helpers.CF("buildpacks", "no-further-args-allowed")
			Eventually(session.Err).Should(Say("Incorrect Usage: unexpected argument"))
			Eventually(session).Should(Say("FAILED"))
			Eventually(session).Should(Say("buildpacks - List all buildpacks"))
			Eventually(session).Should(Exit(1))
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
	})
})
