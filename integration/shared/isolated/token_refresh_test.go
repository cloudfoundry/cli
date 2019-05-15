package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/util/configv3"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Token Refreshing", func() {
	Describe("password grant type", func() {
		BeforeEach(func() {
			helpers.LoginCF()
		})

		When("the token is invalid", func() {
			When("password is explicitly stored as the grant type", func() {
				BeforeEach(func() {
					helpers.SetConfig(func(config *configv3.Config) {
						config.ConfigFile.AccessToken = helpers.ExpiredAccessToken()
						config.ConfigFile.TargetedOrganization.GUID = "fake-org"
						config.ConfigFile.TargetedSpace.GUID = "fake-space"
						config.ConfigFile.UAAGrantType = "password"
					})
				})

				When("running a v7 command", func() {
					When("the cloud controller client encounters an invalid token response", func() {
						It("refreshes the token", func() {
							helpers.SkipIfClientCredentialsTestMode()
							session := helpers.CF("run-task", "app", "'echo banana'")
							Eventually(session.Err).Should(Say("App 'app' not found"))
							Eventually(session).Should(Exit(1))
						})
					})
				})
			})

			When("no grant type is explicitly stored", func() {
				BeforeEach(func() {
					helpers.SetConfig(func(config *configv3.Config) {
						config.ConfigFile.AccessToken = helpers.ExpiredAccessToken()
						config.ConfigFile.TargetedOrganization.GUID = "fake-org"
						config.ConfigFile.TargetedSpace.GUID = "fake-space"
						config.ConfigFile.UAAGrantType = ""
					})
				})

				When("running a v7 command", func() {
					When("the cloud controller client encounters an invalid token response", func() {
						It("refreshes the token", func() {
							helpers.SkipIfClientCredentialsTestMode()
							session := helpers.CF("run-task", "app", "'echo banana'")
							Eventually(session.Err).Should(Say("App 'app' not found"))
							Eventually(session).Should(Exit(1))
						})
					})
				})
			})
		})
	})

	Describe("client grant type", func() {
		BeforeEach(func() {
			helpers.LoginCFWithClientCredentials()
		})

		When("the token is invalid", func() {
			BeforeEach(func() {
				helpers.SetConfig(func(config *configv3.Config) {
					config.ConfigFile.AccessToken = helpers.ExpiredAccessToken()
					config.ConfigFile.TargetedOrganization.GUID = "fake-org"
					config.ConfigFile.TargetedSpace.GUID = "fake-space"
				})
			})

			When("running a v7 command", func() {
				When("the cloud controller client encounters an invalid token response", func() {
					It("displays an error and exits 1", func() {
						session := helpers.CF("run-task", "app", "'echo banana'")
						Eventually(session.Err).Should(Say(`Credentials were rejected, please try again\.`))
						Eventually(session).Should(Exit(1))
					})
				})
			})
		})
	})
})
