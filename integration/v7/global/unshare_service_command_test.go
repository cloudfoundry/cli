package global

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/integration/helpers/servicebrokerstub"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("unshare-service command", func() {
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
	})

	Describe("help", func() {
		When("--help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("unshare-service", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("unshare-service - Unshare a shared service instance from a space"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say(`cf unshare-service SERVICE_INSTANCE -s OTHER_SPACE \[-o OTHER_ORG\] \[-f\]`))
				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say(`-o\s+Org of the other space \(Default: targeted org\)`))
				Eventually(session).Should(Say(`-s\s+Space to unshare the service instance from`))
				Eventually(session).Should(Say(`-f\s+Force unshare without confirmation`))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("delete-service, service, services, share-service, unbind-service"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the service instance name is not provided", func() {
		It("tells the user that the service instance name is required, prints help text, and exits 1", func() {
			session := helpers.CF("unshare-service", "-s", sharedToSpaceName)

			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `SERVICE_INSTANCE` was not provided"))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	When("the space name is not provided", func() {
		It("tells the user that the space name is required, prints help text, and exits 1", func() {
			session := helpers.CF("unshare-service")

			Eventually(session.Err).Should(Say("Incorrect Usage: the required flag `-s' was not specified"))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "unshare-service", serviceInstance, "-s", sharedToSpaceName)
		})

		When("the v3 api version is lower than the minimum version", func() {
			var server *Server

			BeforeEach(func() {
				server = helpers.StartAndTargetMockServerWithAPIVersions(helpers.DefaultV2Version, "3.0.0")
			})

			AfterEach(func() {
				server.Close()
			})

			It("fails with error message that the minimum version is not met", func() {
				session := helpers.CF("unshare-service", serviceInstance, "-s", sharedToSpaceName)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say(`This command requires CF API version 3\.63\.0 or higher\.`))
				Eventually(session).Should(Exit(1))
			})
		})
	})

	When("the environment is set up correctly", func() {
		var (
			service     string
			servicePlan string
		)

		BeforeEach(func() {
			helpers.CreateOrgAndSpace(sharedToOrgName, sharedToSpaceName)
			helpers.SetupCF(sourceOrgName, sourceSpaceName)
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(sourceOrgName)
			helpers.QuickDeleteOrg(sharedToOrgName)
		})

		When("there is a managed service instance in my current targeted space", func() {
			var broker *servicebrokerstub.ServiceBrokerStub

			BeforeEach(func() {
				broker = servicebrokerstub.EnableServiceAccess()
				service = broker.FirstServiceOfferingName()
				servicePlan = broker.FirstServicePlanName()

				Eventually(helpers.CF("create-service", service, servicePlan, serviceInstance)).Should(Exit(0))
			})

			AfterEach(func() {
				broker.Forget()
			})

			When("the service instance has not been shared to this space", func() {
				It("displays info and idempotently exits 0", func() {
					session := helpers.CF("unshare-service", serviceInstance, "-s", sharedToSpaceName, "-o", sharedToOrgName, "-f")
					Eventually(session).Should(Say(`Service instance %s is not shared with space %s in organization %s\.`, serviceInstance, sharedToSpaceName, sharedToOrgName))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))
				})
			})

			When("I have shared my service instance to a space in another org ('-o' flag provided)", func() {
				BeforeEach(func() {
					session := helpers.CF("share-service", serviceInstance, "-s", sharedToSpaceName, "-o", sharedToOrgName)
					Eventually(session).Should(Exit(0))
				})

				When("the org I want to unshare from does not exist", func() {
					It("fails with an org not found error", func() {
						session := helpers.CF("unshare-service", serviceInstance, "-s", sharedToSpaceName, "-o", "missing-org", "-f")
						Eventually(session).Should(Say(`Service instance %s is not shared with space %s in organization missing-org\.`, serviceInstance, sharedToSpaceName))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Exit(0))
					})
				})

				When("the space I want to unshare from does not exist", func() {
					It("fails with a space not found error", func() {
						session := helpers.CF("unshare-service", serviceInstance, "-s", "missing-space", "-o", sharedToOrgName, "-f")
						Eventually(session).Should(Say(`Service instance %s is not shared with space missing-space in organization %s\.`, serviceInstance, sharedToOrgName))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Exit(0))
					})
				})

				When("I want to unshare my service instance from a space and org", func() {
					It("successfully unshares the service instance", func() {
						username, _ := helpers.GetCredentials()
						session := helpers.CF("unshare-service", serviceInstance, "-s", sharedToSpaceName, "-o", sharedToOrgName, "-f")
						Eventually(session).Should(Say(`Unsharing service instance %s from org %s / space %s as %s\.\.\.`, serviceInstance, sharedToOrgName, sharedToSpaceName, username))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Exit(0))
					})
				})
			})

			When("I have shared my service instance to a space within the targeted org ('-o' flag NOT provided)", func() {
				BeforeEach(func() {
					helpers.CreateSpace(sharedToSpaceName)

					session := helpers.CF("share-service", serviceInstance, "-s", sharedToSpaceName)
					Eventually(session).Should(Exit(0))
				})

				When("the space I want to unshare from does not exist", func() {
					It("fails with a space not found error", func() {
						session := helpers.CF("unshare-service", serviceInstance, "-s", "missing-space", "-f")
						Eventually(session).Should(Say(`Service instance %s is not shared with space missing-space in organization %s\.`, serviceInstance, sourceOrgName))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Exit(0))
					})
				})

				When("I want to unshare my service instance from the space", func() {
					It("successfully unshares the service instance when I am admin", func() {
						username, _ := helpers.GetCredentials()
						session := helpers.CF("unshare-service", serviceInstance, "-s", sharedToSpaceName, "-f")
						Eventually(session).Should(Say(`Unsharing service instance %s from org %s / space %s as %s\.\.\.`, serviceInstance, sourceOrgName, sharedToSpaceName, username))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Exit(0))
					})

					When("I have no access to the shared-to space", func() {
						var (
							username string
							password string
						)

						BeforeEach(func() {
							username = helpers.NewUsername()
							password = helpers.NewPassword()
							Eventually(helpers.CF("create-user", username, password)).Should(Exit(0))
							Eventually(helpers.CF("set-space-role", username, sourceOrgName, sourceSpaceName, "SpaceDeveloper")).Should(Exit(0))
							env := map[string]string{
								"CF_USERNAME": username,
								"CF_PASSWORD": password,
							}
							helpers.LogoutCF()
							Eventually(helpers.CFWithEnv(env, "auth")).Should(Exit(0))
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
							session := helpers.CF("unshare-service", serviceInstance, "-s", sharedToSpaceName, "-f")
							Eventually(session).Should(Say("OK"))
							Eventually(session).Should(Exit(0))
						})
					})
				})
			})
		})

		When("the service instance does not exist", func() {
			When("the -f flag is provided", func() {
				It("fails with a service instance not found error", func() {
					session := helpers.CF("unshare-service", serviceInstance, "-s", sharedToSpaceName, "-f")
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say(`Specified instance not found or not a managed service instance\. Sharing is not supported for user provided services\.`))
					Eventually(session).Should(Exit(1))
				})
			})

			When("the -f flag not is provided", func() {
				var buffer *Buffer

				BeforeEach(func() {
					buffer = NewBuffer()
				})

				When("the user enters 'y'", func() {
					BeforeEach(func() {
						_, err := buffer.Write([]byte("y\n"))
						Expect(err).ToNot(HaveOccurred())
					})

					It("fails with a service instance not found error", func() {
						username, _ := helpers.GetCredentials()
						session := helpers.CFWithStdin(buffer, "unshare-service", serviceInstance, "-s", sharedToSpaceName)
						Eventually(session.Err).Should(Say(`WARNING: Unsharing this service instance will remove any existing bindings originating from the service instance in the space "%s". This could cause apps to stop working.`, sharedToSpaceName))
						Eventually(session).Should(Say(`Really unshare the service instance\? \[yN\]`))
						Eventually(session).Should(Say(`Unsharing service instance %s from org %s / space %s as %s\.\.\.`, serviceInstance, sourceOrgName, sharedToSpaceName, username))
						Eventually(session).Should(Say("FAILED"))
						Eventually(session.Err).Should(Say(`Specified instance not found or not a managed service instance\. Sharing is not supported for user provided services\.`))
						Eventually(session).Should(Exit(1))
					})
				})

				When("the user enters 'n'", func() {
					BeforeEach(func() {
						_, err := buffer.Write([]byte("n\n"))
						Expect(err).ToNot(HaveOccurred())
					})

					It("does not attempt to unshare", func() {
						session := helpers.CFWithStdin(buffer, "unshare-service", serviceInstance, "-s", sharedToSpaceName)
						Eventually(session.Err).Should(Say(`WARNING: Unsharing this service instance will remove any existing bindings originating from the service instance in the space "%s". This could cause apps to stop working.`, sharedToSpaceName))
						Eventually(session).Should(Say(`Really unshare the service instance\? \[yN\]`))
						Eventually(session).Should(Say("Unshare cancelled"))
						Eventually(session).Should(Exit(0))
					})
				})

				When("the user enters the default input (hits return)", func() {
					BeforeEach(func() {
						_, err := buffer.Write([]byte("\n"))
						Expect(err).ToNot(HaveOccurred())
					})

					It("does not attempt to unshare", func() {
						session := helpers.CFWithStdin(buffer, "unshare-service", serviceInstance, "-s", sharedToSpaceName)
						Eventually(session.Err).Should(Say(`WARNING: Unsharing this service instance will remove any existing bindings originating from the service instance in the space "%s". This could cause apps to stop working.`, sharedToSpaceName))
						Eventually(session).Should(Say(`Really unshare the service instance\? \[yN\]`))
						Eventually(session).Should(Say("Unshare cancelled"))
						Eventually(session).Should(Exit(0))
					})
				})

				When("the user enters an invalid answer", func() {
					BeforeEach(func() {
						// The second '\n' is intentional. Otherwise the buffer will be
						// closed while the interaction is still waiting for input; it gets
						// an EOF and causes an error.
						_, err := buffer.Write([]byte("wat\n\n"))
						Expect(err).ToNot(HaveOccurred())
					})

					It("asks again", func() {
						session := helpers.CFWithStdin(buffer, "unshare-service", serviceInstance, "-s", sharedToSpaceName)
						Eventually(session.Err).Should(Say(`WARNING: Unsharing this service instance will remove any existing bindings originating from the service instance in the space "%s". This could cause apps to stop working.`, sharedToSpaceName))
						Eventually(session).Should(Say(`Really unshare the service instance\? \[yN\]`))
						Eventually(session).Should(Say(`invalid input \(not y, n, yes, or no\)`))
						Eventually(session).Should(Say(`Really unshare the service instance\? \[yN\]`))
						Eventually(session).Should(Say("Unshare cancelled"))
						Eventually(session).Should(Exit(0))
					})
				})
			})
		})

		When("there is a shared service instance in my currently targeted space", func() {
			var (
				broker   *servicebrokerstub.ServiceBrokerStub
				user     string
				password string
			)

			BeforeEach(func() {
				broker = servicebrokerstub.EnableServiceAccess()
				service = broker.FirstServiceOfferingName()
				servicePlan = broker.FirstServicePlanName()

				user = helpers.NewUsername()
				password = helpers.NewPassword()

				helpers.SetupCF(sourceOrgName, sourceSpaceName)
				Eventually(helpers.CF("enable-service-access", service)).Should(Exit(0))
				Eventually(helpers.CF("create-service", service, servicePlan, serviceInstance)).Should(Exit(0))
				Eventually(helpers.CF("share-service", serviceInstance, "-s", sharedToSpaceName, "-o", sharedToOrgName)).Should(Exit(0))

				Eventually(helpers.CF("create-user", user, password)).Should(Exit(0))
				Eventually(helpers.CF("set-space-role", user, sharedToOrgName, sharedToSpaceName, "SpaceDeveloper")).Should(Exit(0))
			})

			AfterEach(func() {
				helpers.SetupCF(sourceOrgName, sourceSpaceName)
				Eventually(helpers.CF("delete-user", user)).Should(Exit(0))
				broker.Forget()
			})

			Context("and I have no access to the source space", func() {
				BeforeEach(func() {
					env := map[string]string{
						"CF_USERNAME": user,
						"CF_PASSWORD": password,
					}
					helpers.LogoutCF()
					Eventually(helpers.CFWithEnv(env, "auth")).Should(Exit(0))
					Eventually(helpers.CF("target", "-o", sharedToOrgName, "-s", sharedToSpaceName)).Should(Exit(0))
				})

				It("returns a permission error on an attempt to unshare the service", func() {
					session := helpers.CF("unshare-service", serviceInstance, "-s", sharedToSpaceName, "-o", sharedToOrgName, "-f")
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("You are not authorized to perform the requested action"))
					Eventually(session).Should(Exit(1))
				})
			})

			Context("and I have SpaceAuditor access to the source space", func() {
				BeforeEach(func() {
					env := map[string]string{
						"CF_USERNAME": user,
						"CF_PASSWORD": password,
					}
					Eventually(helpers.CF("set-space-role", user, sourceOrgName, sourceSpaceName, "SpaceAuditor")).Should(Exit(0))
					helpers.LogoutCF()
					Eventually(helpers.CFWithEnv(env, "auth")).Should(Exit(0))
					Eventually(helpers.CF("target", "-o", sharedToOrgName, "-s", sharedToSpaceName)).Should(Exit(0))
				})

				It("returns a permission error on an attempt to unshare the service", func() {
					session := helpers.CF("unshare-service", serviceInstance, "-s", sharedToSpaceName, "-o", sharedToOrgName, "-f")
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("You are not authorized to perform the requested action"))
					Eventually(session).Should(Exit(1))
				})
			})
		})
	})
})
