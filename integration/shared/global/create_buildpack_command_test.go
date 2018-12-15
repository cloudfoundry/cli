package global

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("create buildpack command", func() {
	var (
		buildpackName string
		username      string
	)

	BeforeEach(func() {
		buildpackName = helpers.NewBuildpackName()
	})

	Describe("help", func() {
		When("--help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("create-buildpack", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("create-buildpack - Create a buildpack"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say(`cf create-buildpack BUILDPACK PATH POSITION \[--enable|--disable\]`))
				Eventually(session).Should(Say("TIP:"))
				Eventually(session).Should(Say("Path should be a zip file, a url to a zip file, or a local directory. Position is a positive integer, sets priority, and is sorted from lowest to highest."))
				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say(`--disable\s+Disable the buildpack from being used for staging`))
				Eventually(session).Should(Say(`--enable\s+Enable the buildpack to be used for staging`))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("buildpacks, push"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			path, err := os.Getwd()
			Expect(err).ToNot(HaveOccurred())

			helpers.CheckEnvironmentTargetedCorrectly(false, false, ReadOnlyOrg, "create-buildpack", "fake-buildpack", path, "1")
		})
	})

	When("the user is logged in", func() {
		BeforeEach(func() {
			helpers.LoginCF()
			username, _ = helpers.GetCredentials()
		})

		When("uploading from a directory", func() {
			var buildpackDir string

			AfterEach(func() {
				err := os.RemoveAll(buildpackDir)
				Expect(err).ToNot(HaveOccurred())
			})

			When("zipping the directory errors", func() {
				BeforeEach(func() {
					buildpackDir = "some/nonexistent/dir"
				})

				It("returns an error", func() {
					session := helpers.CF("create-buildpack", buildpackName, buildpackDir, "1")
					Eventually(session.Err).Should(Say("Incorrect Usage: The specified path 'some/nonexistent/dir' does not exist."))
					Eventually(session).Should(Say("USAGE:"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("zipping the directory succeeds", func() {
				BeforeEach(func() {
					var err error
					buildpackDir, err = ioutil.TempDir("", "buildpackdir-")
					Expect(err).ToNot(HaveOccurred())
					file, err := ioutil.TempFile(buildpackDir, "myfile-")
					defer file.Close()
					Expect(err).ToNot(HaveOccurred())
				})

				It("successfully uploads a buildpack", func() {
					session := helpers.CF("create-buildpack", buildpackName, buildpackDir, "1")
					Eventually(session).Should(Say(`Creating buildpack %s as %s\.\.\.`, buildpackName, username))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Say(`Uploading buildpack %s as %s\.\.\.`, buildpackName, username))
					Eventually(session).Should(Say("Done uploading"))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))
				})
			})

			When("the specified directory is empty", func() {
				BeforeEach(func() {
					var err error
					buildpackDir, err = ioutil.TempDir("", "empty-")
					Expect(err).ToNot(HaveOccurred())
				})

				It("fails and reports that the directory is empty", func() {
					session := helpers.CF("create-buildpack", buildpackName, buildpackDir, "1")
					Eventually(session).Should(Say(`Creating buildpack %s as %s\.\.\.`, buildpackName, username))
					Eventually(session.Err).Should(Say("The specified path '%s' cannot be an empty directory.", regexp.QuoteMeta(buildpackDir)))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				})
			})
		})

		When("uploading from a zip", func() {
			var stacks []string

			BeforeEach(func() {
				stacks = helpers.EnsureMinimumNumberOfStacks(2)
			})

			When("specifying a valid path", func() {
				When("the new buildpack is unique", func() {
					When("the new buildpack has a nil stack", func() {
						It("successfully uploads a buildpack", func() {
							helpers.BuildpackWithoutStack(func(buildpackPath string) {
								session := helpers.CF("create-buildpack", buildpackName, buildpackPath, "1")
								Eventually(session).Should(Say(`Creating buildpack %s as %s\.\.\.`, buildpackName, username))
								Eventually(session).Should(Say("OK"))
								Eventually(session).Should(Say(`Uploading buildpack %s as %s\.\.\.`, buildpackName, username))
								Eventually(session).Should(Say("Done uploading"))
								Eventually(session).Should(Say("OK"))
								Eventually(session).Should(Exit(0))
							})

							session := helpers.CF("buildpacks")
							Eventually(session).Should(Say(helpers.BuildpacksOutputRegex(helpers.BuildpackFields{
								Name: buildpackName, Position: "1"})))
							Eventually(session).Should(Exit(0))
						})
					})

					When("the new buildpack has a valid stack", func() {
						BeforeEach(func() {
							helpers.SkipIfVersionLessThan(ccversion.MinVersionBuildpackStackAssociationV2)
						})

						It("successfully uploads a buildpack", func() {
							helpers.BuildpackWithStack(func(buildpackPath string) {
								session := helpers.CF("create-buildpack", buildpackName, buildpackPath, "1")
								Eventually(session).Should(Say(`Creating buildpack %s as %s\.\.\.`, buildpackName, username))
								Eventually(session).Should(Say("OK"))
								Eventually(session).Should(Say(`Uploading buildpack %s as %s\.\.\.`, buildpackName, username))
								Eventually(session).Should(Say("Done uploading"))
								Eventually(session).Should(Say("OK"))
								Eventually(session).Should(Exit(0))
							}, stacks[0])

							session := helpers.CF("buildpacks")
							Eventually(session).Should(Say(helpers.BuildpacksOutputRegex(helpers.BuildpackFields{
								Name: buildpackName, Stack: stacks[0], Position: "1",
							})))
							Eventually(session).Should(Exit(0))
						})
					})
				})

				When("the new buildpack has an invalid stack", func() {
					BeforeEach(func() {
						helpers.SkipIfVersionLessThan(ccversion.MinVersionBuildpackStackAssociationV2)
					})

					It("returns the appropriate error", func() {
						helpers.BuildpackWithStack(func(buildpackPath string) {
							session := helpers.CF("create-buildpack", buildpackName, buildpackPath, "1")
							Eventually(session).Should(Say(`Creating buildpack %s as %s\.\.\.`, buildpackName, username))
							Eventually(session).Should(Say("OK"))
							Eventually(session).Should(Say(`Uploading buildpack %s as %s\.\.\.`, buildpackName, username))
							Eventually(session.Err).Should(Say(`Uploaded buildpack stack \(fake-stack\) does not exist`))
							Eventually(session).Should(Exit(1))
						}, "fake-stack")
					})
				})

				When("a buildpack with the same name exists", func() {
					var (
						existingBuildpack string
					)

					BeforeEach(func() {
						existingBuildpack = buildpackName
					})

					When("the new buildpack has a nil stack", func() {
						BeforeEach(func() {
							helpers.SkipIfVersionLessThan(ccversion.MinVersionBuildpackStackAssociationV2)
						})

						When("the existing buildpack does not have a nil stack", func() {
							BeforeEach(func() {
								helpers.BuildpackWithStack(func(buildpackPath string) {
									session := helpers.CF("create-buildpack", existingBuildpack, buildpackPath, "5")
									Eventually(session).Should(Exit(0))
								}, stacks[0])
							})

							It("successfully uploads a buildpack", func() {
								helpers.BuildpackWithStack(func(buildpackPath string) {
									session := helpers.CF("create-buildpack", buildpackName, buildpackPath, "1")
									Eventually(session).Should(Say(`Creating buildpack %s as %s\.\.\.`, buildpackName, username))
									Eventually(session).Should(Say("OK"))
									Eventually(session).Should(Say(`Uploading buildpack %s as %s\.\.\.`, buildpackName, username))
									Eventually(session).Should(Exit(0))
								}, stacks[0])

								session := helpers.CF("buildpacks")
								Eventually(session).Should(Say(helpers.BuildpacksOutputRegex(helpers.BuildpackFields{
									Name: buildpackName, Position: "1"})))
								Eventually(session).Should(Say(helpers.BuildpacksOutputRegex(helpers.BuildpackFields{
									Name: existingBuildpack, Stack: stacks[0], Position: "6"})))
								Eventually(session).Should(Exit(0))
							})
						})

						When("the existing buildpack has a nil stack", func() {
							BeforeEach(func() {
								helpers.BuildpackWithoutStack(func(buildpackPath string) {
									session := helpers.CF("create-buildpack", existingBuildpack, buildpackPath, "5")
									Eventually(session).Should(Exit(0))
								})
							})

							It("prints a warning but exits 0", func() {
								helpers.BuildpackWithoutStack(func(buildpackPath string) {
									session := helpers.CF("create-buildpack", buildpackName, buildpackPath, "1")
									Eventually(session).Should(Say(`Creating buildpack %s as %s\.\.\.`, buildpackName, username))
									Eventually(session.Err).Should(Say("Buildpack %s already exists without a stack", buildpackName))
									Eventually(session).Should(Exit(0))
								})

								session := helpers.CF("buildpacks")
								Eventually(session).Should(Say(`%s\s+5`, existingBuildpack))
								Eventually(session).Should(Exit(0))
							})
						})
					})

					When("the new buildpack has a non-nil stack", func() {
						BeforeEach(func() {
							helpers.SkipIfVersionLessThan(ccversion.MinVersionBuildpackStackAssociationV2)
						})

						When("the existing buildpack has a different non-nil stack", func() {
							BeforeEach(func() {
								helpers.BuildpackWithStack(func(buildpackPath string) {
									session := helpers.CF("create-buildpack", existingBuildpack, buildpackPath, "5")
									Eventually(session).Should(Exit(0))
								}, stacks[1])
							})

							It("successfully uploads a buildpack", func() {
								helpers.BuildpackWithStack(func(buildpackPath string) {
									session := helpers.CF("create-buildpack", buildpackName, buildpackPath, "1")
									Eventually(session).Should(Say(`Creating buildpack %s as %s\.\.\.`, buildpackName, username))
									Eventually(session).Should(Say("OK"))
									Eventually(session).Should(Say(`Uploading buildpack %s as %s\.\.\.`, buildpackName, username))
									Eventually(session).Should(Say("Done uploading"))
									Eventually(session).Should(Say("OK"))
									Eventually(session).Should(Exit(0))
								}, stacks[0])

								session := helpers.CF("buildpacks")
								Eventually(session).Should(Say(helpers.BuildpacksOutputRegex(helpers.BuildpackFields{
									Name: buildpackName, Stack: stacks[0]})))
								Eventually(session).Should(Say(helpers.BuildpacksOutputRegex(helpers.BuildpackFields{
									Name: existingBuildpack, Stack: stacks[1]})))
								Eventually(session).Should(Exit(0))
							})
						})

						When("the existing buildpack has a nil stack", func() {
							BeforeEach(func() {
								helpers.BuildpackWithoutStack(func(buildpackPath string) {
									session := helpers.CF("create-buildpack", existingBuildpack, buildpackPath, "5")
									Eventually(session).Should(Exit(0))
								})
							})

							It("prints a warning and tip but exits 0", func() {
								helpers.BuildpackWithStack(func(buildpackPath string) {
									session := helpers.CF("create-buildpack", buildpackName, buildpackPath, "1")
									Eventually(session).Should(Say(`Creating buildpack %s as %s\.\.\.`, buildpackName, username))
									Eventually(session.Err).Should(Say("Buildpack %s already exists without a stack", buildpackName))
									Eventually(session).Should(Say("TIP: use 'cf buildpacks' and 'cf delete-buildpack' to delete buildpack %s without a stack", buildpackName))
									Eventually(session).Should(Exit(0))
								}, stacks[0])
							})
						})

						When("the existing buildpack has the same non-nil stack", func() {
							BeforeEach(func() {
								helpers.BuildpackWithStack(func(buildpackPath string) {
									session := helpers.CF("create-buildpack", existingBuildpack, buildpackPath, "5")
									Eventually(session).Should(Exit(0))
								}, stacks[0])
							})

							It("prints a warning but doesn't exit 1", func() {
								helpers.BuildpackWithStack(func(buildpackPath string) {
									session := helpers.CF("create-buildpack", buildpackName, buildpackPath, "1")
									Eventually(session).Should(Say(`Creating buildpack %s as %s\.\.\.`, buildpackName, username))
									Eventually(session.Err).Should(Say("The buildpack name %s is already in use for the stack %s", buildpackName, stacks[0]))
									Eventually(session).Should(Say("TIP: use 'cf update-buildpack' to update this buildpack"))
									Eventually(session).Should(Exit(0))
								}, stacks[0])
							})
						})
					})

					When("the API doesn't support stack association", func() {
						BeforeEach(func() {
							helpers.SkipIfVersionAtLeast(ccversion.MinVersionBuildpackStackAssociationV2)

							helpers.BuildpackWithoutStack(func(buildpackPath string) {
								session := helpers.CF("create-buildpack", existingBuildpack, buildpackPath, "5")
								Eventually(session).Should(Exit(0))
							})
						})

						It("prints a warning but doesn't exit 1", func() {
							helpers.BuildpackWithoutStack(func(buildpackPath string) {
								session := helpers.CF("create-buildpack", buildpackName, buildpackPath, "1")
								Eventually(session.Err).Should(Say("Buildpack %s already exists", buildpackName))
								Eventually(session).Should(Say("TIP: use 'cf update-buildpack' to update this buildpack"))
								Eventually(session).Should(Exit(0))
							})
						})
					})
				})
			})

			When("specifying an invalid path", func() {
				It("returns the appropriate error", func() {
					session := helpers.CF("create-buildpack", buildpackName, "bogus-path", "1")

					Eventually(session.Err).Should(Say("Incorrect Usage: The specified path 'bogus-path' does not exist"))
					Eventually(session).Should(Say("USAGE:"))
					Eventually(session).Should(Exit(1))
				})
			})
		})

		When("uploading from a URL", func() {
			var buildpackURL string

			When("specifying a valid URL", func() {
				BeforeEach(func() {
					buildpackURL = "https://github.com/cloudfoundry/binary-buildpack/releases/download/v1.0.21/binary-buildpack-v1.0.21.zip"
				})

				It("successfully uploads a buildpack", func() {
					session := helpers.CF("create-buildpack", buildpackName, buildpackURL, "1")
					Eventually(session).Should(Say(`Creating buildpack %s as %s\.\.\.`, buildpackName, username))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Say(`Uploading buildpack %s as %s\.\.\.`, buildpackName, username))
					Eventually(session).Should(Say("Done uploading"))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))
				})
			})

			When("a 4xx or 5xx HTTP response status is encountered", func() {
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
					Eventually(session).Should(Say(`Creating buildpack %s as %s\.\.\.`, buildpackName, username))
					Eventually(session.Err).Should(Say("Download attempt failed; server returned 404 Not Found"))
					Eventually(session.Err).Should(Say(`Unable to install; buildpack is not available from the given URL\.`))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("specifying an invalid URL", func() {
				BeforeEach(func() {
					buildpackURL = "http://not-a-real-url"
				})

				It("returns the appropriate error", func() {
					session := helpers.CF("create-buildpack", buildpackName, buildpackURL, "1")
					Eventually(session).Should(Say(`Creating buildpack %s as %s\.\.\.`, buildpackName, username))
					Eventually(session.Err).Should(Say("Get %s: dial tcp: lookup", buildpackURL))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				})
			})
		})

		When("specifying the position flag", func() {
			When("position is positive integer", func() {
				It("successfully uploads buildpack in correct position", func() {
					helpers.BuildpackWithoutStack(func(buildpackPath string) {
						session := helpers.CF("create-buildpack", buildpackName, buildpackPath, "3")
						Eventually(session).Should(Say(`Creating buildpack %s as %s\.\.\.`, buildpackName, username))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Say(`Uploading buildpack %s as %s\.\.\.`, buildpackName, username))
						Eventually(session).Should(Say("Done uploading"))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Exit(0))
					})

					session := helpers.CF("buildpacks")
					Eventually(session).Should(Say(`%s\s+3`, buildpackName))
					Eventually(session).Should(Exit(0))
				})
			})
		})

		When("using the enable/disable flags", func() {
			When("specifying disable flag", func() {
				It("disables buildpack", func() {
					helpers.BuildpackWithoutStack(func(buildpackPath string) {
						session := helpers.CF("create-buildpack", buildpackName, buildpackPath, "1", "--disable")
						Eventually(session).Should(Say(`Creating buildpack %s as %s\.\.\.`, buildpackName, username))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Say(`Uploading buildpack %s as %s\.\.\.`, buildpackName, username))
						Eventually(session).Should(Say("Done uploading"))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Exit(0))
					})

					session := helpers.CF("buildpacks")
					Eventually(session).Should(Say(`%s\s+1\s+false`, buildpackName))
					Eventually(session).Should(Exit(0))
				})
			})

			When("specifying both enable and disable flags", func() {
				It("returns the appropriate error", func() {
					helpers.BuildpackWithoutStack(func(buildpackPath string) {
						session := helpers.CF("create-buildpack", buildpackName, buildpackPath, "1", "--enable", "--disable")
						Eventually(session).Should(Say("FAILED"))
						Eventually(session.Err).Should(Say("Incorrect Usage: The following arguments cannot be used together: --enable, --disable"))
						Eventually(session).Should(Exit(1))
					})
				})
			})
		})
	})
})
