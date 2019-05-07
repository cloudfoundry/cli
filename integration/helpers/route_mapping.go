package helpers

import (
	"fmt"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

// MapRouteToApplication maps a route to an app using 'cf map-route' and asserts that
// the mapping exists.
func MapRouteToApplication(app string, domain string, host string, path string) {
	Eventually(CF("map-route", app, domain, "--hostname", host, "--path", path)).Should(Exit(0))
	Eventually(CF("routes")).Should(And(Exit(0), Say(fmt.Sprintf("%s\\s+%s\\s+/%s\\s+%s", host, domain, path, app))))
}

// UnmapRouteFromApplication unmaps a route from an app using 'cf unmap-route' and asserts that
// the mapping gets deleted.
func UnmapRouteFromApplication(app string, domain string, host string, path string) {
	Eventually(CF("unmap-route", app, domain, "--hostname", host, "--path", path)).Should(Exit(0))
	session := CF("routes")
	Eventually(session).Should(Exit(0))
	Eventually(session).ShouldNot(Say(fmt.Sprintf("%s\\s+%s\\s+/%s\\s+%s", host, domain, path, app)))
}
