// +build !partialPush

package global

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("delete-buildpack command", func() {
	var (
		buildpackName string
		stacks        []string
	)

	BeforeEach(func() {
		helpers.LoginCF()
		buildpackName = helpers.NewBuildpackName()
	})

	When("the --help flag is passed", func() {
		It("Displays the appropriate help text", func() {
			session := helpers.CF("delete-buildpack", "--help")
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Say("delete-buildpack - Delete a buildpack"))
			Eventually(session).Should(Say("\n"))
			Eventually(session).Should(Say("USAGE:"))
			Eventually(session).Should(Say(`cf delete-buildpack BUILDPACK \[-f] \[-s STACK]`))
			Eventually(session).Should(Say("\n"))
			Eventually(session).Should(Say("OPTIONS:"))
			Eventually(session).Should(Say(`-f\s+Force deletion without confirmation`))
			Eventually(session).Should(Say(`-s\s+Specify stack to disambiguate buildpacks with the same name. Required when buildpack name is ambiguous`))
			Eventually(session).Should(Say("\n"))
			Eventually(session).Should(Say("SEE ALSO:"))
			Eventually(session).Should(Say("buildpacks"))
			Eventually(session).Should(Exit(0))
		})
	})

	When("the buildpack name is not provided", func() {
		It("displays an error and help", func() {
			session := helpers.CF("delete-buildpack")
			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `BUILDPACK` was not provided"))
			Eventually(session).Should(Say("USAGE"))
			Eventually(session).Should(Exit(1))
		})
	})

	When("the buildpack doesn't exist", func() {
		When("the user does not specify a stack", func() {
			It("displays a warning and exits 0", func() {
				session := helpers.CF("delete-buildpack", "-f", "nonexistent-buildpack")
				Eventually(session).Should(Say(`Deleting buildpack nonexistent-buildpack\.\.\.`))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Say("Buildpack nonexistent-buildpack does not exist."))
				Eventually(session).Should(Exit(0))
			})
		})

		When("the user specifies a stack", func() {
			BeforeEach(func() {
				helpers.SkipIfVersionLessThan(ccversion.MinVersionBuildpackStackAssociationV2)
				stacks = helpers.FetchStacks()
			})

			It("displays a warning and exits 0", func() {
				session := helpers.CF("delete-buildpack", "-f", "nonexistent-buildpack", "-s", stacks[0])
				Eventually(session).Should(Say(`Deleting buildpack nonexistent-buildpack with stack %s\.\.\.`, stacks[0]))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Say("Buildpack nonexistent-buildpack with stack %s not found.", stacks[0]))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("there is exactly one buildpack with the specified name", func() {
		When("the stack is specified", func() {
			BeforeEach(func() {
				helpers.SkipIfVersionLessThan(ccversion.MinVersionBuildpackStackAssociationV2)
				stacks = helpers.FetchStacks()
				helpers.BuildpackWithStack(func(buildpackPath string) {
					session := helpers.CF("create-buildpack", buildpackName, buildpackPath, "1")
					Eventually(session).Should(Exit(0))
				}, stacks[0])
			})

			It("deletes the specified buildpack", func() {
				session := helpers.CF("delete-buildpack", buildpackName, "-s", stacks[0], "-f")
				Eventually(session.Out).Should(Say("OK"))
				Eventually(session).Should(Exit(0))
			})
		})

		When("the stack is not specified", func() {
			It("deletes the specified buildpack", func() {
				session := helpers.CF("delete-buildpack", buildpackName, "-f")
				Eventually(session.Out).Should(Say("OK"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("there are two buildpacks with same name", func() {
		BeforeEach(func() {
			helpers.SkipIfVersionLessThan(ccversion.MinVersionBuildpackStackAssociationV2)
			stacks = helpers.EnsureMinimumNumberOfStacks(2)
		})

		Context("neither buildpack has a nil stack", func() {
			BeforeEach(func() {
				helpers.BuildpackWithStack(func(buildpackPath string) {
					session := helpers.CF("create-buildpack", buildpackName, buildpackPath, "1")
					Eventually(session).Should(Exit(0))
				}, stacks[0])

				helpers.BuildpackWithStack(func(buildpackPath string) {
					session := helpers.CF("create-buildpack", buildpackName, buildpackPath, "1")
					Eventually(session).Should(Exit(0))
				}, stacks[1])
			})

			It("properly handles ambiguity", func() {
				By("failing when no stack specified")

				session := helpers.CF("delete-buildpack", buildpackName, "-f")
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))

				By("succeeding with warning when the buildpack name matches but the stack does not")

				session = helpers.CF("delete-buildpack", buildpackName, "-s", "not-a-real-stack", "-f")
				Eventually(session).Should(Say(`Deleting buildpack %s with stack not-a-real-stack\.\.\.`, buildpackName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Say(`Buildpack %s with stack not-a-real-stack not found\.`, buildpackName))
				Eventually(session).Should(Exit(0))

				By("deleting the buildpack when the associated stack is specified")

				session = helpers.CF("delete-buildpack", buildpackName, "-s", stacks[0], "-f")
				Eventually(session).Should(Say(`Deleting buildpack %s with stack %s\.\.\.`, buildpackName, stacks[0]))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))

				session = helpers.CF("delete-buildpack", buildpackName, "-s", stacks[1], "-f")
				Eventually(session).Should(Say(`Deleting buildpack %s with stack %s\.\.\.`, buildpackName, stacks[1]))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))
			})
		})

		Context("one buildpack has a nil stack", func() {
			BeforeEach(func() {
				helpers.BuildpackWithStack(func(buildpackPath string) {
					session := helpers.CF("create-buildpack", buildpackName, buildpackPath, "1")
					Eventually(session).Should(Exit(0))
				}, stacks[0])

				helpers.BuildpackWithoutStack(func(buildpackPath string) {
					session := helpers.CF("create-buildpack", buildpackName, buildpackPath, "1")
					Eventually(session).Should(Exit(0))
				})
			})

			It("properly handles ambiguity", func() {
				By("deleting nil buildpack when no stack specified")
				session := helpers.CF("delete-buildpack", buildpackName, "-f")
				Eventually(session.Out).Should(Say("OK"))
				Eventually(session).Should(Exit(0))

				By("deleting the remaining buildpack when no stack is specified")
				session = helpers.CF("delete-buildpack", buildpackName, "-f")
				Eventually(session.Out).Should(Say("OK"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the -f flag not is provided", func() {
		var buffer *Buffer

		BeforeEach(func() {
			buffer = NewBuffer()

			helpers.BuildpackWithoutStack(func(buildpackPath string) {
				session := helpers.CF("create-buildpack", buildpackName, buildpackPath, "1")
				Eventually(session).Should(Exit(0))
			})
		})

		When("the user enters 'y'", func() {
			BeforeEach(func() {
				buffer.Write([]byte("y\n"))
			})

			It("deletes the buildpack", func() {
				session := helpers.CFWithStdin(buffer, "delete-buildpack", buildpackName)
				Eventually(session).Should(Say(`Deleting buildpack %s\.\.\.`, buildpackName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))
			})
		})

		When("the user enters 'n'", func() {
			BeforeEach(func() {
				buffer.Write([]byte("n\n"))
			})

			It("does not delete the buildpack", func() {
				session := helpers.CFWithStdin(buffer, "delete-buildpack", buildpackName)
				Eventually(session).Should(Say("Delete cancelled"))
				Eventually(session).Should(Exit(0))

				session = helpers.CF("buildpacks")
				Eventually(session).Should(Say(buildpackName))
				Eventually(session).Should(Exit(0))
			})
		})

		When("the user enters the default input (hits return)", func() {
			BeforeEach(func() {
				buffer.Write([]byte("\n"))
			})

			It("does not delete the buildpack", func() {
				session := helpers.CFWithStdin(buffer, "delete-buildpack", buildpackName)
				Eventually(session).Should(Say("Delete cancelled"))
				Eventually(session).Should(Exit(0))

				session = helpers.CF("buildpacks")
				Eventually(session).Should(Say(buildpackName))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the -f flag is provided", func() {
		It("deletes the org", func() {
			session := helpers.CF("delete-buildpack", buildpackName, "-f")
			Eventually(session).Should(Say(`Deleting buildpack %s\.\.\.`, buildpackName))
			Eventually(session).Should(Say("OK"))
			Eventually(session).Should(Exit(0))
		})
	})

	When("the -s flag is provided", func() {
		When("the API is less than the minimum", func() {
			var server *Server

			BeforeEach(func() {
				server = helpers.StartAndTargetServerWithAPIVersions(ccversion.MinV2ClientVersion, ccversion.MinV3ClientVersion)
			})

			AfterEach(func() {
				server.Close()
			})

			It("fails with no networking api error message", func() {
				session := helpers.CF("delete-buildpack", "potato", "-s", "ahoyhoy")
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("Option '-s' requires CF API version %s or higher. Your target is %s.", ccversion.MinVersionBuildpackStackAssociationV2, ccversion.MinV2ClientVersion))
				Eventually(session).Should(Exit(1))
			})
		})
	})
})
