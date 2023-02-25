package configv3_test

import (
	"os"

	"code.cloudfoundry.org/cli/util/configv3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	var (
		homeDir string
		config  *configv3.Config
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
			config, err = configv3.LoadConfig()
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

	Describe("SetKubernetesAuthInfo", func() {
		BeforeEach(func() {
			var err error
			config, err = configv3.LoadConfig()
			Expect(err).NotTo(HaveOccurred())
		})

		JustBeforeEach(func() {
			config.SetKubernetesAuthInfo("k8s-auth")
		})

		It("sets the cf-on-k8s auth info", func() {
			Expect(config.ConfigFile.CFOnK8s.AuthInfo).To(Equal("k8s-auth"))
		})
	})
})
