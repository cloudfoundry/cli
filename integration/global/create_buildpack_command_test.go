package global

import (
	"io/ioutil"
	"os"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	. "code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("create-buildpack command", func() {
	var (
		bpname string
	)

	BeforeEach(func() {
		LoginCF()
		bpname = NewBuildpack()
	})

	Context("when creating a new buildpack", func() {
		Context("without stack association", func() {
			It("creates the new buildpack", func() {
				session := CF("create-buildpack", bpname, "../assets/test_buildpacks/simple_buildpack-v1.0.0.zip", "1")
				Eventually(session).Should(Exit(0))

				session = CF("buildpacks")
				Eventually(session).Should(Exit(0))
				Expect(session.Out).To(Say(bpname))
			})
		})

		Context("with stack association", func() {
			BeforeEach(func() {
				SkipIfVersionLessThan(ccversion.MinVersionBuildpackStackAssociationV3)
			})

			It("creates the new buildpack and assigns its stack", func() {
				session := CF("create-buildpack", bpname, "../assets/test_buildpacks/simple_buildpack-cflinuxfs2-v1.0.0.zip", "1")
				Eventually(session).Should(Exit(0))

				session = CF("buildpacks")
				Eventually(session).Should(Exit(0))
				Expect(session.Out).To(Say(bpname + ".*cflinuxfs2.*1"))
			})
		})
	})

	Context("when creating a buildpack with no stack that already exists", func() {
		BeforeEach(func() {
			session := CF("create-buildpack", bpname, "../assets/test_buildpacks/simple_buildpack-v1.0.0.zip", "1")
			Eventually(session).Should(Exit(0))
		})

		It("issues a warning and exits 0", func() {
			session := CF("create-buildpack", bpname, "../assets/test_buildpacks/simple_buildpack-v1.0.0.zip", "1")
			Eventually(session).Should(Exit(0))
			Eventually(session.Out).Should(Say("Buildpack %s already exists", bpname))
		})
	})

	Context("when creating a buildpack/stack that already exists", func() {
		BeforeEach(func() {
			SkipIfVersionLessThan(ccversion.MinVersionBuildpackStackAssociationV3)

			session := CF("create-buildpack", bpname, "../assets/test_buildpacks/simple_buildpack-cflinuxfs2-v1.0.0.zip", "1")
			Eventually(session).Should(Exit(0))
		})

		It("issues a warning and exits 0", func() {
			session := CF("create-buildpack", bpname, "../assets/test_buildpacks/simple_buildpack-cflinuxfs2-v1.0.0.zip", "1")
			Eventually(session).Should(Exit(0))
			Eventually(session.Out).Should(Say("The buildpack name " + bpname + " is already in use for the stack cflinuxfs2"))
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
			session := CF("create-buildpack", bpname, filename, "not-an-integer")
			Eventually(session.Err).Should(Say("Incorrect usage: Value for POSITION must be integer"))
			Eventually(session).Should(Say("cf create-buildpack BUILDPACK PATH POSITION")) // help
			Eventually(session).Should(Exit(1))
		})
	})

	Context("when a nonexistent file is provided", func() {
		It("outputs an error message to the user and exits 1", func() {
			session := CF("create-buildpack", bpname, "some-bogus-file", "1")
			Eventually(session.Err).Should(Say("Incorrect Usage: The specified path 'some-bogus-file' does not exist."))
			Eventually(session).Should(Exit(1))
		})
	})

	Context("when a URL is provided as the buildpack", func() {
		It("outputs an error message to the user, provides help text, and exits 1", func() {
			session := CF("create-buildpack", bpname, "https://example.com/bogus.tgz", "1")
			Eventually(session).Should(Say("Failed to create a local temporary zip file for the buildpack"))
			Eventually(session).Should(Say("FAILED"))
			Eventually(session).Should(Say("Couldn't write zip file: zip: not a valid zip file"))
			Eventually(session).Should(Exit(1))
		})
	})
})
