package push

import (
	"path/filepath"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("displaying manifest differences between pushes", func() {
	var (
		appName string
	)

	BeforeEach(func() {
		appName = helpers.PrefixedRandomName("app")
	})
	When("there is an app that has been pushed and we push it again with a different manifest", func() {
		It("displays the diff from the manifest", func() {
			helpers.WithHelloWorldApp(func(dir string) {

				pathToManifest := filepath.Join(dir, "manifest.yml")
				helpers.WriteManifest(pathToManifest, map[string]interface{}{
					"applications": []map[string]interface{}{
						{
							"name":       appName,
							"disk_quota": "1GB",
						},
					},
				})

				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName)
				Eventually(session).Should(Exit(0))
				Expect(session).To(Say("name:\\s+%s", appName))

				helpers.WriteManifest(pathToManifest, map[string]interface{}{
					"applications": []map[string]interface{}{
						{
							"name":       appName,
							"disk_quota": "2GB",
						},
					},
				})

				session = helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName)
				Eventually(session).Should(Exit(0))
				Expect(session).To(Say("name:\\s+%s", appName))
				Expect(session).To(Say(`\-   disk-quota: 1024M`))
				Expect(session).To(Say(`\+   disk-quota: 2028M`))
			})
		})
	})
})
