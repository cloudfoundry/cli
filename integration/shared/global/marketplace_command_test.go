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
			When("there are no accessible services", func() {
				BeforeEach(func() {
					helpers.LogoutCF()
				})

				It("displays a message that no services are available", func() {
					session := helpers.CF("marketplace")
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Say("\n\n"))
					Eventually(session).Should(Say("No service offerings found"))
					Eventually(session).Should(Exit(0))
				})
			})
		})
	})
})
