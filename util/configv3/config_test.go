package configv3_test

import (
	"os"

	. "code.cloudfoundry.org/cli/util/configv3"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	var (
		homeDir string
		config  *Config
	)

	BeforeEach(func() {
		homeDir = setup()
	})

	AfterEach(func() {
		teardown(homeDir)
	})

	Describe("IsTTY", func() {
		BeforeEach(func() {
			Expect(os.Setenv("FORCE_TTY", "true")).ToNot(HaveOccurred())

			var err error
			config, err = LoadConfig()
			Expect(err).ToNot(HaveOccurred())
			Expect(config).ToNot(BeNil())
		})

		AfterEach(func() {
			Expect(os.Unsetenv("FORCE_TTY")).ToNot(HaveOccurred())
		})

		It("overrides specific config values", func() {
			Expect(config.IsTTY()).To(BeTrue())
		})
	})

	Describe("WriteConfig", func() {
		It("Writes to the config file", func() {
			Expect(true).To(BeFalse())

		})
	})
})
