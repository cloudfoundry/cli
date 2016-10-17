package helpers

import (
	"time"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

const (
	CFRouteLongTimeout = 30 * time.Second
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
	Eventually(CF("create-route", r.Space, r.Domain, "--hostname", r.Host, "--path", r.Path), CFRouteLongTimeout).Should(Exit(0))
}

func (r Route) Delete() {
	Eventually(CF("delete-route", r.Domain, "--hostname", r.Host, "--path", r.Path, "-f"), CFRouteLongTimeout).Should(Exit(0))
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
	Eventually(CF("create-domain", d.Org, d.Name), CFRouteLongTimeout).Should(Exit(0))
	Eventually(CF("domains"), CFRouteLongTimeout).Should(And(Exit(0), Say(d.Name)))
}

func (d Domain) Share() {
	Eventually(CF("share-private-domain", d.Org, d.Name), CFRouteLongTimeout).Should(Exit(0))
}

func (d Domain) Delete() {
	Eventually(CF("delete-domain", d.Name, "-f"), CFRouteLongTimeout).Should(Exit(0))
}
