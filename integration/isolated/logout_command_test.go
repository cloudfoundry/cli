package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/util/configv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("logout command", func() {
	var (
		orgName   string
		spaceName string
	)

	BeforeEach(func() {
		helpers.LoginCF()

		orgName = helpers.NewOrgName()
		spaceName = helpers.NewSpaceName()
	})

	Context("help", func() {
		It("displays help", func() {
			session := helpers.CF("logout", "--help")
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Say("   logout - Log user out"))
			Eventually(session).Should(Say("USAGE:"))
			Eventually(session).Should(Say("   cf logout"))
			Eventually(session).Should(Say("ALIAS:"))
			Eventually(session).Should(Say("   lo"))
			Eventually(session).Should(Exit(0))
		})
	})

	Context("when there's user information set in the config", func() {
		BeforeEach(func() {
			helpers.SetConfig(func(conf *configv3.Config) {
				conf.SetAccessToken("some-access-token")
				conf.SetRefreshToken("bb8f7b209ff74409877974bce5752412-r")
				conf.SetOrganizationInformation("some-org-guid", orgName)
				conf.SetSpaceInformation("some-space-guid", spaceName, true)
				conf.SetUAAGrantType("client_credentials")
				conf.SetUAAClientCredentials("potatoface", "acute")
			})
		})

		It("clears out user information in the config", func() {
			session := helpers.CF("logout")

			Eventually(session).Should(Say("Logging out..."))
			Eventually(session).Should(Say("OK"))
			Eventually(session).Should(Exit(0))

			config, err := configv3.LoadConfig()
			Expect(err).ToNot(HaveOccurred())

			Expect(config.ConfigFile.AccessToken).To(BeEmpty())
			Expect(config.ConfigFile.RefreshToken).To(BeEmpty())
			Expect(config.ConfigFile.TargetedOrganization.GUID).To(BeEmpty())
			Expect(config.ConfigFile.TargetedOrganization.Name).To(BeEmpty())
			Expect(config.ConfigFile.TargetedSpace.AllowSSH).To(BeFalse())
			Expect(config.ConfigFile.TargetedSpace.GUID).To(BeEmpty())
			Expect(config.ConfigFile.TargetedSpace.Name).To(BeEmpty())
			Expect(config.ConfigFile.UAAGrantType).To(BeEmpty())
			Expect(config.ConfigFile.UAAOAuthClient).To(Equal("cf"))
			Expect(config.ConfigFile.UAAOAuthClientSecret).To(BeEmpty())
		})
	})
})
