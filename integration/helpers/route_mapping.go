package helpers

import (
	"fmt"
	"time"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

const (
	CFRouteMappingLongTimeout = 30 * time.Second
)

func BindRouteToApplication(app string, domain string, host string, path string) {
	Eventually(CF("map-route", app, domain, "--hostname", host, "--path", path), CFRouteMappingLongTimeout).Should(Exit(0))
	Eventually(CF("routes"), CFRouteMappingLongTimeout).Should(And(Exit(0), Say(fmt.Sprintf("%s\\s+%s\\s+/%s\\s+%s", host, domain, path, app))))
}

func UnbindRouteToApplication(app string, domain string, host string, path string) {
	Eventually(CF("unmap-route", app, domain, "--hostname", host, "--path", path), CFRouteMappingLongTimeout).Should(Exit(0))
	session := CF("routes")
	Eventually(session, CFRouteMappingLongTimeout).Should(Exit(0))
	Eventually(session, CFRouteMappingLongTimeout).ShouldNot(Say(fmt.Sprintf("%s\\s+%s\\s+/%s\\s+%s", host, domain, path, app)))
}
