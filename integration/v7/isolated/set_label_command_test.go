package isolated

import (
	"fmt"
	"regexp"

	"code.cloudfoundry.org/cli/integration/helpers/servicebrokerstub"

	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("set-label command", func() {
	Describe("help", func() {
		When("--help flag is set", func() {
			It("appears in cf help -a", func() {
				session := helpers.CF("help", "-a")
				Eventually(session).Should(Exit(0))
				Expect(session).To(HaveCommandInCategoryWithDescription("set-label", "METADATA", "Set a label (key-value pairs) for an API resource"))
			})

			It("Displays command usage to output", func() {
				session := helpers.CF("set-label", "--help")

				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say(`\s+set-label - Set a label \(key-value pairs\) for an API resource`))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say(`\s+cf set-label RESOURCE RESOURCE_NAME KEY=VALUE\.\.\.`))
				Eventually(session).Should(Say("EXAMPLES:"))
				Eventually(session).Should(Say(`\s+cf set-label app dora env=production`))
				Eventually(session).Should(Say(`\s+cf set-label org business pci=true public-facing=false`))
				Eventually(session).Should(Say(`\s+cf set-label buildpack go_buildpack go=1.12 -s cflinuxfs3`))
				Eventually(session).Should(Say("RESOURCES:"))
				Eventually(session).Should(Say(`\s+app`))
				Eventually(session).Should(Say(`\s+buildpack`))
				Eventually(session).Should(Say(`\s+domain`))
				Eventually(session).Should(Say(`\s+org`))
				Eventually(session).Should(Say(`\s+route`))
				Eventually(session).Should(Say(`\s+service-broker`))
				Eventually(session).Should(Say(`\s+service-instance`))
				Eventually(session).Should(Say(`\s+service-offering`))
				Eventually(session).Should(Say(`\s+service-plan`))
				Eventually(session).Should(Say(`\s+space`))
				Eventually(session).Should(Say(`\s+stack`))
				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say(`\s+--stack, -s\s+Specify stack to disambiguate buildpacks with the same name`))
				Eventually(session).Should(Say(`\s+--broker, -b\s+Specify a service broker to disambiguate service offerings or service plans with the same name`))
				Eventually(session).Should(Say(`\s+--offering, -e\s+Specify a service offering to disambiguate service plans with the same name`))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say(`\s+labels, unset-label`))

				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the environment is set up correctly", func() {
		var (
			orgName       string
			spaceName     string
			username      string
			stackNameBase string
		)

		testExpectedBehaviors := func(resourceType, resourceTypeFormatted, resourceName string) {
			By("checking the behavior when the resource does not exist", func() {
				unknownResourceName := "nonexistent-" + resourceType
				session := helpers.CF("set-label", resourceType, unknownResourceName, "some-key=some-value")
				Eventually(session.Err).Should(Say("%s '%s' not found", resourceTypeFormatted, unknownResourceName))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})

			By("checking the behavior when the label has an empty key and an invalid value", func() {
				session := helpers.CF("set-label", resourceType, resourceName, "=test", "sha2=108&eb90d734")
				Eventually(session.Err).Should(Say("Metadata label key error: key cannot be empty string, Metadata label value error: '108&eb90d734' contains invalid characters"))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})

			By("checking the behavior when the label does not include a '=' to separate the key and value", func() {
				session := helpers.CF("set-label", resourceType, resourceName, "test-label")
				Eventually(session.Err).Should(Say("Metadata error: no value provided for label 'test-label'"))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		}

		BeforeEach(func() {
			username, _ = helpers.GetCredentials()
			orgName = helpers.NewOrgName()
			stackNameBase = helpers.NewStackName()
		})

		When("assigning label to app", func() {
			var appName string

			BeforeEach(func() {
				spaceName = helpers.NewSpaceName()
				appName = helpers.PrefixedRandomName("app")

				helpers.SetupCF(orgName, spaceName)
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CF("push", appName, "-p", appDir, "--no-start")).Should(Exit(0))
				})
			})

			AfterEach(func() {
				helpers.QuickDeleteOrg(orgName)
			})

			It("has the expected shared behaviors", func() {
				testExpectedBehaviors("app", "App", appName)
			})

			It("sets the specified labels on the app", func() {
				session := helpers.CF("set-label", "app", appName, "some-key=some-value", "some-other-key=some-other-value")
				Eventually(session).Should(Say(regexp.QuoteMeta(`Setting label(s) for app %s in org %s / space %s as %s...`), appName, orgName, spaceName, username))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))

				helpers.CheckExpectedLabels(fmt.Sprintf("/v3/apps/%s", helpers.AppGUID(appName)), false, helpers.MetadataLabels{
					"some-key":       "some-value",
					"some-other-key": "some-other-value",
				})
			})

			When("more than one value is provided for the same key", func() {
				It("uses the last value", func() {
					session := helpers.CF("set-label", "app", appName, "owner=sue", "owner=beth")
					Eventually(session).Should(Exit(0))

					helpers.CheckExpectedLabels(fmt.Sprintf("/v3/apps/%s", helpers.AppGUID(appName)), false, helpers.MetadataLabels{
						"owner": "beth",
					})
				})
			})
		})

		When("assigning label to domain", func() {
			var (
				domainName string
				domain     helpers.Domain
			)

			BeforeEach(func() {
				domainName = helpers.NewDomainName("labels")
				domain = helpers.NewDomain(orgName, domainName)

				helpers.SetupCFWithOrgOnly(orgName)
				domain.CreatePrivate()
			})

			AfterEach(func() {
				domain.DeletePrivate()
				helpers.QuickDeleteOrg(orgName)
			})

			It("has the expected shared behaviors", func() {
				testExpectedBehaviors("domain", "Domain", domainName)
			})

			It("sets the specified labels on the domain", func() {
				session := helpers.CF("set-label", "domain", domainName, "some-key=some-value", "some-other-key=some-other-value")
				Eventually(session).Should(Exit(0))
				Expect(session).Should(Say(regexp.QuoteMeta(`Setting label(s) for domain %s as %s...`), domainName, username))
				Expect(session).Should(Say("OK"))

				helpers.CheckExpectedLabels(fmt.Sprintf("/v3/domains?names=%s", domainName), true, helpers.MetadataLabels{
					"some-key":       "some-value",
					"some-other-key": "some-other-value",
				})
			})

			When("more than one value is provided for the same key", func() {
				It("uses the last value", func() {
					session := helpers.CF("set-label", "domain", domainName, "some-key=some-value", "some-key=some-other-value")
					Eventually(session).Should(Exit(0))
					Expect(session).Should(Say(regexp.QuoteMeta(`Setting label(s) for domain %s as %s...`), domainName, username))
					Expect(session).Should(Say("OK"))

					helpers.CheckExpectedLabels(fmt.Sprintf("/v3/domains?names=%s", domainName), true, helpers.MetadataLabels{
						"some-key": "some-other-value",
					})
				})
			})
		})

		When("assigning label to space", func() {
			BeforeEach(func() {
				spaceName = helpers.NewSpaceName()
				helpers.SetupCF(orgName, spaceName)
			})

			AfterEach(func() {
				helpers.QuickDeleteOrg(orgName)
			})

			It("has the expected shared behaviors", func() {
				testExpectedBehaviors("space", "Space", spaceName)
			})

			It("sets the specified labels on the space", func() {
				session := helpers.CF("set-label", "space", spaceName, "some-key=some-value", "some-other-key=some-other-value")
				Eventually(session).Should(Say(regexp.QuoteMeta(`Setting label(s) for space %s in org %s as %s...`), spaceName, orgName, username))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))

				helpers.CheckExpectedLabels(fmt.Sprintf("/v3/spaces/%s", helpers.GetSpaceGUID(spaceName)), false, helpers.MetadataLabels{
					"some-key":       "some-value",
					"some-other-key": "some-other-value",
				})
			})

			When("more than one value is provided for the same key", func() {
				It("uses the last value", func() {
					session := helpers.CF("set-label", "space", spaceName, "owner=sue", "owner=beth")
					Eventually(session).Should(Exit(0))

					helpers.CheckExpectedLabels(fmt.Sprintf("/v3/spaces/%s", helpers.GetSpaceGUID(spaceName)), false, helpers.MetadataLabels{
						"owner": "beth",
					})
				})
			})
		})

		When("assigning label to org", func() {
			BeforeEach(func() {
				helpers.SetupCFWithOrgOnly(orgName)
			})

			AfterEach(func() {
				helpers.QuickDeleteOrg(orgName)
			})

			It("has the expected shared behaviors", func() {
				testExpectedBehaviors("org", "Organization", orgName)
			})

			It("sets the specified labels on the org", func() {
				session := helpers.CF("set-label", "org", orgName, "pci=true", "public-facing=false")
				Eventually(session).Should(Say(regexp.QuoteMeta(`Setting label(s) for org %s as %s...`), orgName, username))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))

				helpers.CheckExpectedLabels(fmt.Sprintf("/v3/organizations/%s", helpers.GetOrgGUID(orgName)), false, helpers.MetadataLabels{
					"pci":           "true",
					"public-facing": "false",
				})
			})

			When("more than one value is provided for the same key", func() {
				It("uses the last value", func() {
					session := helpers.CF("set-label", "org", orgName, "owner=sue", "owner=beth")
					Eventually(session).Should(Exit(0))

					helpers.CheckExpectedLabels(fmt.Sprintf("/v3/organizations/%s", helpers.GetOrgGUID(orgName)), false, helpers.MetadataLabels{
						"owner": "beth",
					})
				})
			})
		})

		When("assigning label to route", func() {
			var (
				orgGUID    string
				routeName  string
				domainName string
				domain     helpers.Domain
			)
			BeforeEach(func() {
				orgName = helpers.NewOrgName()
				spaceName = helpers.NewSpaceName()
				helpers.SetupCF(orgName, spaceName)

				orgGUID = helpers.GetOrgGUID(orgName)
				domainName = helpers.NewDomainName()
				domain = helpers.NewDomain(orgName, domainName)
				domain.Create()
				Eventually(helpers.CF("create-route", domainName)).Should(Exit(0))
				routeName = domainName
			})

			AfterEach(func() {
				Eventually(helpers.CF("delete-route", domainName, "-f")).Should(Exit(0))
				domain.Delete()
				helpers.QuickDeleteOrg(orgName)
			})

			It("has the expected shared behaviors", func() {
				// The Domain is checked first, hence why the error message says 'Domain' and not 'Route'
				testExpectedBehaviors("route", "Domain", routeName)
			})

			When("the route contains a port", func() {
				var tcpDomain helpers.Domain

				BeforeEach(func() {
					tcpDomainName := helpers.NewDomainName()
					tcpDomain = helpers.NewDomain(orgName, tcpDomainName)
					tcpDomain.CreateWithRouterGroup(helpers.FindOrCreateTCPRouterGroup(0))

					Eventually(helpers.CF("create-route", tcpDomainName, "--port", "1028")).Should(Exit(0))
					Eventually(helpers.CF("create-route", tcpDomainName, "--port", "1029")).Should(Exit(0))
					routeName = tcpDomainName + ":1028"
				})

				AfterEach(func() {
					tcpDomain.DeleteShared()
				})

				It("sets the specified labels on the route", func() {
					session := helpers.CF("set-label", "route", routeName, "some-key=some-value", "some-other-key=some-other-value")

					Eventually(session).Should(Say(regexp.QuoteMeta(`Setting label(s) for route %s in org %s / space %s as %s...`), routeName, orgName, spaceName, username))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))

					helpers.CheckExpectedLabels("/v3/routes?ports=1028", true, helpers.MetadataLabels{
						"some-key":       "some-value",
						"some-other-key": "some-other-value",
					})

					helpers.CheckExpectedLabels("/v3/routes?ports=1029", true, helpers.MetadataLabels{})
				})
			})

			When("the route is unknown", func() {
				It("displays an error", func() {
					invalidRoute := "nonexistent-host." + domainName
					session := helpers.CF("set-label", "route", invalidRoute, "some-key=some-value")
					Eventually(session.Err).Should(Say(`Route with host 'nonexistent-host', domain '%s', and path '/' not found\.`, domainName))
					Eventually(session.Out).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				})
			})

			It("sets the specified labels on the route", func() {
				session := helpers.CF("set-label", "route", routeName, "some-key=some-value", "some-other-key=some-other-value")

				Eventually(session).Should(Say(regexp.QuoteMeta(`Setting label(s) for route %s in org %s / space %s as %s...`), routeName, orgName, spaceName, username))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))

				helpers.CheckExpectedLabels(fmt.Sprintf("/v3/routes?organization_guids=%s", orgGUID), true, helpers.MetadataLabels{
					"some-key":       "some-value",
					"some-other-key": "some-other-value",
				})
			})

			When("more than one value is provided for the same key", func() {
				It("uses the last value", func() {
					session := helpers.CF("set-label", "route", routeName, "owner=sue", "owner=beth")
					Eventually(session).Should(Exit(0))

					helpers.CheckExpectedLabels(fmt.Sprintf("/v3/routes?organization_guids=%s", orgGUID), true, helpers.MetadataLabels{
						"owner": "beth",
					})
				})
			})
		})

		When("assigning label to buildpack", func() {
			var (
				buildpackName string
			)

			BeforeEach(func() {
				helpers.LoginCF()
				buildpackName = helpers.NewBuildpackName()
			})

			When("the buildpack exists for at most one stack", func() {
				var (
					currentStack string
				)

				BeforeEach(func() {
					currentStack = helpers.PreferredStack()
					helpers.BuildpackWithStack(func(buildpackPath string) {
						session := helpers.CF("create-buildpack", buildpackName, buildpackPath, "98")
						Eventually(session).Should(Exit(0))
					}, currentStack)
				})

				AfterEach(func() {
					helpers.CF("delete-buildpack", buildpackName, "-f", "-s", currentStack)
				})

				It("has the expected shared behaviors", func() {
					testExpectedBehaviors("buildpack", "Buildpack", buildpackName)
				})

				It("sets the specified labels on the buildpack", func() {
					session := helpers.CF("set-label", "buildpack", buildpackName, "pci=true", "public-facing=false")
					Eventually(session).Should(Say(regexp.QuoteMeta(`Setting label(s) for buildpack %s as %s...`), buildpackName, username))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))

					buildpackGUID := helpers.BuildpackGUIDByNameAndStack(buildpackName, currentStack)
					helpers.CheckExpectedLabels(fmt.Sprintf("/v3/buildpacks/%s", buildpackGUID), false, helpers.MetadataLabels{
						"pci":           "true",
						"public-facing": "false",
					})
				})

				When("the buildpack exists for multiple stacks", func() {
					var stacks []string

					BeforeEach(func() {
						stacks = []string{helpers.PreferredStack(), helpers.CreateStack()}

						helpers.BuildpackWithStack(func(buildpackPath string) {
							createSession := helpers.CF("create-buildpack", buildpackName, buildpackPath, "99")
							Eventually(createSession).Should(Exit(0))
						}, stacks[1])
					})
					AfterEach(func() {
						helpers.CF("delete-buildpack", buildpackName, "-f", "-s", stacks[1])
						helpers.DeleteStack(stacks[1])
					})

					When("stack is not specified", func() {
						It("displays an error", func() {
							session := helpers.CF("set-label", "buildpack", buildpackName, "some-key=some-value")
							Eventually(session.Err).Should(Say(fmt.Sprintf("Multiple buildpacks named %s found. Specify a stack name by using a '-s' flag.", buildpackName)))
							Eventually(session).Should(Say("FAILED"))
							Eventually(session).Should(Exit(1))
						})
					})

					When("stack is specified", func() {
						It("sets the specified labels on the correct buildpack", func() {
							session := helpers.CF("set-label", "buildpack", buildpackName, "pci=true", "public-facing=false", "--stack", stacks[1])
							Eventually(session).Should(Say(regexp.QuoteMeta(`Setting label(s) for buildpack %s with stack %s as %s...`), buildpackName, stacks[1], username))
							Eventually(session).Should(Say("OK"))
							Eventually(session).Should(Exit(0))

							buildpackGUID := helpers.BuildpackGUIDByNameAndStack(buildpackName, stacks[1])
							helpers.CheckExpectedLabels(fmt.Sprintf("/v3/buildpacks/%s", buildpackGUID), false, helpers.MetadataLabels{
								"pci":           "true",
								"public-facing": "false",
							})
						})
					})
				})

				When("the buildpack exists in general but does NOT exist for the specified stack", func() {
					It("displays an error", func() {
						session := helpers.CF("set-label", "buildpack", buildpackName, "some-key=some-value", "--stack", "FAKE")
						Eventually(session.Err).Should(Say(fmt.Sprintf("Buildpack '%s' with stack 'FAKE' not found", buildpackName)))
						Eventually(session).Should(Say("FAILED"))
						Eventually(session).Should(Exit(1))
					})
				})

				When("more than one value is provided for the same key", func() {
					It("uses the last value", func() {
						session := helpers.CF("set-label", "buildpack", buildpackName, "owner=sue", "owner=beth")
						Eventually(session).Should(Exit(0))

						buildpackGUID := helpers.BuildpackGUIDByNameAndStack(buildpackName, currentStack)
						helpers.CheckExpectedLabels(fmt.Sprintf("/v3/buildpacks/%s", buildpackGUID), false, helpers.MetadataLabels{
							"owner": "beth",
						})
					})
				})
			})

			When("the buildpack exists for multiple stacks", func() {
				var (
					stacks             [3]string
					buildpackGUIDs     [3]string
					testWithStackCount int
				)

				BeforeEach(func() {
					stacks[0] = helpers.PreferredStack()
					testWithStackCount += 1
					stacks[1] = helpers.CreateStack(fmt.Sprintf("%s-%d", stackNameBase, testWithStackCount))

					for i := 0; i < 2; i++ {
						helpers.BuildpackWithStack(func(buildpackPath string) {
							createSession := helpers.CF("create-buildpack", buildpackName, buildpackPath,
								fmt.Sprintf("%d", 95+i))
							Eventually(createSession).Should(Exit(0))
							buildpackGUIDs[i] = helpers.BuildpackGUIDByNameAndStack(buildpackName, stacks[i])
						}, stacks[i])
					}
					helpers.CF("curl", "/v3/buildpacks?names="+buildpackName)
				})

				AfterEach(func() {
					helpers.CF("delete-buildpack", buildpackName, "-f", "-s", stacks[0])
					helpers.CF("delete-buildpack", buildpackName, "-f", "-s", stacks[1])
					helpers.DeleteStack(stacks[1])
				})

				When("all buildpacks are stack-scoped", func() {
					When("no stack is specified", func() {
						It("displays an error", func() {
							session := helpers.CF("set-label", "buildpack", buildpackName, "some-key=some-value")
							Eventually(session.Err).Should(Say(fmt.Sprintf("Multiple buildpacks named %s found. Specify a stack name by using a '-s' flag.", buildpackName)))
							Eventually(session).Should(Say("FAILED"))
							Eventually(session).Should(Exit(1))
						})
					})

					When("a nonexistent stack is specified", func() {
						It("displays an error", func() {
							bogusStackName := stacks[0] + "-bogus-" + stacks[1]
							session := helpers.CF("set-label", "buildpack", buildpackName, "olive=3", "mangosteen=4", "--stack", bogusStackName)
							Eventually(session.Err).Should(Say(regexp.QuoteMeta(fmt.Sprintf("Buildpack '%s' with stack '%s' not found", buildpackName, bogusStackName))))
							Eventually(session).Should(Say("FAILED"))
							Eventually(session).Should(Exit(1))
						})
					})

					When("an existing stack is specified", func() {
						It("updates the correct buildpack", func() {
							session := helpers.CF("set-label", "buildpack", buildpackName, "peach=5", "quince=6", "--stack", stacks[0])
							Eventually(session).Should(Say(regexp.QuoteMeta(`Setting label(s) for buildpack %s with stack %s as %s...`), buildpackName, stacks[0], username))
							Eventually(session).Should(Say("OK"))
							Eventually(session).Should(Exit(0))

							helpers.CheckExpectedLabels(fmt.Sprintf("/v3/buildpacks/%s", buildpackGUIDs[0]), false, helpers.MetadataLabels{
								"peach":  "5",
								"quince": "6",
							})
						})
					})
				})

				When("one of the buildpacks is not stack-scoped", func() {
					BeforeEach(func() {
						helpers.BuildpackWithoutStack(func(buildpackPath string) {
							createSession := helpers.CF("create-buildpack", buildpackName, buildpackPath, "97")
							Eventually(createSession).Should(Exit(0))
							buildpackGUIDs[2] = helpers.BuildpackGUIDByNameAndStack(buildpackName, "")
						})
					})
					AfterEach(func() {
						helpers.CF("delete-buildpack", buildpackName, "-f")
					})

					When("no stack is specified", func() {
						It("updates the unscoped buildpack", func() {
							session := helpers.CF("set-label", "buildpack", buildpackName, "mango=1", "figs=2")
							Eventually(session).Should(Say(regexp.QuoteMeta(`Setting label(s) for buildpack %s as %s...`), buildpackName, username))
							Eventually(session).Should(Say("OK"))
							Eventually(session).Should(Exit(0))

							helpers.CheckExpectedLabels(fmt.Sprintf("/v3/buildpacks/%s", buildpackGUIDs[2]), false, helpers.MetadataLabels{
								"mango": "1",
								"figs":  "2",
							})
						})
					})

					When("an existing stack is specified", func() {
						It("updates the correct buildpack", func() {
							session := helpers.CF("set-label", "buildpack", buildpackName, "tangelo=3", "lemon=4", "--stack", stacks[1])
							Eventually(session).Should(Say(regexp.QuoteMeta(`Setting label(s) for buildpack %s with stack %s as %s...`), buildpackName, stacks[1], username))
							Eventually(session).Should(Say("OK"))
							Eventually(session).Should(Exit(0))

							helpers.CheckExpectedLabels(fmt.Sprintf("/v3/buildpacks/%s", buildpackGUIDs[1]), false, helpers.MetadataLabels{
								"tangelo": "3",
								"lemon":   "4",
							})
						})
					})

					When("a nonexistent stack is specified", func() {
						It("displays an error", func() {
							bogusStackName := stacks[0] + "-bogus-" + stacks[1]
							session := helpers.CF("set-label", "buildpack", buildpackName, "olive=3", "mangosteen=4", "--stack", bogusStackName)
							Eventually(session.Err).Should(Say(regexp.QuoteMeta(fmt.Sprintf("Buildpack '%s' with stack '%s' not found", buildpackName, bogusStackName))))
							Eventually(session).Should(Say("FAILED"))
							Eventually(session).Should(Exit(1))
						})
					})
				})
			})
		})

		When("assigning label to stack", func() {
			var (
				stackName string
				stackGUID string
			)

			BeforeEach(func() {
				helpers.LoginCF()
				stackName, stackGUID = helpers.CreateStackWithGUID()
			})

			AfterEach(func() {
				deleteResourceByGUID(stackGUID, "stacks")
			})

			It("has the expected shared behaviors", func() {
				testExpectedBehaviors("stack", "Stack", stackName)
			})

			It("sets the specified labels on the stack", func() {
				session := helpers.CF("set-label", "stack", stackName, "pci=true", "public-facing=false")
				Eventually(session).Should(Say(regexp.QuoteMeta(`Setting label(s) for stack %s as %s...`), stackName, username))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))

				helpers.CheckExpectedLabels(fmt.Sprintf("/v3/stacks/%s", stackGUID), false, helpers.MetadataLabels{
					"pci":           "true",
					"public-facing": "false",
				})
			})

			When("more than one value is provided for the same key", func() {
				It("uses the last value", func() {
					session := helpers.CF("set-label", "stack", stackName, "owner=sue", "owner=beth")
					Eventually(session).Should(Exit(0))

					helpers.CheckExpectedLabels(fmt.Sprintf("/v3/stacks/%s", stackGUID), false, helpers.MetadataLabels{
						"owner": "beth",
					})
				})
			})
		})

		When("assigning label to service-broker", func() {
			var broker *servicebrokerstub.ServiceBrokerStub

			BeforeEach(func() {
				orgName = helpers.NewOrgName()
				spaceName = helpers.NewSpaceName()
				helpers.SetupCF(orgName, spaceName)
				broker = servicebrokerstub.Register()
			})

			AfterEach(func() {
				helpers.QuickDeleteOrg(orgName)
				broker.Forget()
			})

			It("has the expected shared behaviors", func() {
				testExpectedBehaviors("service-broker", "Service broker", broker.Name)
			})

			It("sets the specified labels on the service-broker", func() {
				session := helpers.CF("set-label", "service-broker", broker.Name, "some-key=some-value", "some-other-key=some-other-value")
				Eventually(session).Should(Say(regexp.QuoteMeta(`Setting label(s) for service-broker %s as %s...`), broker.Name, username))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))

				helpers.CheckExpectedLabels(fmt.Sprintf("/v3/service_brokers?names=%s", broker.Name), true, helpers.MetadataLabels{
					"some-key":       "some-value",
					"some-other-key": "some-other-value",
				})
			})

			When("more than one value is provided for the same key", func() {
				It("uses the last value", func() {
					session := helpers.CF("set-label", "service-broker", broker.Name, "owner=sue", "owner=beth")
					Eventually(session).Should(Exit(0))

					helpers.CheckExpectedLabels(fmt.Sprintf("/v3/service_brokers?names=%s", broker.Name), true, helpers.MetadataLabels{
						"owner": "beth",
					})
				})
			})
		})

		When("assigning label to service-instance", func() {
			var serviceInstanceName string

			BeforeEach(func() {
				orgName = helpers.NewOrgName()
				spaceName = helpers.NewSpaceName()
				helpers.SetupCF(orgName, spaceName)

				serviceInstanceName = helpers.NewServiceInstanceName()
				Eventually(helpers.CF("cups", serviceInstanceName)).Should(Exit(0))
			})

			AfterEach(func() {
				helpers.QuickDeleteOrg(orgName)
			})

			It("has the expected shared behaviors", func() {
				testExpectedBehaviors("service-instance", "Service instance", serviceInstanceName)
			})

			It("sets the specified labels on the service-instance", func() {
				session := helpers.CF("set-label", "service-instance", serviceInstanceName, "some-key=some-value", "some-other-key=some-other-value")
				Eventually(session).Should(Say(regexp.QuoteMeta(`Setting label(s) for service-instance %s in org %s / space %s as %s...`), serviceInstanceName, orgName, spaceName, username))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))

				helpers.CheckExpectedLabels(fmt.Sprintf("/v3/service_instances?names=%s", serviceInstanceName), true, helpers.MetadataLabels{
					"some-key":       "some-value",
					"some-other-key": "some-other-value",
				})
			})

			When("more than one value is provided for the same key", func() {
				It("uses the last value", func() {
					session := helpers.CF("set-label", "service-instance", serviceInstanceName, "owner=sue", "owner=beth")
					Eventually(session).Should(Exit(0))

					helpers.CheckExpectedLabels(fmt.Sprintf("/v3/service_instances?names=%s", serviceInstanceName), true, helpers.MetadataLabels{
						"owner": "beth",
					})
				})
			})
		})

		When("assigning label to service-offering", func() {
			var (
				broker              *servicebrokerstub.ServiceBrokerStub
				serviceOfferingName string
			)

			BeforeEach(func() {
				orgName = helpers.NewOrgName()
				spaceName = helpers.NewSpaceName()
				helpers.SetupCF(orgName, spaceName)
				broker = servicebrokerstub.Register()
				serviceOfferingName = broker.Services[0].Name
			})

			AfterEach(func() {
				helpers.QuickDeleteOrg(orgName)
				broker.Forget()
			})

			It("has the expected shared behaviors", func() {
				testExpectedBehaviors("service-offering", "Service offering", serviceOfferingName)
			})

			It("sets the specified labels", func() {
				session := helpers.CF("set-label", "service-offering", serviceOfferingName, "some-key=some-value", "some-other-key=some-other-value")
				Eventually(session).Should(Say(regexp.QuoteMeta(`Setting label(s) for service-offering %s as %s...`), serviceOfferingName, username))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))

				helpers.CheckExpectedLabels(fmt.Sprintf("/v3/service_offerings?names=%s", serviceOfferingName), true, helpers.MetadataLabels{
					"some-key":       "some-value",
					"some-other-key": "some-other-value",
				})
			})

			When("more than one value is provided for the same key", func() {
				It("uses the last value", func() {
					session := helpers.CF("set-label", "service-broker", broker.Name, "owner=sue", "owner=beth")
					Eventually(session).Should(Exit(0))

					helpers.CheckExpectedLabels(fmt.Sprintf("/v3/service_brokers?names=%s", broker.Name), true, helpers.MetadataLabels{
						"owner": "beth",
					})
				})
			})

			When("the service broker name is specified as part of the command", func() {
				It("sets the specified labels", func() {
					session := helpers.CF("set-label", "-b", broker.Name, "service-offering", serviceOfferingName, "some-key=some-value", "some-other-key=some-other-value")
					Eventually(session).Should(Say(regexp.QuoteMeta(`Setting label(s) for service-offering %s from service broker %s as %s...`), serviceOfferingName, broker.Name, username))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))

					helpers.CheckExpectedLabels(fmt.Sprintf("/v3/service_offerings?names=%s", serviceOfferingName), true, helpers.MetadataLabels{
						"some-key":       "some-value",
						"some-other-key": "some-other-value",
					})
				})
			})
		})

		When("assigning label to service-plan", func() {
			var (
				broker              *servicebrokerstub.ServiceBrokerStub
				servicePlanName     string
				serviceOfferingName string
			)

			BeforeEach(func() {
				orgName = helpers.NewOrgName()
				spaceName = helpers.NewSpaceName()
				helpers.SetupCF(orgName, spaceName)
				broker = servicebrokerstub.Register()
				servicePlanName = broker.Services[0].Plans[0].Name
				serviceOfferingName = broker.Services[0].Name
			})

			AfterEach(func() {
				helpers.QuickDeleteOrg(orgName)
				broker.Forget()
			})

			It("has the expected shared behaviors", func() {
				testExpectedBehaviors("service-plan", "Service plan", servicePlanName)
			})

			It("sets the specified labels", func() {
				session := helpers.CF("set-label", "service-plan", servicePlanName, "-b", broker.Name, "-e", serviceOfferingName, "some-key=some-value", "some-other-key=some-other-value")
				Eventually(session).Should(Say(regexp.QuoteMeta(`Setting label(s) for service-plan %s from service offering %s / service broker %s as %s...`), servicePlanName, serviceOfferingName, broker.Name, username))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))

				helpers.CheckExpectedLabels(fmt.Sprintf("/v3/service_plans?names=%s", servicePlanName), true, helpers.MetadataLabels{
					"some-key":       "some-value",
					"some-other-key": "some-other-value",
				})
			})
		})

	})
})

func deleteResourceByGUID(guid string, urlType string) {
	session := helpers.CF("curl", "-v", "-X", "DELETE",
		fmt.Sprintf("/v3/%s/%s", urlType, guid))
	Eventually(session).Should(Exit(0))
	Eventually(session).Should(Say(`(?:204 No Content|202 Accepted)`))
}
