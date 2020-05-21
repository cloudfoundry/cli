package helpers

import (
	"fmt"
	"regexp"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

type RouterGroup struct {
	Name            string
	ReservablePorts string
	guid            string
}

func NewRouterGroup(name, reservablePorts string) RouterGroup {
	return RouterGroup{
		Name:            name,
		ReservablePorts: reservablePorts,
	}
}

func (rg *RouterGroup) Create() {
	session := CF("curl", "-X", "POST", "/routing/v1/router_groups", "-d", fmt.Sprintf(`{
		"name": "%s",
		"type": "tcp",
		"reservable_ports": "%s"
	}`, rg.Name, rg.ReservablePorts))

	Eventually(session).Should(Exit(0))

	// Capture the guid from the response so this group can be deleted later
	stdout := string(session.Out.Contents())
	guidFieldRegex := regexp.MustCompile(`"guid":\s*"([^"]+)"`)
	guidMatch := guidFieldRegex.FindStringSubmatch(stdout)
	Expect(guidMatch).NotTo(BeNil())

	rg.guid = guidMatch[1]
}

func (rg *RouterGroup) Delete() {
	session := CF("curl", "--fail", "-X", "DELETE", fmt.Sprintf("/routing/v1/router_groups/%s", rg.guid))

	Eventually(session).Should(Exit(0))
}
