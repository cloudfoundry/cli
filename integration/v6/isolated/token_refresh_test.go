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

				When("running a v6 command", func() {
					When("the cloud controller client encounters an invalid token response", func() {
						It("refreshes the token", func() {
							session := helpers.CF("unbind-service", "app", "service")
							Eventually(session.Err).Should(Say("App 'app' not found"))
							Eventually(session).Should(Exit(1))
						})
					})

					When("the UAA client encounters an invalid token response", func() {
						It("refreshes the token", func() {
							username, _ := helpers.GetCredentials()
							session := helpers.CF("create-user", username, helpers.NewPassword())
							Eventually(session.Err).Should(Say(fmt.Sprintf("user %s already exists", username)))
							Eventually(session).Should(Exit(0))
						})
					})
				})

				When("running an unrefactored v6 command", func() {
					It("refreshes the token", func() {
						session := helpers.CF("stack", "some-stack")
						Eventually(session).Should(Say("Stack some-stack not found"))
						Eventually(session).Should(Exit(1))
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

				When("running a v6 command", func() {
					When("the cloud controller client encounters an invalid token response", func() {
						It("refreshes the token", func() {
							session := helpers.CF("unbind-service", "app", "service")
							Eventually(session.Err).Should(Say("App 'app' not found"))
							Eventually(session).Should(Exit(1))
						})
					})

					When("the UAA client encounters an invalid token response", func() {
						It("refreshes the token", func() {
							username, _ := helpers.GetCredentials()
							session := helpers.CF("create-user", username, helpers.NewPassword())
							Eventually(session.Err).Should(Say(fmt.Sprintf("user %s already exists", username)))
							Eventually(session).Should(Exit(0))
						})
					})
				})

				When("running an unrefactored v6 command", func() {
					It("refreshes the token", func() {
						session := helpers.CF("stack", "some-stack")
						Eventually(session).Should(Say("Stack some-stack not found"))
						Eventually(session).Should(Exit(1))
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

			When("running a v6 refactored command", func() {
				When("the cloud controller client encounters an invalid token response", func() {
					It("refreshes the token", func() {
						session := helpers.CF("unbind-service", "app", "service")
						Eventually(session.Err).Should(Say("App 'app' not found"))
						Eventually(session).Should(Exit(1))
					})
				})

				When("the UAA client encounters an invalid token response", func() {
					It("refreshes the token", func() {
						username := helpers.NewUsername()
						session := helpers.CF("create-user", username, helpers.NewPassword())
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Exit(0))
					})
				})
			})

			When("running a v6 unrefactored command", func() {
				When("the cloud controller client encounters an invalid token response", func() {
					It("refreshes the token", func() {
						username, _ := helpers.GetCredentials()
						session := helpers.CF("quotas")
						Eventually(session).Should(Say("Getting quotas as %s", username))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Exit(0))
					})
				})
			})
		})
	})
})
