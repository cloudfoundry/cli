package isolated

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
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
	DockerImage           = "cloudfoundry/diego-docker-app-custom"
)

var (
	// Suite Level
	apiURL            string
	skipSSLValidation bool
	ReadOnlyOrg       string
	ReadOnlySpace     string

	// Per Test Level
	homeDir string
)

func TestIsolated(t *testing.T) {
	RegisterFailHandler(Fail)
	reporters := []Reporter{}

	prBuilderReporter := helpers.GetPRBuilderReporter()
	if prBuilderReporter != nil {
		reporters = append(reporters, prBuilderReporter)
	}

	RunSpecsWithDefaultAndCustomReporters(t, "Isolated Integration Suite", reporters)
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

	ReadOnlyOrg, ReadOnlySpace = helpers.SetupReadOnlyOrgAndSpace()
	GinkgoWriter.Write([]byte(fmt.Sprintf("==============================End of Global Node %d Synchronized Before Each==============================", GinkgoParallelNode())))
})

var _ = SynchronizedAfterSuite(func() {
	GinkgoWriter.Write([]byte(fmt.Sprintf("==============================Global Node %d Synchronized After Each==============================", GinkgoParallelNode())))
	homeDir = helpers.SetHomeDir()
	helpers.SetAPI()
	helpers.LoginCF()
	helpers.QuickDeleteOrg(ReadOnlyOrg)
	helpers.DestroyHomeDir(homeDir)
	GinkgoWriter.Write([]byte(fmt.Sprintf("==============================End of Global Node %d Synchronized After Each==============================", GinkgoParallelNode())))
}, func() {
	outputRoot := os.Getenv(helpers.PRBuilderOutputEnvVar)
	if outputRoot != "" {
		writeFailureSummary(outputRoot)
	}
})

func writeFailureSummary(outputRoot string) {
	outfile, err := os.Create(filepath.Join(outputRoot, "summary.txt"))
	failureSummaries := make([]string, 0)
	if err != nil {
		panic(err)
	}
	err = filepath.Walk(outputRoot, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		allFailures, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		failures := strings.Split(string(allFailures), "\n")
		failureSummaries = append(failureSummaries, failures...)
		return nil
	})
	if err != nil {
		panic(err)
	}
	sort.Strings(failureSummaries)
	anyFailures := false
	_, err = outfile.WriteString("### CI Run Summary:\nThe following failures were detected in the pipeline:\n```\n")
	var previousLine string
	for _, line := range failureSummaries {
		if line != "" && line != previousLine {
			anyFailures = true
			previousLine = line
			fmt.Fprintln(outfile, line)
		}
	}
	_, err = outfile.WriteString("```")
	if !anyFailures {
		err = os.Remove(filepath.Join(outputRoot, "summary.txt"))
	}
	if err != nil {
		panic(err)
	}
}

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
