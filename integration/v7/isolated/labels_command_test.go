package isolated

import (
	"fmt"
	"regexp"

	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("labels command", func() {
	When("--help flag is set", func() {
		It("appears in cf help -a", func() {
			session := helpers.CF("help", "-a")
			Eventually(session).Should(Exit(0))
			Expect(session).To(HaveCommandInCategoryWithDescription("labels", "METADATA", "List all labels (key-value pairs) for an API resource"))
		})

		It("Displays command usage", func() {
			session := helpers.CF("labels", "--help")

			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Say(`\s+labels - List all labels \(key-value pairs\) for an API resource`))
			Eventually(session).Should(Say("USAGE:"))
			Eventually(session).Should(Say(`\s+cf labels RESOURCE RESOURCE_NAME`))
			Eventually(session).Should(Say("EXAMPLES:"))
			Eventually(session).Should(Say(`\s+cf labels app dora`))
			Eventually(session).Should(Say("RESOURCES:"))
			Eventually(session).Should(Say(`\s+app`))
			Eventually(session).Should(Say(`\s+buildpack`))
			Eventually(session).Should(Say(`\s+org`))
			Eventually(session).Should(Say(`\s+space`))
			Eventually(session).Should(Say("SEE ALSO:"))
			Eventually(session).Should(Say(`\s+set-label, unset-label`))
			Eventually(session).Should(Exit(0))
		})
	})

	When("the environment is set up correctly", func() {
		var (
			orgName       string
			spaceName     string
			appName       string
			username      string
			buildpackName string
		)

		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			buildpackName = helpers.NewBuildpackName()
			username, _ = helpers.GetCredentials()
			helpers.LoginCF()
			helpers.CreateOrg(orgName)
		})

		Describe("app labels", func() {
			BeforeEach(func() {
				helpers.TargetOrg(orgName)
				spaceName = helpers.NewSpaceName()
				appName = helpers.PrefixedRandomName("app")
				helpers.CreateSpace(spaceName)
				helpers.TargetOrgAndSpace(orgName, spaceName)
				helpers.SetupCF(orgName, spaceName)
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CF("push", appName, "-p", appDir, "--no-start")).Should(Exit(0))
				})
			})

			When("there are labels set on the application", func() {
				BeforeEach(func() {
					session := helpers.CF("set-label", "app", appName, "some-other-key=some-other-value", "some-key=some-value")
					Eventually(session).Should(Exit(0))
				})

				It("lists the labels", func() {
					session := helpers.CF("labels", "app", appName)
					Eventually(session).Should(Say(regexp.QuoteMeta("Getting labels for app %s in org %s / space %s as %s...\n\n"), appName, orgName, spaceName, username))
					Eventually(session).Should(Say(`key\s+value`))
					Eventually(session).Should(Say(`some-key\s+some-value`))
					Eventually(session).Should(Say(`some-other-key\s+some-other-value`))
					Eventually(session).Should(Exit(0))
				})
			})

			When("there are no labels set on the application", func() {
				It("indicates that there are no labels", func() {
					session := helpers.CF("labels", "app", appName)
					Eventually(session).Should(Say(regexp.QuoteMeta("Getting labels for app %s in org %s / space %s as %s...\n\n"), appName, orgName, spaceName, username))
					Expect(session).ToNot(Say(`key\s+value`))
					Eventually(session).Should(Say("No labels found."))
					Eventually(session).Should(Exit(0))
				})
			})

			When("the app does not exist", func() {
				It("displays an error", func() {
					session := helpers.CF("labels", "app", "non-existent-app")
					Eventually(session).Should(Say(regexp.QuoteMeta("Getting labels for app non-existent-app in org %s / space %s as %s...\n\n"), orgName, spaceName, username))
					Eventually(session.Err).Should(Say("App 'non-existent-app' not found"))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				})
			})
		})

		Describe("org labels", func() {
			When("there are labels set on the organization", func() {
				BeforeEach(func() {
					session := helpers.CF("set-label", "org", orgName, "some-other-key=some-other-value", "some-key=some-value")
					Eventually(session).Should(Exit(0))
				})
				It("lists the labels", func() {
					session := helpers.CF("labels", "org", orgName)
					Eventually(session).Should(Say(regexp.QuoteMeta("Getting labels for org %s as %s...\n\n"), orgName, username))
					Eventually(session).Should(Say(`key\s+value`))
					Eventually(session).Should(Say(`some-key\s+some-value`))
					Eventually(session).Should(Say(`some-other-key\s+some-other-value`))
					Eventually(session).Should(Exit(0))
				})
			})

			When("there are no labels set on the organization", func() {
				It("indicates that there are no labels", func() {
					session := helpers.CF("labels", "org", orgName)
					Eventually(session).Should(Say(regexp.QuoteMeta("Getting labels for org %s as %s...\n\n"), orgName, username))
					Expect(session).ToNot(Say(`key\s+value`))
					Eventually(session).Should(Say("No labels found."))
					Eventually(session).Should(Exit(0))
				})
			})

			When("the org does not exist", func() {
				It("displays an error", func() {
					session := helpers.CF("labels", "org", "non-existent-org")
					Eventually(session).Should(Say(regexp.QuoteMeta("Getting labels for org %s as %s...\n\n"), "non-existent-org", username))
					Eventually(session.Err).Should(Say("Organization 'non-existent-org' not found"))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				})
			})
		})

		Describe("space labels", func() {
			BeforeEach(func() {
				helpers.TargetOrg(orgName)
				spaceName = helpers.NewSpaceName()
				helpers.CreateSpace(spaceName)
				helpers.TargetOrgAndSpace(orgName, spaceName)
				helpers.SetupCF(orgName, spaceName)
			})

			When("there are labels set on the space", func() {
				BeforeEach(func() {
					session := helpers.CF("set-label", "space", spaceName, "some-other-key=some-other-value", "some-key=some-value")
					Eventually(session).Should(Exit(0))
				})
				It("lists the labels", func() {
					session := helpers.CF("labels", "space", spaceName)
					Eventually(session).Should(Say(regexp.QuoteMeta("Getting labels for space %s in org %s as %s...\n\n"), spaceName, orgName, username))
					Eventually(session).Should(Say(`key\s+value`))
					Eventually(session).Should(Say(`some-key\s+some-value`))
					Eventually(session).Should(Say(`some-other-key\s+some-other-value`))
					Eventually(session).Should(Exit(0))
				})
			})

			When("there are no labels set on the space", func() {
				It("indicates that there are no labels", func() {
					session := helpers.CF("labels", "space", spaceName)
					Eventually(session).Should(Say(regexp.QuoteMeta("Getting labels for space %s in org %s as %s...\n\n"), spaceName, orgName, username))
					Expect(session).ToNot(Say(`key\s+value`))
					Eventually(session).Should(Say("No labels found."))
					Eventually(session).Should(Exit(0))
				})
			})

			When("the space does not exist", func() {
				It("displays an error", func() {
					session := helpers.CF("labels", "space", "non-existent-space")
					Eventually(session).Should(Say(regexp.QuoteMeta("Getting labels for space %s in org %s as %s...\n\n"), "non-existent-space", orgName, username))
					Eventually(session.Err).Should(Say("Space 'non-existent-space' not found"))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				})
			})
		})

		Describe("buildpack labels", func() {
			When("there is exactly one buildpack with a given name", func() {
				When("the buildpack is not bound to a stack", func() {
					BeforeEach(func() {
						helpers.SetupBuildpackWithoutStack(buildpackName)
					})
					AfterEach(func() {
						session := helpers.CF("delete-buildpack", buildpackName, "-f")
						Eventually(session).Should(Exit(0))
					})

					It("fails if a non-existent stack is specified", func() {
						session := helpers.CF("labels", "buildpack", buildpackName, "-s", "bogus-stack")
						Eventually(session.Err).Should(Say("Buildpack %s with stack bogus-stack not found", buildpackName))
						Eventually(session).Should(Say("FAILED"))
						Eventually(session).Should(Exit(1))
					})

					It("fails if the -s is specified without an argument", func() {
						session := helpers.CF("labels", "buildpack", buildpackName, "-s")
						Eventually(session.Err).Should(Say("Incorrect Usage:"))
						Eventually(session).Should(Exit(1))
					})

					It("indicates that there are no labels", func() {
						session := helpers.CF("labels", "buildpack", buildpackName)
						Eventually(session).Should(Say(regexp.QuoteMeta("Getting labels for buildpack %s as %s...\n\n"), buildpackName, username))
						Expect(session).ToNot(Say(`key\s+value`))
						Eventually(session).Should(Say("No labels found."))
						Eventually(session).Should(Exit(0))
					})

					When("there are labels on the buildpack", func() {
						BeforeEach(func() {
							session := helpers.CF("set-label", "buildpack", buildpackName, "some-other-key=some-other-value", "some-key=some-value")
							Eventually(session).Should(Exit(0))
						})

						It("lists the labels when no -s flag is given", func() {
							session := helpers.CF("labels", "buildpack", buildpackName)
							Eventually(session).Should(Say(regexp.QuoteMeta("Getting labels for buildpack %s as %s...\n\n"), buildpackName, username))
							Eventually(session).Should(Say(`key\s+value`))
							Eventually(session).Should(Say(`some-key\s+some-value`))
							Eventually(session).Should(Say(`some-other-key\s+some-other-value`))
							Eventually(session).Should(Exit(0))
						})

						It("lists the labels when the -s flag is given with an empty-string", func() {
							session := helpers.CF("labels", "buildpack", buildpackName, "-s", "")
							Eventually(session).Should(Say(regexp.QuoteMeta("Getting labels for buildpack %s as %s...\n\n"), buildpackName, username))
							Eventually(session).Should(Say(`key\s+value`))
							Eventually(session).Should(Say(`some-key\s+some-value`))
							Eventually(session).Should(Say(`some-other-key\s+some-other-value`))
							Eventually(session).Should(Exit(0))
						})
					})
				})

				When("the buildpack is bound to a stack", func() {
					BeforeEach(func() {
						helpers.SetupBuildpackWithStack(buildpackName, "cflinuxfs3")
						session := helpers.CF("set-label", "buildpack", buildpackName, "-s", "cflinuxfs3", "some-other-key=some-other-value", "some-key=some-value")
						Eventually(session).Should(Exit(0))
					})
					AfterEach(func() {
						session := helpers.CF("delete-buildpack", buildpackName, "-f", "-s", "cflinuxfs3")
						Eventually(session).Should(Exit(0))
					})

					It("lists the labels when no stack is specified", func() {
						session := helpers.CF("labels", "buildpack", buildpackName)
						Eventually(session).Should(Say(regexp.QuoteMeta("Getting labels for buildpack %s as %s...\n\n"), buildpackName, username))
						Eventually(session).Should(Say(`key\s+value`))
						Eventually(session).Should(Say(`some-key\s+some-value`))
						Eventually(session).Should(Say(`some-other-key\s+some-other-value`))
						Eventually(session).Should(Exit(0))
					})

					It("lists the labels when the stack is specified", func() {
						session := helpers.CF("labels", "buildpack", buildpackName, "-s", "cflinuxfs3")
						Eventually(session).Should(Say(regexp.QuoteMeta("Getting labels for buildpack %s with stack %s as %s...\n\n"), buildpackName, "cflinuxfs3", username))
						Eventually(session).Should(Say(`key\s+value`))
						Eventually(session).Should(Say(`some-key\s+some-value`))
						Eventually(session).Should(Say(`some-other-key\s+some-other-value`))
						Eventually(session).Should(Exit(0))
					})

					It("fails if a non-existent stack is specified", func() {
						session := helpers.CF("labels", "buildpack", buildpackName, "-s", "bogus-stack")
						Eventually(session.Err).Should(Say("Buildpack %s with stack bogus-stack not found", buildpackName))
						Eventually(session).Should(Say("FAILED"))
						Eventually(session).Should(Exit(1))
					})
				})
			})

			When("there are multiple buildpacks with the same name", func() {
				var (
					newStackName string
				)

				BeforeEach(func() {
					newStackName = helpers.NewStackName()
					helpers.CreateStack(newStackName)
					helpers.SetupBuildpackWithStack(buildpackName, newStackName)
					helpers.SetupBuildpackWithStack(buildpackName, "cflinuxfs3")
					session := helpers.CF("set-label", "buildpack", buildpackName, "-s", newStackName,
						"my-stack-some-other-key=some-other-value", "some-key=some-value")
					Eventually(session).Should(Exit(0))
					session = helpers.CF("set-label", "buildpack", buildpackName, "--stack", "cflinuxfs3",
						"cfl2=var2", "cfl1=var1")
					Eventually(session).Should(Exit(0))
				})
				AfterEach(func() {
					session := helpers.CF("delete-buildpack", buildpackName, "-f", "-s", "cflinuxfs3")
					Eventually(session).Should(Exit(0))
					session = helpers.CF("delete-buildpack", buildpackName, "-f", "-s", newStackName)
					Eventually(session).Should(Exit(0))
					helpers.DeleteStack(newStackName)
				})

				It("fails when no stack is given", func() {
					session := helpers.CF("labels", "buildpack", buildpackName)
					Eventually(session.Err).Should(Say(fmt.Sprintf(`Multiple buildpacks named %s found. Specify a stack name by using a '-s' flag.`, buildpackName)))
					Eventually(session).Should(Say(`FAILED`))
					Eventually(session).Should(Exit(1))
				})

				It("fails when an empty-string stack is given", func() {
					session := helpers.CF("labels", "buildpack", buildpackName, "--stack", "")
					Eventually(session.Err).Should(Say(fmt.Sprintf(`Multiple buildpacks named %s found. Specify a stack name by using a '-s' flag.`, buildpackName)))
					Eventually(session).Should(Say(`FAILED`))
					Eventually(session).Should(Exit(1))
				})

				It("fails when a non-existent stack is given", func() {
					session := helpers.CF("labels", "buildpack", buildpackName, "-s", "bogus-stack")
					Eventually(session.Err).Should(Say("Buildpack %s with stack bogus-stack not found", buildpackName))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				})

				It("lists the labels for buildpackName/newStackName", func() {
					session := helpers.CF("labels", "buildpack", buildpackName, "-s", newStackName)
					Eventually(session).Should(Say(regexp.QuoteMeta("Getting labels for buildpack %s with stack %s as %s...\n\n"), buildpackName, newStackName, username))
					Eventually(session).Should(Say(`key\s+value`))
					Eventually(session).Should(Say(`my-stack-some-other-key\s+some-other-value`))
					Eventually(session).Should(Say(`some-key\s+some-value`))
					Eventually(session).Should(Exit(0))
				})

				It("lists the labels for buildpackName/cflinuxfs3", func() {
					session := helpers.CF("labels", "buildpack", buildpackName, "--stack", "cflinuxfs3")
					Eventually(session).Should(Say(regexp.QuoteMeta("Getting labels for buildpack %s with stack cflinuxfs3 as %s...\n\n"), buildpackName, username))
					Eventually(session).Should(Say(`key\s+value`))
					Eventually(session).Should(Say(`cfl1\s+var1`))
					Eventually(session).Should(Say(`cfl2\s+var2`))
					Eventually(session).Should(Exit(0))
				})

				When("there is also a buildpack with the same name but has no stack", func() {
					BeforeEach(func() {
						helpers.SetupBuildpackWithoutStack(buildpackName)
						session := helpers.CF("set-label", "buildpack", buildpackName,
							"nostack1=var1", "nostack2=var2")
						Eventually(session).Should(Exit(0))

					})
					AfterEach(func() {
						session := helpers.CF("delete-buildpack", buildpackName, "-f")
						Eventually(session).Should(Exit(0))
					})

					It("lists the labels of the no-stack buildpack when no stack is specified", func() {
						session := helpers.CF("labels", "buildpack", buildpackName)
						Eventually(session).Should(Say(regexp.QuoteMeta("Getting labels for buildpack %s as %s...\n\n"), buildpackName, username))
						Eventually(session).Should(Say(`key\s+value`))
						Eventually(session).Should(Say(`nostack1\s+var1`))
						Eventually(session).Should(Say(`nostack2\s+var2`))
						Eventually(session).Should(Exit(0))
					})

					It("fails if a non-existent stack is specified", func() {
						session := helpers.CF("labels", "buildpack", buildpackName, "-s", "bogus-stack")
						Eventually(session.Err).Should(Say("Buildpack %s with stack bogus-stack not found", buildpackName))
						Eventually(session).Should(Say("FAILED"))
						Eventually(session).Should(Exit(1))
					})
				})
			})
		})
	})
})
