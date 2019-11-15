package isolated

import (
	"encoding/json"
	"fmt"
	"regexp"

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
				Eventually(session).Should(Say(`\s+space`))
				Eventually(session).Should(Say(`\s+stack`))
				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say(`\s+--stack, -s\s+Specify stack to disambiguate buildpacks with the same name`))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say(`\s+labels, unset-label`))

				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the environment is set up correctly", func() {
		var (
			orgName            string
			spaceName          string
			appName            string
			username           string
			stackNameBase      string
			testWithStackCount int
		)

		type commonResource struct {
			Metadata struct {
				Labels map[string]string
			}
		}

		type resourceCollection struct {
			Resources []commonResource
		}

		BeforeEach(func() {
			username, _ = helpers.GetCredentials()
			orgName = helpers.NewOrgName()
			stackNameBase = helpers.NewStackName()
		})

		When("assigning label to app", func() {
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

			It("sets the specified labels on the app", func() {
				session := helpers.CF("set-label", "app", appName, "some-key=some-value", "some-other-key=some-other-value")
				Eventually(session).Should(Say(regexp.QuoteMeta(`Setting label(s) for app %s in org %s / space %s as %s...`), appName, orgName, spaceName, username))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))
				appGUID := helpers.AppGUID(appName)
				session = helpers.CF("curl", fmt.Sprintf("/v3/apps/%s", appGUID))
				Eventually(session).Should(Exit(0))
				appJSON := session.Out.Contents()
				var app commonResource
				Expect(json.Unmarshal(appJSON, &app)).To(Succeed())
				Expect(len(app.Metadata.Labels)).To(Equal(2))
				Expect(app.Metadata.Labels["some-key"]).To(Equal("some-value"))
				Expect(app.Metadata.Labels["some-other-key"]).To(Equal("some-other-value"))
			})

			When("the app is unknown", func() {
				It("displays an error", func() {
					session := helpers.CF("set-label", "app", "non-existent-app", "some-key=some-value")
					Eventually(session.Err).Should(Say("App 'non-existent-app' not found"))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("the label has an empty key and an invalid value", func() {
				It("displays an error", func() {
					session := helpers.CF("set-label", "app", appName, "=test", "sha2=108&eb90d734")
					Eventually(session.Err).Should(Say("Metadata label key error: key cannot be empty string, Metadata label value error: '108&eb90d734' contains invalid characters"))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("the label does not include a '=' to separate the key and value", func() {
				It("displays an error", func() {
					session := helpers.CF("set-label", "app", appName, "test-label")
					Eventually(session.Err).Should(Say("Metadata error: no value provided for label 'test-label'"))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("more than one value is provided for the same key", func() {
				It("uses the last value", func() {
					session := helpers.CF("set-label", "app", appName, "owner=sue", "owner=beth")
					Eventually(session).Should(Exit(0))
					appGUID := helpers.AppGUID(appName)
					session = helpers.CF("curl", fmt.Sprintf("/v3/apps/%s", appGUID))
					Eventually(session).Should(Exit(0))
					appJSON := session.Out.Contents()
					var app commonResource
					Expect(json.Unmarshal(appJSON, &app)).To(Succeed())
					Expect(len(app.Metadata.Labels)).To(Equal(1))
					Expect(app.Metadata.Labels["owner"]).To(Equal("beth"))
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

			It("sets the specified labels on the domain", func() {
				session := helpers.CF("set-label", "domain", domainName, "some-key=some-value", "some-other-key=some-other-value")
				Eventually(session).Should(Exit(0))
				Expect(session).Should(Say(regexp.QuoteMeta(`Setting label(s) for domain %s as %s...`), domainName, username))
				Expect(session).Should(Say("OK"))

				session = helpers.CF("curl", fmt.Sprintf("/v3/domains?names=%s", domainName))
				Eventually(session).Should(Exit(0))
				domainJSON := session.Out.Contents()
				var domains resourceCollection
				Expect(json.Unmarshal(domainJSON, &domains)).To(Succeed())
				Expect(len(domains.Resources)).To(Equal(1))
				Expect(len(domains.Resources[0].Metadata.Labels)).To(Equal(2))
				Expect(domains.Resources[0].Metadata.Labels["some-key"]).To(Equal("some-value"))
				Expect(domains.Resources[0].Metadata.Labels["some-other-key"]).To(Equal("some-other-value"))
			})

			When("the domain is unknown", func() {
				It("displays an error", func() {
					session := helpers.CF("set-label", "domain", "non-existent-domain.example.com", "some-key=some-value")
					Eventually(session.Err).Should(Say("Domain 'non-existent-domain.example.com' not found"))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("the label has an empty key and an invalid value", func() {
				It("displays an error", func() {
					session := helpers.CF("set-label", "domain", domainName, "=test", "sha2=108&eb90d734")
					Eventually(session.Err).Should(Say("Metadata label key error: key cannot be empty string, Metadata label value error: '108&eb90d734' contains invalid characters"))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("the label does not include a '=' to separate the key and value", func() {
				It("displays an error", func() {
					session := helpers.CF("set-label", "domain", domainName, "test-label")
					Eventually(session.Err).Should(Say("Metadata error: no value provided for label 'test-label'"))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("more than one value is provided for the same key", func() {
				It("uses the last value", func() {
					session := helpers.CF("set-label", "domain", domainName, "some-key=some-value", "some-key=some-other-value")
					Eventually(session).Should(Exit(0))
					Expect(session).Should(Say(regexp.QuoteMeta(`Setting label(s) for domain %s as %s...`), domainName, username))
					Expect(session).Should(Say("OK"))

					session = helpers.CF("curl", fmt.Sprintf("/v3/domains?names=%s", domainName))
					Eventually(session).Should(Exit(0))
					domainJSON := session.Out.Contents()
					var domains resourceCollection
					Expect(json.Unmarshal(domainJSON, &domains)).To(Succeed())
					Expect(len(domains.Resources)).To(Equal(1))
					Expect(len(domains.Resources[0].Metadata.Labels)).To(Equal(1))
					Expect(domains.Resources[0].Metadata.Labels["some-key"]).To(Equal("some-other-value"))
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

			It("sets the specified labels on the space", func() {
				session := helpers.CF("set-label", "space", spaceName, "some-key=some-value", "some-other-key=some-other-value")
				Eventually(session).Should(Say(regexp.QuoteMeta(`Setting label(s) for space %s in org %s as %s...`), spaceName, orgName, username))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))
				spaceGUID := helpers.GetSpaceGUID(spaceName)
				session = helpers.CF("curl", fmt.Sprintf("/v3/spaces/%s", spaceGUID))
				Eventually(session).Should(Exit(0))
				spaceJSON := session.Out.Contents()
				var space commonResource
				Expect(json.Unmarshal(spaceJSON, &space)).To(Succeed())
				Expect(len(space.Metadata.Labels)).To(Equal(2))
				Expect(space.Metadata.Labels["some-key"]).To(Equal("some-value"))
				Expect(space.Metadata.Labels["some-other-key"]).To(Equal("some-other-value"))
			})

			When("the space is unknown", func() {
				It("displays an error", func() {
					session := helpers.CF("set-label", "space", "non-existent-space", "some-key=some-value")
					Eventually(session.Err).Should(Say("Space 'non-existent-space' not found"))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("the label has an empty key and an invalid value", func() {
				It("displays an error", func() {
					session := helpers.CF("set-label", "space", spaceName, "=test", "sha2=108&eb90d734")
					Eventually(session.Err).Should(Say("Metadata label key error: key cannot be empty string, Metadata label value error: '108&eb90d734' contains invalid characters"))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("the label does not include a '=' to separate the key and value", func() {
				It("displays an error", func() {
					session := helpers.CF("set-label", "space", spaceName, "test-label")
					Eventually(session.Err).Should(Say("Metadata error: no value provided for label 'test-label'"))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("more than one value is provided for the same key", func() {
				It("uses the last value", func() {
					session := helpers.CF("set-label", "space", spaceName, "owner=sue", "owner=beth")
					Eventually(session).Should(Exit(0))
					spaceGUID := helpers.GetSpaceGUID(spaceName)
					session = helpers.CF("curl", fmt.Sprintf("/v3/spaces/%s", spaceGUID))
					Eventually(session).Should(Exit(0))
					spaceJSON := session.Out.Contents()
					var space commonResource
					Expect(json.Unmarshal(spaceJSON, &space)).To(Succeed())
					Expect(len(space.Metadata.Labels)).To(Equal(1))
					Expect(space.Metadata.Labels["owner"]).To(Equal("beth"))
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
			It("sets the specified labels on the org", func() {
				session := helpers.CF("set-label", "org", orgName, "pci=true", "public-facing=false")
				Eventually(session).Should(Say(regexp.QuoteMeta(`Setting label(s) for org %s as %s...`), orgName, username))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))

				orgGUID := helpers.GetOrgGUID(orgName)
				session = helpers.CF("curl", fmt.Sprintf("/v3/organizations/%s", orgGUID))
				Eventually(session).Should(Exit(0))
				orgJSON := session.Out.Contents()
				var org commonResource
				Expect(json.Unmarshal(orgJSON, &org)).To(Succeed())
				Expect(len(org.Metadata.Labels)).To(Equal(2))
				Expect(org.Metadata.Labels["pci"]).To(Equal("true"))
				Expect(org.Metadata.Labels["public-facing"]).To(Equal("false"))
			})

			When("the org is unknown", func() {
				It("displays an error", func() {
					session := helpers.CF("set-label", "org", "non-existent-org", "some-key=some-value")
					Eventually(session.Err).Should(Say("Organization 'non-existent-org' not found"))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("the label has an empty key and an invalid value", func() {
				It("displays an error", func() {
					session := helpers.CF("set-label", "org", orgName, "=test", "sha2=108&eb90d734")
					Eventually(session.Err).Should(Say("Metadata label key error: key cannot be empty string, Metadata label value error: '108&eb90d734' contains invalid characters"))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("the label does not include a '=' to separate the key and value", func() {
				It("displays an error", func() {
					session := helpers.CF("set-label", "org", orgName, "test-label")
					Eventually(session.Err).Should(Say("Metadata error: no value provided for label 'test-label'"))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("more than one value is provided for the same key", func() {
				It("uses the last value", func() {
					session := helpers.CF("set-label", "org", orgName, "owner=sue", "owner=beth")
					Eventually(session).Should(Exit(0))
					orgGUID := helpers.GetOrgGUID(orgName)
					session = helpers.CF("curl", fmt.Sprintf("/v3/organizations/%s", orgGUID))
					Eventually(session).Should(Exit(0))
					orgJSON := session.Out.Contents()
					var org commonResource
					Expect(json.Unmarshal(orgJSON, &org)).To(Succeed())
					Expect(len(org.Metadata.Labels)).To(Equal(1))
					Expect(org.Metadata.Labels["owner"]).To(Equal("beth"))
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

			It("sets the specified labels on the route", func() {
				session := helpers.CF("set-label", "route", routeName, "some-key=some-value", "some-other-key=some-other-value")

				Eventually(session).Should(Say(regexp.QuoteMeta(`Setting label(s) for route %s in org %s / space %s as %s...`), routeName, orgName, spaceName, username))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))
				session = helpers.CF("curl", fmt.Sprintf("/v3/routes?organization_guids=%s", orgGUID))
				Eventually(session).Should(Exit(0))
				routeJSON := session.Out.Contents()
				var routes resourceCollection
				Expect(json.Unmarshal(routeJSON, &routes)).To(Succeed())
				Expect(len(routes.Resources)).To(Equal(1))
				Expect(len(routes.Resources[0].Metadata.Labels)).To(Equal(2))
				Expect(routes.Resources[0].Metadata.Labels["some-key"]).To(Equal("some-value"))
				Expect(routes.Resources[0].Metadata.Labels["some-other-key"]).To(Equal("some-other-value"))
			})

			When("the domain is unknown", func() {
				It("displays an error", func() {
					session := helpers.CF("set-label", "route", "non-existent-domain", "some-key=some-value")
					Eventually(session.Err).Should(Say("Domain 'non-existent-domain' not found"))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("the route is unknown", func() {
				It("displays an error", func() {
					invalidRoute := "non-existent-host." + domainName
					session := helpers.CF("set-label", "route", invalidRoute, "some-key=some-value")
					Eventually(session.Err).Should(Say(fmt.Sprintf("Route with host 'non-existent-host' and domain '%s' not found", domainName)))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("the label has an empty key and an invalid value", func() {
				It("displays an error", func() {
					session := helpers.CF("set-label", "route", routeName, "=test", "sha2=108&eb90d734")
					Eventually(session.Err).Should(Say("Metadata label key error: key cannot be empty string, Metadata label value error: '108&eb90d734' contains invalid characters"))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("the label does not include a '=' to separate the key and value", func() {
				It("displays an error", func() {
					session := helpers.CF("set-label", "route", routeName, "test-label")
					Eventually(session.Err).Should(Say("Metadata error: no value provided for label 'test-label'"))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("more than one value is provided for the same key", func() {
				It("uses the last value", func() {
					session := helpers.CF("set-label", "route", routeName, "owner=sue", "owner=beth")
					Eventually(session).Should(Exit(0))

					session = helpers.CF("curl", fmt.Sprintf("/v3/routes?organization_guids=%s", orgGUID))
					Eventually(session).Should(Exit(0))
					routeJSON := session.Out.Contents()
					var routes resourceCollection
					Expect(json.Unmarshal(routeJSON, &routes)).To(Succeed())
					Expect(len(routes.Resources)).To(Equal(1))
					Expect(len(routes.Resources[0].Metadata.Labels)).To(Equal(1))
					Expect(routes.Resources[0].Metadata.Labels["owner"]).To(Equal("beth"))
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

				It("sets the specified labels on the buildpack", func() {
					session := helpers.CF("set-label", "buildpack", buildpackName, "pci=true", "public-facing=false")
					Eventually(session).Should(Say(regexp.QuoteMeta(`Setting label(s) for buildpack %s as %s...`), buildpackName, username))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))

					buildpackGUID := helpers.BuildpackGUIDByNameAndStack(buildpackName, currentStack)
					session = helpers.CF("curl", fmt.Sprintf("/v3/buildpacks/%s", buildpackGUID))
					Eventually(session).Should(Exit(0))
					buildpackJSON := session.Out.Contents()
					var buildpack commonResource
					Expect(json.Unmarshal(buildpackJSON, &buildpack)).To(Succeed())
					Expect(len(buildpack.Metadata.Labels)).To(Equal(2))
					Expect(buildpack.Metadata.Labels["pci"]).To(Equal("true"))
					Expect(buildpack.Metadata.Labels["public-facing"]).To(Equal("false"))
				})

				When("the buildpack is unknown", func() {
					It("displays an error", func() {
						session := helpers.CF("set-label", "buildpack", "non-existent-buildpack", "some-key=some-value")
						Eventually(session.Err).Should(Say("Buildpack non-existent-buildpack not found"))
						Eventually(session).Should(Say("FAILED"))
						Eventually(session).Should(Exit(1))
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
							session = helpers.CF("curl", fmt.Sprintf("/v3/buildpacks/%s", buildpackGUID))
							Eventually(session).Should(Exit(0))
							buildpackJSON := session.Out.Contents()
							var buildpack commonResource
							Expect(json.Unmarshal(buildpackJSON, &buildpack)).To(Succeed())
							Expect(len(buildpack.Metadata.Labels)).To(Equal(2))
							Expect(buildpack.Metadata.Labels["pci"]).To(Equal("true"))
							Expect(buildpack.Metadata.Labels["public-facing"]).To(Equal("false"))
						})
					})
				})

				When("the buildpack exists in general but does NOT exist for the specified stack", func() {
					It("displays an error", func() {
						session := helpers.CF("set-label", "buildpack", buildpackName, "some-key=some-value", "--stack", "FAKE")
						Eventually(session.Err).Should(Say(fmt.Sprintf("Buildpack %s with stack FAKE not found", buildpackName)))
						Eventually(session).Should(Say("FAILED"))
						Eventually(session).Should(Exit(1))
					})
				})

				When("the label has an empty key and an invalid value", func() {
					It("displays an error", func() {
						session := helpers.CF("set-label", "buildpack", buildpackName, "=test", "sha2=108&eb90d734")
						Eventually(session.Err).Should(Say("Metadata label key error: key cannot be empty string, Metadata label value error: '108&eb90d734' contains invalid characters"))
						Eventually(session).Should(Say("FAILED"))
						Eventually(session).Should(Exit(1))
					})
				})

				When("the label does not include a '=' to separate the key and value", func() {
					It("displays an error", func() {
						session := helpers.CF("set-label", "buildpack", buildpackName, "test-label")
						Eventually(session.Err).Should(Say("Metadata error: no value provided for label 'test-label'"))
						Eventually(session).Should(Say("FAILED"))
						Eventually(session).Should(Exit(1))
					})
				})

				When("more than one value is provided for the same key", func() {
					It("uses the last value", func() {
						session := helpers.CF("set-label", "buildpack", buildpackName, "owner=sue", "owner=beth")
						Eventually(session).Should(Exit(0))
						buildpackGUID := helpers.BuildpackGUIDByNameAndStack(buildpackName, currentStack)
						session = helpers.CF("curl", fmt.Sprintf("/v3/buildpacks/%s", buildpackGUID))
						Eventually(session).Should(Exit(0))
						buildpackJSON := session.Out.Contents()
						var buildpack commonResource
						Expect(json.Unmarshal(buildpackJSON, &buildpack)).To(Succeed())
						Expect(len(buildpack.Metadata.Labels)).To(Equal(1))
						Expect(buildpack.Metadata.Labels["owner"]).To(Equal("beth"))
					})
				})
			})

			When("the buildpack exists for multiple stacks", func() {
				var (
					stacks         [3]string
					buildpackGUIDs [3]string
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

					When("a non-existent stack is specified", func() {
						It("displays an error", func() {
							bogusStackName := stacks[0] + "-bogus-" + stacks[1]
							session := helpers.CF("set-label", "buildpack", buildpackName, "olive=3", "mangosteen=4", "--stack", bogusStackName)
							Eventually(session.Err).Should(Say(regexp.QuoteMeta(fmt.Sprintf("Buildpack %s with stack %s not found", buildpackName, bogusStackName))))
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

							session = helpers.CF("curl", fmt.Sprintf("/v3/buildpacks/%s", buildpackGUIDs[0]))
							Eventually(session).Should(Exit(0))
							buildpackJSON := session.Out.Contents()
							var buildpack commonResource
							Expect(json.Unmarshal(buildpackJSON, &buildpack)).To(Succeed())
							Expect(len(buildpack.Metadata.Labels)).To(Equal(2))
							Expect(buildpack.Metadata.Labels["peach"]).To(Equal("5"))
							Expect(buildpack.Metadata.Labels["quince"]).To(Equal("6"))
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

							session = helpers.CF("curl", fmt.Sprintf("/v3/buildpacks/%s", buildpackGUIDs[2]))
							Eventually(session).Should(Exit(0))
							buildpackJSON := session.Out.Contents()
							var buildpack commonResource
							Expect(json.Unmarshal(buildpackJSON, &buildpack)).To(Succeed())
							Expect(len(buildpack.Metadata.Labels)).To(Equal(2))
							Expect(buildpack.Metadata.Labels["mango"]).To(Equal("1"))
							Expect(buildpack.Metadata.Labels["figs"]).To(Equal("2"))
						})
					})

					When("an existing stack is specified", func() {
						It("updates the correct buildpack", func() {
							session := helpers.CF("set-label", "buildpack", buildpackName, "tangelo=3", "lemon=4", "--stack", stacks[1])
							Eventually(session).Should(Say(regexp.QuoteMeta(`Setting label(s) for buildpack %s with stack %s as %s...`), buildpackName, stacks[1], username))
							Eventually(session).Should(Say("OK"))
							Eventually(session).Should(Exit(0))

							session = helpers.CF("curl", fmt.Sprintf("/v3/buildpacks/%s", buildpackGUIDs[1]))
							Eventually(session).Should(Exit(0))
							buildpackJSON := session.Out.Contents()
							var buildpack commonResource
							Expect(json.Unmarshal(buildpackJSON, &buildpack)).To(Succeed())
							Expect(len(buildpack.Metadata.Labels)).To(Equal(2))
							Expect(buildpack.Metadata.Labels["tangelo"]).To(Equal("3"))
							Expect(buildpack.Metadata.Labels["lemon"]).To(Equal("4"))
						})
					})

					When("a non-existent stack is specified", func() {
						It("displays an error", func() {
							bogusStackName := stacks[0] + "-bogus-" + stacks[1]
							session := helpers.CF("set-label", "buildpack", buildpackName, "olive=3", "mangosteen=4", "--stack", bogusStackName)
							Eventually(session.Err).Should(Say(regexp.QuoteMeta(fmt.Sprintf("Buildpack %s with stack %s not found", buildpackName, bogusStackName))))
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

			It("sets the specified labels on the stack", func() {
				session := helpers.CF("set-label", "stack", stackName, "pci=true", "public-facing=false")
				Eventually(session).Should(Say(regexp.QuoteMeta(`Setting label(s) for stack %s as %s...`), stackName, username))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))

				session = helpers.CF("curl", fmt.Sprintf("/v3/stacks/%s", stackGUID))
				Eventually(session).Should(Exit(0))
				stackJSON := session.Out.Contents()
				var stack commonResource
				Expect(json.Unmarshal(stackJSON, &stack)).To(Succeed())
				Expect(len(stack.Metadata.Labels)).To(Equal(2))
				Expect(stack.Metadata.Labels["pci"]).To(Equal("true"))
				Expect(stack.Metadata.Labels["public-facing"]).To(Equal("false"))
			})

			When("the stack is unknown", func() {
				It("displays an error", func() {
					session := helpers.CF("set-label", "stack", "non-existent-stack", "some-key=some-value")
					Eventually(session.Err).Should(Say("Stack 'non-existent-stack' not found"))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("the label has an empty key and an invalid value", func() {
				It("displays an error", func() {
					session := helpers.CF("set-label", "stack", stackName, "=test", "sha2=108&eb90d734")
					Eventually(session.Err).Should(Say("Metadata label key error: key cannot be empty string, Metadata label value error: '108&eb90d734' contains invalid characters"))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("the label does not include a '=' to separate the key and value", func() {
				It("displays an error", func() {
					session := helpers.CF("set-label", "stack", stackName, "test-label")
					Eventually(session.Err).Should(Say("Metadata error: no value provided for label 'test-label'"))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("more than one value is provided for the same key", func() {
				It("uses the last value", func() {
					session := helpers.CF("set-label", "stack", stackName, "owner=sue", "owner=beth")
					Eventually(session).Should(Exit(0))
					session = helpers.CF("curl", fmt.Sprintf("/v3/stacks/%s", stackGUID))
					Eventually(session).Should(Exit(0))
					stackJSON := session.Out.Contents()
					var stack commonResource
					Expect(json.Unmarshal(stackJSON, &stack)).To(Succeed())
					Expect(len(stack.Metadata.Labels)).To(Equal(1))
					Expect(stack.Metadata.Labels["owner"]).To(Equal("beth"))
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
