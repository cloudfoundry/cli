package global

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"

	"code.cloudfoundry.org/cli/integration/helpers"
)

var _ = Describe("set-space-role command", func() {
	var (
		orgName   string
		spaceName string
	)

	When("the set_roles_by_username flag is disabled", func() {
		BeforeEach(func() {
			spaceName = helpers.NewSpaceName()
			helpers.LoginCF()
			orgName = helpers.CreateAndTargetOrg()
			helpers.CreateSpace(spaceName)
			helpers.DisableFeatureFlag("set_roles_by_username")
		})

		AfterEach(func() {
			helpers.EnableFeatureFlag("set_roles_by_username")
		})

		When("the user does not exist", func() {
			It("prints the error from UAA and exits 1", func() {
				session := helpers.CF("set-space-role", "not-exists", orgName, spaceName, "SpaceDeveloper")
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("No user exists with the username 'not-exists'."))
				Eventually(session).Should(Exit(1))
			})
		})
	})
})
