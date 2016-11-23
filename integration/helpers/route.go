package helpers

import (
	"fmt"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

type Route struct {
	Space  string
	Host   string
	Domain string
	Path   string
}

func NewRoute(space string, domain string, hostname string, path string) Route {
	return Route{
		Space:  space,
		Domain: domain,
		Host:   hostname,
		Path:   path,
	}
}

func (r Route) Create() {
	Eventually(CF("create-route", r.Space, r.Domain, "--hostname", r.Host, "--path", r.Path)).Should(Exit(0))
}

func (r Route) Delete() {
	Eventually(CF("delete-route", r.Domain, "--hostname", r.Host, "--path", r.Path, "-f")).Should(Exit(0))
}

func DomainName(prefix ...string) string {
	if len(prefix) > 0 {
		return fmt.Sprintf("integration-%s.com", PrefixedRandomName(prefix[0]))
	}
	return fmt.Sprintf("integration%s.com", PrefixedRandomName(""))
}

type Domain struct {
	Org  string
	Name string
}

func NewDomain(org string, name string) Domain {
	return Domain{
		Org:  org,
		Name: name,
	}
}

func (d Domain) Create() {
	Eventually(CF("create-domain", d.Org, d.Name)).Should(Exit(0))
	Eventually(CF("domains")).Should(And(Exit(0), Say(d.Name)))
}

func (d Domain) CreateShared() {
	Eventually(CF("create-shared-domain", d.Name)).Should(Exit(0))
}

func (d Domain) CreateWithRouterGroup(routerGroup string) {
	Eventually(CF("create-shared-domain", d.Name, "--router-group", routerGroup)).Should(Exit(0))
}

func (d Domain) Share() {
	Eventually(CF("share-private-domain", d.Org, d.Name)).Should(Exit(0))
}

func (d Domain) Delete() {
	Eventually(CF("delete-domain", d.Name, "-f")).Should(Exit(0))
}
