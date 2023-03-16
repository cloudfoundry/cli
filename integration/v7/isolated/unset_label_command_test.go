package isolated

import (
	"fmt"
	"regexp"

	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/integration/helpers/servicebrokerstub"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("unset-label command", func() {
	When("--help flag is set", func() {
		It("appears in cf help -a", func() {
			session := helpers.CF("help", "-a")
			Eventually(session).Should(Exit(0))
			Expect(session).To(HaveCommandInCategoryWithDescription("unset-label", "METADATA", "Unset a label (key-value pairs) for an API resource"))
		})

		It("Displays command usage to output", func() {
			session := helpers.CF("unset-label", "--help")

			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Say(`\s+unset-label - Unset a label \(key-value pairs\) for an API resource`))
			Eventually(session).Should(Say("USAGE:"))
			Eventually(session).Should(Say(`\s+cf unset-label RESOURCE RESOURCE_NAME KEY...`))
			Eventually(session).Should(Say("EXAMPLES:"))
			Eventually(session).Should(Say(`\s+cf unset-label app dora ci_signature_sha2`))
			Eventually(session).Should(Say(`\s+cf unset-label org business pci public-facing`))
			Eventually(session).Should(Say(`\s+cf unset-label buildpack go_buildpack go -s cflinuxfs3`))
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
			Eventually(session).Should(Say(`\s+labels, set-label`))
			Eventually(session).Should(Exit(0))
		})
	})

	When("the environment is set up correctly", func() {
		var (
			orgName   string
			spaceName string
			username  string
		)

		BeforeEach(func() {
			username, _ = helpers.GetCredentials()
			helpers.LoginCF()
			orgName = helpers.NewOrgName()
		})

		When("unsetting labels from an app", func() {
			var appName string

			BeforeEach(func() {
				spaceName = helpers.NewSpaceName()
				appName = helpers.PrefixedRandomName("app")

				helpers.SetupCF(orgName, spaceName)
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CF("push", appName, "-p", appDir, "--no-start")).Should(Exit(0))
				})

				session := helpers.CF("set-label", "app", appName, "some-key=some-value", "some-other-key=some-other-value", "some-third-key=other")
				Eventually(session).Should(Exit(0))
			})
			AfterEach(func() {
				helpers.QuickDeleteOrg(orgName)
			})

			It("unsets the specified labels on the app", func() {
				session := helpers.CF("unset-label", "app", appName, "some-other-key", "some-third-key")
				Eventually(session).Should(Exit(0))
				Expect(session).Should(Say(regexp.QuoteMeta(`Removing label(s) for app %s in org %s / space %s as %s...`), appName, orgName, spaceName, username))
				Expect(session).Should(Say("OK"))

				helpers.CheckExpectedLabels(fmt.Sprintf("/v3/apps/%s", helpers.AppGUID(appName)), false, helpers.MetadataLabels{
					"some-key": "some-value",
				})
			})
		})

		When("unsetting labels from a buildpack", func() {
			var (
				buildpackName string
				buildpackGUID string
				stacks        []string
			)
			BeforeEach(func() {
				helpers.LoginCF()
				buildpackName = helpers.NewBuildpackName()
			})

			When("there is only one instance of the given buildpack", func() {

				BeforeEach(func() {
					stacks = helpers.FetchStacks()
					helpers.BuildpackWithStack(func(buildpackPath string) {
						session := helpers.CF("create-buildpack", buildpackName, buildpackPath, "98")
						Eventually(session).Should(Exit(0))
					}, stacks[0])
					buildpackGUID = helpers.BuildpackGUIDByNameAndStack(buildpackName, stacks[0])
					session := helpers.CF("set-label", "buildpack", buildpackName, "pci=true", "public-facing=false", "a-third-label=some-value")
					Eventually(session).Should(Exit(0))
				})
				AfterEach(func() {
					deleteResourceByGUID(buildpackGUID, "buildpacks")
				})

				It("unsets the specified labels on the buildpack", func() {
					session := helpers.CF("unset-label", "buildpack", buildpackName, "public-facing", "pci")
					Eventually(session).Should(Exit(0))
					Expect(session).Should(Say(regexp.QuoteMeta(`Removing label(s) for buildpack %s as %s...`), buildpackName, username))
					Expect(session).Should(Say("OK"))

					helpers.CheckExpectedLabels(fmt.Sprintf("/v3/buildpacks/%s", buildpackGUID), false, helpers.MetadataLabels{
						"a-third-label": "some-value",
					})
				})
			})

			When("the buildpack is unknown", func() {
				It("displays an error", func() {
					session := helpers.CF("unset-label", "buildpack", "nonexistent-buildpack", "some-key=some-value")
					Eventually(session).Should(Exit(1))
					Expect(session.Err).Should(Say("Buildpack 'nonexistent-buildpack' not found"))
					Expect(session).Should(Say("FAILED"))
				})
			})

			When("the buildpack exists for multiple stacks", func() {
				var buildpackGUIDs [2]string
				BeforeEach(func() {
					stacks = []string{helpers.PreferredStack(), helpers.CreateStack()}
					for i := 0; i < 2; i++ {
						helpers.BuildpackWithStack(func(buildpackPath string) {
							createSession := helpers.CF("create-buildpack", buildpackName, buildpackPath, fmt.Sprintf("%d", 98+i))
							Eventually(createSession).Should(Exit(0))
						}, stacks[i])
						buildpackGUIDs[i] = helpers.BuildpackGUIDByNameAndStack(buildpackName, stacks[i])
						session := helpers.CF("set-label", "buildpack",
							buildpackName, "-s", stacks[i],
							"pci=true",
							fmt.Sprintf("public-facing%d=false", i))
						Eventually(session).Should(Exit(0))
					}
				})
				AfterEach(func() {
					for i := 0; i < 2; i++ {
						deleteResourceByGUID(buildpackGUIDs[i], "buildpacks")
					}
					helpers.DeleteStack(stacks[1])
				})

				When("stack is not specified", func() {
					It("displays an error", func() {
						session := helpers.CF("unset-label", "buildpack", buildpackName, "pci")
						Eventually(session).Should(Exit(1))
						Expect(session.Err).Should(Say(fmt.Sprintf("Multiple buildpacks named %s found. Specify a stack name by using a '-s' flag.", buildpackName)))
						Expect(session).Should(Say("FAILED"))
					})
				})

				When("stack is specified", func() {
					When("the label is invalid", func() {
						It("gives an error message", func() {
							const badLabel = "^^snorky"
							session := helpers.CF("unset-label", "buildpack", buildpackName, badLabel, "--stack", stacks[0])
							Eventually(session).Should(Exit(1))
							Expect(session).Should(Say(regexp.QuoteMeta(fmt.Sprintf("Removing label(s) for buildpack %s with stack %s as %s...", buildpackName, stacks[0], username))))
							Expect(session.Err).Should(Say(regexp.QuoteMeta(fmt.Sprintf("Metadata label key error: '%s' contains invalid characters", badLabel))))
							Expect(session).Should(Say("FAILED"))
						})
					})

					It("deletes the specified labels from the correct buildpack", func() {
						session := helpers.CF("unset-label", "buildpack", buildpackName, "pci", "--stack", stacks[0])
						Eventually(session).Should(Exit(0))
						Expect(session).Should(Say(regexp.QuoteMeta(`Removing label(s) for buildpack %s with stack %s as %s...`), buildpackName, stacks[0], username))
						Expect(session).Should(Say("OK"))

						helpers.CheckExpectedLabels(fmt.Sprintf("/v3/buildpacks/%s", buildpackGUIDs[0]), false, helpers.MetadataLabels{
							"public-facing0": "false",
						})

						helpers.CheckExpectedLabels(fmt.Sprintf("/v3/buildpacks/%s", buildpackGUIDs[1]), false, helpers.MetadataLabels{
							"public-facing1": "false",
							"pci":            "true",
						})
					})
				})
			})
		})

		When("unsetting labels from a domain", func() {

			var (
				domainName string
				domain     helpers.Domain
			)

			BeforeEach(func() {
				domainName = helpers.NewDomainName("labels")
				domain = helpers.NewDomain(orgName, domainName)

				helpers.SetupCFWithOrgOnly(orgName)
				domain.CreatePrivate()

				session := helpers.CF("set-label", "domain", domainName,
					"some-key=some-value", "some-other-key=some-other-value", "some-third-key=other")
				Eventually(session).Should(Exit(0))
			})

			AfterEach(func() {
				domain.DeletePrivate()
				helpers.QuickDeleteOrg(orgName)
			})

			It("unsets the specified labels on the domain", func() {
				session := helpers.CF("unset-label", "domain", domainName, "some-other-key", "some-third-key")
				Eventually(session).Should(Exit(0))
				Expect(session).Should(Say(regexp.QuoteMeta(`Removing label(s) for domain %s as %s...`), domainName, username))
				Expect(session).Should(Say("OK"))

				helpers.CheckExpectedLabels(fmt.Sprintf("/v3/domains?names=%s", domainName), true, helpers.MetadataLabels{
					"some-key": "some-value",
				})
			})
		})

		When("unsetting labels from an org", func() {
			BeforeEach(func() {
				helpers.SetupCFWithOrgOnly(orgName)
				session := helpers.CF("set-label", "org", orgName, "pci=true", "public-facing=false")
				Eventually(session).Should(Exit(0))
			})
			AfterEach(func() {
				helpers.QuickDeleteOrg(orgName)
			})

			It("unsets the specified labels on the org", func() {
				session := helpers.CF("unset-label", "org", orgName, "public-facing")
				Eventually(session).Should(Exit(0))
				Expect(session).Should(Say(regexp.QuoteMeta(`Removing label(s) for org %s as %s...`), orgName, username))
				Expect(session).Should(Say("OK"))

				helpers.CheckExpectedLabels(fmt.Sprintf("/v3/organizations/%s", helpers.GetOrgGUID(orgName)), false, helpers.MetadataLabels{
					"pci": "true",
				})
			})
		})

		When("unsetting labels from a route", func() {
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

				session := helpers.CF("set-label", "route", routeName, "some-key=some-value", "some-other-key=some-other-value")
				Eventually(session).Should(Exit(0))
				Expect(session).Should(Say(regexp.QuoteMeta(`Setting label(s) for route %s in org %s / space %s as %s...`), routeName, orgName, spaceName, username))
				Expect(session).Should(Say("OK"))
			})

			AfterEach(func() {
				Eventually(helpers.CF("delete-route", domainName, "-f")).Should(Exit(0))
				domain.Delete()
				helpers.QuickDeleteOrg(orgName)
			})

			It("unsets the specified labels on the route", func() {
				session := helpers.CF("unset-label", "route", routeName, "some-key")
				Eventually(session).Should(Exit(0))
				Expect(session).Should(Say(regexp.QuoteMeta(`Removing label(s) for route %s in org %s / space %s as %s...`), routeName, orgName, spaceName, username))
				Expect(session).Should(Say("OK"))

				helpers.CheckExpectedLabels(fmt.Sprintf("/v3/routes?organization_guids=%s", orgGUID), true, helpers.MetadataLabels{
					"some-other-key": "some-other-value",
				})
			})
		})

		When("unsetting labels from a space", func() {
			BeforeEach(func() {
				spaceName = helpers.NewSpaceName()
				helpers.SetupCF(orgName, spaceName)
				session := helpers.CF("set-label", "space", spaceName, "pci=true", "public-facing=false")
				Eventually(session).Should(Exit(0))
			})
			AfterEach(func() {
				helpers.QuickDeleteOrg(orgName)
			})

			It("unsets the specified labels on the space", func() {
				session := helpers.CF("unset-label", "space", spaceName, "public-facing")
				Eventually(session).Should(Exit(0))
				Expect(session).Should(Say(regexp.QuoteMeta(`Removing label(s) for space %s in org %s as %s...`), spaceName, orgName, username))
				Expect(session).Should(Say("OK"))

				helpers.CheckExpectedLabels(fmt.Sprintf("/v3/spaces/%s", helpers.GetSpaceGUID(spaceName)), false, helpers.MetadataLabels{
					"pci": "true",
				})
			})
		})

		When("unsetting labels from a stack", func() {
			var (
				stackGUID string
				stackName string
			)

			BeforeEach(func() {
				helpers.LoginCF()
				stackName, stackGUID = helpers.CreateStackWithGUID()
				session := helpers.CF("set-label", "stack", stackName, "pci=true", "public-facing=false")
				Eventually(session).Should(Exit(0))
			})

			AfterEach(func() {
				deleteResourceByGUID(stackGUID, "stacks")
			})

			It("unsets the specified labels on the stack", func() {
				session := helpers.CF("unset-label", "stack", stackName, "public-facing")
				Eventually(session).Should(Exit(0))
				Expect(session).Should(Say(regexp.QuoteMeta(`Removing label(s) for stack %s as %s...`), stackName, username))
				Expect(session).Should(Say("OK"))

				helpers.CheckExpectedLabels(fmt.Sprintf("/v3/stacks/%s", stackGUID), false, helpers.MetadataLabels{
					"pci": "true",
				})
			})
		})

		When("unsetting labels from a service-broker", func() {
			var broker *servicebrokerstub.ServiceBrokerStub

			BeforeEach(func() {
				spaceName = helpers.NewSpaceName()
				helpers.SetupCF(orgName, spaceName)
				broker = servicebrokerstub.Register()

				session := helpers.CF("set-label", "service-broker", broker.Name, "pci=true", "public-facing=false")
				Eventually(session).Should(Exit(0))
			})

			AfterEach(func() {
				helpers.QuickDeleteOrg(orgName)
				broker.Forget()
			})

			It("unsets the specified labels on the service-broker", func() {
				session := helpers.CF("unset-label", "service-broker", broker.Name, "public-facing")
				Eventually(session).Should(Exit(0))
				Expect(session).Should(Say(regexp.QuoteMeta(`Removing label(s) for service-broker %s as %s...`), broker.Name, username))
				Expect(session).Should(Say("OK"))

				helpers.CheckExpectedLabels(fmt.Sprintf("/v3/service_brokers?names=%s", broker.Name), true, helpers.MetadataLabels{
					"pci": "true",
				})
			})
		})

		When("unsetting labels from a service-instance", func() {
			var serviceInstanceName string

			BeforeEach(func() {
				spaceName = helpers.NewSpaceName()
				serviceInstanceName = helpers.NewServiceInstanceName()

				helpers.SetupCF(orgName, spaceName)
				Eventually(helpers.CF("cups", serviceInstanceName)).Should(Exit(0))

				session := helpers.CF("set-label", "service-instance", serviceInstanceName, "some-key=some-value", "some-other-key=some-other-value", "some-third-key=other")
				Eventually(session).Should(Exit(0))
			})

			AfterEach(func() {
				helpers.QuickDeleteOrg(orgName)
			})

			It("unsets the specified labels on the service-instance", func() {
				session := helpers.CF("unset-label", "service-instance", serviceInstanceName, "some-other-key", "some-third-key")
				Eventually(session).Should(Exit(0))
				Expect(session).Should(Say(regexp.QuoteMeta(`Removing label(s) for service-instance %s in org %s / space %s as %s...`), serviceInstanceName, orgName, spaceName, username))
				Expect(session).Should(Say("OK"))

				helpers.CheckExpectedLabels(fmt.Sprintf("/v3/service_instances/%s", helpers.ServiceInstanceGUID(serviceInstanceName)), false, helpers.MetadataLabels{
					"some-key": "some-value",
				})
			})
		})

		When("unsetting labels from a service-offering", func() {
			var (
				broker              *servicebrokerstub.ServiceBrokerStub
				serviceOfferingName string
			)

			BeforeEach(func() {
				spaceName = helpers.NewSpaceName()
				helpers.SetupCF(orgName, spaceName)
				broker = servicebrokerstub.Register()
				serviceOfferingName = broker.Services[0].Name

				session := helpers.CF("set-label", "service-offering", serviceOfferingName, "pci=true", "public-facing=false")
				Eventually(session).Should(Exit(0))
			})

			AfterEach(func() {
				helpers.QuickDeleteOrg(orgName)
				broker.Forget()
			})

			It("unsets the specified labels", func() {
				session := helpers.CF("unset-label", "service-offering", serviceOfferingName, "public-facing")
				Eventually(session).Should(Exit(0))
				Expect(session).Should(Say(regexp.QuoteMeta(`Removing label(s) for service-offering %s as %s...`), serviceOfferingName, username))
				Expect(session).Should(Say("OK"))

				helpers.CheckExpectedLabels(fmt.Sprintf("/v3/service_offerings?names=%s", serviceOfferingName), true, helpers.MetadataLabels{
					"pci": "true",
				})
			})

			When("the service broker name is specified", func() {
				It("unsets the specified labels", func() {
					session := helpers.CF("unset-label", "-b", broker.Name, "service-offering", serviceOfferingName, "public-facing")
					Eventually(session).Should(Exit(0))
					Expect(session).Should(Say(regexp.QuoteMeta(`Removing label(s) for service-offering %s from service broker %s as %s...`), serviceOfferingName, broker.Name, username))
					Expect(session).Should(Say("OK"))

					helpers.CheckExpectedLabels(fmt.Sprintf("/v3/service_offerings?names=%s", serviceOfferingName), true, helpers.MetadataLabels{
						"pci": "true",
					})
				})
			})
		})

		When("unsetting labels from a service-plan", func() {
			var (
				broker              *servicebrokerstub.ServiceBrokerStub
				servicePlanName     string
				serviceOfferingName string
			)

			BeforeEach(func() {
				spaceName = helpers.NewSpaceName()
				helpers.SetupCF(orgName, spaceName)
				broker = servicebrokerstub.Register()
				servicePlanName = broker.Services[0].Plans[0].Name
				serviceOfferingName = broker.Services[0].Name

				session := helpers.CF("set-label", "service-plan", servicePlanName, "pci=true", "public-facing=false")
				Eventually(session).Should(Exit(0))
			})

			AfterEach(func() {
				helpers.QuickDeleteOrg(orgName)
				broker.Forget()
			})

			It("unsets the specified labels", func() {
				session := helpers.CF("unset-label", "service-plan", servicePlanName, "-b", broker.Name, "-e", serviceOfferingName, "public-facing")
				Eventually(session).Should(Say(regexp.QuoteMeta(`Removing label(s) for service-plan %s from service offering %s / service broker %s as %s...`), servicePlanName, serviceOfferingName, broker.Name, username))
				Eventually(session).Should(Exit(0))
				Expect(session).Should(Say("OK"))

				helpers.CheckExpectedLabels(fmt.Sprintf("/v3/service_plans?names=%s", servicePlanName), true, helpers.MetadataLabels{
					"pci": "true",
				})
			})
		})
	})
})
