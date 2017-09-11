package isolated

import (
	"fmt"
	"io/ioutil"
	"os"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Push with manifest", func() {
	var (
		appName  string
		tempFile string
	)

	const (
		dockerImage = "cloudfoundry/diego-docker-app-custom"
	)

	BeforeEach(func() {
		orgName := helpers.NewOrgName()
		spaceName := helpers.NewSpaceName()
		setupCF(orgName, spaceName)

		appName = helpers.PrefixedRandomName("app")
		f, err := ioutil.TempFile("", "combination-manifest-with-p")
		Expect(err).ToNot(HaveOccurred())
		Expect(f.Close()).To(Succeed())
		tempFile = f.Name()
	})

	AfterEach(func() {
		Expect(os.Remove(tempFile)).ToNot(HaveOccurred())
	})

	Context("when the specified manifest file does not exist", func() {
		It("displays a path does not exist error, help, and exits 1", func() {
			session := helpers.CF("push", "-f", "./non-existent-file")
			Eventually(session.Err).Should(Say("Incorrect Usage: The specified path './non-existent-file' does not exist."))
			Eventually(session.Out).Should(Say("NAME:"))
			Eventually(session.Out).Should(Say("USAGE:"))
			Eventually(session).Should(Exit(1))
		})
	})

	Context("manifest contains 'docker' and the '-o' flag is provided", func() {
		BeforeEach(func() {
			manifestContents := []byte(fmt.Sprintf(`
---
applications:
- name: %s
  docker:
    image: some-image
`, appName))
			Expect(ioutil.WriteFile(tempFile, manifestContents, 0666)).To(Succeed())
		})

		It("overrides 'docker' in the manifest with the '-o' flag value", func() {
			helpers.WithHelloWorldApp(func(dir string) {
				Eventually(helpers.CF("push", "-o", dockerImage, "-f", tempFile)).Should(Exit(0))

				appGUID := helpers.AppGUID(appName)
				// TODO: replace this with 'cf app' once #146661157 is complete
				session := helpers.CF("curl", fmt.Sprintf("/v2/apps/%s", appGUID))
				Eventually(session.Out).Should(Say(dockerImage))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("manifest contains both 'buildpack' and 'docker'", func() {
		BeforeEach(func() {
			manifestContents := []byte(fmt.Sprintf(`
---
applications:
- name: %s
  buildpack: staticfile_buildpack
  docker:
   image: %s
`, appName, dockerImage))
			Expect(ioutil.WriteFile(tempFile, manifestContents, 0666)).To(Succeed())
		})

		It("displays an error and exits 1", func() {
			helpers.WithHelloWorldApp(func(dir string) {
				session := helpers.CF("push", appName, "-f", tempFile)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Say("Invalid application configuration:"))
				Eventually(session).Should(Say("Application %s must not be configured with both 'buildpack' and 'docker'", appName))
				Eventually(session).Should(Exit(1))
			})
		})
	})

	Context("manifest contains both 'docker' and 'path'", func() {
		BeforeEach(func() {
			manifestContents := []byte(fmt.Sprintf(`
---
applications:
- name: %s
  path: .
  docker:
   image: %s
`, appName, dockerImage))
			Expect(ioutil.WriteFile(tempFile, manifestContents, 0666)).To(Succeed())
		})

		It("displays an error and exits 1", func() {
			helpers.WithHelloWorldApp(func(dir string) {
				session := helpers.CF("push", appName, "-f", tempFile)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Say("Invalid application configuration:"))
				Eventually(session).Should(Say("Application %s must not be configured with both 'docker' and 'path'", appName))
				Eventually(session).Should(Exit(1))
			})
		})
	})
})
