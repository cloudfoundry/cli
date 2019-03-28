package global

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
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
			Eventually(session).Should(Say(`--force, -f\s+Force deletion without confirmation`))
			Eventually(session).Should(Say(`--stack, -s\s+Specify stack to disambiguate buildpacks with the same name. Required when buildpack name is ambiguous`))
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
				Eventually(session.Err).Should(Say("Buildpack nonexistent-buildpack does not exist."))
				Eventually(session).Should(Exit(0))
			})
		})

		When("the user specifies a stack", func() {
			BeforeEach(func() {
				stacks = helpers.FetchStacks()
			})

			It("displays a warning and exits 0", func() {
				session := helpers.CF("delete-buildpack", "-f", "nonexistent-buildpack", "-s", stacks[0])
				Eventually(session).Should(Say(`Deleting buildpack nonexistent-buildpack with stack %s\.\.\.`, stacks[0]))
				Eventually(session).Should(Say("OK"))
				Eventually(session.Err).Should(Say("Buildpack nonexistent-buildpack with stack %s not found.", stacks[0]))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("there is exactly one buildpack with the specified name", func() {
		When("the stack is specified", func() {
			BeforeEach(func() {
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
				Eventually(session.Err).Should(Say("Multiple buildpacks named %s found. Specify a stack name by using a '-s' flag.", buildpackName))
				Eventually(session).Should(Exit(1))

				By("succeeding with warning when the buildpack name matches but the stack does not")

				session = helpers.CF("delete-buildpack", buildpackName, "-s", "not-a-real-stack", "-f")
				Eventually(session).Should(Say(`Deleting buildpack %s with stack not-a-real-stack\.\.\.`, buildpackName))
				Eventually(session).Should(Say("OK"))
				Eventually(session.Err).Should(Say(`Buildpack %s with stack not-a-real-stack not found\.`, buildpackName))
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
				By("deleting the nil stack buildpack when no stack specified")
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
				_, err := buffer.Write([]byte("y\n"))
				Expect(err).ToNot(HaveOccurred())
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
				_, err := buffer.Write([]byte("n\n"))
				Expect(err).ToNot(HaveOccurred())
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
				_, err := buffer.Write([]byte("\n"))
				Expect(err).ToNot(HaveOccurred())
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

	When("the --force flag is provided", func() {
		It("deletes the buildpack without asking for confirmation", func() {
			session := helpers.CF("delete-buildpack", buildpackName, "--force")
			Eventually(session).Should(Say(`Deleting buildpack %s\.\.\.`, buildpackName))
			Eventually(session).Should(Say("OK"))
			Eventually(session).Should(Exit(0))
		})
	})
})
