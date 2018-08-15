package configv3_test

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	. "code.cloudfoundry.org/cli/util/configv3"

	. "github.com/onsi/ginkgo"
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

	Describe("WriteConfig", func() {
		var config *Config

		BeforeEach(func() {
			config = &Config{
				ConfigFile: JSONConfig{
					ConfigVersion: 3,
					Target:        "foo.com",
					ColorEnabled:  "true",
				},
				ENV: EnvOverride{
					CFColor: "false",
				},
			}
		})

		When("no errors are encountered", func() {
			It("writes ConfigFile to homeDir/.cf/config.json", func() {
				err := WriteConfig(config)
				Expect(err).ToNot(HaveOccurred())

				file, err := ioutil.ReadFile(filepath.Join(homeDir, ".cf", "config.json"))
				Expect(err).ToNot(HaveOccurred())

				var writtenCFConfig JSONConfig
				err = json.Unmarshal(file, &writtenCFConfig)
				Expect(err).ToNot(HaveOccurred())

				Expect(writtenCFConfig.ConfigVersion).To(Equal(config.ConfigFile.ConfigVersion))
				Expect(writtenCFConfig.Target).To(Equal(config.ConfigFile.Target))
				Expect(writtenCFConfig.ColorEnabled).To(Equal(config.ConfigFile.ColorEnabled))
			})
		})
	})
})
