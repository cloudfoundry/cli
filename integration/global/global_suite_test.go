package global

import (
	"time"

	"code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

const (
	CFEventuallyTimeout = 30 * time.Second
)

var (
	// Per Test Level
	homeDir string
)

func TestGlobal(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Global Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	// Ginkgo Globals
	SetDefaultEventuallyTimeout(CFEventuallyTimeout)

	// Setup common environment variables
	helpers.TurnOffColors()
	return nil
}, func(_ []byte) {
	if GinkgoParallelNode() != 1 {
		Fail("Test suite cannot run in parallel")
	}
})

var _ = BeforeEach(func() {
	homeDir = helpers.SetHomeDir()
	helpers.SetAPI()
})

var _ = AfterEach(func() {
	helpers.DestroyHomeDir(homeDir)
})

func setupCF(org string, space string) {
	helpers.LoginCF()
	helpers.CreateOrgAndSpace(org, space)
	helpers.TargetOrgAndSpace(org, space)
}
