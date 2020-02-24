package commonisolated

import (
	"fmt"
	"testing"
	"time"

	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/integration/helpers/fakeservicebroker"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	CFEventuallyTimeout   = 300 * time.Second
	CFConsistentlyTimeout = 500 * time.Millisecond
	RealIsolationSegment  = "persistent_isolation_segment"
	DockerImage           = "cloudfoundry/diego-docker-app-custom"
)

func CommonTestIsolated(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Isolated Integration Suite")
}

func CommonGinkgoSetup(
	// Per suite Level
	failureSummaryFilename string,
	apiURL *string,
	skipSSLValidation *bool,
	readOnlyOrg *string,
	readOnlySpace *string,

	// Per test level
	homeDir *string,
) interface{} {
	var _ = SynchronizedBeforeSuite(func() []byte {
		_, _ = GinkgoWriter.Write([]byte("==============================Global FIRST Node Synchronized Before Each=============================="))
		SetDefaultEventuallyTimeout(CFEventuallyTimeout)
		SetDefaultConsistentlyDuration(CFConsistentlyTimeout)

		helpers.SetupSynchronizedSuite(func() {
			helpers.EnableFeatureFlag("diego_docker")
		})

		fakeservicebroker.Setup()

		_, _ = GinkgoWriter.Write([]byte("==============================End of Global FIRST Node Synchronized Before Each=============================="))

		return nil
	}, func(_ []byte) {
		_, _ = GinkgoWriter.Write([]byte(fmt.Sprintf("==============================Global Node %d Synchronized Before Each==============================", GinkgoParallelNode())))
		// Ginkgo Globals
		SetDefaultEventuallyTimeout(CFEventuallyTimeout)
		SetDefaultConsistentlyDuration(CFConsistentlyTimeout)

		// Setup common environment variables
		helpers.TurnOffColors()

		*readOnlyOrg, *readOnlySpace = helpers.SetupReadOnlyOrgAndSpace()
		_, _ = GinkgoWriter.Write([]byte(fmt.Sprintf("==============================End of Global Node %d Synchronized Before Each==============================", GinkgoParallelNode())))
	})

	var _ = SynchronizedAfterSuite(func() {
		_, _ = GinkgoWriter.Write([]byte(fmt.Sprintf("==============================Global Node %d Synchronized After Each==============================", GinkgoParallelNode())))
		*homeDir = helpers.SetHomeDir()
		helpers.SetAPI()
		helpers.LoginCF()
		fakeservicebroker.Cleanup()
		helpers.QuickDeleteOrg(*readOnlyOrg)
		helpers.DestroyHomeDir(*homeDir)
		_, _ = GinkgoWriter.Write([]byte(fmt.Sprintf("==============================End of Global Node %d Synchronized After Each==============================", GinkgoParallelNode())))
	}, func() {})

	var _ = BeforeEach(func() {
		_, _ = GinkgoWriter.Write([]byte("==============================Global Before Each=============================="))
		*homeDir = helpers.SetHomeDir()
		*apiURL, *skipSSLValidation = helpers.SetAPI()
		_, _ = GinkgoWriter.Write([]byte("==============================End of Global Before Each=============================="))
	})

	var _ = AfterEach(func() {
		_, _ = GinkgoWriter.Write([]byte("==============================Global After Each==============================\n"))
		helpers.DestroyHomeDir(*homeDir)
		_, _ = GinkgoWriter.Write([]byte("==============================End of Global After Each=============================="))
	})

	return nil
}
