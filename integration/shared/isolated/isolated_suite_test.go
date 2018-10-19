package isolated

import (
	"fmt"
	"testing"
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
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
	reporters := []Reporter{}

	honeyCombReporter := helpers.GetHoneyCombReporter("Isolated Test Suit")

	if honeyCombReporter != nil {
		reporters = append(reporters, honeyCombReporter)
	}

	RunSpecsWithDefaultAndCustomReporters(t, "Isolated Integration Suite", reporters)
}

var _ = SynchronizedBeforeSuite(func() []byte {
	GinkgoWriter.Write([]byte("==============================Global FIRST Node Synchronized Before Each=============================="))
	SetDefaultEventuallyTimeout(CFEventuallyTimeout)
	SetDefaultConsistentlyDuration(CFConsistentlyTimeout)

	helpers.SetupSynchronizedSuite(func() {
		helpers.EnableFeatureFlag("diego_docker")

		if helpers.IsVersionMet(ccversion.MinVersionShareServiceV3) {
			helpers.EnableFeatureFlag("service_instance_sharing")
		}
	})
	GinkgoWriter.Write([]byte("==============================End of Global FIRST Node Synchronized Before Each=============================="))

	return nil
}, func(_ []byte) {
	GinkgoWriter.Write([]byte(fmt.Sprintf("==============================Global Node %d Synchronized Before Each==============================", GinkgoParallelNode())))
	// Ginkgo Globals
	SetDefaultEventuallyTimeout(CFEventuallyTimeout)
	SetDefaultConsistentlyDuration(CFConsistentlyTimeout)

	// Setup common environment variables
	helpers.TurnOffColors()

	ReadOnlyOrg, ReadOnlySpace = helpers.SetupReadOnlyOrgAndSpace()
	GinkgoWriter.Write([]byte(fmt.Sprintf("==============================End of Global Node %d Synchronized Before Each==============================", GinkgoParallelNode())))
})

var _ = SynchronizedAfterSuite(func() {
	GinkgoWriter.Write([]byte(fmt.Sprintf("==============================Global Node %d Synchronized After Each==============================", GinkgoParallelNode())))
	homeDir = helpers.SetHomeDir()
	helpers.SetAPI()
	helpers.LoginCF()
	helpers.QuickDeleteOrg(ReadOnlyOrg)
	helpers.DestroyHomeDir(homeDir)
	GinkgoWriter.Write([]byte(fmt.Sprintf("==============================End of Global Node %d Synchronized After Each==============================", GinkgoParallelNode())))
}, func() {
})

var _ = BeforeEach(func() {
	GinkgoWriter.Write([]byte("==============================Global Before Each=============================="))
	homeDir = helpers.SetHomeDir()
	apiURL, skipSSLValidation = helpers.SetAPI()
	GinkgoWriter.Write([]byte("==============================End of Global Before Each=============================="))
})

var _ = AfterEach(func() {
	GinkgoWriter.Write([]byte("==============================Global After Each==============================\n"))
	helpers.DestroyHomeDir(homeDir)
	GinkgoWriter.Write([]byte("==============================End of Global After Each=============================="))
})
