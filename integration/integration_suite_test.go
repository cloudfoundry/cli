package integration

import (
	"io/ioutil"
	"os"
	"strconv"
	"time"

	"code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"

	"testing"
)

const (
	CFLongTimeout = 30 * time.Second
)

var (
	// Suite Level
	apiURL            string
	skipSSLValidation string
	originalColor     string

	// Per Test Level
	homeDir string
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	return nil
}, func(_ []byte) {
	// Ginkgo Globals
	SetDefaultEventuallyTimeout(5 * time.Second)

	// Setup common environment variables
	apiURL = os.Getenv("CF_API")
	turnOffColors()
})

var _ = SynchronizedAfterSuite(func() {},
	func() {
		setColor()
	})

var _ = BeforeEach(func() {
	setHomeDir()
	setSkipSSLValidation()
	setAPI()
})

var _ = AfterEach(func() {
	destroyHomeDir()
})

func setHomeDir() {
	var err error
	homeDir, err = ioutil.TempDir("", "cli-gats-test")
	Expect(err).NotTo(HaveOccurred())

	os.Setenv("CF_HOME", homeDir)
}

func setSkipSSLValidation() {
	if skip, err := strconv.ParseBool(os.Getenv("SKIP_SSL_VALIDATION")); err == nil && !skip {
		skipSSLValidation = ""
		return
	}
	skipSSLValidation = "--skip-ssl-validation"
}

func getAPI() string {
	if apiURL == "" {
		apiURL = "api.bosh-lite.com"
	}
	return apiURL
}

func setAPI() {
	Eventually(helpers.CF("api", getAPI(), skipSSLValidation)).Should(Exit(0))
}

func unsetAPI() {
	Eventually(helpers.CF("api", "--unset")).Should(Exit(0))
}

func destroyHomeDir() {
	if homeDir != "" {
		os.RemoveAll(homeDir)
	}
}

func turnOffColors() {
	originalColor = os.Getenv("CF_COLOR")
	os.Setenv("CF_COLOR", "false")
}

func setColor() {
	os.Setenv("CF_COLOR", originalColor)
}

func getCredentials() (string, string) {
	username := os.Getenv("CF_USERNAME")
	if username == "" {
		username = "admin"
	}
	password := os.Getenv("CF_PASSWORD")
	if password == "" {
		password = "admin"
	}
	return username, password
}

func loginCF() {
	username, password := getCredentials()
	Eventually(helpers.CF("auth", username, password)).Should(Exit(0))
}

func logoutCF() {
	Eventually(helpers.CF("logout")).Should(Exit(0))
}

func createOrgAndSpace(org string, space string) {
	Eventually(helpers.CF("create-org", org)).Should(Exit(0))
	Eventually(helpers.CF("create-space", space, "-o", org)).Should(Exit(0))
}

func createSpace(space string) {
	Eventually(helpers.CF("create-space", space)).Should(Exit(0))
}

func targetOrgAndSpace(org string, space string) {
	Eventually(helpers.CF("target", "-o", org, "-s", space)).Should(Exit(0))
}

func targetOrg(org string) {
	Eventually(helpers.CF("target", "-o", org)).Should(Exit(0))
}

func setupCF(org string, space string) {
	loginCF()
	createOrgAndSpace(org, space)
	targetOrgAndSpace(org, space)
}
