package integration

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/utils/configv3"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Token Refreshing", func() {
	Context("when running a v2 command with an invalid token", func() {
		BeforeEach(func() {
			Skip("skip #133310639")
			loginCF()

			config, err := configv3.LoadConfig()
			Expect(err).ToNot(HaveOccurred())
			config.ConfigFile.AccessToken = config.ConfigFile.AccessToken + "foo"
			config.ConfigFile.TargetedOrganization.GUID = "fake-org"
			config.ConfigFile.TargetedSpace.GUID = "fake-space"
			err = configv3.WriteConfig(config)
			Expect(err).ToNot(HaveOccurred())
		})

		It("refreshes the token", func() {
			session := helpers.CF("unbind-service", "app", "service")
			Eventually(session.Err).Should(Say("App app not found"))
			Eventually(session).Should(Exit(1))
		})
	})

	Context("when running a v3 command with an invalid token", func() {
		BeforeEach(func() {
			loginCF()

			config, err := configv3.LoadConfig()
			Expect(err).ToNot(HaveOccurred())
			config.ConfigFile.AccessToken = config.ConfigFile.AccessToken + "foo"
			config.ConfigFile.TargetedOrganization.GUID = "fake-org"
			config.ConfigFile.TargetedSpace.GUID = "fake-space"
			err = configv3.WriteConfig(config)
			Expect(err).ToNot(HaveOccurred())
		})

		It("refreshes the token", func() {
			session := helpers.CF("-v", "run-task", "app", "'echo banana'")
			Eventually(session.Err).Should(Say("App app not found"))
			Eventually(session).Should(Exit(1))
		})
	})
})
