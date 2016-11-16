package helpers

import (
	"fmt"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

func BindRouteToApplication(app string, domain string, host string, path string) {
	Eventually(CF("map-route", app, domain, "--hostname", host, "--path", path)).Should(Exit(0))
	Eventually(CF("routes")).Should(And(Exit(0), Say(fmt.Sprintf("%s\\s+%s\\s+/%s\\s+%s", host, domain, path, app))))
}

func UnbindRouteToApplication(app string, domain string, host string, path string) {
	Eventually(CF("unmap-route", app, domain, "--hostname", host, "--path", path)).Should(Exit(0))
	session := CF("routes")
	Eventually(session).Should(Exit(0))
	Eventually(session).ShouldNot(Say(fmt.Sprintf("%s\\s+%s\\s+/%s\\s+%s", host, domain, path, app)))
}
