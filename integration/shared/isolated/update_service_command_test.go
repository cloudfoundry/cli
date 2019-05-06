package isolated

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("update-service command", func() {
	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			// the upgrade flag is passed here to exercise a particular code path before refactoring
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "update-service", "foo", "--upgrade")
		})
	})

	When("an api is targeted, the user is logged in, and an org and space are targeted", func() {
		var (
			orgName string
		)

		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			var spaceName = helpers.NewSpaceName()
			helpers.SetupCF(orgName, spaceName)
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		When("there are no service instances", func() {
			When("upgrading", func() {
				It("displays an informative error before prompting and exits 1", func() {
					session := helpers.CF("update-service", "non-existent-service", "--upgrade")
					Eventually(session.Err).Should(Say("Service instance non-existent-service not found"))
					Eventually(session).Should(Exit(1))
				})
			})
		})

		When("providing other arguments while upgrading", func() {
			It("displays an informative error message and exits 1", func() {
				session := helpers.CF("update-service", "irrelevant", "--upgrade", "-c", "{\"hello\": \"world\"}")
				Eventually(session.Err).Should(Say("Incorrect Usage: The following arguments cannot be used together: --upgrade, -t, -c, -p"))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Exit(1))
			})
		})

		When("there is a service instance", func() {
			var (
				service             string
				servicePlan         string
				broker              helpers.ServiceBroker
				serviceInstanceName string
				username            string
			)

			BeforeEach(func() {
				var domain = helpers.DefaultSharedDomain()
				service = helpers.PrefixedRandomName("SERVICE")
				servicePlan = helpers.PrefixedRandomName("SERVICE-PLAN")
				broker = helpers.CreateBroker(domain, service, servicePlan)

				Eventually(helpers.CF("service-access")).Should(Say(service))
				Eventually(helpers.CF("enable-service-access", service)).Should(Exit(0))

				serviceInstanceName = helpers.PrefixedRandomName("SI")
				Eventually(helpers.CF("create-service", service, servicePlan, serviceInstanceName)).Should(Exit(0))

				username, _ = helpers.GetCredentials()
			})

			AfterEach(func() {
				Eventually(helpers.CF("delete-service", serviceInstanceName, "-f")).Should(Exit(0))
				broker.Destroy()
			})

			When("updating to a service plan that does not exist", func() {
				It("displays an informative error message, exits 1", func() {
					session := helpers.CF("update-service", serviceInstanceName, "-p", "non-existing-service-plan")
					Eventually(session).Should(Say("Plan does not exist for the %s service", service))
					Eventually(session).Should(Exit(1))
				})
			})

			When("updating to the same service plan (no-op)", func() {
				It("displays an informative success message, exits 0", func() {
					session := helpers.CF("update-service", serviceInstanceName, "-p", servicePlan)
					Eventually(session).Should(Say("Updating service instance %s as %s...", serviceInstanceName, username))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))
				})
			})

			When("upgrading", func() {
				var buffer *Buffer

				BeforeEach(func() {
					buffer = NewBuffer()
				})

				When("cancelling the update", func() {
					BeforeEach(func() {
						_, err := buffer.Write([]byte("n\n"))
						Expect(err).ToNot(HaveOccurred())
					})

					It("does not proceed", func() {
						session := helpers.CFWithStdin(buffer, "update-service", serviceInstanceName, "--upgrade")
						Eventually(session).Should(Say("You are about to update %s", serviceInstanceName))
						Eventually(session).Should(Say("Warning: This operation may be long running and will block further operations on the service until complete."))
						Eventually(session).Should(Say("Really update service %s\\? \\[yN\\]:", serviceInstanceName))
						Eventually(session).Should(Say("Update cancelled"))
						Eventually(session).Should(Exit(0))
					})
				})

				When("proceeding with the update", func() {
					BeforeEach(func() {
						helpers.SkipIfVersionLessThan(ccversion.MinVersionUpdateServiceInstanceMaintenanceInfoV2)
						_, err := buffer.Write([]byte("y\n"))
						Expect(err).ToNot(HaveOccurred())
					})

					It("updates the service", func() {
						session := helpers.CFWithStdin(buffer, "update-service", serviceInstanceName, "--upgrade")

						By("displaying an informative message")
						Eventually(session).Should(Say("You are about to update %s", serviceInstanceName))
						Eventually(session).Should(Say("Warning: This operation may be long running and will block further operations on the service until complete."))
						Eventually(session).Should(Say("Really update service %s\\? \\[yN\\]:", serviceInstanceName))
						Eventually(session).Should(Exit(0))

						By("requesting an upgrade from the platform")
						session = helpers.CF("service", serviceInstanceName)
						Eventually(session).Should(Say("status:\\s+update succeeded"))
					})
				})
			})
		})
	})
})
