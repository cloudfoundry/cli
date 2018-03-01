package isolated

import (
	"testing"
	"time"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	CFEventuallyTimeout   = 300 * time.Second
	CFConsistentlyTimeout = 500 * time.Millisecond
	RealIsolationSegment  = "persistent_isolation_segment"
	DockerImage           = "cloudfoundry/diego-docker-app-custom"
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
	SetDefaultConsistentlyDuration(CFConsistentlyTimeout)

	// Setup common environment variables
	helpers.TurnOffColors()

	helpers.SetupSynchronizedSuite(func() {
		helpers.EnableFeatureFlag("diego_docker")
		helpers.EnableFeatureFlag("service_instance_sharing")
	})

	ReadOnlyOrg, ReadOnlySpace = helpers.SetupReadOnlyOrgAndSpace()
})

var _ = SynchronizedAfterSuite(func() {
	helpers.SetAPI()
	helpers.LoginCF()
	helpers.QuickDeleteOrg(ReadOnlyOrg)
}, func() {
})

var _ = BeforeEach(func() {
	homeDir = helpers.SetHomeDir()
	apiURL, skipSSLValidation = helpers.SetAPI()
})

var _ = AfterEach(func() {
	GinkgoWriter.Write([]byte("==============================Global After Each=============================="))
	helpers.DestroyHomeDir(homeDir)
})

func setupCF(org string, space string) {
	helpers.LoginCF()
	if org != ReadOnlyOrg && space != ReadOnlySpace {
		helpers.CreateOrgAndSpace(org, space)
	}
	helpers.TargetOrgAndSpace(org, space)
}
