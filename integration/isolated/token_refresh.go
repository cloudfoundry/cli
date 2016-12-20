package isolated

import (
	"fmt"

	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/util/configv3"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Token Refreshing", func() {
	Context("when running a v2 command with an invalid token", func() {
		BeforeEach(func() {
			helpers.LoginCF()

			helpers.SetConfig(func(config *configv3.Config) {
				config.ConfigFile.AccessToken = config.ConfigFile.AccessToken + "foo"
				config.ConfigFile.TargetedOrganization.GUID = "fake-org"
				config.ConfigFile.TargetedSpace.GUID = "fake-space"
			})
		})

		Context("when the cloud controller client encounters an invalid token response", func() {
			It("refreshes the token", func() {
				session := helpers.CF("unbind-service", "app", "service")
				Eventually(session.Err).Should(Say("App app not found"))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the UAA client encounters an invalid token response", func() {
			It("refreshes the token", func() {
				username, _ := helpers.GetCredentials()
				session := helpers.CF("create-user", username, helpers.RandomPassword())
				Eventually(session.Err).Should(Say(fmt.Sprintf("user %s already exists", username)))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("when running a v3 command with an invalid token", func() {
		BeforeEach(func() {
			helpers.LoginCF()

			helpers.SetConfig(func(config *configv3.Config) {
				config.ConfigFile.AccessToken = config.ConfigFile.AccessToken + "foo"
				config.ConfigFile.TargetedOrganization.GUID = "fake-org"
				config.ConfigFile.TargetedSpace.GUID = "fake-space"
			})
		})

		Context("when the cloud controller client encounters an invalid token response", func() {
			It("refreshes the token", func() {
				session := helpers.CF("-v", "run-task", "app", "'echo banana'")
				Eventually(session.Err).Should(Say("App app not found"))
				Eventually(session).Should(Exit(1))
			})
		})
	})
})
