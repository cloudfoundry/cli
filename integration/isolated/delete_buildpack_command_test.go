package isolated

import (
	"io/ioutil"
	"path/filepath"

	"path/filepath"

	"code.cloudfoundry.org/cli/integration/helpers"

	. "code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("delete-buildpack command", func() {
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
		BeforeEach(func() {
			helpers.LoginCF()
		})

		It("displays a warning and exits 0", func() {
			session := helpers.CF("delete-buildpack", "-f", "nonexistent-buildpack")
			Eventually(session).Should(Say("Deleting buildpack nonexistent-buildpack"))
			Eventually(session).Should(Say("OK"))
			Eventually(session).Should(Say("Buildpack nonexistent-buildpack does not exist."))
			Eventually(session).Should(Exit(0))
		})
	})

	Context("when the buildpack exists", func() {
		var buildpackName string

		BeforeEach(func() {
			helpers.LoginCF()

			dir, err := ioutil.TempDir("", "update-buildpack-test")
			Expect(err).ToNot(HaveOccurred())

			filename := "some-file"
			tempfile := filepath.Join(dir, filename)
			err = ioutil.WriteFile(tempfile, []byte{}, 0400)
			Expect(err).ToNot(HaveOccurred())

			buildpackName = helpers.NewBuildpack()
			session := helpers.CF("create-buildpack", buildpackName, dir, "1")
			Eventually(session).Should(Exit(0))
		})

		Context("when the -f flag not is provided", func() {
			var buffer *Buffer

			BeforeEach(func() {
				buffer = NewBuffer()
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

				It("does not delete the org", func() {
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

})
