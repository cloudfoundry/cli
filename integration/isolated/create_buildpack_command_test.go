package isolated

import (
	"os"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = FDescribe("create buildpack command", func() {
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

					// TODO: fix according https://www.pivotaltracker.com/story/show/158479997
					XContext("when the new buildpack has an invalid stack", func() {
						It("returns the appropriate error", func() {
							helpers.BuildpackWithStack(func(buildpackPath string) {
								session := helpers.CF("create-buildpack", buildpackName, buildpackPath, "1")
								Eventually(session).Should(Exit(1))
								Eventually(session.Err).Should(Say("Buildpack stack 'bogus-stack' does not exist"))
							}, "bogus-stack")
						})
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

							It("returns the appropriate error", func() {
								helpers.BuildpackWithStack(func(buildpackPath string) {
									session := helpers.CF("create-buildpack", buildpackName, buildpackPath, "1")
									Eventually(session).Should(Exit(0))
									Eventually(session).Should(Say("Buildpack %s already exists without a stack", buildpackName))
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
									Eventually(session).Should(Exit(1))
								}, stacks[0])

								session := helpers.CF("buildpacks")
								Eventually(session).Should(Exit(0))
								Expect(session.Out).To(Say(`%s\s+%s\s+1`, buildpackName, stacks[0]))
								Expect(session.Out).To(Say(`%s\s+%s\s+6`, existingBuildpack, stacks[1]))
							})
						})

						// TODO: Fix this with https://www.pivotaltracker.com/story/show/158479997
						XContext("when the existing buildpack has a nil stack", func() {
							BeforeEach(func() {
								helpers.BuildpackWithStack(func(buildpackPath string) {
									session := helpers.CF("create-buildpack", existingBuildpack, buildpackPath, "5")
									Eventually(session).Should(Exit(0))
								}, "")
							})

							It("returns the appropriate error", func() {
								helpers.BuildpackWithStack(func(buildpackPath string) {
									session := helpers.CF("create-buildpack", buildpackName, buildpackPath, "1")
									Eventually(session).Should(Exit(0))
								}, stacks[0])

								session := helpers.CF("buildpacks")
								Eventually(session).Should(Exit(0))
								Expect(session.Out).To(Say(`%s\s+5`, existingBuildpack))
							})
						})

						Context("when the existing buildpack has the same non-nil stack", func() {
							BeforeEach(func() {
								helpers.BuildpackWithStack(func(buildpackPath string) {
									session := helpers.CF("create-buildpack", existingBuildpack, buildpackPath, "5")
									Eventually(session).Should(Exit(0))
								}, stacks[0])
							})
							It("returns the appropriate error", func() {
								helpers.BuildpackWithStack(func(buildpackPath string) {
									session := helpers.CF("create-buildpack", buildpackName, buildpackPath, "1")
									Eventually(session).Should(Exit(0))
									Eventually(session).Should(Say("Buildpack %s already exists with stack %s", buildpackName, stacks[0]))
								}, stacks[0])

								session := helpers.CF("buildpacks")
								Eventually(session).Should(Exit(0))
								Expect(session.Out).To(Say(`%s\s+%s\s+6`, existingBuildpack, stacks[0]))
							})
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
			var (
			// buildpackName string
			)
			BeforeEach(func() {
				// buildpackName = helpers.NewBuildpack()
			})
			Context("when specifying a valid path", func() {
				It("successfully uploads a buildpack", func() {
				})
			})
			Context("when specifying an invalid path", func() {
				It("returns the appropriate error", func() {
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
					Expect(session.Out).To(Say(`%s\s+3`, buildpackName))
				})
			})

			Context("when position is negative integer", func() {
				It("uploads buildpack in first position", func() {
					helpers.BuildpackWithStack(func(buildpackPath string) {
						session := helpers.CF("create-buildpack", buildpackName, buildpackPath, "-1")
						Eventually(session).Should(Exit(1))
						Eventually(session.Err).Should(Say("Position must be a positive integer"))
					}, "")
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
						Expect(session.Err).To(Say("Incorrect Usage: The following arguments cannot be used together: enable, disable"))
					}, "")
				})
			})
		})
	})
})
