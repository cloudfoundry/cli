package plugin

import (
	"os"
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
	reporters := []Reporter{}

	prBuilderReporter := helpers.GetPRBuilderReporter()
	if prBuilderReporter != nil {
		reporters = append(reporters, prBuilderReporter)
	}

	RunSpecsWithDefaultAndCustomReporters(t, "Plugin Suite", reporters)
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
	testPluginPath, err = Build("code.cloudfoundry.org/cli/integration/assets/test_plugin")
	Expect(err).ToNot(HaveOccurred())

	overrideTestPluginPath, err = Build("code.cloudfoundry.org/cli/integration/assets/test_plugin_with_command_overrides")
	Expect(err).ToNot(HaveOccurred())

	panicTestPluginPath, err = Build("code.cloudfoundry.org/cli/integration/assets/test_plugin_with_panic")
	Expect(err).ToNot(HaveOccurred())
})

var _ = SynchronizedAfterSuite(func() {
	CleanupBuildArtifacts()
},
	func() {
		outputRoot := os.Getenv(helpers.PRBuilderOutputEnvVar)
		if outputRoot != "" {
			helpers.WriteFailureSummary(outputRoot, "summary_isplugins.txt")
		}
	},
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
	helpers.CreateOrgAndSpace(org, space)
	helpers.TargetOrgAndSpace(org, space)
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
