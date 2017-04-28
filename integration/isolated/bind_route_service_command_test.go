package isolated

import (
	. "code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("bind-route-service command", func() {
	Describe("help", func() {
		It("includes a description of the options", func() {
			session := CF("help", "bind-route-service")
			Eventually(session).Should(Say("-c\\s+Valid JSON object containing service-specific configuration parameters, provided inline or in a file. For a list of supported configuration parameters, see documentation for the particular service offering."))
			Eventually(session).Should(Say("--hostname, -n\\s+Hostname used in combination with DOMAIN to specify the route to bind"))
			Eventually(session).Should(Say("--path\\s+Path used in combination with HOSTNAME and DOMAIN to specify the route to bind"))
			Eventually(session).Should(Exit(0))
		})
	})
})
