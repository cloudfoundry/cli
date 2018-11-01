package push

import (
	"fmt"
	"io/ioutil"
	"os"
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
	PushCommandName       = "push"
	PublicDockerImage     = "cloudfoundry/diego-docker-app-custom"
)

var (
	// Suite Level
	organization string
	space        string
	realDir      string

	// Per Test Level
	homeDir string
)

func TestPush(t *testing.T) {
	RegisterFailHandler(Fail)
	reporters := []Reporter{}

	honeyCombReporter := helpers.GetHoneyCombReporter("Push Integration Suite")

	if honeyCombReporter != nil {
		reporters = append(reporters, honeyCombReporter)
	}

	RunSpecsWithDefaultAndCustomReporters(t, "Push Integration Suite", reporters)
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

	homeDir = helpers.SetHomeDir()
	helpers.SetAPI()
	helpers.LoginCF()

	organization = helpers.NewOrgName()
	helpers.CreateOrg(organization)
	helpers.TargetOrg(organization)
	helpers.CreateSpace("empty-space")
	helpers.DestroyHomeDir(homeDir)

	var err error
	realDir, err = ioutil.TempDir("", "push-real-dir")
	Expect(err).ToNot(HaveOccurred())
	GinkgoWriter.Write([]byte(fmt.Sprintf("==============================End of Global Node %d Synchronized Before Each==============================", GinkgoParallelNode())))
})

var _ = SynchronizedAfterSuite(func() {
	GinkgoWriter.Write([]byte(fmt.Sprintf("==============================Global Node %d Synchronized After Each==============================", GinkgoParallelNode())))
	homeDir = helpers.SetHomeDir()
	helpers.SetAPI()
	helpers.LoginCF()
	helpers.QuickDeleteOrg(organization)
	Expect(os.RemoveAll(realDir)).ToNot(HaveOccurred())
	helpers.DestroyHomeDir(homeDir)
	GinkgoWriter.Write([]byte(fmt.Sprintf("==============================End of Global Node %d Synchronized After Each==============================", GinkgoParallelNode())))
}, func() {
})

var _ = BeforeEach(func() {
	GinkgoWriter.Write([]byte("==============================Global Before Each=============================="))
	homeDir = helpers.SetHomeDir()
	helpers.SetAPI()
	space = helpers.NewSpaceName()
	helpers.SetupCF(organization, space)
	GinkgoWriter.Write([]byte("==============================End of Global Before Each=============================="))
})

var _ = AfterEach(func() {
	GinkgoWriter.Write([]byte("==============================Global After Each=============================="))
	helpers.SetAPI()
	helpers.SetupCF(organization, space)
	helpers.QuickDeleteSpace(space)
	helpers.DestroyHomeDir(homeDir)
})
