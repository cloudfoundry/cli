package configv3_test

import (
	"time"

	. "code.cloudfoundry.org/cli/util/configv3"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	var (
		config *Config
	)

	BeforeEach(func() {
		config = &Config{}
	})

	When("there are environment variables set", func() {
		BeforeEach(func() {
			config.ENV = EnvOverride{
				CFDialTimeout:    "1234",
				CFPassword:       "I am password.",
				CFStagingTimeout: "8675",
				CFStartupTimeout: "309",
				CFUsername:       "i-R-user",
				DockerPassword:   "banana",
				HTTPSProxy:       "proxy.com",
			}
		})

		It("overrides specific config values using those variables", func() {
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
		BeforeEach(func() {
			config.ENV.BinaryName = "potatoman"
		})

		It("returns the name used to invoke", func() {
			Expect(config.BinaryName()).To(Equal("potatoman"))
		})
	})

	Describe("BinaryVersion", func() {
		It("returns back version.BinaryVersion", func() {
			conf := Config{}
			Expect(conf.BinaryVersion()).To(Equal("0.0.0-unknown-version"))
		})
	})

	Describe("DialTimeout", func() {
		When("no DialTimeout is set in the env", func() {
			BeforeEach(func() {
				config.ENV.CFDialTimeout = ""
			})

			It("uses the default dial timeout", func() {
				Expect(config.DialTimeout()).To(Equal(DefaultDialTimeout))
			})
		})
	})

	DescribeTable("Experimental",
		func(envVal string, expected bool) {
			config.ENV.Experimental = envVal
			Expect(config.Experimental()).To(Equal(expected))
		},

		Entry("uses default value of false if environment value is not set", "", false),
		Entry("uses environment value if a valid environment value is set", "true", true),
		Entry("uses default value of false if an invalid environment value is set", "something-invalid", false),
	)

	DescribeTable("ExperimentalLogin",
		func(envVal string, expected bool) {
			config.ENV.ExperimentalLogin = envVal
			Expect(config.ExperimentalLogin()).To(Equal(expected))
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

	Describe("StagingTimeout", func() {
		When("no StagingTimeout is set in the env", func() {
			BeforeEach(func() {
				config.ENV.CFStagingTimeout = ""
			})

			It("uses the default staging timeout", func() {
				Expect(config.StagingTimeout()).To(Equal(DefaultStagingTimeout))
			})
		})
	})

	Describe("StartupTimeout", func() {
		When("no StartupTimeout is set in the env", func() {
			BeforeEach(func() {
				config.ENV.CFStartupTimeout = ""
			})

			It("uses the default startup timeout", func() {
				Expect(config.StartupTimeout()).To(Equal(DefaultStartupTimeout))
			})
		})
	})
})
