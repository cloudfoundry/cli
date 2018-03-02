package push

import (
	"io/ioutil"
	"os"
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
	RunSpecs(t, "Push Integration Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	return nil
}, func(_ []byte) {
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

	helpers.SetupSynchronizedSuite(func() {
		helpers.EnableFeatureFlag("diego_docker")
		helpers.EnableFeatureFlag("service_instance_sharing")
	})

	var err error
	realDir, err = ioutil.TempDir("", "push-real-dir")
	Expect(err).ToNot(HaveOccurred())
})

var _ = SynchronizedAfterSuite(func() {
	helpers.SetAPI()
	helpers.LoginCF()
	helpers.QuickDeleteOrg(organization)
	Expect(os.RemoveAll(realDir)).ToNot(HaveOccurred())
}, func() {
})

var _ = BeforeEach(func() {
	homeDir = helpers.SetHomeDir()
	helpers.SetAPI()
	space = helpers.NewSpaceName()
	helpers.SetupCF(organization, space)
})

var _ = AfterEach(func() {
	GinkgoWriter.Write([]byte("==============================Global After Each=============================="))
	helpers.QuickDeleteSpace(space)
	helpers.DestroyHomeDir(homeDir)
})
