package push

import (
	"fmt"
	"path/filepath"

	"code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("push with buildpacks in manifest", func() {
	var (
		appName string
	)

	BeforeEach(func() {
		appName = helpers.NewAppName()
	})

	Context("when buildpacks (plural) is provided", func() {
		It("sets all buildpacks correctly for the pushed app", func() {
			helpers.WithHelloWorldApp(func(dir string) {
				helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
					"applications": []map[string]interface{}{
						{
							"name": appName,
							"buildpacks": []string{
								"https://github.com/cloudfoundry/ruby-buildpack",
								"https://github.com/cloudfoundry/staticfile-buildpack",
							},
						},
					},
				})
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName)
				Eventually(session).Should(Exit(0))
			})

			session := helpers.CF("curl", fmt.Sprintf("v3/apps/%s", helpers.AppGUID(appName)))

			Eventually(session).Should(Say(`https://github.com/cloudfoundry/ruby-buildpack"`))
			Eventually(session).Should(Say(`https://github.com/cloudfoundry/staticfile-buildpack"`))
			Eventually(session).Should(Exit(0))
		})
	})
})
