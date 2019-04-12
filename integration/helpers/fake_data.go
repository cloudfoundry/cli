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

func AddFiftyOneOrgs(server *ghttp.Server) {
	AddHandler(server,
		http.MethodGet,
		"/v3/organizations",
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

func fixtureData(name string) []byte {
	wd := os.Getenv("GOPATH")
	fp := filepath.Join(wd, "src", "code.cloudfoundry.org", "cli", "integration", "helpers", "fixtures", name)
	b, err := ioutil.ReadFile(fp)
	Expect(err).ToNot(HaveOccurred())
	return b
}
