package push

import (
	"io/ioutil"
	"os"
	"regexp"
	"testing"
	"time"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

const (
	CFEventuallyTimeout  = 300 * time.Second
	RealIsolationSegment = "persistent_isolation_segment"
	PushCommandName      = "push"
	PublicDockerImage    = "cloudfoundry/diego-docker-app-custom"
)

var (
	// Suite Level
	organization       string
	space              string
	foundDefaultDomain string
	realDir            string

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
	SetDefaultConsistentlyDuration(CFEventuallyTimeout)
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

	helpers.EnableDockerSupport()

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
	setupCF(organization, space)
})

var _ = AfterEach(func() {
	GinkgoWriter.Write([]byte("==============================Global After Each=============================="))
	helpers.QuickDeleteSpace(space)
	helpers.DestroyHomeDir(homeDir)
})

func defaultSharedDomain() string {
	// TODO: Move this into helpers when other packages need it, figure out how
	// to cache cuz this is a wacky call otherwise
	if foundDefaultDomain == "" {
		session := helpers.CF("domains")
		Eventually(session).Should(Exit(0))

		regex, err := regexp.Compile(`(.+?)\s+shared`)
		Expect(err).ToNot(HaveOccurred())

		matches := regex.FindStringSubmatch(string(session.Out.Contents()))
		Expect(matches).To(HaveLen(2))

		foundDefaultDomain = matches[1]
	}
	return foundDefaultDomain
}

func setupCF(org string, space string) {
	helpers.LoginCF()
	helpers.TargetOrg(org)
	helpers.CreateSpace(space)
	helpers.TargetOrgAndSpace(org, space)
}
