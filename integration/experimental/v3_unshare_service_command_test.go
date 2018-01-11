package experimental

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	. "github.com/onsi/gomega/ghttp"
)

var _ = PDescribe("v3-unshare-service command", func() {
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
				session := helpers.CF("v3-unshare-service", "--help")
				Eventually(session.Out).Should(Say("NAME:"))
				Eventually(session.Out).Should(Say("v3-unshare-service - Unshare a shared service instance from a space"))
				Eventually(session.Out).Should(Say("USAGE:"))
				Eventually(session.Out).Should(Say("cf v3-unshare-service SERVICE_INSTANCE -s OTHER_SPACE \\[-o OTHER_ORG\\] \\[-f\\]"))
				Eventually(session.Out).Should(Say("OPTIONS:"))
				Eventually(session.Out).Should(Say("-o\\s+Org of the other space \\(Default: targeted org\\)"))
				Eventually(session.Out).Should(Say("-s\\s+Space to unshare the service instance from"))
				Eventually(session.Out).Should(Say("-f\\s+Force unshare without confirmation"))
				Eventually(session.Out).Should(Say("SEE ALSO:"))
				Eventually(session.Out).Should(Say("delete-service, service, services, unbind-service, v3-share-service"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("when the service instance name is not provided", func() {
		It("tells the user that the service instance name is required, prints help text, and exits 1", func() {
			session := helpers.CF("v3-unshare-service", "-s", sharedToSpaceName)

			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `SERVICE_INSTANCE` was not provided"))
			Eventually(session.Out).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	Context("when the space name is not provided", func() {
		It("tells the user that the space name is required, prints help text, and exits 1", func() {
			session := helpers.CF("v3-unshare-service")

			Eventually(session.Err).Should(Say("Incorrect Usage: the required flag `-s' was not specified"))
			Eventually(session.Out).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	It("displays the experimental warning", func() {
		session := helpers.CF("v3-unshare-service", serviceInstance, "-s", sharedToSpaceName)
		Eventually(session.Out).Should(Say("This command is in EXPERIMENTAL stage and may change without notice"))
		Eventually(session).Should(Exit())
	})

	Context("when the environment is not setup correctly", func() {
		Context("when no API endpoint is set", func() {
			BeforeEach(func() {
				helpers.UnsetAPI()
			})

			It("fails with no API endpoint set message", func() {
				session := helpers.CF("v3-unshare-service", serviceInstance, "-s", sharedToSpaceName)
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
				session := helpers.CF("v3-unshare-service", serviceInstance, "-s", sharedToSpaceName)
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
				session := helpers.CF("v3-unshare-service", serviceInstance, "-s", sharedToSpaceName)
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
				session := helpers.CF("v3-unshare-service", serviceInstance, "-s", sharedToSpaceName)
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
				session := helpers.CF("v3-unshare-service", serviceInstance, "-s", sharedToSpaceName)
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
				session := helpers.CF("v3-unshare-service", serviceInstance, "-s", sharedToSpaceName)
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

			Context("when a service instance has not been shared to a space", func() {
				It("returns an error on an attempt to unshare it from that space", func() {
					session := helpers.CF("v3-unshare-service", serviceInstance, "-s", sharedToSpaceName, "-o", sharedToOrgName, "-f")
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("Failed to unshare service instance '%s'. Ensure the space and specified org exist and that the service instance has been shared to this space.", serviceInstance))
					Eventually(session).Should(Exit(1))
				})
			})

			Context("when I have shared my service instance to a space in another org", func() {
				BeforeEach(func() {
					session := helpers.CF("v3-share-service", serviceInstance, "-s", sharedToSpaceName, "-o", sharedToOrgName)
					Eventually(session).Should(Exit(0))
				})

				Context("when the org I want to unshare from does not exist", func() {
					It("fails with an org not found error", func() {
						session := helpers.CF("v3-unshare-service", serviceInstance, "-s", sharedToSpaceName, "-o", "missing-org", "-f")
						Eventually(session).Should(Say("FAILED"))
						Eventually(session.Err).Should(Say("Failed to unshare service instance '%s'. Ensure the space and specified org exist and that the service instance has been shared to this space.", serviceInstance))
						Eventually(session).Should(Exit(1))
					})
				})

				Context("when the space I want to unshare from does not exist", func() {
					It("fails with a space not found error", func() {
						session := helpers.CF("v3-unshare-service", serviceInstance, "-s", "missing-space", "-o", sharedToOrgName, "-f")
						Eventually(session).Should(Say("FAILED"))
						Eventually(session.Err).Should(Say("Failed to unshare service instance '%s'. Ensure the space and specified org exist and that the service instance has been shared to this space.", serviceInstance))
						Eventually(session).Should(Exit(1))
					})
				})

				Context("when I want to unshare my service instance from a space and org", func() {
					It("successfully unshares the service instance", func() {
						session := helpers.CF("v3-unshare-service", serviceInstance, "-s", sharedToSpaceName, "-o", sharedToOrgName, "-f")
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Exit(0))
					})
				})
			})

			Context("when I have shared my service instance to a space within the targeted org", func() {
				BeforeEach(func() {
					helpers.CreateSpace(sharedToSpaceName)

					session := helpers.CF("v3-share-service", serviceInstance, "-s", sharedToSpaceName)
					Eventually(session).Should(Exit(0))
				})

				Context("when the space I want to unshare from does not exist", func() {
					It("fails with a space not found error", func() {
						session := helpers.CF("v3-unshare-service", serviceInstance, "-s", "missing-space", "-f")
						Eventually(session).Should(Say("FAILED"))
						Eventually(session.Err).Should(Say("Failed to unshare service instance '%s'. Ensure the space and specified org exist and that the service instance has been shared to this space.", serviceInstance))
						Eventually(session).Should(Exit(1))
					})
				})

				Context("when I want to unshare my service instance from the space", func() {
					It("successfully unshares the service instance when I am admin", func() {
						session := helpers.CF("v3-unshare-service", serviceInstance, "-s", sharedToSpaceName, "-f")
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Exit(0))
					})

					Context("when I have no access to the shared-to space", func() {
						var (
							username string
							password string
						)

						BeforeEach(func() {
							username = helpers.NewUsername()
							password = helpers.NewPassword()
							Eventually(helpers.CF("create-user", username, password)).Should(Exit(0))
							Eventually(helpers.CF("set-space-role", username, sourceOrgName, sourceSpaceName, "SpaceDeveloper")).Should(Exit(0))
							Eventually(helpers.CF("auth", username, password)).Should(Exit(0))
							helpers.TargetOrgAndSpace(sourceOrgName, sourceSpaceName)
						})

						AfterEach(func() {
							helpers.LoginCF()
							helpers.TargetOrgAndSpace(sourceOrgName, sourceSpaceName)
							session := helpers.CF("delete-user", username, "-f")
							Eventually(session).Should(Say("OK"))
							Eventually(session).Should(Exit(0))
						})

						It("successfully unshares the service instance", func() {
							session := helpers.CF("v3-unshare-service", serviceInstance, "-s", sharedToSpaceName, "-f")
							Eventually(session).Should(Say("OK"))
							Eventually(session).Should(Exit(0))
						})
					})
				})
			})
		})

		Context("when the service instance does not exist", func() {
			Context("when the -f flag is provided", func() {
				It("fails with a service instance not found error", func() {
					session := helpers.CF("v3-unshare-service", serviceInstance, "-s", sharedToSpaceName, "-f")
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("Service instance %s not found", serviceInstance))
					Eventually(session).Should(Exit(1))
				})
			})

			Context("when the -f flag not is provided", func() {
				var buffer *Buffer

				BeforeEach(func() {
					buffer = NewBuffer()
				})

				Context("when the user enters 'y'", func() {
					BeforeEach(func() {
						buffer.Write([]byte("y\n"))
					})

					It("fails with a service instance not found error", func() {
						username, _ := helpers.GetCredentials()
						session := helpers.CFWithStdin(buffer, "v3-unshare-service", serviceInstance, "-s", sharedToSpaceName)
						Eventually(session.Err).Should(Say("WARNING: Unsharing this service instance will remove any service bindings that exist in any spaces that this instance is shared into. This could cause applications to stop working."))
						Eventually(session.Out).Should(Say("Really unshare the service instance\\? \\[yN\\]"))
						Eventually(session.Out).Should(Say("Unsharing service instance %s from org %s / space %s as %s...", serviceInstance, sourceOrgName, sharedToSpaceName, username))
						Eventually(session.Out).Should(Say("FAILED"))
						Eventually(session.Err).Should(Say("Service instance %s not found", serviceInstance))
						Eventually(session).Should(Exit(1))
					})
				})

				Context("when the user enters 'n'", func() {
					BeforeEach(func() {
						buffer.Write([]byte("n\n"))
					})

					It("does not attempt to unshare", func() {
						session := helpers.CFWithStdin(buffer, "v3-unshare-service", serviceInstance, "-s", sharedToSpaceName)
						Eventually(session.Err).Should(Say("WARNING: Unsharing this service instance will remove any service bindings that exist in any spaces that this instance is shared into. This could cause applications to stop working."))
						Eventually(session.Out).Should(Say("Really unshare the service instance\\? \\[yN\\]"))
						Eventually(session.Out).Should(Say("Unshare cancelled"))
						Eventually(session).Should(Exit(0))
					})
				})

				Context("when the user enters the default input (hits return)", func() {
					BeforeEach(func() {
						buffer.Write([]byte("\n"))
					})

					It("does not attempt to unshare", func() {
						session := helpers.CFWithStdin(buffer, "v3-unshare-service", serviceInstance, "-s", sharedToSpaceName)
						Eventually(session.Err).Should(Say("WARNING: Unsharing this service instance will remove any service bindings that exist in any spaces that this instance is shared into. This could cause applications to stop working."))
						Eventually(session.Out).Should(Say("Really unshare the service instance\\? \\[yN\\]"))
						Eventually(session.Out).Should(Say("Unshare cancelled"))
						Eventually(session).Should(Exit(0))
					})
				})

				Context("when the user enters an invalid answer", func() {
					BeforeEach(func() {
						// The second '\n' is intentional. Otherwise the buffer will be
						// closed while the interaction is still waiting for input; it gets
						// an EOF and causes an error.
						buffer.Write([]byte("wat\n\n"))
					})

					It("asks again", func() {
						session := helpers.CFWithStdin(buffer, "v3-unshare-service", serviceInstance, "-s", sharedToSpaceName)
						Eventually(session.Err).Should(Say("WARNING: Unsharing this service instance will remove any service bindings that exist in any spaces that this instance is shared into. This could cause applications to stop working."))
						Eventually(session.Out).Should(Say("Really unshare the service instance\\? \\[yN\\]"))
						Eventually(session.Out).Should(Say("invalid input \\(not y, n, yes, or no\\)"))
						Eventually(session.Out).Should(Say("Really unshare the service instance\\? \\[yN\\]"))
						Eventually(session.Out).Should(Say("Unshare cancelled"))
						Eventually(session).Should(Exit(0))
					})
				})
			})
		})
	})
})
