package configv3_test

import (
	"os"
	"time"

	. "code.cloudfoundry.org/cli/util/configv3"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	var (
		config  *Config
		homeDir string
	)

	BeforeEach(func() {
		homeDir = setup()
	})

	AfterEach(func() {
		teardown(homeDir)
	})

	Context("when there are environment variables", func() {
		BeforeEach(func() {
			Expect(os.Setenv("CF_DIAL_TIMEOUT", "1234")).ToNot(HaveOccurred())
			Expect(os.Setenv("CF_DOCKER_PASSWORD", "banana")).ToNot(HaveOccurred())
			Expect(os.Setenv("CF_PASSWORD", "I am password.")).ToNot(HaveOccurred())
			Expect(os.Setenv("CF_STAGING_TIMEOUT", "8675")).ToNot(HaveOccurred())
			Expect(os.Setenv("CF_STARTUP_TIMEOUT", "309")).ToNot(HaveOccurred())
			Expect(os.Setenv("CF_USERNAME", "i-R-user")).ToNot(HaveOccurred())
			Expect(os.Setenv("https_proxy", "proxy.com")).ToNot(HaveOccurred())

			var err error
			config, err = LoadConfig()
			Expect(err).ToNot(HaveOccurred())
			Expect(config).ToNot(BeNil())
		})

		AfterEach(func() {
			Expect(os.Unsetenv("CF_DIAL_TIMEOUT")).ToNot(HaveOccurred())
			Expect(os.Unsetenv("CF_DOCKER_PASSWORD")).ToNot(HaveOccurred())
			Expect(os.Unsetenv("CF_PASSWORD")).ToNot(HaveOccurred())
			Expect(os.Unsetenv("CF_STAGING_TIMEOUT")).ToNot(HaveOccurred())
			Expect(os.Unsetenv("CF_STARTUP_TIMEOUT")).ToNot(HaveOccurred())
			Expect(os.Unsetenv("CF_USERNAME")).ToNot(HaveOccurred())
			Expect(os.Unsetenv("https_proxy")).ToNot(HaveOccurred())
		})

		It("overrides specific config values", func() {
			Expect(config.CFUsername()).To(Equal("i-R-user"))
			Expect(config.CFPassword()).To(Equal("I am password."))
			Expect(config.DialTimeout()).To(Equal(1234 * time.Second))
			Expect(config.DockerPassword()).To(Equal("banana"))
			Expect(config.HTTPSProxy()).To(Equal("proxy.com"))
			Expect(config.StagingTimeout()).To(Equal(time.Duration(8675) * time.Minute))
			Expect(config.StartupTimeout()).To(Equal(time.Duration(309) * time.Minute))
		})
	})

	Describe("BinaryName", func() {
		It("returns the name used to invoke", func() {
			config, err := LoadConfig()
			Expect(err).ToNot(HaveOccurred())
			Expect(config).ToNot(BeNil())

			// Ginkgo will uses a config file as the first test argument, so that
			// will be considered the binary name
			Expect(config.BinaryName()).To(Equal("configv3.test"))
		})
	})

	Describe("BinaryVersion", func() {
		It("returns back version.BinaryVersion", func() {
			conf := Config{}
			Expect(conf.BinaryVersion()).To(Equal("0.0.0-unknown-version"))
		})
	})

	DescribeTable("Experimental",
		func(envVal string, expected bool) {
			setConfig(homeDir, `{}`)

			defer os.Unsetenv("CF_CLI_EXPERIMENTAL")
			Expect(os.Unsetenv("CF_CLI_EXPERIMENTAL")).ToNot(HaveOccurred())
			if envVal != "" {
				Expect(os.Setenv("CF_CLI_EXPERIMENTAL", envVal)).ToNot(HaveOccurred())
			}

			config, err := LoadConfig()
			Expect(err).ToNot(HaveOccurred())
			Expect(config).ToNot(BeNil())

			Expect(config.Experimental()).To(Equal(expected))
		},

		Entry("uses default value of false if environment value is not set", "", false),
		Entry("uses environment value if a valid environment value is set", "true", true),
		Entry("uses default value of false if an invalid environment value is set", "something-invalid", false),
	)

	DescribeTable("LogLevel",
		func(envVal string, expectedLevel int) {
			config := Config{ENV: EnvOverride{CFLogLevel: envVal}}
			Expect(config.LogLevel()).To(Equal(expectedLevel))
		},

		Entry("Default to 0", "", 0),
		Entry("panic returns 0", "panic", 0),
		Entry("fatal returns 1", "fatal", 1),
		Entry("error returns 2", "error", 2),
		Entry("warn returns 3", "warn", 3),
		Entry("info returns 4", "info", 4),
		Entry("debug returns 5", "debug", 5),
		Entry("dEbUg returns 5", "dEbUg", 5),
	)
})
