package configv3_test

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"

	"code.cloudfoundry.org/cli/util/configv3"
	"gopkg.in/yaml.v2"

	. "github.com/onsi/ginkgo/v2"
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
		var (
			config *configv3.Config
			file   []byte
		)

		BeforeEach(func() {
			config = &configv3.Config{
				ConfigFile: configv3.JSONConfig{
					ConfigVersion: 3,
					Target:        "foo.com",
					ColorEnabled:  "true",
				},
				ENV: configv3.EnvOverride{
					CFColor: "false",
				},
			}
			err := config.WriteConfig()
			Expect(err).ToNot(HaveOccurred())

			file, err = ioutil.ReadFile(filepath.Join(homeDir, ".cf", "config.json"))
			Expect(err).ToNot(HaveOccurred())
		})

		It("writes ConfigFile to homeDir/.cf/config.json", func() {
			var writtenCFConfig configv3.JSONConfig
			err := json.Unmarshal(file, &writtenCFConfig)
			Expect(err).ToNot(HaveOccurred())

			Expect(writtenCFConfig.ConfigVersion).To(Equal(config.ConfigFile.ConfigVersion))
			Expect(writtenCFConfig.Target).To(Equal(config.ConfigFile.Target))
			Expect(writtenCFConfig.ColorEnabled).To(Equal(config.ConfigFile.ColorEnabled))
		})

		It("writes the top-level keys in alphabetical order", func() {
			// we use yaml.MapSlice here to preserve the original order
			// https://github.com/golang/go/issues/27179
			var ms yaml.MapSlice
			err := yaml.Unmarshal(file, &ms)
			Expect(err).ToNot(HaveOccurred())

			keys := make([]string, len(ms))
			for i, item := range ms {
				keys[i] = item.Key.(string)
			}
			caseInsensitive := func(i, j int) bool { return strings.ToLower(keys[i]) < strings.ToLower(keys[j]) }
			Expect(sort.SliceIsSorted(keys, caseInsensitive)).To(BeTrue())
		})
	})
})
