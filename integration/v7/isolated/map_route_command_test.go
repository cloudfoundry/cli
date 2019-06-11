package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("map-route command", func() {
	Describe("help", func() {
		When("--help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("map-route", "--help")
				Eventually(session).Should(Say(`NAME:`))
				Eventually(session).Should(Say(`map-route - Map a route to an app\n`))
				Eventually(session).Should(Say(`\n`))

				Eventually(session).Should(Say(`USAGE:`))
				Eventually(session).Should(Say(`cf map-route APP_NAME DOMAIN \[--hostname HOSTNAME\] \[--path PATH\]\n`))
				Eventually(session).Should(Say(`\n`))

				Eventually(session).Should(Say(`EXAMPLES:`))
				Eventually(session).Should(Say(`cf map-route my-app example.com\s+# example.com`))
				Eventually(session).Should(Say(`cf map-route my-app example.com --hostname myhost\s+# myhost.example.com`))
				Eventually(session).Should(Say(`cf map-route my-app example.com --hostname myhost --path foo\s+# myhost.example.com/foo`))
				Eventually(session).Should(Say(`\n`))

				Eventually(session).Should(Say(`OPTIONS:`))
				Eventually(session).Should(Say(`--hostname, -n\s+Hostname for the HTTP route \(required for shared domains\)`))
				Eventually(session).Should(Say(`--path\s+Path for the HTTP route`))
				Eventually(session).Should(Say(`\n`))

				Eventually(session).Should(Say(`SEE ALSO:`))
				Eventually(session).Should(Say(`create-route, routes, unmap-route`))

				Eventually(session).Should(Exit(0))
			})
		})
	})
})
