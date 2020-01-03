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
			Eventually(session).Should(Say(`\s+space`))
			Eventually(session).Should(Say(`\s+stack`))
			Eventually(session).Should(Say("OPTIONS:"))
			Eventually(session).Should(Say(`\s+--stack, -s\s+Specify stack to disambiguate buildpacks with the same name`))
			Eventually(session).Should(Say("SEE ALSO:"))
			Eventually(session).Should(Say(`\s+labels, set-label`))
			Eventually(session).Should(Exit(0))
		})
	})

	When("the environment is set up correctly", func() {
		var (
			orgName   string
			spaceName string
			appName   string
			username  string
			stackName string
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
			helpers.LoginCF()
			orgName = helpers.NewOrgName()
		})

		When("unsetting labels from an app", func() {
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

				helpers.CheckExpectedMetadata(fmt.Sprintf("/v3/apps/%s", helpers.AppGUID(appName)), false, helpers.MetadataLabels{
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
					session := helpers.CF("set-label", "buildpack", buildpackName, "pci=true", "public-facing=false")
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

					// verify the labels are deleted
					session = helpers.CF("curl", fmt.Sprintf("/v3/buildpacks/%s", buildpackGUID))
					Eventually(session).Should(Exit(0))
					buildpackJSON := session.Out.Contents()

					var buildpack commonResource
					Expect(json.Unmarshal(buildpackJSON, &buildpack)).To(Succeed())
					Expect(len(buildpack.Metadata.Labels)).To(Equal(0))
				})
			})

			When("the buildpack is unknown", func() {
				It("displays an error", func() {
					session := helpers.CF("unset-label", "buildpack", "non-existent-buildpack", "some-key=some-value")
					Eventually(session).Should(Exit(1))
					Expect(session.Err).Should(Say("Buildpack 'non-existent-buildpack' not found"))
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
							fmt.Sprintf("pci%d=true", i),
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
						session := helpers.CF("unset-label", "buildpack", buildpackName, "pci1")
						Eventually(session).Should(Exit(1))
						Expect(session.Err).Should(Say(fmt.Sprintf("Multiple buildpacks named %s found. Specify a stack name by using a '-s' flag.", buildpackName)))
						Expect(session).Should(Say("FAILED"))
					})
				})

				When("stack is specified", func() {
					When("the label is invalid", func() {
						It("gives an error message", func() {
							badLabel := "^^snorky"
							session := helpers.CF("unset-label", "buildpack", buildpackName, badLabel, "--stack", stacks[0])
							Eventually(session).Should(Exit(1))
							Expect(session).Should(Say(regexp.QuoteMeta(fmt.Sprintf("Removing label(s) for buildpack %s with stack %s as %s...", buildpackName, stacks[0], username))))
							Expect(session.Err).Should(Say(regexp.QuoteMeta(fmt.Sprintf("Metadata label key error: '%s' contains invalid characters", badLabel))))
							Expect(session).Should(Say("FAILED"))
						})
					})

					It("deletes the specified labels from the correct buildpack", func() {
						var buildpack commonResource

						session := helpers.CF("unset-label", "buildpack", buildpackName, "pci0", "--stack", stacks[0])
						Eventually(session).Should(Exit(0))
						Expect(session).Should(Say(regexp.QuoteMeta(`Removing label(s) for buildpack %s with stack %s as %s...`), buildpackName, stacks[0], username))
						Expect(session).Should(Say("OK"))

						session = helpers.CF("curl", fmt.Sprintf("/v3/buildpacks/%s", buildpackGUIDs[0]))
						Eventually(session).Should(Exit(0))
						buildpackJSON := session.Out.Contents()
						Expect(json.Unmarshal(buildpackJSON, &buildpack)).To(Succeed())
						Expect(len(buildpack.Metadata.Labels)).To(Equal(1))
						Expect(buildpack.Metadata.Labels["public-facing0"]).To(Equal("false"))

						session = helpers.CF("curl", fmt.Sprintf("/v3/buildpacks/%s", buildpackGUIDs[1]))
						Eventually(session).Should(Exit(0))
						buildpackJSON = session.Out.Contents()
						buildpack = commonResource{}
						Expect(json.Unmarshal(buildpackJSON, &buildpack)).To(Succeed())
						Expect(len(buildpack.Metadata.Labels)).To(Equal(2))
						Expect(buildpack.Metadata.Labels["pci1"]).To(Equal("true"))
						Expect(buildpack.Metadata.Labels["public-facing1"]).To(Equal("false"))

						session = helpers.CF("unset-label", "buildpack", buildpackName, "pci1", "--stack", stacks[1])
						Eventually(session).Should(Exit(0))
						Expect(session).Should(Say(regexp.QuoteMeta(`Removing label(s) for buildpack %s with stack %s as %s...`), buildpackName, stacks[1], username))
						Expect(session).Should(Say("OK"))

						session = helpers.CF("curl", fmt.Sprintf("/v3/buildpacks/%s", buildpackGUIDs[1]))
						Eventually(session).Should(Exit(0))
						buildpackJSON = session.Out.Contents()
						buildpack = commonResource{}
						Expect(json.Unmarshal(buildpackJSON, &buildpack)).To(Succeed())
						Expect(len(buildpack.Metadata.Labels)).To(Equal(1))
						Expect(buildpack.Metadata.Labels["public-facing1"]).To(Equal("false"))
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
				session := helpers.CF("unset-label", "domain", domainName,
					"some-other-key", "some-third-key")
				Eventually(session).Should(Exit(0))
				Expect(session).Should(Say(regexp.QuoteMeta(`Removing label(s) for domain %s as %s...`), domainName, username))
				Expect(session).Should(Say("OK"))

				session = helpers.CF("curl", fmt.Sprintf("/v3/domains?names=%s", domainName))
				Eventually(session).Should(Exit(0))
				domainJSON := session.Out.Contents()
				var domains resourceCollection
				Expect(json.Unmarshal(domainJSON, &domains)).To(Succeed())
				Expect(len(domains.Resources)).To(Equal(1))
				Expect(len(domains.Resources[0].Metadata.Labels)).To(Equal(1))
				Expect(domains.Resources[0].Metadata.Labels["some-key"]).To(Equal("some-value"))
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

				orgGUID := helpers.GetOrgGUID(orgName)
				session = helpers.CF("curl", fmt.Sprintf("/v3/organizations/%s", orgGUID))
				Eventually(session).Should(Exit(0))
				orgJSON := session.Out.Contents()

				var org commonResource
				Expect(json.Unmarshal(orgJSON, &org)).To(Succeed())
				Expect(len(org.Metadata.Labels)).To(Equal(1))
				Expect(org.Metadata.Labels["pci"]).To(Equal("true"))
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

				session = helpers.CF("curl", fmt.Sprintf("/v3/routes?organization_guids=%s", orgGUID))
				Eventually(session).Should(Exit(0))

				routeJSON := session.Out.Contents()
				var routes resourceCollection

				Expect(json.Unmarshal(routeJSON, &routes)).To(Succeed())
				Expect(len(routes.Resources)).To(Equal(1))
				Expect(len(routes.Resources[0].Metadata.Labels)).To(Equal(1))
				Expect(routes.Resources[0].Metadata.Labels["some-other-key"]).To(Equal("some-other-value"))

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

				spaceGUID := helpers.GetSpaceGUID(spaceName)
				session = helpers.CF("curl", fmt.Sprintf("/v3/spaces/%s", spaceGUID))
				Eventually(session).Should(Exit(0))
				spaceJSON := session.Out.Contents()

				var space commonResource
				Expect(json.Unmarshal(spaceJSON, &space)).To(Succeed())
				Expect(len(space.Metadata.Labels)).To(Equal(1))
				Expect(space.Metadata.Labels["pci"]).To(Equal("true"))
			})
		})

		When("unsetting labels from a stack", func() {
			var stackGUID string

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

				session = helpers.CF("curl", fmt.Sprintf("/v3/stacks/%s", stackGUID))
				Eventually(session).Should(Exit(0))
				stackJSON := session.Out.Contents()

				var stack commonResource
				Expect(json.Unmarshal(stackJSON, &stack)).To(Succeed())
				Expect(len(stack.Metadata.Labels)).To(Equal(1))
				Expect(stack.Metadata.Labels["pci"]).To(Equal("true"))
			})
		})
	})
})
