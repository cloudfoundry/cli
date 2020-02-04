package plugin

import (
	"regexp"

	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/util/configv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("plugin API", func() {
	BeforeEach(func() {
		installTestPlugin()
	})

	AfterEach(func() {
		uninstallTestPlugin()
	})

	Describe("AccessToken", func() {
		It("returns the access token", func() {
			confirmTestPluginOutput("AccessToken", `bearer [\w\d\.]+`)
		})
	})

	Describe("ApiEndpoint", func() {
		It("returns the API endpoint", func() {
			confirmTestPluginOutput("ApiEndpoint", apiURL)
		})
	})

	Describe("CliCommand", func() {
		Describe("stdout output", func() {
			It("outputs the result of the command to stdout", func() {
				confirmTestPluginOutputWithArg("CliCommand", "target", "api endpoint",
					"@@ plugin CliCommand response")
			})
		})

		Describe("returned value from CliCommand", func() {
			It("gets a slice of lines", func() {
				confirmTestPluginOutputWithArg("CliCommand", "target",
					"@@ plugin CliCommand response", "0: api endpoint", "1: api version")
			})

			It("does not see an empty line at the end", func() {
				session := helpers.CF("CliCommand", "target")
				Eventually(session).Should(Exit(0))
				Expect(session).ShouldNot(Say(`\n[0-9]+:\s*\z`))
			})
		})
	})

	Describe("GetApp", func() {
		var appName string
		BeforeEach(func() {
			createTargetedOrgAndSpace()
			appName = helpers.PrefixedRandomName("APP")
			helpers.WithHelloWorldApp(func(appDir string) {
				Eventually(helpers.CF("push", appName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
			})
		})

		It("gets application information", func() {
			confirmTestPluginOutputWithArg("GetApp", appName, appName)
		})
	})

	Describe("GetApps", func() {
		var appName [2]string
		BeforeEach(func() {
			createTargetedOrgAndSpace()
			// Verify apps come back in order of creation, not alphabetically
			appName[0] = "Z" + helpers.PrefixedRandomName("APP")
			appName[1] = "A" + helpers.PrefixedRandomName("APP")
			helpers.WithHelloWorldApp(func(appDir string) {
				Eventually(helpers.CF("push", appName[0], "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
			})
			helpers.WithHelloWorldApp(func(appDir string) {
				Eventually(helpers.CF("push", appName[1], "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
			})
		})

		It("gets application information", func() {
			confirmTestPluginOutputWithArg("GetApps", appName[0], appName[1])
		})
	})

	Describe("GetOrg", func() {
		var orgName string
		var orgGUID string
		var domainName string
		var domain helpers.Domain

		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			helpers.CreateOrg(orgName)
			helpers.TargetOrg(orgName)
			orgGUID = helpers.GetOrgGUID(orgName)

			domainName = helpers.DomainName("get-org-test")
			domain = helpers.NewDomain(orgName, domainName)
			domain.Create()
		})

		AfterEach(func() {
			domain.Delete()
		})

		It("gets the organization information", func() {
			confirmTestPluginOutputWithArg("GetOrg", orgName, orgName, orgGUID, domainName)
		})

		When("the org has metadata", func() {
			BeforeEach(func() {
				Eventually(helpers.CF("set-label", "org", orgName, "orgType=production")).Should(Exit(0))
			})
			It("Displays the metadata correctly", func() {
				confirmTestPluginOutputWithArg("GetOrg", orgName, regexp.QuoteMeta("Labels:map[orgType:{Value:production"))
			})
		})

		When("the org does not exist", func() {
			It("Displays a useful error message", func() {
				confirmTestPluginOutputWithArg("GetOrg", "blahblahblah", "Error GetOrg: Organization 'blahblahblah' not found")
			})
		})
	})

	Describe("GetSpace", func() {
		var orgName string
		var spaceName string
		var spaceGUID string

		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()
			helpers.CreateOrgAndSpace(orgName, spaceName)
			helpers.TargetOrg(orgName)
			spaceGUID = helpers.GetSpaceGUID(spaceName)
		})
		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		It("gets the space information", func() {
			confirmTestPluginOutputWithArg("GetSpace", spaceName, spaceGUID)
		})

		When("the space has metadata", func() {
			BeforeEach(func() {
				Eventually(helpers.CF("set-label", "space", spaceName, "spaceType=production")).Should(Exit(0))
			})
			It("Displays the metadata correctly", func() {
				confirmTestPluginOutputWithArg("GetSpace", spaceName, "spaceType", "production")
			})
		})

		When("the space does not exist", func() {
			It("Displays a useful error message", func() {
				confirmTestPluginOutputWithArg("GetSpace", "blahblahblah", "Error GetSpace: Space 'blahblahblah' not found")
			})
		})
	})

	Describe("GetSpaces", func() {
		var orgName string

		BeforeEach(func() {
			orgName = helpers.CreateAndTargetOrg()
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		Context("when there are no spaces", func() {
			It("gets no spaces", func() {
				confirmTestPluginOutputWithArg("GetSpaces", "No spaces found")
			})
		})

		Context("when there are spaces", func() {
			var spaceName string
			var spaceGUID string
			var spaceName2 string
			var spaceGUID2 string

			BeforeEach(func() {
				spaceName = helpers.NewSpaceName()
				helpers.CreateSpace(spaceName)
				spaceName2 = helpers.NewSpaceName()
				helpers.CreateSpace(spaceName2)
				spaceGUID = helpers.GetSpaceGUID(spaceName)
				spaceGUID2 = helpers.GetSpaceGUID(spaceName2)
			})

			It("gets the spaces' information", func() {
				// Results are in random order so make two separate calls
				confirmTestPluginOutputWithArg("GetSpaces", spaceName, spaceGUID)
				confirmTestPluginOutputWithArg("GetSpaces", spaceName2, spaceGUID2)
			})

			When("the spaces have metadata", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("set-label", "space", spaceName, "spaceType=production")).Should(Exit(0))
					Eventually(helpers.CF("set-label", "space", spaceName2, "squadron=goose")).Should(Exit(0))
				})
				It("Displays the metadata", func() {
					// Results are in random order so make two separate calls
					confirmTestPluginOutputWithArg("GetSpaces", spaceName, "spaceType", "production")
					confirmTestPluginOutputWithArg("GetSpaces", spaceName2, "squadron", "goose")
				})
			})
		})
	})

	Describe("GetCurrentSpace", func() {
		It("gets the current targeted Space", func() {
			_, space := createTargetedOrgAndSpace()
			confirmTestPluginOutput("GetCurrentSpace", space)
		})
	})

	Describe("GetCurrentOrg", func() {
		It("gets the current targeted Org", func() {
			org, _ := createTargetedOrgAndSpace()
			confirmTestPluginOutput("GetCurrentOrg", org)
		})
	})

	Describe("IsLoggedIn", func() {
		When("logged in", func() {
			It("returns true", func() {
				confirmTestPluginOutput("IsLoggedIn", "true")
			})
		})
		When("logged out", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
			})
			It("returns false", func() {
				confirmTestPluginOutput("IsLoggedIn", "false")
			})
		})
	})

	Describe("Username", func() {
		It("gets the current user's name", func() {
			username := getUsername()
			confirmTestPluginOutput("Username", username)
		})

		When("not logged in", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
			})
			It("returns an error", func() {
				confirmTestPluginOutput("Username", "not logged in")
			})
		})

		When("the token is invalid", func() {
			var accessToken string

			BeforeEach(func() {
				helpers.SetConfig(func(config *configv3.Config) {
					accessToken = config.ConfigFile.AccessToken
					config.ConfigFile.AccessToken = accessToken + "***"
				})
			})
			AfterEach(func() {
				helpers.SetConfig(func(config *configv3.Config) {
					config.ConfigFile.AccessToken = accessToken
				})
			})

			When("running a v7 plugin command", func() {
				It("complains about the token", func() {
					session := helpers.CF("Username")
					Eventually(session).Should(Say("illegal base64 data at input byte 342"))
				})
			})
		})
	})

	Describe("IsSkipSSLValidation", func() {
		When("--skip-ssl-validation is not specified", func() {
			BeforeEach(func() {
				if helpers.SkipSSLValidation() {
					Skip("Test is being run with skip ssl validation")
				}
			})
			It("returns false", func() {
				confirmTestPluginOutput("IsSkipSSLValidation", "false")
			})
		})
		When("--skip-ssl-validation is specified", func() {
			BeforeEach(func() {
				Eventually(helpers.CF("api", apiURL, "--skip-ssl-validation")).Should(Exit(0))
			})
			It("returns true", func() {
				confirmTestPluginOutput("IsSkipSSLValidation", "true")
			})
		})
	})

})
