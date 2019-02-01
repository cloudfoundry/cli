package global

import (
	"code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("marketplace command", func() {
	When("an API endpoint is set", func() {
		When("not logged in", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
			})

			It("displays an informative message and exits 0", func() {
				session := helpers.CF("marketplace")
				Eventually(session).Should(Say("Getting all services from marketplace"))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
