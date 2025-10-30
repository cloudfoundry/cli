package helpers

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"

	"code.cloudfoundry.org/cli/v8/resources"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

// MinTestPort should be defined by the CF router group for integration tests.
const MinTestPort = 1024

// MaxTestPort should be defined by the CF router group for integration tests.
const MaxTestPort = 1034

// FindOrCreateTCPRouterGroup uses the routing API to find a router group with name
// INTEGRATION-TCP-NODE-<node>, or create one if it does not exist. Returns the name of
// the router group.
func FindOrCreateTCPRouterGroup(node int) string {
	routerGroupName := fmt.Sprintf("INTEGRATION-TCP-NODE-%d", node)

	session := CF("curl", fmt.Sprintf("/routing/v1/router_groups?name=%s", routerGroupName))
	Eventually(session).Should(Exit(0))
	doesNotExist := regexp.MustCompile("ResourceNotFoundError")
	if doesNotExist.Match(session.Out.Contents()) {
		jsonBody := fmt.Sprintf(`{"name": "%s", "type": "tcp", "reservable_ports": "%d-%d"}`, routerGroupName, MinTestPort, MaxTestPort)
		session := CF("curl", "-d", jsonBody, "-X", "POST", "/routing/v1/router_groups")
		Eventually(session).Should(Say(`"name":\s*"%s"`, routerGroupName))
		Eventually(session).Should(Say(`"type":\s*"tcp"`))
		Eventually(session).Should(Exit(0))
	}

	return routerGroupName
}

// Route represents a route.
type Route struct {
	Domain  string
	Host    string
	Path    string
	Port    int
	Space   string
	Options map[string]*string
}

// NewRoute constructs a route with given space, domain, hostname, and path.
func NewRoute(space string, domain string, hostname string, path string, options map[string]*string) Route {
	return Route{
		Space:   space,
		Domain:  domain,
		Host:    hostname,
		Path:    path,
		Options: options,
	}
}

// NewTCPRoute constructs a TCP route with given space, domain, and port.
func NewTCPRoute(space string, domain string, port int, options map[string]*string) Route {
	return Route{
		Space:   space,
		Domain:  domain,
		Port:    port,
		Options: options,
	}
}

// Create creates a route using the 'cf create-route' command.
func (r Route) Create() {
	r.Options = make(map[string]*string)
	if r.Port != 0 {
		Eventually(CF("create-route", r.Space, r.Domain, "--port", fmt.Sprint(r.Port), "--option", fmt.Sprint(r.Options))).Should(Exit(0))
	} else {
		Eventually(CF("create-route", r.Space, r.Domain, "--hostname", r.Host, "--path", r.Path, "--option", fmt.Sprint(r.Options))).Should(Exit(0))
	}
}

// Create creates a route using the 'cf create-route' command.
func (r Route) V7Create() {
	if r.Port != 0 {
		Eventually(CF("create-route", r.Domain, "--port", fmt.Sprint(r.Port))).Should(Exit(0))
	} else {
		Eventually(CF("create-route", r.Domain, "--hostname", r.Host, "--path", r.Path)).Should(Exit(0))
	}
}

// Delete deletes a route using the 'cf delete-route' command.
func (r Route) Delete() {
	if r.Port != 0 {
		Eventually(CF("delete-route", r.Domain, "--port", fmt.Sprint(r.Port))).Should(Exit(0))
	} else {
		Eventually(CF("delete-route", r.Domain, "--hostname", r.Host, "--path", r.Path, "-f")).Should(Exit(0))
	}
}

// String stringifies a route (e.g. "host.domain.com:port/path")
func (r Route) String() string {
	routeString := r.Domain

	if r.Port != 0 {
		routeString = fmt.Sprintf("%s:%d", routeString, r.Port)
	}

	if r.Host != "" {
		routeString = fmt.Sprintf("%s.%s", r.Host, routeString)
	}

	if r.Path != "" {
		routeString = path.Join(routeString, r.Path)
	}

	return routeString
}

func (r Route) GUID() string {
	var domainReceiver struct {
		Domains []resources.Domain `json:"resources"`
	}
	Curlf(&domainReceiver, "/v3/domains?names=%s", r.Domain)
	Expect(domainReceiver.Domains).To(HaveLen(1))

	query := []string{fmt.Sprintf("domain_guids=%s", domainReceiver.Domains[0].GUID)}
	if r.Host != "" {
		query = append(query, fmt.Sprintf("hosts=%s", r.Host))
	}
	if r.Path != "" {
		path := r.Path
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		query = append(query, fmt.Sprintf("paths=%s", path))
	}
	if r.Port != 0 {
		query = append(query, fmt.Sprintf("ports=%d", r.Port))
	}

	var routeReceiver struct {
		Routes []resources.Route `json:"resources"`
	}
	Curlf(&routeReceiver, "/v3/routes?%s", strings.Join(query, "&"))
	Expect(routeReceiver.Routes).To(HaveLen(1))

	return routeReceiver.Routes[0].GUID
}

func GetPort() int {
	port := 1024
	testName := os.Getenv("CF_INT_TEST_NAME")
	if testName == "cc" {
		port = 1025
	}
	return port
}
