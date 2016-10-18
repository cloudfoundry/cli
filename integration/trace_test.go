package integration

import (
	"path"
	"runtime"

	. "code.cloudfoundry.org/cli/integration/helpers"

	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("trace", func() {
	var orgName string

	BeforeEach(func() {
		orgName = PrefixedRandomName("ORG")
		spaceName := PrefixedRandomName("SPACE")

		setupCF(orgName, spaceName)
	})

	AfterEach(func() {
		Eventually(CF("delete-org", "-f", orgName), CFLongTimeout).Should(Exit(0))
	})

	Context("writing the trace to the filesystem", func() {
		var (
			prevCfTrace string
			tempDir     string
			traceFile   string
		)

		BeforeEach(func() {
			prevCfTrace = os.Getenv("CF_TRACE")
			var err error
			// cannot use ioutil.TempFile because it sets the permissions to 0600
			// itself
			tempDir, err = ioutil.TempDir("", "cf-trace")
			Expect(err).ToNot(HaveOccurred())
			traceFile = path.Join(tempDir, "cf-trace")

			Expect(err).ToNot(HaveOccurred())
			os.Setenv("CF_TRACE", traceFile)
		})

		AfterEach(func() {
			err := os.Setenv("CF_TRACE", prevCfTrace)
			Expect(err).ToNot(HaveOccurred())
			err = os.RemoveAll(tempDir)
			Expect(err).ToNot(HaveOccurred())
		})

		It("creates the file with 0600 permission", func() {
			Eventually(CF("apps"), CFLongTimeout).Should(Exit(0))
			stat, err := os.Stat(traceFile)
			Expect(err).ToNot(HaveOccurred())
			if runtime.GOOS == "windows" {
				Expect(stat.Mode().String()).To(Equal(os.FileMode(0666).String()))
			} else {
				Expect(stat.Mode().String()).To(Equal(os.FileMode(0600).String()))
			}
		})
	})
})
