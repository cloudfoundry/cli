package helpers

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var noSpaces string = `{
	"next_url": null,
	"prev_url": null,
	"resources": null,
	"total_pages": 1,
	"total_results": 0
	}`

func AddFiftyOneOrgs(server *ghttp.Server, orgName string) {
	AddHandler(server,
		http.MethodGet,
		"/v2/organizations?order-by=name",
		http.StatusOK,
		[]byte(fmt.Sprintf(string(fixtureData("fifty-orgs-page-1.json")), orgName)),
	)

	AddHandler(server,
		http.MethodGet,
		"/v2/organizations?order-by=name&order-direction=asc&page=2&results-per-page=50",
		http.StatusOK,
		fixtureData("fifty-orgs-page-2.json"),
	)

	AddHandler(server,
		http.MethodGet,
		"/v2/spaces?order-by=name&q=organization_guid%3A6f30e06d-360e-4cd7-9849-01f28109bc37",
		http.StatusOK,
		[]byte(noSpaces),
	)

	AddHandler(server,
		http.MethodGet,
		strings.ToLower(fmt.Sprintf("/v2/organizations?q=name%%3A%[1]s&inline-relations-depth=1", orgName)),
		http.StatusOK,
		[]byte(fmt.Sprintf(string(fixtureData("empty-org-depth-1.json")), orgName)),
	)

	AddHandler(server,
		http.MethodGet,
		"/v2/organizations/some-org-guid/spaces?order-by=name&inline-relations-depth=1",
		http.StatusOK,
		[]byte(noSpaces),
	)
}

func fixtureData(name string) []byte {
	wd := os.Getenv("GOPATH")
	fp := filepath.Join(wd, "src", "code.cloudfoundry.org", "cli", "integration", "helpers", "fixtures", name)
	b, err := ioutil.ReadFile(fp)
	Expect(err).ToNot(HaveOccurred())
	return b
}
