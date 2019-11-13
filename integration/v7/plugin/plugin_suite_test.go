package plugin

import (
	"testing"
	"time"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

const (
	CFEventuallyTimeout   = 180 * time.Second
	CFConsistentlyTimeout = 500 * time.Millisecond
)

var (
	// Suite Level
	testPluginPath         string
	overrideTestPluginPath string
	panicTestPluginPath    string
	apiURL                 string
	skipSSLValidation      bool

	// Per Test Level
	homeDir string
)

func TestPlugin(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Plugin Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	return nil
}, func(path []byte) {
	// Ginkgo Globals
	SetDefaultEventuallyTimeout(CFEventuallyTimeout)
	SetDefaultConsistentlyDuration(CFConsistentlyTimeout)

	// Setup common environment variables
	helpers.TurnOffColors()

	var err error
	testPluginPath, err = Build("code.cloudfoundry.org/cli/integration/assets/test_plugin_v7", "-tags=V7")
	Expect(err).ToNot(HaveOccurred())

	overrideTestPluginPath, err = Build("code.cloudfoundry.org/cli/integration/assets/test_plugin_with_command_overrides")
	Expect(err).ToNot(HaveOccurred())

	panicTestPluginPath, err = Build("code.cloudfoundry.org/cli/integration/assets/test_plugin_with_panic_v7", "-tags=V7")
	Expect(err).ToNot(HaveOccurred())
})

var _ = SynchronizedAfterSuite(func() {
	CleanupBuildArtifacts()
},
	func() {},
)

var _ = BeforeEach(func() {
	homeDir = helpers.SetHomeDir()
	apiURL, skipSSLValidation = helpers.SetAPI()
	helpers.LoginCF()
	Eventually(helpers.CF("remove-plugin-repo", "CF-Community")).Should(Exit(0))
})

var _ = AfterEach(func() {
	GinkgoWriter.Write([]byte("==============================Global After Each=============================="))
	helpers.DestroyHomeDir(homeDir)
})

func installTestPlugin() {
	session := helpers.CF("install-plugin", "-f", testPluginPath)
	Eventually(session).Should(Exit(0))
}

func uninstallTestPlugin() {
	session := helpers.CF("uninstall-plugin", "CF-CLI-Integration-Test-Plugin")
	Eventually(session).Should(Exit(0))
}

func createTargetedOrgAndSpace() (string, string) {
	org := helpers.NewOrgName()
	space := helpers.NewSpaceName()
	username, _ := helpers.GetCredentials()
	helpers.CreateOrgAndSpace(org, space)
	helpers.TargetOrgAndSpace(org, space)
	helpers.SetOrgRole(username, org, "OrgManager", helpers.ClientCredentialsTestMode())
	helpers.SetSpaceRole(username, org, space, "SpaceDeveloper", helpers.ClientCredentialsTestMode())
	helpers.SetSpaceRole(username, org, space, "SpaceManager", helpers.ClientCredentialsTestMode())
	return org, space
}

func confirmTestPluginOutput(command string, output ...string) {
	session := helpers.CF(command)
	for _, val := range output {
		Eventually(session).Should(Say(val))
	}
	Eventually(session).Should(Exit(0))
}

func confirmTestPluginOutputWithArg(command string, arg string, output ...string) {
	session := helpers.CF(command, arg)
	for _, val := range output {
		Eventually(session).Should(Say(val))
	}
	Eventually(session).Should(Exit(0))
}
