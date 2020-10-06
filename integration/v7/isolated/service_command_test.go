package isolated

import (
	"fmt"
	"strings"
	"time"

	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/integration/helpers/servicebrokerstub"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("service command", func() {
	const serviceCommand = "v3-service"

	Describe("help", func() {
		const serviceInstanceName = "fake-service-instance-name"

		matchHelpMessage := SatisfyAll(
			Say(`NAME:\n`),
			Say(fmt.Sprintf(`\s+%s - Show service instance info\n`, serviceCommand)),
			Say(`\n`),
			Say(`USAGE:\n`),
			Say(`\s+cf service SERVICE_INSTANCE\n`),
			Say(`\n`),
			Say(`OPTIONS:\n`),
			Say(`\s+--guid\s+Retrieve and display the given service's guid. All other output for the service is suppressed.\n`),
			Say(`\n`),
			Say(`SEE ALSO:\n`),
			Say(`\s+bind-service, rename-service, update-service\n`),
			Say(`$`),
		)

		When("the -h flag is specified", func() {
			It("succeeds and prints help", func() {
				session := helpers.CF(serviceCommand, serviceInstanceName, "-h")
				Eventually(session).Should(Exit(0))
				Expect(session.Out).To(matchHelpMessage)
			})
		})

		When("the service instance name is missing", func() {
			It("fails with an error and prints help", func() {
				session := helpers.CF(serviceCommand)
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Incorrect Usage: the required argument `SERVICE_INSTANCE` was not provided"))
				Expect(session.Out).To(matchHelpMessage)
			})
		})

		When("an extra parameter is specified", func() {
			It("fails with an error and prints help", func() {
				session := helpers.CF(serviceCommand, serviceInstanceName, "anotherRandomParameter")
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say(`Incorrect Usage: unexpected argument "anotherRandomParameter"`))
				Expect(session.Out).To(SatisfyAll(
					Say(`FAILED\n\n`),
					matchHelpMessage,
				))
			})
		})

		When("an extra flag is specified", func() {
			It("fails with an error and prints help", func() {
				session := helpers.CF(serviceCommand, serviceInstanceName, "--anotherRandomFlag")
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Incorrect Usage: unknown flag `anotherRandomFlag'"))
				Expect(session.Out).To(matchHelpMessage)
			})
		})
	})

	When("environment is not set up", func() {
		It("displays an error and exits 1", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, serviceCommand, "serviceInstanceName")
		})
	})

	When("user is logged in and targeting a space", func() {
		var (
			serviceInstanceName string
			orgName             string
			spaceName           string
		)

		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()
			helpers.SetupCF(orgName, spaceName)

			serviceInstanceName = helpers.NewServiceInstanceName()
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		When("the service instance is user-provided", func() {
			const (
				routeServiceURL = "https://route.com"
				syslogURL       = "https://syslog.com"
				tags            = "foo, bar"
			)

			BeforeEach(func() {
				command := []string{
					"create-user-provided-service", serviceInstanceName,
					"-r", routeServiceURL,
					"-l", syslogURL,
					"-t", tags,
				}
				Eventually(helpers.CF(command...)).Should(Exit(0))
			})

			It("can show the GUID", func() {
				session := helpers.CF(serviceCommand, serviceInstanceName, "--guid")
				Eventually(session).Should(Exit(0))
				Expect(strings.TrimSpace(string(session.Out.Contents()))).To(HaveLen(36), "GUID wrong length")
			})

			It("can show the service instance details", func() {
				session := helpers.CF(serviceCommand, serviceInstanceName)
				Eventually(session).Should(Exit(0))

				username, _ := helpers.GetCredentials()
				Expect(session).To(SatisfyAll(
					Say(`Showing info of service %s in org %s / space %s as %s...\n`, serviceInstanceName, orgName, spaceName, username),
					Say(`\n`),
					Say(`name:\s+%s\n`, serviceInstanceName),
					Say(`guid:\s+\S{36}\n`),
					Say(`type:\s+user-provided`),
					Say(`tags:\s+%s\n`, tags),
					Say(`route service url:\s+%s\n`, routeServiceURL),
					Say(`syslog drain url:\s+%s\n`, syslogURL),
					Say(`\n`),
					Say(`Bound apps:\n`),
					Say(`There are no bound apps for this service instance\.\n`),
				))
			})

			When("bound to apps", func() {
				var (
					appName1     string
					appName2     string
					bindingName1 string
					bindingName2 string
				)

				BeforeEach(func() {
					appName1 = helpers.PrefixedRandomName("APP1")
					appName2 = helpers.PrefixedRandomName("APP2")
					bindingName1 = helpers.RandomName()
					bindingName2 = helpers.RandomName()

					helpers.WithHelloWorldApp(func(appDir string) {
						Eventually(helpers.CF("push", appName1, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
						Eventually(helpers.CF("push", appName2, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
					})
					Eventually(helpers.CF("bind-service", appName1, serviceInstanceName, "--binding-name", bindingName1)).Should(Exit(0))
					Eventually(helpers.CF("bind-service", appName2, serviceInstanceName, "--binding-name", bindingName2)).Should(Exit(0))
				})

				It("displays the bound apps", func() {
					session := helpers.CF(serviceCommand, serviceInstanceName, "-v")
					Eventually(session).Should(Exit(0))

					Expect(session).To(SatisfyAll(
						Say(`Bound apps:\n`),
						Say(`name\s+binding name\s+status\s+message\n`),
						Say(`%s\s+%s\s+create succeeded\s*\n`, appName1, bindingName1),
						Say(`%s\s+%s\s+create succeeded\s*\n`, appName2, bindingName2),
					))
				})
			})
		})

		When("the service instance is managed by a broker", func() {
			const (
				testPollingInterval = time.Second
				testTimeout         = time.Minute
			)

			var broker *servicebrokerstub.ServiceBrokerStub

			AfterEach(func() {
				broker.Forget()
			})

			When("created successfully", func() {
				const tags = "foo, bar"

				BeforeEach(func() {
					broker = servicebrokerstub.New().WithAsyncDelay(time.Nanosecond).EnableServiceAccess()

					helpers.CreateManagedServiceInstance(
						broker.FirstServiceOfferingName(),
						broker.FirstServicePlanName(),
						serviceInstanceName,
						"-t", tags,
					)
				})

				It("can show the service instance details", func() {
					session := helpers.CF(serviceCommand, serviceInstanceName)
					Eventually(session).Should(Exit(0))

					username, _ := helpers.GetCredentials()
					Expect(session).To(SatisfyAll(
						Say(`Showing info of service %s in org %s / space %s as %s...\n`, serviceInstanceName, orgName, spaceName, username),
						Say(`\n`),
						Say(`name:\s+%s\n`, serviceInstanceName),
						Say(`guid:\s+\S+\n`),
						Say(`type:\s+managed`),
						Say(`broker:\s+%s`, broker.Name),
						Say(`offering:\s+%s`, broker.FirstServiceOfferingName()),
						Say(`plan:\s+%s`, broker.FirstServicePlanName()),
						Say(`tags:\s+%s\n`, tags),
						Say(`offering tags:\s+%s\n`, strings.Join(broker.Services[0].Tags, ", ")),
						Say(`description:\s+%s\n`, broker.Services[0].Description),
						Say(`documentation:\s+%s\n`, broker.Services[0].DocumentationURL),
						Say(`dashboard url:\s+http://example.com\n`),
						Say(`\n`),
						Say(`Showing status of last operation from service instance %s...\n`, serviceInstanceName),
						Say(`\n`),
						Say(`status:\s+create succeeded\n`),
						Say(`message:\s+very happy service\n`),
						Say(`started:\s+%s\n`, helpers.TimestampRegex),
						Say(`updated:\s+%s\n`, helpers.TimestampRegex),
						Say(`\n`),
						Say(`Bound apps:\n`),
						Say(`There are no bound apps for this service instance\.\n`),
						Say(`\n`),
						Say(`No parameters are set for service instance %s.\n`, serviceInstanceName),
						Say(`\n`),
						Say(`Sharing:\n`),
						Say(`This service instance is not currently being shared.`),
						Say(`\n`),
						Say(`Upgrades are not supported by this broker.\n`),
					))
				})
			})

			When("creation is in progress", func() {
				const (
					tags             = "foo, bar"
					brokerAsyncDelay = time.Second
				)

				BeforeEach(func() {
					broker = servicebrokerstub.New().WithAsyncDelay(brokerAsyncDelay).EnableServiceAccess()
					command := []string{
						"create-service",
						broker.FirstServiceOfferingName(),
						broker.FirstServicePlanName(),
						serviceInstanceName,
						"-t", tags,
					}
					Eventually(helpers.CF(command...)).Should(Exit(0))
				})

				It("can show the GUID immediately", func() {
					session := helpers.CF(serviceCommand, serviceInstanceName, "--guid")
					Eventually(session).Should(Exit(0))
					Expect(strings.TrimSpace(string(session.Out.Contents()))).To(HaveLen(36), "GUID wrong length")
				})

				It("can show the service instance details", func() {
					session := helpers.CF(serviceCommand, serviceInstanceName)
					Eventually(session).Should(Exit(0))

					username, _ := helpers.GetCredentials()
					Expect(session).To(SatisfyAll(
						Say(`Showing info of service %s in org %s / space %s as %s...\n`, serviceInstanceName, orgName, spaceName, username),
						Say(`\n`),
						Say(`name:\s+%s\n`, serviceInstanceName),
						Say(`guid:\s+\S+\n`),
						Say(`type:\s+managed`),
						Say(`broker:\s+%s`, broker.Name),
						Say(`offering:\s+%s`, broker.FirstServiceOfferingName()),
						Say(`plan:\s+%s`, broker.FirstServicePlanName()),
						Say(`tags:\s+%s\n`, tags),
						Say(`offering tags:\s+%s\n`, strings.Join(broker.Services[0].Tags, ", ")),
						Say(`description:\s+%s\n`, broker.Services[0].Description),
						Say(`documentation:\s+%s\n`, broker.Services[0].DocumentationURL),
						Say(`dashboard url:\s+http://example.com\n`),
						Say(`\n`),
						Say(`Showing status of last operation from service instance %s...\n`, serviceInstanceName),
						Say(`\n`),
						Say(`status:\s+create in progress\n`),
						Say(`message:\s+very happy service\n`),
						Say(`started:\s+%s\n`, helpers.TimestampRegex),
						Say(`updated:\s+%s\n`, helpers.TimestampRegex),
						Say(`\n`),
						Say(`Bound apps:\n`),
						Say(`There are no bound apps for this service instance\.\n`),
						Say(`\n`),
						Say(`Unable to show parameters: An operation for service instance %s is in progress.`, serviceInstanceName),
						Say(`\n`),
						Say(`Sharing:\n`),
						Say(`This service instance is not currently being shared.`),
						Say(`\n`),
						Say(`Upgrading:\n`),
						Say(`Upgrades are not supported by this broker.\n`),
					))
				})
			})

			When("service instance parameters have been set", func() {
				var parameters string

				BeforeEach(func() {
					broker = servicebrokerstub.EnableServiceAccess()
					parameters = fmt.Sprintf(`{"foo":"%s"}`, helpers.RandomName())
					command := []string{
						"create-service",
						broker.FirstServiceOfferingName(),
						broker.FirstServicePlanName(),
						serviceInstanceName,
						"-c", parameters,
					}
					Eventually(helpers.CF(command...)).Should(Exit(0))
					Eventually(helpers.CF(serviceCommand, serviceInstanceName)).Should(Say(`status:\s+create succeeded`))
				})

				It("reports the service instance parameters", func() {
					session := helpers.CF(serviceCommand, serviceInstanceName)
					Eventually(session).Should(Exit(0))

					Expect(session).To(SatisfyAll(
						Say(`Showing parameters for service instance %s...\n`, serviceInstanceName),
						Say(`\n`),
						Say(`%s\n`, parameters),
						Say(`\n`),
					))
				})
			})

			When("the instance is shared with another space", func() {
				var sharedToSpaceName string

				BeforeEach(func() {
					sharedToSpaceName = helpers.NewSpaceName()
					helpers.CreateSpace(sharedToSpaceName)

					broker = servicebrokerstub.New().EnableServiceAccess()
					command := []string{
						"create-service",
						broker.FirstServiceOfferingName(),
						broker.FirstServicePlanName(),
						serviceInstanceName,
					}
					Eventually(helpers.CF(command...)).Should(Exit(0))

					output := func() *Buffer {
						session := helpers.CF(serviceCommand, serviceInstanceName)
						session.Wait()
						return session.Out
					}

					Eventually(output, testTimeout, testPollingInterval).Should(Say(`status:\s+create succeeded\n`))

					command = []string{
						"share-service",
						serviceInstanceName,
						"-s",
						sharedToSpaceName,
					}
					Eventually(helpers.CF(command...)).Should(Exit(0))
				})

				AfterEach(func() {
					command := []string{
						"unshare-service",
						serviceInstanceName,
						"-s", sharedToSpaceName,
						"-f",
					}
					Eventually(helpers.CF(command...)).Should(Exit(0))

					helpers.QuickDeleteSpace(sharedToSpaceName)
				})

				It("can show that the service is being shared", func() {
					session := helpers.CF(serviceCommand, serviceInstanceName)
					Eventually(session).Should(Exit(0))

					Expect(session).To(SatisfyAll(
						Say(`Sharing:\n`),
						Say(`Shared with spaces:\n`),
						Say(`org\s+space\s+bindings\n`),
						Say(`%s\s+%s\s+0\s*\n`, orgName, sharedToSpaceName),
					))
				})
			})

			When("the instance is being accessed form shared to space", func() {
				var sharedToSpaceName string

				BeforeEach(func() {
					sharedToSpaceName = helpers.NewSpaceName()
					helpers.CreateSpace(sharedToSpaceName)

					broker = servicebrokerstub.New().EnableServiceAccess()
					command := []string{
						"create-service",
						broker.FirstServiceOfferingName(),
						broker.FirstServicePlanName(),
						serviceInstanceName,
					}
					Eventually(helpers.CF(command...)).Should(Exit(0))

					output := func() *Buffer {
						session := helpers.CF(serviceCommand, serviceInstanceName)
						session.Wait()
						return session.Out
					}

					Eventually(output, testTimeout, testPollingInterval).Should(Say(`status:\s+create succeeded\n`))

					command = []string{
						"share-service",
						serviceInstanceName,
						"-s",
						sharedToSpaceName,
					}
					Eventually(helpers.CF(command...)).Should(Exit(0))

					helpers.TargetOrgAndSpace(orgName, sharedToSpaceName)
				})

				AfterEach(func() {
					helpers.TargetOrgAndSpace(orgName, spaceName)

					command := []string{
						"unshare-service",
						serviceInstanceName,
						"-s", sharedToSpaceName,
						"-f",
					}
					Eventually(helpers.CF(command...)).Should(Exit(0))

					helpers.QuickDeleteSpace(sharedToSpaceName)
				})

				It("can show that the service has been shared", func() {
					session := helpers.CF(serviceCommand, serviceInstanceName)
					Eventually(session).Should(Exit(0))

					Expect(session).To(SatisfyAll(
						Say(`Sharing:\n`),
						Say(`This service instance is shared from space %s of org %s.\n`, spaceName, orgName),
					))
				})
			})

			When("the broker supports maintenance info", func() {
				When("the service is up to date", func() {
					var serviceInstanceName string

					BeforeEach(func() {
						serviceInstanceName = helpers.NewServiceInstanceName()
						broker = servicebrokerstub.New().WithMaintenanceInfo("1.2.3").EnableServiceAccess()
						helpers.CreateManagedServiceInstance(broker.FirstServiceOfferingName(), broker.FirstServicePlanName(), serviceInstanceName)
					})

					It("says that the service has no upgrades available", func() {
						session := helpers.CF(serviceCommand, serviceInstanceName)
						Eventually(session).Should(Exit(0))

						Expect(session).To(SatisfyAll(
							Say(`Upgrading:\n`),
							Say(`There is no upgrade available for this service.\n`),
						))
					})
				})

				When("an update is available", func() {
					var serviceInstanceName string

					BeforeEach(func() {
						serviceInstanceName = helpers.NewServiceInstanceName()
						broker = servicebrokerstub.New().WithMaintenanceInfo("1.2.3").EnableServiceAccess()
						helpers.CreateManagedServiceInstance(broker.FirstServiceOfferingName(), broker.FirstServicePlanName(), serviceInstanceName)

						broker.WithMaintenanceInfo("1.2.4", "really cool improvement").Configure().Register()
					})

					It("displays information about the upgrade", func() {
						session := helpers.CF(serviceCommand, serviceInstanceName)
						Eventually(session).Should(Exit(0))

						Expect(session).To(SatisfyAll(
							Say(`Upgrading:\n`),
							Say(`Showing available upgrade details for this service...\n`),
							Say(`\n`),
							Say(`Upgrade description: really cool improvement\n`),
							Say(`\n`),
							Say(`TIP: You can upgrade using 'cf upgrade-service %s'\n`, serviceInstanceName),
						))
					})
				})
			})

			When("bound to apps", func() {
				var (
					appName1     string
					appName2     string
					bindingName1 string
					bindingName2 string
				)

				BeforeEach(func() {
					appName1 = helpers.PrefixedRandomName("APP1")
					appName2 = helpers.PrefixedRandomName("APP2")
					bindingName1 = helpers.RandomName()
					bindingName2 = helpers.RandomName()

					const asyncDelay = time.Minute // Forces bind to be "in progress" for predictable output
					broker = servicebrokerstub.New().WithAsyncDelay(asyncDelay).EnableServiceAccess()

					helpers.CreateManagedServiceInstance(
						broker.FirstServiceOfferingName(),
						broker.FirstServicePlanName(),
						serviceInstanceName,
					)

					helpers.WithHelloWorldApp(func(appDir string) {
						Eventually(helpers.CF("push", appName1, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
						Eventually(helpers.CF("push", appName2, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
					})
					Eventually(helpers.CF("bind-service", appName1, serviceInstanceName, "--binding-name", bindingName1)).Should(Exit(0))
					Eventually(helpers.CF("bind-service", appName2, serviceInstanceName, "--binding-name", bindingName2)).Should(Exit(0))
				})

				It("displays the bound apps", func() {
					session := helpers.CF(serviceCommand, serviceInstanceName, "-v")
					Eventually(session).Should(Exit(0))

					Expect(session).To(SatisfyAll(
						Say(`Bound apps:\n`),
						Say(`name\s+binding name\s+status\s+message\n`),
						Say(`%s\s+%s\s+create in progress\s*\n`, appName1, bindingName1),
						Say(`%s\s+%s\s+create in progress\s*\n`, appName2, bindingName2),
					))
				})
			})
		})
	})
})
