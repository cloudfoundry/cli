package global

import (
	"io/ioutil"
	"os"

	"code.cloudfoundry.org/cli/integration/helpers"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("create-buildpack command", func() {
	var (
		buildpackName string
		stacks        []string
	)

	BeforeEach(func() {
		helpers.LoginCF()
		buildpackName = helpers.NewBuildpack()
	})

	Context("when creating a new buildpack", func() {
		Context("without stack association", func() {
			It("creates the new buildpack", func() {
				helpers.BuildpackWithStack(func(buildpackPath string) {
					session := helpers.CF("create-buildpack", buildpackName, buildpackPath, "1")
					Eventually(session).Should(Exit(0))
				}, "")

				session := helpers.CF("buildpacks")
				Eventually(session).Should(Exit(0))
				Expect(session.Out).To(Say(buildpackName))
			})
		})

		Context("with stack association", func() {
			BeforeEach(func() {
				helpers.SkipIfVersionLessThan(ccversion.MinVersionBuildpackStackAssociationV3)
				stacks = helpers.FetchStacks()
			})

			It("creates the new buildpack and assigns its stack", func() {
				helpers.BuildpackWithStack(func(buildpackPath string) {
					session := helpers.CF("create-buildpack", buildpackName, buildpackPath, "1")
					Eventually(session).Should(Exit(0))
				}, stacks[0])

				session := helpers.CF("buildpacks")
				Eventually(session).Should(Exit(0))
				Expect(session.Out).To(Say(buildpackName + ".*cflinuxfs2.*1"))
			})
		})
	})

	Context("when creating a buildpack with no stack that already exists", func() {
		BeforeEach(func() {
			stacks = helpers.FetchStacks()

			helpers.BuildpackWithStack(func(buildpackPath string) {
				session := helpers.CF("create-buildpack", buildpackName, buildpackPath, "1")
				Eventually(session).Should(Exit(0))
			}, "")
		})

		It("issues a warning and exits 0", func() {
			helpers.BuildpackWithStack(func(buildpackPath string) {
				session := helpers.CF("create-buildpack", buildpackName, buildpackPath, "1")
				Eventually(session).Should(Exit(0))
				Eventually(session.Out).Should(Say("Buildpack %s already exists", buildpackName))
			}, "")
		})
	})

	Context("when creating a buildpack/stack that already exists", func() {
		BeforeEach(func() {
			helpers.SkipIfVersionLessThan(ccversion.MinVersionBuildpackStackAssociationV3)

			helpers.BuildpackWithStack(func(buildpackPath string) {
				session := helpers.CF("create-buildpack", buildpackName, buildpackPath, "1")
				Eventually(session).Should(Exit(0))
			}, stacks[0])
		})

		It("issues a warning and exits 0", func() {
			helpers.BuildpackWithStack(func(buildpackPath string) {
				session := helpers.CF("create-buildpack", buildpackName, buildpackPath, "1")
				Eventually(session).Should(Exit(0))
				Eventually(session.Out).Should(Say("The buildpack name " + buildpackName + " is already in use for the stack " + stacks[0]))
			}, stacks[0])
		})
	})

	Context("when the wrong data type is provided as the position argument", func() {
		var (
			filename string
		)

		BeforeEach(func() {
			// The args take a filepath. Creating a real file will avoid the file
			// does not exist error, and trigger the correct error case we are
			// testing.
			f, err := ioutil.TempFile("", "create-buildpack-invalid")
			Expect(err).NotTo(HaveOccurred())
			f.Close()

			filename = f.Name()
		})

		AfterEach(func() {
			err := os.Remove(filename)
			Expect(err).NotTo(HaveOccurred())
		})

		It("outputs an error message to the user, provides help text, and exits 1", func() {
			session := helpers.CF("create-buildpack", buildpackName, filename, "not-an-integer")
			Eventually(session.Err).Should(Say("Incorrect usage: Value for POSITION must be integer"))
			Eventually(session).Should(Say("cf create-buildpack BUILDPACK PATH POSITION")) // help
			Eventually(session).Should(Exit(1))
		})
	})

	Context("when a nonexistent file is provided", func() {
		It("outputs an error message to the user and exits 1", func() {
			session := helpers.CF("create-buildpack", buildpackName, "some-bogus-file", "1")
			Eventually(session.Err).Should(Say("Incorrect Usage: The specified path 'some-bogus-file' does not exist."))
			Eventually(session).Should(Exit(1))
		})
	})

	Context("when a URL is provided as the buildpack", func() {
		It("outputs an error message to the user, provides help text, and exits 1", func() {
			session := helpers.CF("create-buildpack", buildpackName, "https://example.com/bogus.tgz", "1")
			Eventually(session).Should(Say("Failed to create a local temporary zip file for the buildpack"))
			Eventually(session).Should(Say("FAILED"))
			Eventually(session).Should(Say("Couldn't write zip file: zip: not a valid zip file"))
			Eventually(session).Should(Exit(1))
		})
	})
})
