package isolated

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("create-app-manifest command", func() {
	Context("when app has no hostname", func() {
		var (
			domain  helpers.Domain
			appName string
			tmpDir  string
		)
		BeforeEach(func() {
			orgName := helpers.NewOrgName()
			spaceName := helpers.PrefixedRandomName("SPACE")

			setupCF(orgName, spaceName)
			domain = helpers.NewDomain(orgName, helpers.DomainName(""))
			domain.Create()

			appName = "hello"
			helpers.WithHelloWorldApp(func(appDir string) {
				Eventually(helpers.CF("push", appName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-hostname", "-d", domain.Name)).Should(Exit(0))
			})

			var err error
			tmpDir, err = ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())

		})
		AfterEach(func() {
			os.RemoveAll(tmpDir)
		})

		It("contains routes without hostnames", func() {
			manifestPath := filepath.Join(tmpDir, "manifest.yml")
			session := helpers.CF("create-app-manifest", appName, "-p", manifestPath)
			Eventually(session).Should(Say("Manifest file created successfully"))
			Eventually(session).Should(Exit(0))

			manifestContents, err := ioutil.ReadFile(manifestPath)
			Expect(err).NotTo(HaveOccurred())

			manifest := string(manifestContents)
			Expect(manifest).To(ContainSubstring("routes"))
			Expect(manifest).To(ContainSubstring(domain.Name))
		})
	})
})
