package isolated

import (
	"fmt"
	"time"

	"code.cloudfoundry.org/cli/v9/integration/helpers"
	"code.cloudfoundry.org/cli/v9/integration/helpers/servicebrokerstub"
	"code.cloudfoundry.org/cli/v9/resources"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

// Note that this test requires cloud_controller_ng property "cc.max_service_credential_bindings_per_app_service_instance" to be >= 2
var _ = Describe("cleanup-outdated-service-bindings command", func() {
	const command = "cleanup-outdated-service-bindings"

	Describe("help", func() {
		matchHelpMessage := SatisfyAll(
			Say(`NAME:\n`),
			Say(`\s+cleanup-outdated-service-bindings - Cleans up old service bindings for an app, keeping only the most recent binding for each service instance\n`),
			Say(`\n`),
			Say(`USAGE:\n`),
			Say(`\s+cf cleanup-outdated-service-bindings APP_NAME \[--keep-last N\] \[--service-instance SERVICE_INSTANCE_NAME\] \[--force\] \[--wait\]\n`),
			Say(`\n`),
			Say(`EXAMPLES:\n`),
			Say(`\s+cf cleanup-outdated-service-bindings myapp\n`),
			Say(`\s+cf cleanup-outdated-service-bindings myapp --keep-last 2 --service-instance myinstance --wait\n`),
			Say(`\n`),
			Say(`OPTIONS:\n`),
			Say(`\s+--force, -f\s+Force deletion without confirmation\n`),
			Say(`\s+--keep-last\s+Keep the last N service bindings \(default: 1\)\n`),
			Say(`\s+--service-instance\s+Only delete service bindings for the specified service instance\n`),
			Say(`\s+--wait, -w\s+Wait for the operation\(s\) to complete\n`),
			Say(`\n`),
			Say(`SEE ALSO:\n`),
			Say(`\s+bind-service, unbind-service\n`),
		)

		When("the -h flag is specified", func() {
			It("succeeds and prints help", func() {
				session := helpers.CF(command, "-h")
				Eventually(session).Should(Exit(0))
				Expect(session.Out).To(matchHelpMessage)
			})
		})

		When("the --help flag is specified", func() {
			It("succeeds and prints help", func() {
				session := helpers.CF(command, "--help")
				Eventually(session).Should(Exit(0))
				Expect(session.Out).To(matchHelpMessage)
			})
		})

		When("no arguments are provided", func() {
			It("displays a warning, the help text, and exits 1", func() {
				session := helpers.CF(command)
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Incorrect Usage: the required argument `APP_NAME` was not provided"))
				Expect(session.Out).To(matchHelpMessage)
			})
		})

		When("unknown flag is passed", func() {
			It("displays a warning, the help text, and exits 1", func() {
				session := helpers.CF(command, "-u")
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Incorrect Usage: unknown flag `u"))
				Expect(session.Out).To(matchHelpMessage)
			})
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "cleanup-outdated-service-bindings", "app-name")
		})
	})

	Describe("cleaning up service bindings", func() {
		var (
			orgName   string
			spaceName string
			username  string
			broker    *servicebrokerstub.ServiceBrokerStub
			input     *Buffer
		)

		getBindingCount := func(serviceInstanceName string) int {
			var receiver struct {
				Resources []resources.ServiceCredentialBinding `json:"resources"`
			}
			helpers.Curlf(&receiver, "/v3/service_credential_bindings?service_instance_names=%s", serviceInstanceName)
			return len(receiver.Resources)
		}

		createServiceInstanceWithTwoBindings := func(appName string) (serviceInstanceName, oldServiceBindingGUID string) {
			serviceInstanceName = helpers.NewServiceInstanceName()
			helpers.CreateManagedServiceInstance(broker.FirstServiceOfferingName(), broker.FirstServicePlanName(), serviceInstanceName)

			Eventually(helpers.CF("bind-service", appName, serviceInstanceName, "--wait")).Should(Exit(0))

			appGUID := helpers.AppGUID(appName)

			var receiver struct {
				Resources []resources.ServiceCredentialBinding `json:"resources"`
			}
			helpers.Curlf(&receiver, "/v3/service_credential_bindings?app_guids=%s&service_instance_names=%s", appGUID, serviceInstanceName)
			Expect(receiver.Resources).To(HaveLen(1))

			oldServiceBindingGUID = receiver.Resources[0].GUID

			jsonBody := fmt.Sprintf(`
{
    "type": "app",
    "relationships": {
      "service_instance": {
        "data": {
          "guid": "%s"
        }
      },
      "app": {
        "data": {
          "guid": "%s"
        }
      }
    },
    "strategy": "multiple"
}
`, helpers.ServiceInstanceGUID(serviceInstanceName), appGUID)
			helpers.CF("curl", "-d", jsonBody, "-X", "POST", "/v3/service_credential_bindings")
			// TODO uncomment and remove previous curl
			//Eventually(helpers.CF("bind-service", appName, serviceInstanceName, "--wait", "--strategy=multiple")).Should(Exit(0))
			return
		}

		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()
			helpers.SetupCF(orgName, spaceName)
			username, _ = helpers.GetCredentials()
			broker = servicebrokerstub.EnableServiceAccess()
			input = NewBuffer()
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
			broker.Forget()
		})

		Context("one service binding", func() {
			var (
				serviceInstanceName string
				appName             string
			)

			BeforeEach(func() {
				serviceInstanceName = helpers.NewServiceInstanceName()
				helpers.CreateManagedServiceInstance(broker.FirstServiceOfferingName(), broker.FirstServicePlanName(), serviceInstanceName)

				appName = helpers.NewAppName()
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CF("push", appName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
				})

				Eventually(helpers.CF("bind-service", appName, serviceInstanceName, "--wait")).Should(Exit(0))
			})

			It("does nothing", func() {
				session := helpers.CF(command, appName)
				Eventually(session).Should(Exit(0))

				Expect(session.Out).To(SatisfyAll(
					Say(`Cleaning up outdated service bindings for app %s in org %s / space %s as %s\.\.\.\n`, appName, orgName, spaceName, username),
					Say(`No outdated service bindings found\.`),
					Say(`OK\n`),
				))

				Expect(string(session.Err.Contents())).To(BeEmpty())
				Expect(getBindingCount(serviceInstanceName)).To(Equal(1))
			})
		})

		Context("two service bindings", func() {
			var (
				serviceInstanceName   string
				appName               string
				oldServiceBindingGUID string
			)

			BeforeEach(func() {
				appName = helpers.NewAppName()
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CF("push", appName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
				})

				serviceInstanceName, oldServiceBindingGUID = createServiceInstanceWithTwoBindings(appName)
			})

			It("deletes the oldest binding", func() {
				_, err := input.Write([]byte("y\n"))
				Expect(err).ToNot(HaveOccurred())

				session := helpers.CFWithStdin(input, command, appName)
				Eventually(session).Should(Exit(0))

				Expect(session.Out).To(SatisfyAll(
					Say(`Cleaning up outdated service bindings for app %s in org %s / space %s as %s\.\.\.\n`, appName, orgName, spaceName, username),
					Say(`Found 1 outdated service binding\.`),
					Say("Really delete all outdated service bindings?"),
					Say(`Deleting service binding %s...`, oldServiceBindingGUID),
					Say(`OK\n`),
				))

				Expect(string(session.Err.Contents())).To(BeEmpty())
				Expect(getBindingCount(serviceInstanceName)).To(Equal(1))
			})

			When("the user inputs 'N' to confirmation", func() {
				It("deletes nothing", func() {
					_, err := input.Write([]byte("N\n"))
					Expect(err).ToNot(HaveOccurred())

					session := helpers.CFWithStdin(input, command, appName)
					Eventually(session).Should(Exit(0))

					Expect(session.Out).To(SatisfyAll(
						Say(`Cleaning up outdated service bindings for app %s in org %s / space %s as %s\.\.\.\n`, appName, orgName, spaceName, username),
						Say(`Found 1 outdated service binding\.`),
						Say("Really delete all outdated service bindings?"),
						Say("Outdated service bindings have not been deleted."),
					))

					Expect(string(session.Err.Contents())).To(BeEmpty())
					Expect(getBindingCount(serviceInstanceName)).To(Equal(2))
				})
			})

			When("the --force flag is set", func() {
				It("deletes the oldest binding without asking for confirmation", func() {
					session := helpers.CFWithStdin(input, command, appName, "--force")
					Eventually(session).Should(Exit(0))

					Expect(session.Out).To(SatisfyAll(
						Say(`Cleaning up outdated service bindings for app %s in org %s / space %s as %s\.\.\.\n`, appName, orgName, spaceName, username),
						Say(`Found 1 outdated service binding\.`),
						Say(`Deleting service binding %s...`, oldServiceBindingGUID),
						Say(`OK\n`),
					))

					Expect(string(session.Err.Contents())).To(BeEmpty())
					Expect(getBindingCount(serviceInstanceName)).To(Equal(1))
				})
			})

			When("--keep-last 2 flag is specified", func() {
				It("does nothing", func() {
					session := helpers.CF(command, appName, "--keep-last", "2")
					Eventually(session).Should(Exit(0))

					Expect(session.Out).To(SatisfyAll(
						Say(`Cleaning up outdated service bindings for app %s in org %s / space %s as %s\.\.\.\n`, appName, orgName, spaceName, username),
						Say(`No outdated service bindings found.`),
						Say(`OK\n`),
					))

					Expect(string(session.Err.Contents())).To(BeEmpty())
					Expect(getBindingCount(serviceInstanceName)).To(Equal(2))
				})
			})

			Context("asynchronous broker response", func() {
				BeforeEach(func() {
					broker.WithAsyncDelay(time.Second).Configure()
				})

				It("starts to delete the oldest binding", func() {
					_, err := input.Write([]byte("y\n"))
					Expect(err).ToNot(HaveOccurred())

					session := helpers.CFWithStdin(input, command, appName)
					Eventually(session).Should(Exit(0))

					Expect(session.Out).To(SatisfyAll(
						Say(`Cleaning up outdated service bindings for app %s in org %s / space %s as %s\.\.\.\n`, appName, orgName, spaceName, username),
						Say(`Found 1 outdated service binding\.`),
						Say(`Deleting service binding %s...`, oldServiceBindingGUID),
						Say(`OK\n`),
						Say(`Unbinding in progress. Use 'cf service %s' to check operation status\.\n`, serviceInstanceName),
					))

					Expect(string(session.Err.Contents())).To(BeEmpty())
					Expect(getBindingCount(serviceInstanceName)).To(Equal(2))
				})

				When("--wait flag is specified", func() {
					It("waits for completion", func() {
						_, err := input.Write([]byte("y\n"))
						Expect(err).ToNot(HaveOccurred())

						session := helpers.CFWithStdin(input, command, appName, "--wait")
						Eventually(session).Should(Exit(0))

						Expect(session.Out).To(SatisfyAll(
							Say(`Cleaning up outdated service bindings for app %s in org %s / space %s as %s\.\.\.\n`, appName, orgName, spaceName, username),
							Say(`Found 1 outdated service binding\.`),
							Say(`Deleting service binding %s...`, oldServiceBindingGUID),
							Say(`Waiting for the operation to complete\.+\n`),
							Say(`OK\n`),
						))

						Expect(string(session.Err.Contents())).To(BeEmpty())
						Expect(getBindingCount(serviceInstanceName)).To(Equal(1))
					})
				})
			})

			When("a service instance name is specified", func() {
				Context("two service instances with two bindings each", func() {
					var (
						serviceInstance2Name   string
						oldServiceBinding2GUID string
					)
					BeforeEach(func() {
						serviceInstance2Name, oldServiceBinding2GUID = createServiceInstanceWithTwoBindings(appName)
					})

					It("deletes the oldest binding only of the specified service instance", func() {
						_, err := input.Write([]byte("y\n"))
						Expect(err).ToNot(HaveOccurred())

						session := helpers.CFWithStdin(input, command, appName, "--service-instance", serviceInstance2Name)
						Eventually(session).Should(Exit(0))

						Expect(session.Out).To(SatisfyAll(
							Say(`Cleaning up outdated service bindings for app %s in org %s / space %s as %s\.\.\.\n`, appName, orgName, spaceName, username),
							Say(`Found 1 outdated service binding\.`),
							Say(`Deleting service binding %s...`, oldServiceBinding2GUID),
							Say(`OK\n`),
						))

						Expect(string(session.Err.Contents())).To(BeEmpty())
						Expect(getBindingCount(serviceInstanceName)).To(Equal(2))
						Expect(getBindingCount(serviceInstance2Name)).To(Equal(1))
					})
				})

				Context("service instance does not exist", func() {
					It("displays FAILED and service not found", func() {
						session := helpers.CF(command, appName, "--service-instance", "does-not-exist")
						Eventually(session).Should(Exit(1))
						Expect(session.Out).To(Say("FAILED"))
						Expect(session.Err).To(Say("Service instance 'does-not-exist' not found"))
					})
				})
			})
		})

		Context("app does not exist", func() {
			It("displays FAILED and app not found", func() {
				session := helpers.CF(command, "does-not-exist")
				Eventually(session).Should(Exit(1))
				Expect(session.Out).To(Say("FAILED"))
				Expect(session.Err).To(Say("App 'does-not-exist' not found"))
			})
		})
	})
})
