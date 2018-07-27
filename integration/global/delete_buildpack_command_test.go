package global

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
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
		buildpackName = helpers.NewBuildpack()
	})

	Context("when the environment is not setup correctly", func() {
		XIt("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(false, false, ReadOnlyOrg, "delete-buildpack", "nonexistent-buildpack")
		})
	})

	Context("when the buildpack name is not provided", func() {
		It("displays an error and help", func() {
			session := helpers.CF("delete-buildpack")
			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `BUILDPACK` was not provided"))
			Eventually(session).Should(Say("USAGE"))
			Eventually(session).Should(Exit(1))
		})
	})

	Context("when the buildpack doesn't exist", func() {
		It("displays a warning and exits 0", func() {
			session := helpers.CF("delete-buildpack", "-f", "nonexistent-buildpack")
			Eventually(session).Should(Say("Deleting buildpack nonexistent-buildpack"))
			Eventually(session).Should(Say("OK"))
			Eventually(session).Should(Say("Buildpack nonexistent-buildpack does not exist."))
			Eventually(session).Should(Exit(0))
		})
	})

	PContext("there is exactly one buildpack with the specified name", func() {

		BeforeEach(func() {
			stacks = helpers.FetchStacks()
			helpers.BuildpackWithStack(func(buildpackPath string) {
				session := helpers.CF("create-buildpack", buildpackName, buildpackPath, "1")
				Eventually(session).Should(Exit(0))
			}, stacks[0])
		})

		Context("when the stack is specified", func() {
			It("deletes the specified buildpack", func() {
				session := helpers.CF("delete-buildpack", buildpackName, "-s", stacks[0], "-f")
				Eventually(session).Should(Exit(0))
				Eventually(session.Out).Should(Say("OK"))
			})
		})

		Context("when the stack is not specified", func() {
			It("deletes the specified buildpack", func() {
				session := helpers.CF("delete-buildpack", buildpackName, "-f")
				Eventually(session).Should(Exit(0))
				Eventually(session.Out).Should(Say("OK"))
			})
		})
	})

	Context("there are two buildpacks with same name", func() {

		BeforeEach(func() {
			helpers.SkipIfVersionLessThan(ccversion.MinVersionBuildpackStackAssociationV3)
			helpers.SkipIfOneStack()
		})

		Context("neither buildpack has a nil stack", func() {
			BeforeEach(func() {
				stacks = helpers.FetchStacks()

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
				Eventually(session).Should(Exit(1))
				Eventually(session.Out).Should(Say("FAILED"))

				By("deleting the buildpack when the associated stack is specified")

				session = helpers.CF("delete-buildpack", buildpackName, "-s", stacks[0], "-f")
				Eventually(session).Should(Exit(0))
				Eventually(session.Out).Should(Say("OK"))

				session = helpers.CF("delete-buildpack", buildpackName, "-s", stacks[1], "-f")
				Eventually(session).Should(Exit(0))
				Eventually(session.Out).Should(Say("OK"))
			})
		})

		Context("one buildpack has a nil stack", func() {

			BeforeEach(func() {
				stacks = helpers.FetchStacks()

				helpers.BuildpackWithStack(func(buildpackPath string) {
					session := helpers.CF("create-buildpack", buildpackName, buildpackPath, "1")
					Eventually(session).Should(Exit(0))
				}, stacks[0])

				helpers.BuildpackWithStack(func(buildpackPath string) {
					session := helpers.CF("create-buildpack", buildpackName, buildpackPath, "1")
					Eventually(session).Should(Exit(0))
				}, "")
			})

			It("properly handles ambiguity", func() {
				By("deleting nil buildpack when no stack specified")
				session := helpers.CF("delete-buildpack", buildpackName, "-f")
				Eventually(session).Should(Exit(0))
				Eventually(session.Out).Should(Say("OK"))

				By("deleting the remaining buildpack when no stack is specified")
				session = helpers.CF("delete-buildpack", buildpackName, "-f")
				Eventually(session).Should(Exit(0))
				Eventually(session.Out).Should(Say("OK"))
			})
		})
	})

	Context("when the -f flag not is provided", func() {
		var buffer *Buffer

		BeforeEach(func() {
			buffer = NewBuffer()

			helpers.BuildpackWithStack(func(buildpackPath string) {
				session := helpers.CF("create-buildpack", buildpackName, buildpackPath, "1")
				Eventually(session).Should(Exit(0))
			}, "")
		})

		Context("when the user enters 'y'", func() {
			BeforeEach(func() {
				buffer.Write([]byte("y\n"))
			})

			It("deletes the buildpack", func() {
				session := helpers.CFWithStdin(buffer, "delete-buildpack", buildpackName)
				Eventually(session).Should(Say("Deleting buildpack %s", buildpackName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))
			})
		})

		Context("when the user enters 'n'", func() {
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

		Context("when the user enters the default input (hits return)", func() {
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

	Context("when the -f flag is provided", func() {
		It("deletes the org", func() {
			session := helpers.CF("delete-buildpack", buildpackName, "-f")
			Eventually(session).Should(Say("Deleting buildpack %s", buildpackName))
			Eventually(session).Should(Say("OK"))
			Eventually(session).Should(Exit(0))
		})
	})
})
