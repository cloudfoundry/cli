package isolated

import (
	"regexp"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("unbind-route-service command", func() {
	BeforeEach(func() {
		helpers.SkipIfClientCredentialsTestMode()
	})
	Describe("help", func() {
		It("includes a description of the options", func() {
			session := helpers.CF("help", "unbind-route-service")
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Say("unbind-route-service - Unbind a service instance from an HTTP route"))
			Eventually(session).Should(Say("USAGE:"))
			Eventually(session).Should(Say(regexp.QuoteMeta("cf unbind-route-service DOMAIN [--hostname HOSTNAME] [--path PATH] SERVICE_INSTANCE [-f]")))
			Eventually(session).Should(Say("EXAMPLES:"))
			Eventually(session).Should(Say("cf unbind-route-service example.com --hostname myapp --path foo myratelimiter"))
			Eventually(session).Should(Say("ALIAS:"))
			Eventually(session).Should(Say("urs"))
			Eventually(session).Should(Say("OPTIONS:"))
			Eventually(session).Should(Say(`-f\s+Force unbinding without confirmation`))
			Eventually(session).Should(Say(`--hostname, -n\s+Hostname used in combination with DOMAIN to specify the route to unbind`))
			Eventually(session).Should(Say(`--path\s+Path used in combination with HOSTNAME and DOMAIN to specify the route to unbind`))
			Eventually(session).Should(Say("SEE ALSO:"))
			Eventually(session).Should(Say("delete-service, routes, services"))
			Eventually(session).Should(Exit(0))
		})
	})
})
