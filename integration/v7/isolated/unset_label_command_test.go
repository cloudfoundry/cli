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
			Eventually(session).Should(Say(`\s+cf unset-label RESOURCE RESOURCE_NAME KEY`))
			Eventually(session).Should(Say("EXAMPLES:"))
			Eventually(session).Should(Say(`\s+cf unset-label app dora ci_signature_sha2`))
			Eventually(session).Should(Say("RESOURCES:"))
			Eventually(session).Should(Say(`\s+app`))
			Eventually(session).Should(Say(`\s+buildpack`))
			Eventually(session).Should(Say(`\s+org`))
			Eventually(session).Should(Say(`\s+space`))
			Eventually(session).Should(Say("SEE ALSO:"))
			Eventually(session).Should(Say(`\s+set-label, labels`))
			Eventually(session).Should(Exit(0))
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

			It("unsets the specified labels on the app", func() {
				session := helpers.CF("unset-label", "app", appName, "some-other-key", "some-third-key")
				Eventually(session).Should(Say(regexp.QuoteMeta(`Removing label(s) for app %s in org %s / space %s as %s...`), appName, orgName, spaceName, username))
				Consistently(session).ShouldNot(Say("\n\nOK"))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))

				appGuid := helpers.AppGUID(appName)
				session = helpers.CF("curl", fmt.Sprintf("/v3/apps/%s", appGuid))
				Eventually(session).Should(Exit(0))
				appJSON := session.Out.Contents()

				var app commonResource
				Expect(json.Unmarshal(appJSON, &app)).To(Succeed())
				Expect(len(app.Metadata.Labels)).To(Equal(1))
				Expect(app.Metadata.Labels["some-key"]).To(Equal("some-value"))
			})
		})

		When("unsetting labels from a buildpack", func() {
			var (
				buildpackName string
				buildpackGUID string
				stacks        []string
			)
			BeforeEach(func() {
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
					Eventually(session).Should(Say(regexp.QuoteMeta(`Removing label(s) for buildpack %s as %s...`), buildpackName, username))
					Consistently(session).ShouldNot(Say("\n\nOK"))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))

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
					Eventually(session.Err).Should(Say("Buildpack non-existent-buildpack not found"))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("the buildpack exists for multiple stacks", func() {
				var buildpackGUIDs [2]string
				BeforeEach(func() {
					stacks = helpers.EnsureMinimumNumberOfStacks(2)
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
				})

				When("stack is not specified", func() {
					It("displays an error", func() {
						session := helpers.CF("unset-label", "buildpack", buildpackName, "pci1")
						Eventually(session.Err).Should(Say(fmt.Sprintf("Multiple buildpacks named %s found. Specify a stack name by using a '-s' flag.", buildpackName)))
						Eventually(session).Should(Say("FAILED"))
						Eventually(session).Should(Exit(1))
					})
				})

				When("stack is specified", func() {
					When("the label is invalid", func() {
						It("gives an error message", func() {
							badLabel := "^^snorky"
							session := helpers.CF("unset-label", "buildpack", buildpackName, badLabel, "--stack", stacks[0])
							Eventually(session).Should(Say(regexp.QuoteMeta(fmt.Sprintf("Removing label(s) for buildpack %s with stack %s as %s...", buildpackName, stacks[0], username))))
							Eventually(session.Err).Should(Say(regexp.QuoteMeta(fmt.Sprintf("Metadata label key error: '%s' contains invalid characters", badLabel))))
							Eventually(session).Should(Say("FAILED"))
							Eventually(session).Should(Exit(1))
						})
					})

					It("deletes the specified labels from the correct buildpack", func() {
						var buildpack commonResource

						session := helpers.CF("unset-label", "buildpack", buildpackName, "pci0", "--stack", stacks[0])
						Eventually(session).Should(Say(regexp.QuoteMeta(`Removing label(s) for buildpack %s with stack %s as %s...`), buildpackName, stacks[0], username))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Exit(0))

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
						Eventually(session).Should(Say(regexp.QuoteMeta(`Removing label(s) for buildpack %s with stack %s as %s...`), buildpackName, stacks[1], username))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Exit(0))

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

		When("unsetting labels from an org", func() {
			BeforeEach(func() {
				session := helpers.CF("set-label", "org", orgName, "pci=true", "public-facing=false")
				Eventually(session).Should(Exit(0))
			})

			It("unsets the specified labels on the org", func() {
				session := helpers.CF("unset-label", "org", orgName, "public-facing")
				Eventually(session).Should(Say(regexp.QuoteMeta(`Removing label(s) for org %s as %s...`), orgName, username))
				Consistently(session).ShouldNot(Say("\n\nOK"))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))

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

		When("unsetting labels from a space", func() {
			BeforeEach(func() {
				spaceName = helpers.NewSpaceName()
				helpers.SetupCF(orgName, spaceName)
				session := helpers.CF("set-label", "space", spaceName, "pci=true", "public-facing=false")
				Eventually(session).Should(Exit(0))
			})

			It("unsets the specified labels on the space", func() {
				session := helpers.CF("unset-label", "space", spaceName, "public-facing")
				Eventually(session).Should(Say(regexp.QuoteMeta(`Removing label(s) for space %s in org %s as %s...`), spaceName, orgName, username))
				Consistently(session).ShouldNot(Say("\n\nOK"))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))

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
	})
})
