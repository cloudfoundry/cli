package global

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	. "github.com/onsi/gomega/ghttp"
)

var _ = PDescribe("update-buildpack command", func() {
	var (
		buildpackName string
		username      string
		// stacks        []string
	)

	BeforeEach(func() {
		buildpackName = helpers.NewBuildpack()
		username, _ = helpers.GetCredentials()
	})

	When("--help flag is set", func() {
		It("Displays command usage to output", func() {
			session := helpers.CF("update-buildpack", "--help")
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Say("update-buildpack - Update a buildpack"))
			Eventually(session).Should(Say("USAGE:"))
			Eventually(session).Should(Say("cf update-buildpack BUILDPACK [-p PATH] [-i POSITION] [-s STACK] [--enable|--disable] [--lock|--unlock]"))
			Eventually(session).Should(Say("TIP:"))
			Eventually(session).Should(Say("Path should be a zip file, a url to a zip file, or a local directory. Position is a positive integer, sets priority, and is sorted from lowest to highest."))
			Eventually(session).Should(Say("OPTIONS:"))
			Eventually(session).Should(Say("--disable\\s+Disable the buildpack from being used for staging"))
			Eventually(session).Should(Say("--enable\\s+Enable the buildpack to be used for staging"))
			Eventually(session).Should(Say("-i\\s+The order in which the buildpacks are checked during buildpack auto-detection"))
			Eventually(session).Should(Say("--lock\\s+Lock the buildpack to prevent updates"))
			Eventually(session).Should(Say("-p\\s+Path to directory or zip file"))
			Eventually(session).Should(Say("--unlock\\s+Unlock the buildpack to enable updates"))
			Eventually(session).Should(Say("-s\\s+Specify stack to disambiguate buildpacks with the same name"))
			Eventually(session).Should(Say("SEE ALSO:"))
			Eventually(session).Should(Say("buildpacks, rename-buildpack"))
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

		When("the buildpack is not provided", func() {
			It("returns a buildpack argument not provided error", func() {
				session := helpers.CF("update-buildpack", "-p", ".")

				Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `BUILDPACK` was not provided"))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the buildpack name is provided", func() {
			When("the buildpack exists", func() {
				Context("regardless of the buildpack being enabled or locked", func() {
					BeforeEach(func() {
						helpers.BuildpackWithStack(func(buildpackPath string) {
							session := helpers.CF("create-buildpack", buildpackName, buildpackPath, "99")
							Eventually(session).Should(Exit(0))
						}, "")
					})

					When("the -p flag is provided", func() {
						When("the path is local", func() {
							When("the buildpack path exists", func() {
								When("uploading from a directory", func() {
									var buildpackDir string

									BeforeEach(func() {
										var err error
										buildpackDir, err = ioutil.TempDir("", "create-buildpack-test-")
										Expect(err).ToNot(HaveOccurred())
										file, err := ioutil.TempFile(buildpackDir, "")
										defer file.Close()
										Expect(err).ToNot(HaveOccurred())
									})

									AfterEach(func() {
										err := os.RemoveAll(buildpackDir)
										Expect(err).ToNot(HaveOccurred())
									})

									It("updates the buildpack with the given bits", func() {
										session := helpers.CF("update-buildpack", buildpackName, "-p", buildpackDir)
										Eventually(session).Should(Say("Updating buildpack %s...", buildpackName))
										Eventually(session).Should(Say("OK"))
										Eventually(session).Should(Exit(0))
									})
								})

								When("uploading from a zip", func() {
									var filename string

									BeforeEach(func() {
										file, err := ioutil.TempFile("", "create-buildpack-test-")
										defer file.Close()
										Expect(err).ToNot(HaveOccurred())
										filename = file.Name() + ".tgz"
										err = os.Rename(file.Name(), filename)
										Expect(err).ToNot(HaveOccurred())
									})

									AfterEach(func() {
										err := os.Remove(filename)
										Expect(err).NotTo(HaveOccurred())
									})

									It("updates the buildpack with the given bits", func() {
										helpers.BuildpackWithStack(func(buildpackPath string) {
											session := helpers.CF("update-buildpack", buildpackName, buildpackPath, "1")
											Eventually(session).Should(Exit(0))
										}, "")

										session := helpers.CF("buildpacks")
										Eventually(session).Should(Say(`%s\s+1`, buildpackName))
										Eventually(session).Should(Exit(0))
									})
								})
							})

							When("the buildpack path does not exist", func() {
								It("returns a buildpack does not exist error", func() {
									session := helpers.CF("update-buildpack", buildpackName, "-p", "this-is-a-bogus-path")

									Eventually(session.Err).Should(Say("Incorrect Usage: The specified path 'this-is-a-bogus-path' does not exist."))
									Eventually(session).Should(Exit(1))
								})
							})
						})

						When("path is a URL", func() {
							var buildpackURL string

							When("specifying a valid URL", func() {
								BeforeEach(func() {
									buildpackURL = "https://github.com/cloudfoundry/binary-buildpack/releases/download/v1.0.21/binary-buildpack-v1.0.21.zip"
								})

								It("successfully uploads a buildpack", func() {
									session := helpers.CF("update-buildpack", buildpackName, buildpackURL, "1")
									Eventually(session).Should(Say("Creating buildpack %s as %s...", buildpackName, username))
									Eventually(session).Should(Say("OK"))
									Eventually(session).Should(Say("Uploading buildpack %s as %s...", buildpackName, username))
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
									session := helpers.CF("update-buildpack", buildpackName, server.URL(), "10")
									Eventually(session.Err).Should(Say("Download attempt failed; server returned 404 Not Found"))
									Eventually(session.Err).Should(Say("Unable to install; buildpack is not available from the given URL\\."))
									Eventually(session).Should(Say("FAILED"))
									Eventually(session).Should(Exit(1))
								})
							})

							When("specifying an invalid URL", func() {
								BeforeEach(func() {
									buildpackURL = "http://not-a-real-url"
								})

								It("returns the appropriate error", func() {
									session := helpers.CF("update-buildpack", buildpackName, "-p", "https://example.com/bogus.tgz")
									Eventually(session.Err).Should(Say("Failed to create a local temporary zip file for the buildpack"))
									Eventually(session.Err).Should(Say("Couldn't write zip file: zip: not a valid zip file"))
									Eventually(session).Should(Say("FAILED"))
									Eventually(session).Should(Exit(1))
								})
							})
						})
					})

					When("the -i flag is provided", func() {
						When("position is a negative integer", func() {
							It("successfully uploads buildpack as the first position", func() {
								session := helpers.CF("update-buildpack", buildpackName, "-i", "-3")
								Eventually(session).Should(Say("Updating buildpack %s as %s...", buildpackName, username))
								Eventually(session).Should(Say("OK", buildpackName, username))
								Eventually(session).Should(Exit(0))

								session = helpers.CF("buildpacks")
								Eventually(session).Should(Say(`%s\s+1`, buildpackName))
								Eventually(session).Should(Exit(0))
							})
						})

						When("position is positive integer", func() {
							It("successfully uploads buildpack in the provided position", func() {
								session := helpers.CF("update-buildpack", buildpackName, "-i", "3")
								Eventually(session).Should(Exit(0))

								session = helpers.CF("buildpacks")
								Eventually(session).Should(Say(`%s\s+3`, buildpackName))
								Eventually(session).Should(Exit(0))
							})
						})
					})

					When("specifying both enable and disable flags", func() {
						It("returns the appropriate error", func() {
							session := helpers.CF("update-buildpack", buildpackName, "--enable", "--disable")
							Eventually(session.Err).Should(Say("Incorrect Usage: The following arguments cannot be used together: --enable, --disable"))
							Eventually(session).Should(Say("FAILED"))
							Eventually(session).Should(Exit(1))
						})
					})

					When("specifying both lock and unlock flags", func() {
						It("returns the appropriate error", func() {
							session := helpers.CF("update-buildpack", buildpackName, "--lock", "--unlock")
							Eventually(session.Err).Should(Say("Incorrect Usage: The following arguments cannot be used together: --lock, --unlock"))
							Eventually(session).Should(Say("FAILED"))
							Eventually(session).Should(Exit(1))
						})
					})

					When("specifying both lock then unlock flags", func() {
						It("locks and then unlocks the buildpack", func() {
							session := helpers.CF("update-buildpack", buildpackName, "--lock")
							Eventually(session).Should(Say("Updating buildpack %s as %s...", buildpackName, username))
							Eventually(session).Should(Say("OK", buildpackName, username))
							Eventually(session).Should(Exit(0))

							session = helpers.CF("buildpacks")
							Eventually(session).Should(Say(`%s\s+\d+\s+true\s+true`, buildpackName))
							Eventually(session).Should(Exit(0))

							session = helpers.CF("update-buildpack", buildpackName, "--unlock")
							Eventually(session).Should(Say("Updating buildpack %s as %s...", buildpackName, username))
							Eventually(session).Should(Say("OK", buildpackName, username))
							Eventually(session).Should(Exit(0))

							session = helpers.CF("buildpacks")
							Eventually(session).Should(Say(`%s\s+\d+\s+true\s+false`, buildpackName))
							Eventually(session).Should(Exit(0))
						})
					})
				})

				When("the existing buildpack is disabled", func() {
					BeforeEach(func() {
						helpers.BuildpackWithStack(func(buildpackPath string) {
							session := helpers.CF("create-buildpack", buildpackName, buildpackPath, "1", "--disable")
							Eventually(session).Should(Exit(0))
						}, "")
					})

					When("specifying enable flag", func() {
						It("enables buildpack", func() {
							session := helpers.CF("update-buildpack", buildpackName, "--enable")
							Eventually(session).Should(Say("Updating buildpack %s as %s...", buildpackName, username))
							Eventually(session).Should(Say("OK", buildpackName, username))
							Eventually(session).Should(Exit(0))

							session = helpers.CF("buildpacks")
							Eventually(session).Should(Say(`%s\s+1\s+true`, buildpackName))
							Eventually(session).Should(Exit(0))
						})
					})
				})

				When("the existing buildpack is enabled", func() {
					BeforeEach(func() {
						helpers.BuildpackWithStack(func(buildpackPath string) {
							session := helpers.CF("create-buildpack", buildpackName, buildpackPath, "1")
							Eventually(session).Should(Exit(0))
						}, "")
					})

					When("specifying disable flag", func() {
						It("disables buildpack", func() {
							helpers.BuildpackWithStack(func(buildpackPath string) {
								session := helpers.CF("update-buildpack", buildpackName, buildpackPath, "1", "--disable")
								Eventually(session).Should(Say("Updating buildpack %s as %s...", buildpackName, username))
								Eventually(session).Should(Say("OK", buildpackName, username))
								Eventually(session).Should(Exit(0))
							}, "")

							session := helpers.CF("buildpacks")
							Eventually(session).Should(Say(`%s\s+1\s+false`, buildpackName))
							Eventually(session).Should(Exit(0))
						})
					})

				})

				When("the existing buildpack is locked", func() {
					var buildpackURL string

					BeforeEach(func() {
						helpers.BuildpackWithStack(func(buildpackPath string) {
							session := helpers.CF("create-buildpack", buildpackName, buildpackPath, "1")
							Eventually(session).Should(Exit(0))
							session = helpers.CF("update-buidpack", buildpackName, "--lock")
							Eventually(session).Should(Exit(0))
						}, "")
						buildpackURL = "https://github.com/cloudfoundry/binary-buildpack/releases/download/v1.0.21/binary-buildpack-v1.0.21.zip"
					})

					Context("the -p is provided", func() {
						It("enables buildpack", func() {
							session := helpers.CF("update-buildpack", buildpackName, "-p", buildpackURL)
							Eventually(session).Should(Say("Updating buildpack %s as %s...", buildpackName, username))
							Eventually(session).Should(Say("FAILED"))
							Eventually(session.Err).Should(Say("The buildpack is locked"))
							Eventually(session).Should(Exit(0))
						})
					})
				})

				When("multiple buildpacks with the same name exist", func() {
					// var existingBuildpack string

					// BeforeEach(func() {
					// 	helpers.SkipIfVersionLessThan(ccversion.MinVersionBuildpackStackAssociationV3)
					// 	helpers.SkipIfOneStack()

					// 	stacks = helpers.FetchStacks()
					// 	helpers.BuildpackWithStack(func(buildpackPath string) {
					// 		session := helpers.CF("create-buildpack", buildpackName, buildpackPath, "1")
					// 		Eventually(session).Should(Exit(0))
					// 	}, stacks[0])

					// 	helpers.BuildpackWithStack(func(buildpackPath string) {
					// 		session := helpers.CF("create-buildpack", buildpackName, buildpackPath, "1")
					// 		Eventually(session).Should(Exit(0))
					// 	}, stacks[1])
					// 	existingBuildpack = buildpackName
					// })

					// It("fails when no stack is specified", func() {
					// 	session := helpers.CF("update-buildpack", buildpackName, "-i", "999")
					// 	Eventually(session).Should(Exit(1))
					// 	Eventually(session).Should(Say("FAILED"))
					// })

					// It("fails when a nonexistent stack is specified", func() {
					// 	session := helpers.CF("update-buildpack", buildpackName, "-i", "999", "-s", "bogus-stack")
					// 	Eventually(session).Should(Exit(1))
					// 	Eventually(session).Should(Say("FAILED"))
					// })

					// It("succeeds when a stack associated with that buildpack name is specified", func() {
					// 	session := helpers.CF("update-buildpack", buildpackName, "-s", stacks[0], "-i", "999")
					// 	Consistently(session.Err).ShouldNot(Say("Incorrect Usage:"))
					// 	Eventually(session).Should(Say("OK"))
					// 	Eventually(session).Should(Exit(0))
					// })

					// When("the new buildpack has a nil stack", func() {
					// 	When("the existing buildpack does not have a nil stack", func() {
					// 		BeforeEach(func() {
					// 			helpers.BuildpackWithStack(func(buildpackPath string) {
					// 				session := helpers.CF("update-buildpack", existingBuildpack, buildpackPath, "5")
					// 				Eventually(session).Should(Exit(0))
					// 			}, stacks[0])
					// 		})

					// 		It("successfully uploads a buildpack", func() {
					// 			helpers.BuildpackWithStack(func(buildpackPath string) {
					// 				session := helpers.CF("update-buildpack", buildpackName, buildpackPath, "1")
					// 				Eventually(session).Should(Exit(0))
					// 			}, stacks[0])

					// 			session := helpers.CF("buildpacks")
					// 			Eventually(session).Should(Exit(0))
					// 			Expect(session).To(Say(`%s\s+1`, buildpackName))
					// 			Expect(session).To(Say(`%s\s+%s\s+6`, existingBuildpack, stacks[0]))
					// 		})
					// 	})

					// 	When("the existing buildpack has a nil stack", func() {
					// 		BeforeEach(func() {
					// 			helpers.BuildpackWithStack(func(buildpackPath string) {
					// 				session := helpers.CF("update-buildpack", existingBuildpack, buildpackPath, "5")
					// 				Eventually(session).Should(Exit(0))
					// 			}, "")
					// 		})

					// 		It("prints a warning", func() {
					// 			helpers.BuildpackWithStack(func(buildpackPath string) {
					// 				session := helpers.CF("update-buildpack", buildpackName, buildpackPath, "1")
					// 				Eventually(session).Should(Exit(0))
					// 				Eventually(session.Err).Should(Say("Buildpack %s already exists without a stack", buildpackName))
					// 			}, "")

					// 			session := helpers.CF("buildpacks")
					// 			Eventually(session).Should(Exit(0))
					// 			Expect(session).To(Say(`%s\s+5`, existingBuildpack))
					// 		})
					// 	})
					// })

					// When("the new buildpack has a non-nil stack", func() {
					// 	BeforeEach(func() {
					// 		helpers.SkipIfVersionLessThan(ccversion.MinVersionBuildpackStackAssociationV3)
					// 	})

					// 	When("the existing buildpack has a different non-nil stack", func() {
					// 		BeforeEach(func() {
					// 			helpers.SkipIfOneStack()
					// 			helpers.BuildpackWithStack(func(buildpackPath string) {
					// 				session := helpers.CF("update-buildpack", existingBuildpack, buildpackPath, "5")
					// 				Eventually(session).Should(Exit(0))
					// 			}, stacks[1])
					// 		})

					// 		It("successfully uploads a buildpack", func() {
					// 			helpers.BuildpackWithStack(func(buildpackPath string) {
					// 				session := helpers.CF("update-buildpack", buildpackName, buildpackPath, "1")
					// 				Eventually(session).Should(Exit(0))
					// 			}, stacks[0])

					// 			session := helpers.CF("buildpacks")
					// 			Eventually(session).Should(Exit(0))
					// 			Expect(session).To(Say(`%s\s+%s\s+1`, buildpackName, stacks[0]))
					// 			Expect(session).To(Say(`%s\s+%s\s+6`, existingBuildpack, stacks[1]))
					// 		})
					// 	})

					// 	When("the existing buildpack has a nil stack", func() {
					// 		BeforeEach(func() {
					// 			helpers.BuildpackWithStack(func(buildpackPath string) {
					// 				session := helpers.CF("update-buildpack", existingBuildpack, buildpackPath, "5")
					// 				Eventually(session).Should(Exit(0))
					// 			}, "")
					// 		})

					// 		It("prints a warning and tip but doesn't exit 1", func() {
					// 			helpers.BuildpackWithStack(func(buildpackPath string) {
					// 				session := helpers.CF("update-buildpack", buildpackName, buildpackPath, "1")
					// 				Eventually(session).Should(Exit(0))
					// 				Eventually(session.Err).Should(Say("Buildpack %s already exists without a stack", buildpackName))
					// 				Eventually(session).Should(Say("TIP: use 'cf buildpacks' and 'cf delete-buildpack' to delete buildpack %s without a stack", buildpackName))
					// 			}, stacks[0])

					// 		})
					// 	})

					// 	When("the existing buildpack has the same non-nil stack", func() {
					// 		BeforeEach(func() {
					// 			helpers.BuildpackWithStack(func(buildpackPath string) {
					// 				session := helpers.CF("update-buildpack", existingBuildpack, buildpackPath, "5")
					// 				Eventually(session).Should(Exit(0))
					// 			}, stacks[0])

					// 		})

					// 		It("prints a warning but doesn't exit 1", func() {
					// 			helpers.BuildpackWithStack(func(buildpackPath string) {
					// 				session := helpers.CF("update-buildpack", buildpackName, buildpackPath, "1")
					// 				Eventually(session).Should(Exit(0))
					// 				Expect(session.Err).To(Say("The buildpack name %s is already in use for the stack %s", buildpackName, stacks[0]))
					// 				Expect(session).To(Say("TIP: use 'cf update-buildpack' to update this buildpack"))
					// 			}, stacks[0])
					// 		})
					// 	})
					// })
				})
			})

			When("the buildpack does not exist", func() {
				It("returns an error", func() {
					session := helpers.CF("update-buildpack", buildpackName, "https://does.not.matter.com", "1")
					Eventually(session.Err).Should(Say("Buildpack %s not found", buildpackName))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				})
			})
		})
	})
})
