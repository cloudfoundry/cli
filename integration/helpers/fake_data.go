package helpers

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

// AddFiftyOneOrgs adds a mock handler to the given server which returns
// 51 orgs on GET requests to /v3/organizations?order_by=name. It also
// paginates, so page 2 can be requested with /v3/organizations?page=2&per_page=50.
func AddFiftyOneOrgs(server *ghttp.Server) {
	AddHandler(server,
		http.MethodGet,
		"/v3/organizations?order_by=name",
		http.StatusOK,
		[]byte(fmt.Sprintf(string(fixtureData("fifty-orgs-page-1.json")), server.URL())),
	)

	AddHandler(server,
		http.MethodGet,
		"/v3/organizations?page=2&per_page=50",
		http.StatusOK,
		fixtureData("fifty-orgs-page-2.json"),
	)
}

// AddFiftyOneSpaces adds mock handlers to the given http server which includes
// an organization which will contain 51 spaces
func AddFiftyOneSpaces(server *ghttp.Server) {
	AddHandler(server,
		http.MethodGet,
		"/v3/organizations?order_by=name",
		http.StatusOK,
		[]byte(fmt.Sprintf(string(fixtureData("fifty-spaces-org.json")), server.URL())),
	)

	AddHandler(server,
		http.MethodGet,
		"/v3/spaces?organization_guids=4305313e-d34e-4015-9a57-5550235cd6b0",
		http.StatusOK,
		[]byte(fmt.Sprintf(string(fixtureData("fifty-spaces-page-1.json")), server.URL())),
	)

	AddHandler(server,
		http.MethodGet,
		"/v3/spaces?organization_guids=4305313e-d34e-4015-9a57-5550235cd6b0&page=2&per_page=50",
		http.StatusOK,
		[]byte(fmt.Sprintf(string(fixtureData("fifty-spaces-page-2.json")), server.URL())),
	)
}

// AddEmptyPaginatedResponse adds a mock handler to the given server which returns
// a response with no resources.
func AddEmptyPaginatedResponse(server *ghttp.Server, path string) {
	AddHandler(server,
		http.MethodGet,
		path,
		http.StatusOK,
		fixtureData("empty-paginated-response.json"),
	)
}

func fixtureData(name string) []byte {
	wd := os.Getenv("GOPATH")
	fp := filepath.Join(wd, "src", "code.cloudfoundry.org", "cli", "integration", "helpers", "fixtures", name)
	b, err := ioutil.ReadFile(fp)
	Expect(err).ToNot(HaveOccurred())
	return b
}
