package global

import (
	"os"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("update-buildpack command", func() {
	var (
		buildpackName string
		stacks        []string
	)

	BeforeEach(func() {
		helpers.LoginCF()
		buildpackName = helpers.NewBuildpack()
	})

	Context("when the buildpack is not provided", func() {
		It("returns a buildpack argument not provided error", func() {
			session := helpers.CF("update-buildpack", "-p", ".")
			Eventually(session).Should(Exit(1))

			Expect(session.Err.Contents()).To(BeEquivalentTo("Incorrect Usage: the required argument `BUILDPACK` was not provided\n\n"))
		})
	})

	Context("when the buildpack's path does not exist", func() {
		It("returns a buildpack does not exist error", func() {
			session := helpers.CF("update-buildpack", "integration-buildpack-update-bogus", "-p", "this-is-a-bogus-path")

			Eventually(session.Err).Should(Say("Incorrect Usage: The specified path 'this-is-a-bogus-path' does not exist."))
			Eventually(session).Should(Exit(1))
		})
	})

	Context("when the wrong data type is provided as the position argument", func() {
		It("outputs an error message to the user, provides help text, and exits 1", func() {
			session := helpers.CF("update-buildpack", "integration-buildpack-not-an-integer", "-i", "not-an-integer")
			Eventually(session.Err).Should(Say("Incorrect Usage: invalid argument for flag `-i' \\(expected int\\)"))
			Eventually(session).Should(Say("cf update-buildpack BUILDPACK")) // help
			Eventually(session).Should(Exit(1))
		})
	})

	PContext("when not using stack association", func() {
		BeforeEach(func() {
			helpers.BuildpackWithStack(func(buildpackPath string) {
				session := helpers.CF("create-buildpack", buildpackName, buildpackPath, "1")
				Eventually(session).Should(Exit(0))
			}, "")
		})

		It("updates buildpack properly", func() {
			session := helpers.CF("update-buildpack", buildpackName, "-i", "999")
			Eventually(session).Should(Exit(0))
			Eventually(session).Should(Say("Updating buildpack %s\\.\\.\\.", buildpackName))
		})
	})

	Context("when multiple buildpacks have the same name", func() {

		BeforeEach(func() {
			helpers.SkipIfVersionLessThan(ccversion.MinVersionBuildpackStackAssociationV3)
			helpers.SkipIfOneStack()

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

		It("handles ambiguity properly", func() {
			By("failing when no stack is specified")
			session := helpers.CF("update-buildpack", buildpackName, "-i", "999")
			Eventually(session).Should(Exit(1))
			Eventually(session.Out).Should(Say("FAILED"))

			By("failing when updating buildpack with non-existent stack")
			session = helpers.CF("update-buildpack", buildpackName, "-i", "999", "-s", "bogus-stack")
			Eventually(session).Should(Exit(1))
			Eventually(session.Out).Should(Say("FAILED"))

			By("updating when a stack associated with that buildpack name is specified")
			session = helpers.CF("update-buildpack", buildpackName, "-s", stacks[0], "-i", "999")
			Eventually(session).Should(Exit(0))
			Eventually(session.Out).Should(Say("OK"))
			Expect(session.Err).NotTo(Say("Incorrect Usage:"))
		})
	})

	Context("when a URL is provided as the buildpack", func() {
		var dir string

		BeforeEach(func() {
			helpers.BuildpackWithStack(func(buildpackPath string) {
				session := helpers.CF("create-buildpack", buildpackName, buildpackPath, "1")
				Eventually(session).Should(Exit(0))
			}, stacks[1])
		})

		AfterEach(func() {
			Expect(os.RemoveAll(dir)).To(Succeed())
		})

		It("outputs an error message to the user, provides help text, and exits 1", func() {
			session := helpers.CF("update-buildpack", buildpackName, "-p", "https://example.com/bogus.tgz")
			Eventually(session).Should(Say("Failed to create a local temporary zip file for the buildpack"))
			Eventually(session).Should(Say("FAILED"))
			Eventually(session).Should(Say("Couldn't write zip file: zip: not a valid zip file"))
			Eventually(session).Should(Exit(1))
		})
	})
})
