package global

import (
	"bytes"
	"log"
	"net/http"
	"os"
	"regexp"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("update-buildpack command", func() {
	var (
		buildpackName string
		username      string
	)

	BeforeEach(func() {
		buildpackName = helpers.NewBuildpackName()
		username, _ = helpers.GetCredentials()
	})

	When("--help flag is set", func() {
		It("Displays command usage to output", func() {
			session := helpers.CF("update-buildpack", "--help")
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Say("update-buildpack - Update a buildpack"))
			Eventually(session).Should(Say("USAGE:"))
			Eventually(session).Should(Say(regexp.QuoteMeta(`cf update-buildpack BUILDPACK [-p PATH | -s STACK | --assign-stack NEW_STACK] [-i POSITION] [--rename NEW_NAME] [--enable|--disable] [--lock|--unlock]`)))
			Eventually(session).Should(Say("TIP:"))
			Eventually(session).Should(Say("Path should be a zip file, a url to a zip file, or a local directory. Position is a positive integer, sets priority, and is sorted from lowest to highest.\n\n"))
			Eventually(session).Should(Say("Use '--assign-stack' with caution. Associating a buildpack with a stack that it does not support may result in undefined behavior. Additionally, changing this association once made may require a local copy of the buildpack.\n\n"))
			Eventually(session).Should(Say("OPTIONS:"))
			Eventually(session).Should(Say(`--assign-stack\s+Assign a stack to a buildpack that does not have a stack association`))
			Eventually(session).Should(Say(`--disable\s+Disable the buildpack from being used for staging`))
			Eventually(session).Should(Say(`--enable\s+Enable the buildpack to be used for staging`))
			Eventually(session).Should(Say(`--lock\s+Lock the buildpack to prevent updates`))
			Eventually(session).Should(Say(`--path, -p\s+Path to directory or zip file`))
			Eventually(session).Should(Say(`--position, -i\s+The order in which the buildpacks are checked during buildpack auto-detection`))
			Eventually(session).Should(Say(`--rename\s+Rename an existing buildpack`))
			Eventually(session).Should(Say(`--stack, -s\s+Specify stack to disambiguate buildpacks with the same name`))
			Eventually(session).Should(Say(`--unlock\s+Unlock the buildpack to enable updates`))
			Eventually(session).Should(Say("SEE ALSO:"))
			Eventually(session).Should(Say("buildpacks, create-buildpack, delete-buildpack"))
			Eventually(session).Should(Exit(0))
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(false, false, ReadOnlyOrg, "update-buildpack", "fake-buildpack")
		})
	})

	When("the user is logged in", func() {
		BeforeEach(func() {
			helpers.LoginCF()
		})

		AfterEach(func() {
			helpers.DeleteBuildpackIfOnOldCCAPI(buildpackName)
		})

		When("the buildpack is not provided", func() {
			It("returns a buildpack argument not provided error", func() {
				session := helpers.CF("update-buildpack", "-p", ".")

				Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `BUILDPACK` was not provided"))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the buildpack name is provided", func() {
			When("the buildpack does not exist", func() {
				It("returns a buildpack not found error", func() {
					session := helpers.CF("update-buildpack", buildpackName)
					Eventually(session.Err).Should(Say("Buildpack '%s' not found", buildpackName))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				})
			})

			Describe("stack association", func() {
				var stacks []string

				BeforeEach(func() {
					stacks = helpers.EnsureMinimumNumberOfStacks(2)
				})

				When("multiple buildpacks with the same name exist in enabled and unlocked state, and one has nil stack", func() {
					BeforeEach(func() {
						helpers.BuildpackWithStack(func(buildpackArchive string) {
							createSession := helpers.CF("create-buildpack", buildpackName, buildpackArchive, "99")
							Eventually(createSession).Should(Exit(0))
						}, stacks[0])

						helpers.BuildpackWithoutStack(func(buildpackArchive string) {
							createSession := helpers.CF("create-buildpack", buildpackName, buildpackArchive, "100")
							Eventually(createSession).Should(Exit(0))
						})

						listSession := helpers.CF("buildpacks")
						Eventually(listSession).Should(Say(helpers.BuildpacksOutputRegex(helpers.BuildpackFields{
							Name: buildpackName, Stack: stacks[0]})))
						Eventually(listSession).Should(Say(helpers.BuildpacksOutputRegex(helpers.BuildpackFields{Name: buildpackName})))
						Eventually(listSession).Should(Exit(0))
					})

					When("no stack association is specified", func() {
						It("acts on the buildpack with the nil stack", func() {
							session := helpers.CF("update-buildpack", buildpackName)

							Eventually(session).Should(Say("Updating buildpack %s as %s...", buildpackName, username))
							Eventually(session).Should(Say("OK"))
							Eventually(session).Should(Exit(0))
						})
					})

					When("the user specifies a stack association not matching any of the existing buildpacks with this name", func() {
						It("reports that it couldn't find the buildpack", func() {
							nonexistentStack := "some-incorrect-stack-name"
							session := helpers.CF("update-buildpack", buildpackName, "-s", nonexistentStack)

							Eventually(session.Err).Should(Say("Buildpack '%s' with stack '%s' not found", buildpackName, nonexistentStack))
							Eventually(session).Should(Say("FAILED"))
							Eventually(session).Should(Exit(1))
						})
					})

					When("the user specifies a stack association matching one of the existing buildpacks with this name", func() {
						It("finds the buildpack with the stack specified (and not the buildpack with the nil stack)", func() {
							session := helpers.CF("update-buildpack", buildpackName, "-s", stacks[0])

							Eventually(session).Should(Say("Updating buildpack %s with stack %s as %s...",
								buildpackName, stacks[0], username,
							))
							Eventually(session).Should(Say("OK"))
							Eventually(session).Should(Exit(0))
						})
					})
				})

				When("multiple buildpacks with the same name exist in enabled and unlocked state, and all have stacks", func() {
					BeforeEach(func() {
						helpers.BuildpackWithStack(func(buildpackArchive string) {
							createSession := helpers.CF("create-buildpack", buildpackName, buildpackArchive, "98")
							Eventually(createSession).Should(Exit(0))
						}, stacks[0])

						helpers.BuildpackWithStack(func(buildpackArchive string) {
							createSession := helpers.CF("create-buildpack", buildpackName, buildpackArchive, "99")
							Eventually(createSession).Should(Exit(0))
						}, stacks[1])

						listSession := helpers.CF("buildpacks")
						Eventually(listSession).Should(Say(helpers.BuildpacksOutputRegex(helpers.BuildpackFields{
							Name: buildpackName, Stack: stacks[0]})))
						Eventually(listSession).Should(Say(helpers.BuildpacksOutputRegex(helpers.BuildpackFields{
							Name: buildpackName, Stack: stacks[1]})))
						Eventually(listSession).Should(Exit(0))
					})

					When("no stack association is specified", func() {
						It("displays an error saying that multiple buildpacks were found", func() {
							session := helpers.CF("update-buildpack", buildpackName)

							Eventually(session.Err).Should(Say(`Multiple buildpacks named %s found\. Specify a stack name by using a '-s' flag\.`, buildpackName))
							Eventually(session).Should(Say("FAILED"))
							Eventually(session).Should(Exit(1))
						})
					})

					When("the user specifies a stack association not matching any of the existing buildpacks with this name", func() {
						It("reports that it couldn't find the buildpack", func() {
							nonexistentStack := "some-incorrect-stack-name"
							session := helpers.CF("update-buildpack", buildpackName, "-s", nonexistentStack)

							Eventually(session.Err).Should(Say("Buildpack '%s' with stack '%s' not found", buildpackName, nonexistentStack))
							Eventually(session).Should(Say("FAILED"))
							Eventually(session).Should(Exit(1))
						})
					})

					When("the user specifies a stack association matching one of the existing buildpacks with this name", func() {
						It("finds the buildpack with the stack specified (and not the buildpack with the other stack)", func() {
							session := helpers.CF("update-buildpack", buildpackName, "-s", stacks[0])

							Eventually(session).Should(Say("Updating buildpack %s with stack %s as %s...",
								buildpackName, stacks[0], username,
							))
							Eventually(session).Should(Say("OK"))
							Eventually(session).Should(Exit(0))
						})
					})
				})

				When("one buildpack with the given name exists in enabled and unlocked state with a stack association", func() {
					BeforeEach(func() {
						helpers.BuildpackWithStack(func(buildpackArchive string) {
							createSession := helpers.CF("create-buildpack", buildpackName, buildpackArchive, "99")
							Eventually(createSession).Should(Exit(0))
						}, stacks[0])

						listSession := helpers.CF("buildpacks")
						Eventually(listSession).Should(Say(helpers.BuildpacksOutputRegex(helpers.BuildpackFields{
							Name: buildpackName, Stack: stacks[0]})))
						Eventually(listSession).Should(Exit(0))
					})

					When("no stack association is specified", func() {
						It("updates the only buildpack with that name", func() {
							session := helpers.CF("update-buildpack", buildpackName)

							Eventually(session).Should(Say("Updating buildpack %s as %s...", buildpackName, username))
							Eventually(session).Should(Say("OK"))
							Eventually(session).Should(Exit(0))
						})
					})

					When("the user specifies a stack association matching the buildpack with this name", func() {
						It("finds the buildpack with the stack specified", func() {
							session := helpers.CF("update-buildpack", buildpackName, "-s", stacks[0])

							Eventually(session).Should(Say("Updating buildpack %s with stack %s as %s...",
								buildpackName, stacks[0], username,
							))
							Eventually(session).Should(Say("OK"))
							Eventually(session).Should(Exit(0))
						})
					})
				})
			})

			When("one buildpack with given name exists in enabled and unlocked state with no stack association", func() {
				BeforeEach(func() {
					helpers.BuildpackWithoutStack(func(buildpackArchive string) {
						createSession := helpers.CF("create-buildpack", buildpackName, buildpackArchive, "99")
						Eventually(createSession).Should(Exit(0))
					})

					listSession := helpers.CF("buildpacks")
					Eventually(listSession).Should(Say(helpers.BuildpacksOutputRegex(helpers.BuildpackFields{Name: buildpackName})))
					Eventually(listSession).Should(Exit(0))
				})

				When("only a name is provided", func() {
					It("prints a success message", func() {
						session := helpers.CF("update-buildpack", buildpackName)

						Eventually(session).Should(Say("Updating buildpack %s as %s...", buildpackName, username))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Exit(0))
					})
				})

				When("the -s flag is provided", func() {
					var (
						stackName string
						session   *Session
					)

					JustBeforeEach(func() {
						stackName = "some-stack"
						session = helpers.CF("update-buildpack", buildpackName, "-s", stackName)
					})

					It("returns a buildpack with stack not found error", func() {
						Eventually(session.Err).Should(Say("Buildpack '%s' with stack '%s' not found", buildpackName, stackName))
						Eventually(session).Should(Say("FAILED"))
						Eventually(session).Should(Exit(1))
					})
				})

				When("the -p flag is provided", func() {
					var (
						buildpackPath string
						session       *Session
					)

					JustBeforeEach(func() {
						session = helpers.CF("update-buildpack", buildpackName, "-p", buildpackPath)
					})

					When("the path is local", func() {
						When("the buildpack path exists", func() {
							When("the buildpack path is an empty directory", func() {
								BeforeEach(func() {
									var err error
									buildpackPath, err = os.MkdirTemp("", "create-buildpack-test-")
									Expect(err).ToNot(HaveOccurred())
								})

								AfterEach(func() {
									err := os.RemoveAll(buildpackPath)
									Expect(err).ToNot(HaveOccurred())
								})

								It("prints an error message", func() {
									Eventually(session).Should(Say("Updating buildpack %s as %s...", buildpackName, username))
									Eventually(session.Err).Should(Say("The specified path '%s' cannot be an empty directory.", regexp.QuoteMeta(buildpackPath)))
									Eventually(session).Should(Exit(1))
								})
							})

							When("the path is a buildpack directory", func() {
								BeforeEach(func() {
									var err error
									buildpackPath, err = os.MkdirTemp("", "create-buildpack-test-")
									Expect(err).ToNot(HaveOccurred())
									file, err := os.CreateTemp(buildpackPath, "")
									Expect(err).ToNot(HaveOccurred())
									defer file.Close()
								})

								AfterEach(func() {
									err := os.RemoveAll(buildpackPath)
									Expect(err).ToNot(HaveOccurred())
								})

								It("updates the buildpack with the given bits", func() {
									Eventually(session).Should(Say("Updating buildpack %s as %s...", buildpackName, username))
									Eventually(session).Should(Say("OK"))
									Eventually(session).Should(Exit(0))
								})
							})

							When("uploading from a zip", func() {
								BeforeEach(func() {
									buildpackPath = helpers.MakeBuildpackArchive("")
								})

								AfterEach(func() {
									err := os.Remove(buildpackPath)
									Expect(err).NotTo(HaveOccurred())
								})

								It("updates the buildpack with the given bits", func() {
									Eventually(session).Should(Say("Updating buildpack %s as %s...", buildpackName, username))
									Eventually(session).Should(Say("OK"))
									Eventually(session).Should(Exit(0))
								})
							})
						})

						When("the buildpack path does not exist", func() {
							BeforeEach(func() {
								buildpackPath = "this-is-a-bogus-path"
							})

							It("returns a buildpack does not exist error", func() {
								Eventually(session.Err).Should(Say("Incorrect Usage: The specified path 'this-is-a-bogus-path' does not exist."))
								Eventually(session).Should(Exit(1))
							})
						})
					})

					When("path is a URL", func() {
						When("specifying a valid URL", func() {
							BeforeEach(func() {
								buildpackPath = "https://github.com/cloudfoundry/binary-buildpack/releases/download/v1.0.21/binary-buildpack-v1.0.21.zip"
							})

							It("successfully uploads a buildpack", func() {
								Eventually(session).Should(Say("Updating buildpack %s as %s...", buildpackName, username))
								Eventually(session).Should(Say("OK"))
								Eventually(session).Should(Say("Uploading buildpack %s as %s...", buildpackName, username))
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
								buildpackPath = server.URL()
							})

							AfterEach(func() {
								server.Close()
							})

							It("displays an appropriate error", func() {
								Eventually(session.Err).Should(Say("Download attempt failed; server returned 404 Not Found"))
								Eventually(session.Err).Should(Say(`Unable to install; buildpack is not available from the given URL\.`))
								Eventually(session).Should(Say("FAILED"))
								Eventually(session).Should(Exit(1))
							})
						})

						When("specifying an invalid URL", func() {
							BeforeEach(func() {
								buildpackPath = "http://not-a-real-url"
							})

							It("returns the appropriate error", func() {
								Eventually(session).Should(Exit(1))
								Expect(session.Err).To(Say(`Get "%s": dial tcp: lookup`, buildpackPath))
								Expect(session).To(Say("FAILED"))
							})
						})
					})
				})

				When("the -i flag is provided", func() {
					// TODO: These are tests that require global integration scoping, break these into their own test and isolate the others
					var (
						buildpackPosition string
						session           *Session
					)

					JustBeforeEach(func() {
						session = helpers.CF("update-buildpack", buildpackName, "-i", buildpackPosition)
					})

					When("position is a negative integer", func() {
						BeforeEach(func() {
							buildpackPosition = "-3"
						})

						It("fails with error that position must be at least 1", func() {
							Eventually(session).Should(Say("Updating buildpack %s as %s...", buildpackName, username))
							Eventually(session.Err).Should(Say("Position must be greater than or equal to 1"))
							Eventually(session).Should(Exit(1))
						})
					})

					When("position is positive integer", func() {
						BeforeEach(func() {
							buildpackPosition = "3"
						})

						It("successfully uploads buildpack in the provided position", func() {
							Eventually(session).Should(Exit(0))

							listSession := helpers.CF("buildpacks")
							Eventually(listSession).Should(Say(`\s+3\s+%s`, buildpackName))
							Eventually(listSession).Should(Exit(0))
						})
					})
				})

				When("the --assign-stack flag is provided", func() {
					var (
						stacks []string
					)

					When("the user assigns a stack that exists on the system", func() {
						BeforeEach(func() {
							stacks = helpers.EnsureMinimumNumberOfStacks(2)
						})

						It("successfully assigns the stack to the buildpack", func() {
							session := helpers.CF("update-buildpack", buildpackName, "--assign-stack", stacks[0])

							Eventually(session).Should(Say("Assigning stack %s to %s as %s...", stacks[0], buildpackName, username))
							Eventually(session).Should(Say("OK"))
							Eventually(session).Should(Exit(0))

							listSession := helpers.CF("buildpacks")
							Eventually(listSession).Should(Say(helpers.BuildpacksOutputRegex(helpers.BuildpackFields{
								Name: buildpackName, Stack: stacks[0]})))
							Eventually(listSession).Should(Exit(0))
						})

						When("the buildpack already has a stack associated to it", func() {
							BeforeEach(func() {
								assignStackSession := helpers.CF("update-buildpack", buildpackName, "--assign-stack", stacks[0])
								Eventually(assignStackSession).Should(Exit(0))
							})

							It("displays an error that the buildpack already has a stack association", func() {
								session := helpers.CF("update-buildpack", buildpackName, "--assign-stack", stacks[1])
								Eventually(session.Err).Should(Say("Buildpack stack cannot be changed"))
								Eventually(session).Should(Say("FAILED"))
								Eventually(session).Should(Exit(1))
							})
						})
					})

					When("the user assigns a stack that does NOT exist on the system", func() {
						It("displays an error that the stack isn't found", func() {
							session := helpers.CF("update-buildpack", buildpackName, "--assign-stack", "nonexistent-stack")
							Eventually(session.Err).Should(Say("Stack 'nonexistent-stack' does not exist"))
							Eventually(session).Should(Say("FAILED"))
							Eventually(session).Should(Exit(1))
						})
					})
				})

				When("the --lock is provided", func() {
					It("locks the buildpack", func() {
						session := helpers.CF("update-buildpack", buildpackName, "--lock")
						Eventually(session).Should(Say("Updating buildpack %s as %s...", buildpackName, username))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Exit(0))

						session = helpers.CF("buildpacks")
						Eventually(session).Should(Say(helpers.BuildpacksOutputRegex(helpers.BuildpackFields{
							Name:   buildpackName,
							Locked: "true",
						})))
						Eventually(session).Should(Exit(0))
					})
				})

				When("the --disable is provided", func() {
					It("disables buildpack", func() {
						session := helpers.CF("update-buildpack", buildpackName, "--disable")
						Eventually(session).Should(Say("Updating buildpack %s as %s...", buildpackName, username))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Exit(0))

						session = helpers.CF("buildpacks")
						Eventually(session).Should(Say(`%s\s+false`, buildpackName))
						Eventually(session).Should(Exit(0))
					})
				})

				When("the --rename flag is provided", func() {
					var (
						newBuildpackName string
					)

					BeforeEach(func() {
						newBuildpackName = helpers.NewBuildpackName()
					})

					When("a buildpack with the new name does not already exist", func() {
						It("renames the buildpack", func() {
							session := helpers.CF("update-buildpack", buildpackName, "--rename", newBuildpackName)
							Eventually(session).Should(Say("Renaming buildpack %s to %s as %s...", buildpackName, newBuildpackName, username))
							Eventually(session).Should(Say("OK"))
							Eventually(session).Should(Exit(0))

							session = helpers.CF("buildpacks")
							Eventually(session).Should(Say(`%s`, newBuildpackName))
							Eventually(session).ShouldNot(Say(`%s`, buildpackName))
							Eventually(session).Should(Exit(0))
						})
					})

					When("a buildpack with the new name already exists", func() {
						BeforeEach(func() {
							helpers.BuildpackWithoutStack(func(buildpackArchive string) {
								createSession := helpers.CF("create-buildpack", newBuildpackName, buildpackArchive, "99")
								Eventually(createSession).Should(Exit(0))
							})

							listSession := helpers.CF("buildpacks")
							Eventually(listSession).Should(Say(helpers.BuildpacksOutputRegex(helpers.BuildpackFields{Name: newBuildpackName})))
							Eventually(listSession).Should(Exit(0))
						})

						It("fails to rename the buildpack", func() {
							session := helpers.CF("update-buildpack", buildpackName, "--rename", newBuildpackName)
							Eventually(session.Err).Should(Say("Buildpack with name '%s' and an unassigned stack already exists", newBuildpackName))
							Eventually(session).Should(Say("FAILED"))
							Eventually(session).Should(Exit(1))
						})
					})
				})
			})

			When("the buildpack exists and is disabled", func() {
				BeforeEach(func() {
					helpers.BuildpackWithoutStack(func(buildpackPath string) {
						session := helpers.CF("create-buildpack", buildpackName, buildpackPath, "1", "--disable")
						Eventually(session).Should(Exit(0))
					})
				})

				When("specifying enable flag", func() {
					It("enables buildpack", func() {
						session := helpers.CF("update-buildpack", buildpackName, "--enable")
						Eventually(session).Should(Say("Updating buildpack %s as %s...", buildpackName, username))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Exit(0))

						session = helpers.CF("buildpacks")
						Eventually(session).Should(Say(helpers.BuildpacksOutputRegex(helpers.BuildpackFields{Name: buildpackName})))
						Eventually(session).Should(Exit(0))
					})
				})
			})

			When("the buildpack exists and is locked", func() {
				var buildpackURL string

				BeforeEach(func() {
					helpers.BuildpackWithoutStack(func(buildpackPath string) {
						session := helpers.CF("create-buildpack", buildpackName, buildpackPath, "1")
						Eventually(session).Should(Exit(0))
						session = helpers.CF("update-buildpack", buildpackName, "--lock")
						Eventually(session).Should(Exit(0))
					})
					buildpackURL = "https://github.com/cloudfoundry/binary-buildpack/releases/download/v1.0.21/binary-buildpack-v1.0.21.zip"
				})

				Context("specifying -p argument", func() {
					It("fails to update buildpack", func() {
						session := helpers.CF("update-buildpack", buildpackName, "-p", buildpackURL)
						Eventually(session).Should(Say("Updating buildpack %s as %s...", buildpackName, username))
						Eventually(session).Should(Say("FAILED"))
						Eventually(session.Err).Should(Say("Buildpack is locked"))
						Eventually(session).Should(Exit(1))
					})
				})

				Context("specifying unlock flag", func() {
					It("unlocks the buildpack", func() {
						session := helpers.CF("update-buildpack", buildpackName, "--unlock")
						Eventually(session).Should(Say("Updating buildpack %s as %s...", buildpackName, username))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Exit(0))

						session = helpers.CF("buildpacks")
						Eventually(session).Should(Say(helpers.BuildpacksOutputRegex(helpers.BuildpackFields{Name: buildpackName})))
						Eventually(session).Should(Exit(0))
					})
				})
			})
		})
	})
})
