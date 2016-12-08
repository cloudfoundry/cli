package isolated

import (
	"regexp"
	"testing"
	"time"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

const (
	CFEventuallyTimeout = 60 * time.Second
)

var (
	// Suite Level
	apiURL            string
	skipSSLValidation string
	ReadOnlyOrg       string
	ReadOnlySpace     string

	// Per Test Level
	homeDir string
)

func TestIsolated(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Isolated Integration Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	return nil
}, func(_ []byte) {
	// Ginkgo Globals
	SetDefaultEventuallyTimeout(CFEventuallyTimeout)

	// Setup common environment variables
	helpers.TurnOffColors()

	ReadOnlyOrg, ReadOnlySpace = helpers.SetupReadOnlyOrgAndSpace()
})

var _ = BeforeEach(func() {
	homeDir = helpers.SetHomeDir()
	apiURL, skipSSLValidation = helpers.SetAPI()
})

var _ = AfterEach(func() {
	helpers.DestroyHomeDir(homeDir)
})

var foundDefaultDomain string

func defaultSharedDomain() string {
	// TODO: Move this into helpers when other packages need it, figure out how
	// to cache cuz this is a wacky call otherwise
	if foundDefaultDomain == "" {
		session := helpers.CF("domains")
		Eventually(session).Should(Exit(0))

		regex, err := regexp.Compile(`(.+?)\s+shared`)
		Expect(err).ToNot(HaveOccurred())

		matches := regex.FindStringSubmatch(string(session.Out.Contents()))
		Expect(matches).To(HaveLen(2))

		foundDefaultDomain = matches[1]
	}
	return foundDefaultDomain
}

func setupCF(org string, space string) {
	helpers.LoginCF()
	if org != ReadOnlyOrg && space != ReadOnlySpace {
		helpers.CreateOrgAndSpace(org, space)
	}
	helpers.TargetOrgAndSpace(org, space)
}
