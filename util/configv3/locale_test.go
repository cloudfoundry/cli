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

	DescribeTable("Locale",
		func(langVal string, lcAllVall string, configVal string, expected string) {
			rawConfig := fmt.Sprintf(`{"Locale":"%s"}`, configVal)
			setConfig(homeDir, rawConfig)

			defer os.Unsetenv("LANG")
			if langVal == "" {
				os.Unsetenv("LANG")
			} else {
				os.Setenv("LANG", langVal)
			}

			defer os.Unsetenv("LC_ALL")
			if lcAllVall == "" {
				os.Unsetenv("LC_ALL")
			} else {
				os.Setenv("LC_ALL", lcAllVall)
			}

			config, err := LoadConfig()
			Expect(err).ToNot(HaveOccurred())
			Expect(config).ToNot(BeNil())

			Expect(config.Locale()).To(Equal(expected))
		},

		Entry("LANG=ko-KO.UTF-8 LC_ALL=empty       config=empty ko-KO", "ko-KO.UTF-8", "", "", "ko-KO"),
		Entry("LANG=ko-KO.UTF-8 LC_ALL=fr_FR.UTF-8 config=empty fr-FR", "ko-KO.UTF-8", "fr_FR.UTF-8", "", "fr-FR"),
		Entry("LANG=ko-KO.UTF-8 LC_ALL=fr_FR.UTF-8 config=pt-BR pt-BR", "ko-KO.UTF-8", "fr_FR.UTF-8", "pt-BR", "pt-BR"),

		Entry("config=empty LANG=empty       LC_ALL=empty       default", "", "", "", DefaultLocale),
	)
})
