package configv3_test

import (
	"fmt"
	"os"

	. "code.cloudfoundry.org/cli/util/configv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	var homeDir string

	BeforeEach(func() {
		homeDir = setup()
	})

	AfterEach(func() {
		teardown(homeDir)
	})

	DescribeTable("ColorEnabled",
		func(configVal string, envVal string, expected ColorSetting) {
			rawConfig := fmt.Sprintf(`{"ColorEnabled":"%s"}`, configVal)
			setConfig(homeDir, rawConfig)

			defer os.Unsetenv("CF_COLOR")
			if envVal == "" {
				os.Unsetenv("CF_COLOR")
			} else {
				os.Setenv("CF_COLOR", envVal)
			}

			config, err := LoadConfig()
			Expect(err).ToNot(HaveOccurred())
			Expect(config).ToNot(BeNil())

			Expect(config.ColorEnabled()).To(Equal(expected))
		},
		Entry("config=true  env=true  enabled", "true", "true", ColorEnabled),
		Entry("config=true  env=false disabled", "true", "false", ColorDisabled),
		Entry("config=false env=true  enabled", "false", "true", ColorEnabled),
		Entry("config=false env=false disabled", "false", "false", ColorDisabled),

		Entry("config=unset env=false disabled", "", "false", ColorDisabled),
		Entry("config=unset env=true  enabled", "", "true", ColorEnabled),
		Entry("config=false env=unset disabled", "false", "", ColorDisabled),
		Entry("config=true  env=unset disabled", "true", "", ColorEnabled),

		Entry("config=unset env=unset falls back to default", "", "", ColorEnabled),
	)
})
