package readonly

import (
	"fmt"
	"testing"
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
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
	var reporters []Reporter

	RunSpecsWithDefaultAndCustomReporters(t, "Read Only Integration Suite", reporters)
}

func CommonGinkgoSetup(
	// Per suite Level
	failureSummaryFilename string,
	apiURL *string,
	skipSSLValidation *bool,

	// Per test level
	homeDir *string,
) interface{} {
	var _ = SynchronizedBeforeSuite(func() {
		_, _ = GinkgoWriter.Write([]byte("==============================Global Read Only Setup=============================="))
		SetDefaultEventuallyTimeout(CFEventuallyTimeout)
		SetDefaultConsistentlyDuration(CFConsistentlyTimeout)

		helpers.SetupSynchronizedSuite(func() {
			helpers.EnableFeatureFlag("diego_docker")

			if helpers.IsVersionMet(ccversion.MinVersionShareServiceV3) {
				helpers.EnableFeatureFlag("service_instance_sharing")
			}
		})

		fakeservicebroker.Setup()

		err := SetupReadOnlySuite()
		if err != nil {
			painc("Read only setup has failed!")
		}

		_, _ = GinkgoWriter.Write([]byte("==============================End of Global Read Only Setup=============================="))
	}, nil)

	var _ = SynchronizedAfterSuite(func() {
		_, _ = GinkgoWriter.Write([]byte(fmt.Sprintf("==============================Global Node %d Synchronized After Each==============================", GinkgoParallelNode())))
		*homeDir = helpers.SetHomeDir()
		helpers.SetAPI()
		helpers.LoginCF()
		fakeservicebroker.Cleanup()


		helpers.DestroyHomeDir(*homeDir)
		_, _ = GinkgoWriter.Write([]byte(fmt.Sprintf("==============================End of Global Node %d Synchronized After Each==============================", GinkgoParallelNode())))
	}, func() {
		CleanupReadOnlyIntegration()
	})

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
