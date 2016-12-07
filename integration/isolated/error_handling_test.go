package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("curl command", func() {
	It("returns the expected request", func() {
		session := helpers.CF("curl", "/v2/banana")
		Eventually(session).Should(Say(`"error_code": "CF-NotFound"`))
		Eventually(session).Should(Exit(0))
	})

	Context("when using -v", func() {
		It("returns the expected request with verbose output", func() {
			session := helpers.CF("curl", "-v", "/v2/banana")
			Eventually(session).Should(Say("GET /v2/banana HTTP/1.1"))
			Eventually(session).Should(Say(`"error_code": "CF-NotFound"`))
			Eventually(session).Should(Exit(0))
		})
	})
})
