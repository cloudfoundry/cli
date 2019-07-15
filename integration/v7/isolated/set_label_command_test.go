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
				Eventually(session).Should(Say(`\s+cf set-label space business_space public-facing=false owner=jane_doe`))
				Eventually(session).Should(Say("RESOURCES:"))
				Eventually(session).Should(Say(`\s+app`))
				Eventually(session).Should(Say(`\s+buildpack`))
				Eventually(session).Should(Say(`\s+org`))
				Eventually(session).Should(Say(`\s+space`))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say(`\s+unset-label, labels`))

				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the environment is set up correctly", func() {
		var (
			orgName   string
			spaceName string
			appName   string
			username  string
		)

		type commonResource struct {
			Metadata struct {
				Labels map[string]string
			}
		}

		BeforeEach(func() {
			username, _ = helpers.GetCredentials()
			helpers.LoginCF()
			orgName = helpers.NewOrgName()
			helpers.CreateOrg(orgName)
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

		When("assigning label to space", func() {
			BeforeEach(func() {
				spaceName = helpers.NewSpaceName()

				helpers.SetupCF(orgName, spaceName)
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

		When("assigning label to buildpack", func() {
			var (
				buildpackName string
			)

			BeforeEach(func() {
				buildpackName = helpers.NewBuildpackName()
				stacks := helpers.FetchStacks()
				helpers.BuildpackWithStack(func(buildpackPath string) {
					session := helpers.CF("create-buildpack", buildpackName, buildpackPath, "98")
					Eventually(session).Should(Exit(0))
				}, stacks[0])
			})

			It("sets the specified labels on the buildpack", func() {
				session := helpers.CF("set-label", "buildpack", buildpackName, "pci=true", "public-facing=false")
				Eventually(session).Should(Say(regexp.QuoteMeta(`Setting label(s) for buildpack %s as %s...`), buildpackName, username))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))

				buildpackGUID := helpers.BuildpackGUIDByNameAndStack(buildpackName, "")
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
					stacks = helpers.EnsureMinimumNumberOfStacks(2)

					helpers.BuildpackWithStack(func(buildpackPath string) {
						createSession := helpers.CF("create-buildpack", buildpackName, buildpackPath, "99")
						Eventually(createSession).Should(Exit(0))
					}, stacks[1])
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
						Eventually(session).Should(Say(regexp.QuoteMeta(`Setting label(s) for buildpack %s as %s...`), buildpackName, username))
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
					buildpackGUID := helpers.BuildpackGUIDByNameAndStack(buildpackName, "")
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
	})
})
