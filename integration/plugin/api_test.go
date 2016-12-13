package plugin

import (
	"fmt"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("plugin API", func() {
	Describe("AccessToken", func() {
		It("returns the access token", func() {
			confirmTestPluginOutput("AccessToken", "bearer [\\w\\d\\.]+")
		})
	})

	Describe("ApiEndpoint", func() {
		It("returns the API endpoint", func() {
			confirmTestPluginOutput("ApiEndpoint", apiURL)
		})
	})

	Describe("ApiVersion", func() {
		It("returns the API version", func() {
			confirmTestPluginOutput("ApiVersion", "2\\.\\d+\\.\\d+")
		})
	})

	Describe("CliCommand", func() {
		It("calls the core cli command and outputs to terminal", func() {
			confirmTestPluginOutput("CliCommand", "API endpoint", "API endpoint")
		})
	})

	Describe("CliCommandWithoutTerminalOutput", func() {
		It("calls the core cli command and without outputting to the terminal", func() {
			session := helpers.CF("CliCommandWithoutTerminalOutput", "target")
			Eventually(session).Should(Say("API endpoint"))
			Consistently(session).ShouldNot(Say("API endpoint"))
			Eventually(session).Should(Exit(0))
		})
	})

	Describe("DopplerEndpoint", func() {
		It("gets Doppler Endpoint", func() {
			confirmTestPluginOutput("DopplerEndpoint", "wss://doppler")
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
		var appName1, appName2 string
		BeforeEach(func() {
			createTargetedOrgAndSpace()
			appName1 = helpers.PrefixedRandomName("APP")
			helpers.WithHelloWorldApp(func(appDir string) {
				Eventually(helpers.CF("push", appName1, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
			})
			appName2 = helpers.PrefixedRandomName("APP")
			helpers.WithHelloWorldApp(func(appDir string) {
				Eventually(helpers.CF("push", appName2, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
			})
		})

		It("gets information for multiple applications", func() {
			appNameRegexp := fmt.Sprintf("(?:%s|%s)", appName1, appName2)
			confirmTestPluginOutput("GetApps", appNameRegexp, appNameRegexp)
		})
	})

	Describe("GetCurrentOrg", func() {
		It("gets the current targeted org", func() {
			org, _ := createTargetedOrgAndSpace()
			confirmTestPluginOutput("GetCurrentOrg", org)
		})
	})

	Describe("GetCurrentSpace", func() {
		It("gets the current targeted Space", func() {
			_, space := createTargetedOrgAndSpace()
			confirmTestPluginOutput("GetCurrentSpace", space)
		})
	})

	Describe("GetOrg", func() {
		It("gets the given org", func() {
			org, _ := createTargetedOrgAndSpace()
			confirmTestPluginOutputWithArg("GetOrg", org, org)
		})
	})

	Describe("GetOrgs", func() {
		It("gets information for multiple orgs", func() {
			org1, _ := createTargetedOrgAndSpace()
			org2, _ := createTargetedOrgAndSpace()
			orgNameRegexp := fmt.Sprintf("(?:%s|%s)", org1, org2)
			confirmTestPluginOutput("GetOrgs", orgNameRegexp, orgNameRegexp)
		})
	})

	Describe("GetOrgUsers", func() {
		It("returns the org users", func() {
			org, _ := createTargetedOrgAndSpace()
			username, _ := helpers.GetCredentials()
			confirmTestPluginOutputWithArg("GetOrgUsers", org, username)
		})
	})

	Describe("GetOrgUsers", func() {
		It("returns the org users", func() {
			org, _ := createTargetedOrgAndSpace()
			username, _ := helpers.GetCredentials()
			confirmTestPluginOutputWithArg("GetOrgUsers", org, username)
		})
	})

	Describe("GetService", func() {
		var (
			serviceInstance string
			broker          helpers.ServiceBroker
		)
		BeforeEach(func() {
			createTargetedOrgAndSpace()
			domain := defaultSharedDomain()
			service := helpers.PrefixedRandomName("SERVICE")
			servicePlan := helpers.PrefixedRandomName("SERVICE-PLAN")
			serviceInstance = helpers.PrefixedRandomName("SI")
			broker = helpers.NewServiceBroker(helpers.PrefixedRandomName("SERVICE-BROKER"), helpers.NewAssets().ServiceBroker, domain, service, servicePlan)
			broker.Push()
			broker.Configure()
			broker.Create()

			Eventually(helpers.CF("enable-service-access", service)).Should(Exit(0))
			Eventually(helpers.CF("create-service", service, servicePlan, serviceInstance)).Should(Exit(0))
		})

		AfterEach(func() {
			broker.Destroy()
		})

		It("gets the given service instance", func() {
			confirmTestPluginOutputWithArg("GetService", serviceInstance, serviceInstance)
		})
	})

	Describe("GetServices", func() {
		var (
			serviceInstance1 string
			serviceInstance2 string
			broker           helpers.ServiceBroker
		)
		BeforeEach(func() {
			createTargetedOrgAndSpace()
			domain := defaultSharedDomain()
			service := helpers.PrefixedRandomName("SERVICE")
			servicePlan := helpers.PrefixedRandomName("SERVICE-PLAN")
			serviceInstance1 = helpers.PrefixedRandomName("SI1")
			serviceInstance2 = helpers.PrefixedRandomName("SI2")
			broker = helpers.NewServiceBroker(helpers.PrefixedRandomName("SERVICE-BROKER"), helpers.NewAssets().ServiceBroker, domain, service, servicePlan)
			broker.Push()
			broker.Configure()
			broker.Create()

			Eventually(helpers.CF("enable-service-access", service)).Should(Exit(0))
			Eventually(helpers.CF("create-service", service, servicePlan, serviceInstance1)).Should(Exit(0))
			Eventually(helpers.CF("create-service", service, servicePlan, serviceInstance2)).Should(Exit(0))
		})

		AfterEach(func() {
			broker.Destroy()
		})

		It("returns a list of services instances", func() {
			servicesNameRegexp := fmt.Sprintf("(?:%s|%s)", serviceInstance1, serviceInstance2)
			confirmTestPluginOutput("GetServices", servicesNameRegexp, servicesNameRegexp)
		})
	})

	Describe("GetSpace", func() {
		It("gets the given space", func() {
			_, space := createTargetedOrgAndSpace()
			confirmTestPluginOutputWithArg("GetSpace", space, space)
		})
	})

	Describe("GetSpaces", func() {
		var space1, space2 string

		BeforeEach(func() {
			_, space1 = createTargetedOrgAndSpace()
			space2 = helpers.PrefixedRandomName("SPACE")
			helpers.CreateSpace(space2)
		})

		It("gets information for multiple spaces", func() {
			spaceNameRegexp := fmt.Sprintf("(?:%s|%s)", space1, space2)
			confirmTestPluginOutput("GetSpaces", spaceNameRegexp, spaceNameRegexp)
		})
	})

	Describe("GetSpaceUsers", func() {
		It("returns the space users", func() {
			org, space := createTargetedOrgAndSpace()
			username, _ := helpers.GetCredentials()
			session := helpers.CF("GetSpaceUsers", org, space)
			Eventually(session).Should(Say(username))
			Eventually(session).Should(Exit(0))
		})
	})

	Describe("HasAPIEndpoint", func() {
		It("returns true", func() {
			confirmTestPluginOutput("HasAPIEndpoint", "true")
		})
	})

	Describe("HasOrganization", func() {
		It("returns true", func() {
			createTargetedOrgAndSpace()
			confirmTestPluginOutput("HasOrganization", "true")
		})
	})

	Describe("HasSpace", func() {
		It("returns true", func() {
			createTargetedOrgAndSpace()
			confirmTestPluginOutput("HasSpace", "true")
		})
	})

	Describe("IsLoggedIn", func() {
		It("returns a true", func() {
			confirmTestPluginOutput("IsLoggedIn", "true")
		})
	})

	Describe("IsSSLDisabled", func() {
		It("returns a true or false", func() {
			if skipSSLValidation == "" {
				confirmTestPluginOutput("IsSSLDisabled", "false")
			} else {
				confirmTestPluginOutput("IsSSLDisabled", "true")
			}
		})
	})

	Describe("LoggregatorEndpoint", func() {
		It("gets Loggregator Endpoint", func() {
			confirmTestPluginOutput("LoggregatorEndpoint", "wss://loggregator")
		})
	})

	Describe("UserEmail", func() {
		It("gets the current user's Email", func() {
			username, _ := helpers.GetCredentials()
			confirmTestPluginOutput("UserEmail", username)
		})
	})

	Describe("UserGuid", func() {
		It("gets the current user's GUID", func() {
			confirmTestPluginOutput("UserGuid", "[\\w\\d]+-[\\w\\d]+-[\\w\\d]+-[\\w\\d]+-[\\w\\d]+")
		})
	})

	Describe("Username", func() {
		It("gets the current username", func() {
			username, _ := helpers.GetCredentials()
			confirmTestPluginOutput("Username", username)
		})
	})
})
