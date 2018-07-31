package experimental

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("create buildpack command", func() {
	var buildpackName string

	BeforeEach(func() {

		buildpackName = helpers.NewBuildpack()
	})

	Describe("help", func() {
		Context("when --help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("create-buildpack", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("create-buildpack - Create a buildpack"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf create-buildpack BUILDPACK PATH POSITION \\[--enable|--disable\\]"))
				Eventually(session).Should(Say("TIP:"))
				Eventually(session).Should(Say("Path should be a zip file, a url to a zip file, or a local directory. Position is a positive integer, sets priority, and is sorted from lowest to highest."))
				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say("--disable\\s+Disable the buildpack from being used for staging"))
				Eventually(session).Should(Say("--enable\\s+Enable the buildpack to be used for staging"))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("buildpacks, push"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("when the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			path, err := os.Getwd()
			Expect(err).ToNot(HaveOccurred())

			helpers.CheckEnvironmentTargetedCorrectly(false, false, ReadOnlyOrg, "create-buildpack", "fake-buildpack", path, "1")
		})
	})

	Context("when the user is logged in", func() {
		BeforeEach(func() {
			helpers.LoginCF()
		})

		Context("when uploading from a directory", func() {
			var buildpackDir string

			Context("when zipping the directory errors", func() {
				BeforeEach(func() {
					buildpackDir = "some/nonexistent/dir"
				})

				AfterEach(func() {
					Expect(os.RemoveAll(buildpackDir)).ToNot(HaveOccurred())
				})

				It("returns an error", func() {
					session := helpers.CF("create-buildpack", buildpackName, buildpackDir, "1")
					Eventually(session).Should(Exit(1))
					Expect(session.Err).To(Say("Incorrect Usage: The specified path 'some/nonexistent/dir' does not exist."))
				})
			})

			Context("when zipping the directory succeeds", func() {
				BeforeEach(func() {
					var err error
					buildpackDir, err = ioutil.TempDir("", "buildpackdir-")
					Expect(err).ToNot(HaveOccurred())
				})

				AfterEach(func() {
					Expect(os.RemoveAll(buildpackDir)).ToNot(HaveOccurred())
				})

				It("successfully uploads a buildpack", func() {
					username, _ := helpers.GetCredentials()
					session := helpers.CF("create-buildpack", buildpackName, buildpackDir, "1")
					Eventually(session).Should(Say("Creating buildpack %s as %s...", buildpackName, username))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Say("Uploading buildpack %s as %s...", buildpackName, username))
					Eventually(session).Should(Say("Done uploading"))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))
				})
			})
		})

		Context("when uploading from a zip", func() {
			var stacks []string

			BeforeEach(func() {
				stacks = helpers.FetchStacks()
			})

			Context("when specifying a valid path", func() {
				Context("when the new buildpack is unique", func() {
					Context("when the new buildpack has a nil stack", func() {
						It("successfully uploads a buildpack", func() {
							helpers.BuildpackWithStack(func(buildpackPath string) {
								session := helpers.CF("create-buildpack", buildpackName, buildpackPath, "1")
								Eventually(session).Should(Exit(0))
							}, "")

							session := helpers.CF("buildpacks")
							Eventually(session).Should(Exit(0))
							Expect(session.Out).To(Say(`%s\s+1`, buildpackName))
						})
					})

					Context("when the new buildpack has a valid stack", func() {
						BeforeEach(func() {
							helpers.SkipIfVersionLessThan(ccversion.MinVersionBuildpackStackAssociationV3)
						})

						It("successfully uploads a buildpack", func() {
							helpers.BuildpackWithStack(func(buildpackPath string) {
								session := helpers.CF("create-buildpack", buildpackName, buildpackPath, "1")
								Eventually(session).Should(Exit(0))
							}, stacks[0])

							session := helpers.CF("buildpacks")
							Eventually(session).Should(Exit(0))
							Expect(session.Out).To(Say(`%s\s+%s\s+1`, buildpackName, stacks[0]))
						})
					})
				})

				Context("when the new buildpack has an invalid stack", func() {
					BeforeEach(func() {
						helpers.SkipIfVersionLessThan(ccversion.MinVersionBuildpackStackAssociationV3)
					})

					It("returns the appropriate error", func() {
						helpers.BuildpackWithStack(func(buildpackPath string) {
							session := helpers.CF("create-buildpack", buildpackName, buildpackPath, "1")
							Eventually(session).Should(Exit(1))
							Expect(session.Err).To(Say("Uploaded buildpack stack \\(fake-stack\\) does not exist"))
						}, "fake-stack")
					})
				})

				Context("when a buildpack with the same name exists", func() {
					var (
						existingBuildpack string
					)

					BeforeEach(func() {
						existingBuildpack = buildpackName
					})

					Context("when the new buildpack has a nil stack", func() {
						BeforeEach(func() {
							helpers.SkipIfVersionLessThan(ccversion.MinVersionBuildpackStackAssociationV3)
						})

						Context("when the existing buildpack does not have a nil stack", func() {
							BeforeEach(func() {
								helpers.BuildpackWithStack(func(buildpackPath string) {
									session := helpers.CF("create-buildpack", existingBuildpack, buildpackPath, "5")
									Eventually(session).Should(Exit(0))
								}, stacks[0])
							})

							It("successfully uploads a buildpack", func() {
								helpers.BuildpackWithStack(func(buildpackPath string) {
									session := helpers.CF("create-buildpack", buildpackName, buildpackPath, "1")
									Eventually(session).Should(Exit(0))
								}, stacks[0])

								session := helpers.CF("buildpacks")
								Eventually(session).Should(Exit(0))
								Expect(session.Out).To(Say(`%s\s+1`, buildpackName))
								Expect(session.Out).To(Say(`%s\s+%s\s+6`, existingBuildpack, stacks[0]))
							})
						})

						Context("when the existing buildpack has a nil stack", func() {
							BeforeEach(func() {
								helpers.BuildpackWithStack(func(buildpackPath string) {
									session := helpers.CF("create-buildpack", existingBuildpack, buildpackPath, "5")
									Eventually(session).Should(Exit(0))
								}, "")
							})

							It("prints a warning", func() {
								helpers.BuildpackWithStack(func(buildpackPath string) {
									session := helpers.CF("create-buildpack", buildpackName, buildpackPath, "1")
									Eventually(session).Should(Exit(0))
									Eventually(session.Err).Should(Say("Buildpack %s already exists without a stack", buildpackName))
								}, "")

								session := helpers.CF("buildpacks")
								Eventually(session).Should(Exit(0))
								Expect(session).To(Say(`%s\s+5`, existingBuildpack))
							})
						})
					})

					Context("when the new buildpack has a non-nil stack", func() {
						BeforeEach(func() {
							helpers.SkipIfVersionLessThan(ccversion.MinVersionBuildpackStackAssociationV3)
						})

						Context("when the existing buildpack has a different non-nil stack", func() {
							BeforeEach(func() {
								helpers.SkipIfOneStack()
								helpers.BuildpackWithStack(func(buildpackPath string) {
									session := helpers.CF("create-buildpack", existingBuildpack, buildpackPath, "5")
									Eventually(session).Should(Exit(0))
								}, stacks[1])
							})

							It("successfully uploads a buildpack", func() {
								helpers.BuildpackWithStack(func(buildpackPath string) {
									session := helpers.CF("create-buildpack", buildpackName, buildpackPath, "1")
									Eventually(session).Should(Exit(0))
								}, stacks[0])

								session := helpers.CF("buildpacks")
								Eventually(session).Should(Exit(0))
								Expect(session.Out).To(Say(`%s\s+%s\s+1`, buildpackName, stacks[0]))
								Expect(session.Out).To(Say(`%s\s+%s\s+6`, existingBuildpack, stacks[1]))
							})
						})

						Context("when the existing buildpack has a nil stack", func() {
							BeforeEach(func() {
								helpers.BuildpackWithStack(func(buildpackPath string) {
									session := helpers.CF("create-buildpack", existingBuildpack, buildpackPath, "5")
									Eventually(session).Should(Exit(0))
								}, "")
							})

							It("prints a warning and tip but doesn't exit 1", func() {
								helpers.BuildpackWithStack(func(buildpackPath string) {
									session := helpers.CF("create-buildpack", buildpackName, buildpackPath, "1")
									Eventually(session).Should(Exit(0))
									Eventually(session.Err).Should(Say("Buildpack %s already exists without a stack", buildpackName))
									Eventually(session.Out).Should(Say("TIP: use 'cf buildpacks' and 'cf delete-buildpack' to delete buildpack %s without a stack", buildpackName))
								}, stacks[0])

							})
						})

						Context("when the existing buildpack has the same non-nil stack", func() {
							BeforeEach(func() {
								helpers.BuildpackWithStack(func(buildpackPath string) {
									session := helpers.CF("create-buildpack", existingBuildpack, buildpackPath, "5")
									Eventually(session).Should(Exit(0))
								}, stacks[0])

							})

							It("prints a warning but doesn't exit 1", func() {
								helpers.BuildpackWithStack(func(buildpackPath string) {
									session := helpers.CF("create-buildpack", buildpackName, buildpackPath, "1")
									Eventually(session).Should(Exit(0))
									Expect(session.Err).To(Say("The buildpack name %s is already in use for the stack %s", buildpackName, stacks[0]))
									Expect(session.Out).To(Say("TIP: use 'cf update-buildpack' to update this buildpack"))
								}, stacks[0])
							})
						})
					})

					Context("when the API doesn't support stack association", func() {
						BeforeEach(func() {
							helpers.SkipIfVersionAtLeast(ccversion.MinVersionBuildpackStackAssociationV3)

							helpers.BuildpackWithStack(func(buildpackPath string) {
								session := helpers.CF("create-buildpack", existingBuildpack, buildpackPath, "5")
								Eventually(session).Should(Exit(0))
							}, "")
						})

						It("prints a warning but doesn't exit 1", func() {
							helpers.BuildpackWithStack(func(buildpackPath string) {
								session := helpers.CF("create-buildpack", "-v", buildpackName, buildpackPath, "1")
								Eventually(session).Should(Exit(0))
								Eventually(session.Err).Should(Say("Buildpack %s already exists", buildpackName))
								Eventually(session.Out).Should(Say("TIP: use 'cf buildpacks' and 'cf delete-buildpack' to delete buildpack %s", buildpackName))
							}, "")
						})
					})
				})
			})

			Context("when specifying an invalid path", func() {
				It("returns the appropriate error", func() {
					session := helpers.CF("create-buildpack", buildpackName, "bogus-path", "1")
					Eventually(session).Should(Exit(1))

					Expect(session.Err).To(Say("Incorrect Usage: The specified path 'bogus-path' does not exist"))
				})
			})
		})

		Context("when uploading from a URL", func() {
			var buildpackURL string

			Context("when specifying a valid URL", func() {
				BeforeEach(func() {
					buildpackURL = "https://github.com/cloudfoundry/binary-buildpack/releases/download/v1.0.21/binary-buildpack-v1.0.21.zip"
				})

				It("successfully uploads a buildpack", func() {
					username, _ := helpers.GetCredentials()
					session := helpers.CF("create-buildpack", buildpackName, buildpackURL, "1")
					Eventually(session).Should(Say("Creating buildpack %s as %s...", buildpackName, username))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Say("Uploading buildpack %s as %s...", buildpackName, username))
					Eventually(session).Should(Say("Done uploading"))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when a 4xx or 5xx HTTP response status is encountered", func() {
				var server *Server

				BeforeEach(func() {
					server = NewServer()
					// Suppresses ginkgo server logs
					server.HTTPTestServer.Config.ErrorLog = log.New(&bytes.Buffer{}, "", 0)
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/"),
							RespondWith(http.StatusNotFound, nil),
						),
					)
				})

				AfterEach(func() {
					server.Close()
				})

				It("displays an appropriate error", func() {
					session := helpers.CF("create-buildpack", buildpackName, server.URL(), "10")
					Eventually(session.Err).Should(Say("Download attempt failed; server returned 404 Not Found"))
					Eventually(session.Err).Should(Say("Unable to install; buildpack is not available from the given URL\\."))
					Eventually(session).Should(Exit(1))
				})
			})

			Context("when specifying an invalid URL", func() {
				BeforeEach(func() {
					buildpackURL = "http://not-a-real-url"
				})

				It("returns the appropriate error", func() {
					session := helpers.CF("create-buildpack", buildpackName, buildpackURL, "1")
					Eventually(session.Err).Should(Say("Get %s: dial tcp: lookup", buildpackURL))
					Eventually(session).Should(Exit(1))
				})
			})
		})

		Context("when specifying the position flag", func() {
			Context("when position is positive integer", func() {
				It("successfully uploads buildpack in correct position", func() {
					helpers.BuildpackWithStack(func(buildpackPath string) {
						session := helpers.CF("create-buildpack", buildpackName, buildpackPath, "3")
						Eventually(session).Should(Exit(0))
					}, "")

					session := helpers.CF("buildpacks")
					Eventually(session).Should(Exit(0))
					Expect(session.Out).To(Say(`%s\s+4`, buildpackName))
				})
			})
		})

		Context("when using the enable/disable flags", func() {
			Context("when specifying disable flag", func() {
				It("disables buildpack", func() {
					helpers.BuildpackWithStack(func(buildpackPath string) {
						session := helpers.CF("create-buildpack", buildpackName, buildpackPath, "1", "--disable")
						Eventually(session).Should(Exit(0))
					}, "")

					session := helpers.CF("buildpacks")
					Eventually(session).Should(Exit(0))
					Expect(session.Out).To(Say(`%s\s+1\s+false`, buildpackName))
				})
			})

			Context("when specifying both enable and disable flags", func() {
				It("returns the appropriate error", func() {
					helpers.BuildpackWithStack(func(buildpackPath string) {
						session := helpers.CF("create-buildpack", buildpackName, buildpackPath, "1", "--enable", "--disable")
						Eventually(session).Should(Exit(1))
						Expect(session.Out).To(Say("FAILED"))
						Expect(session.Err).To(Say("Incorrect Usage: The following arguments cannot be used together: --enable, --disable"))
					}, "")
				})
			})
		})
	})
})
