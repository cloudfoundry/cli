package performance_test

import (
	"fmt"
	"testing"
	"time"

	"code.cloudfoundry.org/cli/integration/helpers"
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
	perfOrg           string
	perfSpace         string

	// Per Test Level
	homeDir string
)

// TODO: Performance tests should probably run serially and in the same order every time

func TestPerformance(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Performance Integration Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	GinkgoWriter.Write([]byte("==============================Global FIRST Node Synchronized Before Each=============================="))
	GinkgoWriter.Write([]byte("==============================End of Global FIRST Node Synchronized Before Each=============================="))

	return nil
}, func(_ []byte) {
	GinkgoWriter.Write([]byte(fmt.Sprintf("==============================Global Node %d Synchronized Before Each==============================", GinkgoParallelNode())))
	// Ginkgo Globals
	SetDefaultEventuallyTimeout(CFEventuallyTimeout)
	SetDefaultConsistentlyDuration(CFConsistentlyTimeout)

	// Setup common environment variables
	helpers.TurnOffColors()
	perfOrg, perfSpace = helpers.SetupReadOnlyOrgAndSpace()

	GinkgoWriter.Write([]byte(fmt.Sprintf("==============================End of Global Node %d Synchronized Before Each==============================", GinkgoParallelNode())))
})

var _ = SynchronizedAfterSuite(func() {
	GinkgoWriter.Write([]byte(fmt.Sprintf("==============================Global Node %d Synchronized After Each==============================", GinkgoParallelNode())))
	homeDir = helpers.SetHomeDir()
	helpers.SetAPI()
	helpers.LoginCF()
	helpers.QuickDeleteOrg(perfOrg)
	helpers.DestroyHomeDir(homeDir)
	GinkgoWriter.Write([]byte(fmt.Sprintf("==============================End of Global Node %d Synchronized After Each==============================", GinkgoParallelNode())))
}, func() {})

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
