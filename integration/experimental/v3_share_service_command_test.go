package experimental

import (
	"fmt"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	. "github.com/onsi/gomega/ghttp"
)

var _ = PDescribe("v3-share-service command", func() {
	var (
		sourceOrgName     string
		sourceSpaceName   string
		sharedToOrgName   string
		sharedToSpaceName string
		serviceInstance   string
	)

	BeforeEach(func() {
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
		Context("when --help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("v3-share-service", "--help")
				Eventually(session.Out).Should(Say("NAME:"))
				Eventually(session.Out).Should(Say("v3-share-service - Share a service instance with another space"))
				Eventually(session.Out).Should(Say("USAGE:"))
				Eventually(session.Out).Should(Say("cf v3-share-service SERVICE_INSTANCE -s OTHER_SPACE \\[-o OTHER_ORG\\]"))
				Eventually(session.Out).Should(Say("OPTIONS:"))
				Eventually(session.Out).Should(Say("-o\\s+Org of the other space \\(Default: targeted org\\)"))
				Eventually(session.Out).Should(Say("-s\\s+Space to share the service instance into"))
				Eventually(session.Out).Should(Say("SEE ALSO:"))
				Eventually(session.Out).Should(Say("bind-service, service, services, v3-unshare-service"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("when the service instance name is not provided", func() {
		It("tells the user that the service instance name is required, prints help text, and exits 1", func() {
			session := helpers.CF("v3-share-service", "-s", sharedToSpaceName)

			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `SERVICE_INSTANCE` was not provided"))
			Eventually(session.Out).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	Context("when the space name is not provided", func() {
		It("tells the user that the space name is required, prints help text, and exits 1", func() {
			session := helpers.CF("v3-share-service")

			Eventually(session.Err).Should(Say("Incorrect Usage: the required flag `-s' was not specified"))
			Eventually(session.Out).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	It("displays the experimental warning", func() {
		session := helpers.CF("v3-share-service", serviceInstance, "-s", sharedToSpaceName)
		Eventually(session.Out).Should(Say("This command is in EXPERIMENTAL stage and may change without notice"))
		Eventually(session).Should(Exit())
	})

	Context("when the environment is not setup correctly", func() {
		Context("when no API endpoint is set", func() {
			BeforeEach(func() {
				helpers.UnsetAPI()
			})

			It("fails with no API endpoint set message", func() {
				session := helpers.CF("v3-share-service", serviceInstance, "-s", sharedToSpaceName)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("No API endpoint set\\. Use 'cf login' or 'cf api' to target an endpoint\\."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the v3 api does not exist", func() {
			var server *Server

			BeforeEach(func() {
				server = helpers.StartAndTargetServerWithoutV3API()
			})

			AfterEach(func() {
				server.Close()
			})

			It("fails with error message that the minimum version is not met", func() {
				session := helpers.CF("v3-share-service", serviceInstance, "-s", sharedToSpaceName)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("This command requires CF API version 3\\.34\\.0 or higher\\."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the v3 api version is lower than the minimum version", func() {
			var server *Server

			BeforeEach(func() {
				server = helpers.StartAndTargetServerWithV3Version("3.0.0")
			})

			AfterEach(func() {
				server.Close()
			})

			It("fails with error message that the minimum version is not met", func() {
				session := helpers.CF("v3-share-service", serviceInstance, "-s", sharedToSpaceName)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("This command requires CF API version 3\\.34\\.0 or higher\\."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when not logged in", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
			})

			It("fails with not logged in message", func() {
				session := helpers.CF("v3-share-service", serviceInstance, "-s", sharedToSpaceName)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("Not logged in\\. Use 'cf login' to log in\\."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when there is no org set", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
				helpers.LoginCF()
			})

			It("fails with no org targeted error message", func() {
				session := helpers.CF("v3-share-service", serviceInstance, "-s", sharedToSpaceName)
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("No org targeted, use 'cf target -o ORG' to target an org\\."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when there is no space set", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
				helpers.LoginCF()
				helpers.TargetOrg(ReadOnlyOrg)
			})

			It("fails with no space targeted error message", func() {
				session := helpers.CF("v3-share-service", serviceInstance, "-s", sharedToSpaceName)
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("No space targeted, use 'cf target -s SPACE' to target a space\\."))
				Eventually(session).Should(Exit(1))
			})
		})
	})

	Context("when the environment is set up correctly", func() {
		var (
			domain      string
			service     string
			servicePlan string
		)

		BeforeEach(func() {
			service = helpers.PrefixedRandomName("SERVICE")
			servicePlan = helpers.PrefixedRandomName("SERVICE-PLAN")

			helpers.CreateOrgAndSpace(sharedToOrgName, sharedToSpaceName)
			setupCF(sourceOrgName, sourceSpaceName)

			domain = defaultSharedDomain()
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(sourceOrgName)
			helpers.QuickDeleteOrg(sharedToOrgName)
		})

		Context("when there is a managed service instance in my current targeted space", func() {
			var broker helpers.ServiceBroker

			BeforeEach(func() {
				broker = helpers.NewServiceBroker(helpers.NewServiceBrokerName(), helpers.NewAssets().ServiceBroker, domain, service, servicePlan)
				broker.Push()
				broker.Configure()
				broker.Create()

				Eventually(helpers.CF("enable-service-access", service)).Should(Exit(0))
				Eventually(helpers.CF("create-service", service, servicePlan, serviceInstance)).Should(Exit(0))
			})

			AfterEach(func() {
				broker.Destroy()
			})

			Context("when I want to share my service instance to a space in another org", func() {
				It("shares the service instance from my targeted space with the share-to org/space", func() {
					session := helpers.CF("v3-share-service", serviceInstance, "-s", sharedToSpaceName, "-o", sharedToOrgName)
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when I want to share my service instance into another space in my targeted org", func() {
				BeforeEach(func() {
					helpers.CreateSpace(sharedToSpaceName)
				})

				It("shares the service instance from my targeted space with the share-to space", func() {
					session := helpers.CF("v3-share-service", serviceInstance, "-s", sharedToSpaceName)
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when the org I want to share into does not exist", func() {
				It("fails with an org not found error", func() {
					session := helpers.CF("v3-share-service", serviceInstance, "-s", sharedToSpaceName, "-o", "missing-org")
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("Organization 'missing-org' not found"))
					Eventually(session).Should(Exit(1))
				})
			})

			Context("when the space I want to share into does not exist", func() {
				It("fails with a space not found error", func() {
					session := helpers.CF("v3-share-service", serviceInstance, "-s", "missing-space")
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("Space 'missing-space' not found"))
					Eventually(session).Should(Exit(1))
				})
			})

			Context("when I am a SpaceAuditor in the space I want to share into", func() {
				BeforeEach(func() {
					user := helpers.NewUsername()
					password := helpers.NewPassword()
					Eventually(helpers.CF("create-user", user, password)).Should(Exit(0))
					Eventually(helpers.CF("set-space-role", user, sourceOrgName, sourceSpaceName, "SpaceDeveloper")).Should(Exit(0))
					Eventually(helpers.CF("set-space-role", user, sharedToOrgName, sharedToSpaceName, "SpaceAuditor")).Should(Exit(0))
					Eventually(helpers.CF("auth", user, password)).Should(Exit(0))
					Eventually(helpers.CF("target", "-o", sourceOrgName, "-s", sourceSpaceName)).Should(Exit(0))
				})

				AfterEach(func() {
					setupCF(sourceOrgName, sourceSpaceName)
				})

				It("fails with an unauthorized error", func() {
					session := helpers.CF("v3-share-service", serviceInstance, "-s", sharedToSpaceName, "-o", sharedToOrgName)
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("You are not authorized to perform the requested action"))
					Eventually(session).Should(Exit(1))
				})
			})

			Context("when my targeted space is the same as my share-to space", func() {
				It("fails with a cannot share to self error", func() {
					session := helpers.CF("v3-share-service", serviceInstance, "-s", sourceSpaceName)
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("Service instances cannot be shared into the space where they were created"))
					Eventually(session).Should(Exit(1))
				})
			})

			Context("when a service instance with the same name exists in the shared-to space", func() {
				BeforeEach(func() {
					helpers.CreateSpace(sharedToSpaceName)
					helpers.TargetOrgAndSpace(sourceOrgName, sharedToSpaceName)
					Eventually(helpers.CF("create-service", service, servicePlan, serviceInstance)).Should(Exit(0))
					helpers.TargetOrgAndSpace(sourceOrgName, sourceSpaceName)
				})

				It("fails with a name clash error", func() {
					session := helpers.CF("v3-share-service", serviceInstance, "-s", sharedToSpaceName)
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say(fmt.Sprintf("A service instance called %s already exists in %s", serviceInstance, sharedToSpaceName)))
					Eventually(session).Should(Exit(1))
				})
			})
		})

		Context("when the service instance does not exist", func() {
			It("fails with a service instance not found error", func() {
				session := helpers.CF("v3-share-service", serviceInstance, "-s", sharedToSpaceName)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("Specified instance not found or not a managed service instance. Sharing is not supported for user provided services."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when I try to share a user-provided-service", func() {
			BeforeEach(func() {
				helpers.CF("create-user-provided-service", serviceInstance, "-p", "\"foo, bar\"")
			})

			It("fails with only managed services can be shared", func() {
				session := helpers.CF("v3-share-service", serviceInstance, "-s", sharedToSpaceName)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("Specified instance not found or not a managed service instance. Sharing is not supported for user provided services."))
				Eventually(session).Should(Exit(1))
			})
		})
	})
})
