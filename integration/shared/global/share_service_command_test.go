// +build !partialPush

package global

import (
	"fmt"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("share-service command", func() {
	var (
		sourceOrgName     string
		sourceSpaceName   string
		sharedToOrgName   string
		sharedToSpaceName string
		serviceInstance   string
	)

	BeforeEach(func() {
		helpers.SkipIfVersionLessThan(ccversion.MinVersionShareServiceV3)

		sourceOrgName = helpers.NewOrgName()
		sourceSpaceName = helpers.NewSpaceName()
		sharedToOrgName = helpers.NewOrgName()
		sharedToSpaceName = helpers.NewSpaceName()
		serviceInstance = helpers.PrefixedRandomName("svc-inst")

		helpers.LoginCF()
		session := helpers.CF("enable-feature-flag", "service_instance_sharing")
		Eventually(session).Should(Exit(0))
	})

	Describe("help", func() {
		When("--help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("share-service", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("share-service - Share a service instance with another space"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say(`cf share-service SERVICE_INSTANCE -s OTHER_SPACE \[-o OTHER_ORG\]`))
				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say(`-o\s+Org of the other space \(Default: targeted org\)`))
				Eventually(session).Should(Say(`-s\s+Space to share the service instance into`))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("bind-service, service, services, unshare-service"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the service instance name is not provided", func() {
		It("tells the user that the service instance name is required, prints help text, and exits 1", func() {
			session := helpers.CF("share-service", "-s", sharedToSpaceName)

			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `SERVICE_INSTANCE` was not provided"))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	When("the space name is not provided", func() {
		It("tells the user that the space name is required, prints help text, and exits 1", func() {
			session := helpers.CF("share-service")

			Eventually(session.Err).Should(Say("Incorrect Usage: the required flag `-s' was not specified"))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "share-service", serviceInstance, "-s", sharedToSpaceName)
		})

		When("the v3 api version is lower than the minimum version", func() {
			var server *Server

			BeforeEach(func() {
				server = helpers.StartAndTargetServerWithAPIVersions(helpers.DefaultV2Version, ccversion.MinV3ClientVersion)
			})

			AfterEach(func() {
				server.Close()
			})

			It("fails with error message that the minimum version is not met", func() {
				session := helpers.CF("share-service", serviceInstance, "-s", sharedToSpaceName)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say(`This command requires CF API version 3\.36\.0 or higher\.`))
				Eventually(session).Should(Exit(1))
			})
		})
	})

	When("the environment is set up correctly", func() {
		var (
			domain      string
			service     string
			servicePlan string
		)

		BeforeEach(func() {
			service = helpers.PrefixedRandomName("SERVICE")
			servicePlan = helpers.PrefixedRandomName("SERVICE-PLAN")

			helpers.CreateOrgAndSpace(sharedToOrgName, sharedToSpaceName)
			helpers.SetupCF(sourceOrgName, sourceSpaceName)

			domain = helpers.DefaultSharedDomain()
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(sharedToOrgName)
			helpers.QuickDeleteOrg(sourceOrgName)
		})

		When("there is a managed service instance in my current targeted space", func() {
			var broker helpers.ServiceBroker

			BeforeEach(func() {
				broker = helpers.NewServiceBroker(helpers.NewServiceBrokerName(), helpers.NewAssets().ServiceBroker, domain, service, servicePlan)
				broker.Push()
				broker.Configure(true)
				broker.Create()

				Eventually(helpers.CF("enable-service-access", service)).Should(Exit(0))
				Eventually(helpers.CF("create-service", service, servicePlan, serviceInstance)).Should(Exit(0))
			})

			AfterEach(func() {
				broker.Destroy()
			})

			When("I want to share my service instance to a space in another org", func() {
				AfterEach(func() {
					Eventually(helpers.CF("unshare-service", serviceInstance, "-s", sharedToSpaceName, "-o", sharedToOrgName, "-f")).Should(Exit(0))
				})

				It("shares the service instance from my targeted space with the share-to org/space", func() {
					username, _ := helpers.GetCredentials()
					session := helpers.CF("share-service", serviceInstance, "-s", sharedToSpaceName, "-o", sharedToOrgName)
					Eventually(session.Out).Should(Say(`Sharing service instance %s into org %s / space %s as %s\.\.\.`, serviceInstance, sharedToOrgName, sharedToSpaceName, username))
					Eventually(session.Out).Should(Say("OK"))
					Eventually(session).Should(Exit(0))
				})

				When("the service instance is already shared with that space", func() {
					BeforeEach(func() {
						Eventually(helpers.CF("share-service", serviceInstance, "-s", sharedToSpaceName, "-o", sharedToOrgName)).Should(Exit(0))
					})

					It("displays a warning and exits 0", func() {
						session := helpers.CF("share-service", serviceInstance, "-s", sharedToSpaceName, "-o", sharedToOrgName)
						Consistently(session.Out).ShouldNot(Say("FAILED"))
						Eventually(session.Out).Should(Say(`Service instance %s is already shared with that space\.`, serviceInstance))
						Eventually(session.Out).Should(Say("OK"))
						Eventually(session).Should(Exit(0))
					})
				})
			})

			When("I want to share my service instance into another space in my targeted org", func() {
				BeforeEach(func() {
					helpers.CreateSpace(sharedToSpaceName)
				})

				AfterEach(func() {
					Eventually(helpers.CF("unshare-service", serviceInstance, "-s", sharedToSpaceName, "-f")).Should(Exit(0))
				})

				It("shares the service instance from my targeted space with the share-to space", func() {
					username, _ := helpers.GetCredentials()
					session := helpers.CF("share-service", serviceInstance, "-s", sharedToSpaceName)
					Eventually(session.Out).Should(Say(`Sharing service instance %s into org %s / space %s as %s\.\.\.`, serviceInstance, sourceOrgName, sharedToSpaceName, username))
					Eventually(session.Out).Should(Say("OK"))
					Eventually(session).Should(Exit(0))
				})

				When("the service instance is already shared with that space", func() {
					BeforeEach(func() {
						Eventually(helpers.CF("share-service", serviceInstance, "-s", sharedToSpaceName)).Should(Exit(0))
					})

					It("displays a warning and exits 0", func() {
						session := helpers.CF("share-service", serviceInstance, "-s", sharedToSpaceName)
						Consistently(session.Out).ShouldNot(Say("FAILED"))
						Eventually(session.Out).Should(Say(`Service instance %s is already shared with that space\.`, serviceInstance))
						Eventually(session.Out).Should(Say("OK"))
						Eventually(session).Should(Exit(0))
					})
				})
			})

			When("the org I want to share into does not exist", func() {
				It("fails with an org not found error", func() {
					session := helpers.CF("share-service", serviceInstance, "-s", sharedToSpaceName, "-o", "missing-org")
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("Organization 'missing-org' not found"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("the space I want to share into does not exist", func() {
				It("fails with a space not found error", func() {
					session := helpers.CF("share-service", serviceInstance, "-s", "missing-space")
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("Space 'missing-space' not found"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("I am a SpaceAuditor in the space I want to share into", func() {
				var sharedToSpaceGUID string
				BeforeEach(func() {
					user := helpers.NewUsername()
					password := helpers.NewPassword()
					Eventually(helpers.CF("create-user", user, password)).Should(Exit(0))
					Eventually(helpers.CF("set-space-role", user, sourceOrgName, sourceSpaceName, "SpaceDeveloper")).Should(Exit(0))
					Eventually(helpers.CF("set-space-role", user, sharedToOrgName, sharedToSpaceName, "SpaceAuditor")).Should(Exit(0))
					env := map[string]string{
						"CF_USERNAME": user,
						"CF_PASSWORD": password,
					}
					Eventually(helpers.CFWithEnv(env, "auth")).Should(Exit(0))
					helpers.TargetOrgAndSpace(sharedToOrgName, sharedToSpaceName)
					sharedToSpaceGUID = helpers.GetSpaceGUID(sharedToSpaceName)
					helpers.TargetOrgAndSpace(sourceOrgName, sourceSpaceName)
				})

				AfterEach(func() {
					helpers.SetupCF(sourceOrgName, sourceSpaceName)
				})

				It("fails with an unauthorized error", func() {
					session := helpers.CF("share-service", serviceInstance, "-s", sharedToSpaceName, "-o", sharedToOrgName)
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say(`Unable to share service instance %s with spaces \['%s'\].`, serviceInstance, sharedToSpaceGUID))
					Eventually(session.Err).Should(Say("Write permission is required in order to share a service instance with a space"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("my targeted space is the same as my share-to space", func() {
				It("fails with a cannot share to self error", func() {
					session := helpers.CF("share-service", serviceInstance, "-s", sourceSpaceName)
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("Service instances cannot be shared into the space where they were created"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("a service instance with the same name exists in the shared-to space", func() {
				BeforeEach(func() {
					helpers.CreateSpace(sharedToSpaceName)
					helpers.TargetOrgAndSpace(sourceOrgName, sharedToSpaceName)
					Eventually(helpers.CF("create-service", service, servicePlan, serviceInstance)).Should(Exit(0))
					helpers.TargetOrgAndSpace(sourceOrgName, sourceSpaceName)
				})

				It("fails with a name clash error", func() {
					session := helpers.CF("share-service", serviceInstance, "-s", sharedToSpaceName)
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say(fmt.Sprintf("A service instance called %s already exists in %s", serviceInstance, sharedToSpaceName)))
					Eventually(session).Should(Exit(1))
				})
			})

			When("the service instance is NOT shareable", func() {
				Context("due to global settings", func() {
					BeforeEach(func() {
						helpers.DisableFeatureFlag("service_instance_sharing")
					})

					AfterEach(func() {
						helpers.EnableFeatureFlag("service_instance_sharing")
					})

					It("should display that the service instance feature flag is disabled and exit 1", func() {
						session := helpers.CF("share-service", serviceInstance, "-s", sharedToSpaceName, "-o", sharedToOrgName)
						Eventually(session.Err).Should(Say(`The "service_instance_sharing" feature flag is disabled for this Cloud Foundry platform.`))
						Eventually(session).Should(Exit(1))
					})
				})

				Context("due to service broker settings", func() {
					BeforeEach(func() {
						broker.Configure(false)
						broker.Update()
					})

					It("should display that service instance sharing is disabled for this service and exit 1", func() {
						session := helpers.CF("share-service", serviceInstance, "-s", sharedToSpaceName, "-o", sharedToOrgName)
						Eventually(session.Err).Should(Say("Service instance sharing is disabled for this service."))
						Eventually(session).Should(Exit(1))
					})
				})

				Context("due to global settings AND service broker settings", func() {
					BeforeEach(func() {
						helpers.DisableFeatureFlag("service_instance_sharing")
						broker.Configure(false)
						broker.Update()
					})

					AfterEach(func() {
						helpers.EnableFeatureFlag("service_instance_sharing")
					})

					It("should display that service instance sharing is disabled for this service and exit 1", func() {
						session := helpers.CF("share-service", serviceInstance, "-s", sharedToSpaceName, "-o", sharedToOrgName)
						Eventually(session.Err).Should(Say(`The "service_instance_sharing" feature flag is disabled for this Cloud Foundry platform. Also, service instance sharing is disabled for this service.`))
						Eventually(session).Should(Exit(1))
					})
				})
			})
		})

		When("the service instance does not exist", func() {
			It("fails with a service instance not found error", func() {
				session := helpers.CF("share-service", serviceInstance, "-s", sharedToSpaceName)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("Specified instance not found or not a managed service instance. Sharing is not supported for user provided services."))
				Eventually(session).Should(Exit(1))
			})
		})

		When("I try to share a user-provided-service", func() {
			BeforeEach(func() {
				helpers.CF("create-user-provided-service", serviceInstance, "-p", `{"username":"admin","password":"pa55woRD"}`)
			})

			It("fails with only managed services can be shared", func() {
				session := helpers.CF("share-service", serviceInstance, "-s", sharedToSpaceName, "-o", sharedToOrgName)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("User-provided services cannot be shared"))
				Eventually(session).Should(Exit(1))
			})
		})
	})
})
