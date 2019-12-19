package isolated

import (
	helpers "code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/util/configv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	Describe("Version Management", func() {
		var oldTarget string
		var oldVersion int

		BeforeEach(func() {
			config := helpers.GetConfig()
			oldTarget = config.Target()
			oldVersion = config.ConfigFile.ConfigVersion
		})
		AfterEach(func() {
			helpers.SetConfig(func(config *configv3.Config) {
				config.ConfigFile.ConfigVersion = oldVersion
				config.ConfigFile.Target = oldTarget
			})
		})

		It("reset config to default if version mismatch", func() {
			helpers.SetConfig(func(config *configv3.Config) {
				config.ConfigFile.ConfigVersion = configv3.CurrentConfigVersion - 1
				config.ConfigFile.Target = "api.my-target"
			})
			helpers.LoginCF()
			config := helpers.GetConfig()
			Expect(config.ConfigFile.ConfigVersion).To(Equal(configv3.CurrentConfigVersion))
			Expect(config.ConfigFile.Target).To(Equal(""))
		})
	})
})
