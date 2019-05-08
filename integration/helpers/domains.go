package helpers

import (
	"fmt"
	"regexp"
	"strings"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

// DomainName returns a random domain name, with a given prefix if provided.
func DomainName(prefix ...string) string {
	if len(prefix) > 0 {
		return fmt.Sprintf("integration-%s.com", PrefixedRandomName(prefix[0]))
	}
	return fmt.Sprintf("integration%s.com", PrefixedRandomName(""))
}

// Domain represents a domain scoped to an organization.
type Domain struct {
	Org  string
	Name string
}

// NewDomain constructs a new Domain with given owning organization and name.
func NewDomain(org string, name string) Domain {
	return Domain{
		Org:  org,
		Name: name,
	}
}

// globally cached
var foundDefaultDomain string

// DefaultSharedDomain runs 'cf domains' to find the default domain, caching
// the result so that the same domain is returned each time it is called.
func DefaultSharedDomain() string {
	if foundDefaultDomain == "" {
		session := CF("domains")
		Eventually(session).Should(Exit(0))

		regex := regexp.MustCompile(`(.+?)\s+shared`)

		output := strings.Split(string(session.Out.Contents()), "\n")
		for _, line := range output {
			if line != "" && !strings.HasPrefix(line, "integration-") {
				matches := regex.FindStringSubmatch(line)
				if len(matches) == 2 {
					foundDefaultDomain = matches[1]
					break
				}
			}
		}
		Expect(foundDefaultDomain).ToNot(BeEmpty())
	}
	return foundDefaultDomain
}

// Create uses 'cf create-domain' to create the domain in org d.Org with name
// d.Name.
func (d Domain) Create() {
	Eventually(CF("create-domain", d.Org, d.Name)).Should(Exit(0))
	Eventually(CF("domains")).Should(And(Exit(0), Say(d.Name)))
}

// CreatePrivate uses 'cf create-private-domain' to create the domain in org
// d.Org with name d.Name.
func (d Domain) CreatePrivate() {
	Eventually(CF("create-private-domain", d.Org, d.Name)).Should(Exit(0))
}

// CreateShared uses 'cf create-shared-domain' to create an shared domain
// with name d.Name.
func (d Domain) CreateShared() {
	Eventually(CF("create-shared-domain", d.Name)).Should(Exit(0))
}

// CreateWithRouterGroup uses 'cf create-shared-domain' to create a shared
// domain with name d.Name and given router group.
func (d Domain) CreateWithRouterGroup(routerGroup string) {
	Eventually(CF("create-shared-domain", d.Name, "--router-group", routerGroup)).Should(Exit(0))
}

// CreateInternal uses 'cf create-shared-domain' to create an shared,
// internal domain with name d.Name.
func (d Domain) CreateInternal() {
	Eventually(CF("create-shared-domain", d.Name, "--internal")).Should(Exit(0))
}

// V7Share uses 'cf share-private-domain' to share the domain with the given
// org.
func (d Domain) V7Share(orgName string) {
	Eventually(CF("share-private-domain", orgName, d.Name)).Should(Exit(0))
}

// Delete uses 'cf delete-domain' to delete the domain without asking for
// confirmation.
func (d Domain) Delete() {
	Eventually(CF("delete-domain", d.Name, "-f")).Should(Exit(0))
}

// DeleteShared uses 'cf delete-shared-domain' to delete the shared domain
// without asking for confirmation.
func (d Domain) DeleteShared() {
	Eventually(CF("delete-shared-domain", d.Name, "-f")).Should(Exit(0))
}
