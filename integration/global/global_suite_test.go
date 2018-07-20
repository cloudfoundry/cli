package global

import (
	"fmt"
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

const (
	CFEventuallyTimeout   = 300 * time.Second
	CFConsistentlyTimeout = 500 * time.Millisecond
)

var (
	// Per Test Level
	homeDir       string
	ReadOnlyOrg   string
	ReadOnlySpace string
)

func TestGlobal(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Global Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	GinkgoWriter.Write([]byte("==============================Global FIRST Node Synchronized Before Each=============================="))
	// Ginkgo Globals
	SetDefaultEventuallyTimeout(CFEventuallyTimeout)
	SetDefaultConsistentlyDuration(CFConsistentlyTimeout)

	// Setup common environment variables
	helpers.TurnOffColors()

	helpers.SetupSynchronizedSuite(func() {
		helpers.EnableFeatureFlag("diego_docker")

		if helpers.IsVersionMet(ccversion.MinVersionShareServiceV3) {
			helpers.EnableFeatureFlag("service_instance_sharing")
		}
	})

	ReadOnlyOrg, ReadOnlySpace = helpers.SetupReadOnlyOrgAndSpace()

	GinkgoWriter.Write([]byte("==============================End of Global FIRST Node Synchronized Before Each=============================="))
	return nil
}, func(_ []byte) {
	GinkgoWriter.Write([]byte(fmt.Sprintf("==============================Global Node %d Synchronized Before Each==============================", GinkgoParallelNode())))
	if GinkgoParallelNode() != 1 {
		Fail("Test suite cannot run in parallel")
	}
	GinkgoWriter.Write([]byte(fmt.Sprintf("==============================End of Global Node %d Synchronized Before Each==============================", GinkgoParallelNode())))
})

var _ = BeforeEach(func() {
	GinkgoWriter.Write([]byte("==============================Global Before Each=============================="))
	homeDir = helpers.SetHomeDir()
	helpers.SetAPI()
	GinkgoWriter.Write([]byte("==============================End of Global Before Each=============================="))
})

var _ = AfterEach(func() {
	GinkgoWriter.Write([]byte("==============================Global After Each=============================="))
	helpers.DestroyHomeDir(homeDir)
	GinkgoWriter.Write([]byte("==============================End of Global After Each=============================="))
})
