package plugin

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("plugin API", func() {
	BeforeEach(func() {
		installTestPlugin()
	})

	AfterEach(func() {
		uninstallTestPlugin()
	})

	Describe("AccessToken", func() {
		It("returns the access token", func() {
			confirmTestPluginOutput("AccessToken", `bearer [\w\d\.]+`)
		})
	})

	Describe("ApiEndpoint", func() {
		It("returns the API endpoint", func() {
			confirmTestPluginOutput("ApiEndpoint", apiURL)
		})
	})

	Describe("GetApp", func() {
		var appName string
		BeforeEach(func() {
			createTargetedOrgAndSpace()
			appName = helpers.PrefixedRandomName("APP")
			helpers.WithHelloWorldApp(func(appDir string) {
				Eventually(helpers.CF("push", appName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
			})
		})

		It("gets application information", func() {
			confirmTestPluginOutputWithArg("GetApp", appName, appName)
		})
	})

	Describe("GetCurrentSpace", func() {
		It("gets the current targeted Space", func() {
			_, space := createTargetedOrgAndSpace()
			confirmTestPluginOutput("GetCurrentSpace", space)
		})
	})

	Describe("Username", func() {
		It("gets the current user's name", func() {
			username := getUsername()
			confirmTestPluginOutput("Username", username)
		})
	})

})
