package experimental

import (
	"fmt"
	"testing"
	"time"

	"code.cloudfoundry.org/cli/v9/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	CFEventuallyTimeout   = 300 * time.Second
	CFConsistentlyTimeout = 500 * time.Millisecond
)

var (
	// Suite Level
	apiURL            string
	skipSSLValidation bool

	// Per Test Level
	homeDir string
)

func TestExperimental(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Experimental Integration Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	GinkgoWriter.Write([]byte("==============================Global FIRST Node Synchronized Before Each=============================="))
	SetDefaultEventuallyTimeout(CFEventuallyTimeout)
	SetDefaultConsistentlyDuration(CFConsistentlyTimeout)

	helpers.SetupSynchronizedSuite(func() {
		helpers.EnableFeatureFlag("diego_docker")
		helpers.EnableFeatureFlag("diego_cnb")
		helpers.EnableFeatureFlag("service_instance_sharing")
	})

	GinkgoWriter.Write([]byte("==============================End of Global FIRST Node Synchronized Before Each=============================="))

	return nil
}, func(_ []byte) {
	GinkgoWriter.Write([]byte(fmt.Sprintf("==============================Global Node %d Synchronized Before Each==============================", GinkgoParallelProcess())))
	// Ginkgo Globals
	SetDefaultEventuallyTimeout(CFEventuallyTimeout)
	SetDefaultConsistentlyDuration(CFConsistentlyTimeout)

	// Setup common environment variables
	helpers.TurnOffColors()

	// Enable Experimental Flag
	helpers.TurnOnExperimental()

	GinkgoWriter.Write([]byte(fmt.Sprintf("==============================End of Global Node %d Synchronized Before Each==============================", GinkgoParallelProcess())))
})

var _ = SynchronizedAfterSuite(func() {
	GinkgoWriter.Write([]byte(fmt.Sprintf("==============================Global Node %d Synchronized After Each==============================", GinkgoParallelProcess())))
	suiteHomeDir := helpers.SetHomeDir()
	helpers.SetAPI()
	helpers.LoginCF()
	helpers.DestroyHomeDir(suiteHomeDir)
	GinkgoWriter.Write([]byte(fmt.Sprintf("==============================End of Global Node %d Synchronized After Each==============================", GinkgoParallelProcess())))
}, func() {})

var _ = BeforeEach(func() {
	GinkgoWriter.Write([]byte("==============================Global Before Each=============================="))
	homeDir = helpers.SetHomeDir()
	apiURL, skipSSLValidation = helpers.SetAPI()

	GinkgoWriter.Write([]byte("==============================End of Global Before Each=============================="))
})

var _ = AfterEach(func() {
	GinkgoWriter.Write([]byte("==============================Global After Each=============================="))
	helpers.DestroyHomeDir(homeDir)
	GinkgoWriter.Write([]byte("==============================End of Global After Each=============================="))
})
