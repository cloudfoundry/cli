package global

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("rename buildpack command", func() {
	Describe("help", func() {
		When("--help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("rename-buildpack", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("rename-buildpack - Rename a buildpack"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say(`cf rename-buildpack BUILDPACK_NAME NEW_BUILDPACK_NAME \[-s STACK\]`))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("update-buildpack"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(false, false, ReadOnlyOrg, "rename-buildpack", "fake-buildpack", "some-name")
		})
	})

	When("the user is logged in", func() {
		var (
			oldBuildpackName string
			newBuildpackName string
			stacks           []string
			username         string
		)

		BeforeEach(func() {
			helpers.LoginCF()
			oldBuildpackName = helpers.NewBuildpackName()
			newBuildpackName = helpers.NewBuildpackName()
			stacks = helpers.EnsureMinimumNumberOfStacks(2)

			username, _ = helpers.GetCredentials()
		})
		AfterEach(func() {
			helpers.DeleteBuildpackIfOnOldCCAPI(oldBuildpackName)
			helpers.DeleteBuildpackIfOnOldCCAPI(newBuildpackName)
		})

		Context("when the user provides a stack", func() {
			var session *Session

			JustBeforeEach(func() {
				session = helpers.CF("rename-buildpack", oldBuildpackName, newBuildpackName, "-s", stacks[0])
			})

			When("no buildpack with the name/stack combo is found", func() {
				When("no buildpacks with the same name exist", func() {
					It("returns a buildpack not found error", func() {
						Eventually(session).Should(Say(`Renaming buildpack %s to %s with stack %s as %s\.\.\.`, oldBuildpackName, newBuildpackName, stacks[0], username))
						Eventually(session).Should(Say("FAILED"))
						Eventually(session.Err).Should(Say("Buildpack '%s' with stack '%s' not found", oldBuildpackName, stacks[0]))
						Eventually(session).Should(Exit(1))
					})
				})

				When("no buildpacks with the same name and stack exist", func() {
					BeforeEach(func() {
						helpers.SetupBuildpackWithoutStack(oldBuildpackName)
					})

					It("returns a buildpack not found error", func() {
						Eventually(session).Should(Say(`Renaming buildpack %s to %s with stack %s as %s\.\.\.`, oldBuildpackName, newBuildpackName, stacks[0], username))
						Eventually(session).Should(Say("FAILED"))
						Eventually(session.Err).Should(Say("Buildpack '%s' with stack '%s' not found", oldBuildpackName, stacks[0]))
						Eventually(session).Should(Exit(1))
					})
				})
			})

			When("there are multiple existing buildpacks with the specified old name", func() {
				When("one of the existing buildpacks has an empty stack", func() {
					BeforeEach(func() {
						helpers.SetupBuildpackWithStack(oldBuildpackName, stacks[0])
						helpers.SetupBuildpackWithoutStack(oldBuildpackName)
					})

					When("renaming to unique name", func() {
						It("successfully renames the buildpack", func() {
							Eventually(session).Should(Say(`Renaming buildpack %s to %s with stack %s as %s\.\.\.`, oldBuildpackName, newBuildpackName, stacks[0], username))
							Eventually(session).Should(Say("OK"))
							Eventually(session).Should(Exit(0))
						})
					})

					When("renaming to the same name as another buildpack", func() {
						When("the existing existing buildpack with the new name has the same stack", func() {
							BeforeEach(func() {
								helpers.SetupBuildpackWithStack(newBuildpackName, stacks[0])
							})

							It("returns an error", func() {
								Eventually(session).Should(Say(`Renaming buildpack %s to %s with stack %s as %s\.\.\.`, oldBuildpackName, newBuildpackName, stacks[0], username))
								Eventually(session).Should(Say("FAILED"))
								Eventually(session.Err).Should(Say("%s is already in use", newBuildpackName))
								Eventually(session).Should(Exit(1))
							})
						})

						When("the existing buildpack with the new name has a different stack", func() {
							BeforeEach(func() {
								helpers.SetupBuildpackWithStack(newBuildpackName, stacks[1])
							})

							It("successfully renames the buildpack", func() {
								Eventually(session).Should(Say(`Renaming buildpack %s to %s with stack %s as %s\.\.\.`, oldBuildpackName, newBuildpackName, stacks[0], username))
								Eventually(session).Should(Say("OK"))
								Eventually(session).Should(Exit(0))
							})
						})

						When("the existing existing buildpack with the new name has an empty stack", func() {
							BeforeEach(func() {
								helpers.SetupBuildpackWithoutStack(newBuildpackName)
							})

							It("successfully renames the buildpack", func() {
								Eventually(session).Should(Say(`Renaming buildpack %s to %s with stack %s as %s\.\.\.`, oldBuildpackName, newBuildpackName, stacks[0], username))
								Eventually(session).Should(Say("OK"))
								Eventually(session).Should(Exit(0))
							})
						})
					})

					When("the new name is invalid", func() {
						BeforeEach(func() {
							newBuildpackName += ".dots!"
						})
						It("returns an invalid-name error", func() {
							Eventually(session).Should(Say(`Renaming buildpack %s to %s with stack %s as %s\.\.\.`, oldBuildpackName, newBuildpackName, stacks[0], username))
							Eventually(session).Should(Say("FAILED"))
							Eventually(session.Err).Should(Say("Buildpack is invalid: (?:name ){1,2}can only contain alphanumeric characters"))
							Eventually(session).Should(Exit(1))
						})
					})
				})

				When("neither of the existing buildpacks has an empty stack", func() {
					BeforeEach(func() {
						helpers.SetupBuildpackWithStack(oldBuildpackName, stacks[0])
						helpers.SetupBuildpackWithStack(oldBuildpackName, stacks[1])
					})

					When("renaming to unique name", func() {
						It("successfully renames the buildpack", func() {
							Eventually(session).Should(Say(`Renaming buildpack %s to %s with stack %s as %s\.\.\.`, oldBuildpackName, newBuildpackName, stacks[0], username))
							Eventually(session).Should(Say("OK"))
							Eventually(session).Should(Exit(0))
						})
					})
				})
			})

			When("just one buildpack is found with the name/stack combo", func() {
				BeforeEach(func() {
					helpers.SetupBuildpackWithStack(oldBuildpackName, stacks[0])
				})

				When("renaming to unique name", func() {
					It("successfully renames the buildpack", func() {
						Eventually(session).Should(Say(`Renaming buildpack %s to %s with stack %s as %s\.\.\.`, oldBuildpackName, newBuildpackName, stacks[0], username))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Exit(0))
					})
				})

				When("renaming to the same name as another buildpack", func() {
					When("the existing buildpack with the new name has the same stack", func() {
						BeforeEach(func() {
							helpers.SetupBuildpackWithStack(newBuildpackName, stacks[0])
						})

						It("returns a buildpack name/stack taken error", func() {
							Eventually(session).Should(Say(`Renaming buildpack %s to %s with stack %s as %s\.\.\.`, oldBuildpackName, newBuildpackName, stacks[0], username))
							Eventually(session).Should(Say("FAILED"))
							Eventually(session.Err).Should(Say("%s is already in use", newBuildpackName))
							Eventually(session).Should(Exit(1))
						})
					})

					When("the existing buildpack with the new name has a different stack", func() {
						BeforeEach(func() {
							helpers.SetupBuildpackWithStack(newBuildpackName, stacks[1])
						})

						It("successfully renames the buildpack", func() {
							Eventually(session).Should(Say(`Renaming buildpack %s to %s with stack %s as %s\.\.\.`, oldBuildpackName, newBuildpackName, stacks[0], username))
							Eventually(session).Should(Say("OK"))
							Eventually(session).Should(Exit(0))
						})
					})

					When("the existing buildpack with the new name has an empty stack", func() {
						BeforeEach(func() {
							helpers.SetupBuildpackWithoutStack(newBuildpackName)
						})

						It("successfully renames the buildpack", func() {
							Eventually(session).Should(Say(`Renaming buildpack %s to %s with stack %s as %s\.\.\.`, oldBuildpackName, newBuildpackName, stacks[0], username))
							Eventually(session).Should(Say("OK"))
							Eventually(session).Should(Exit(0))
						})
					})
				})
			})
		})

		//If the user does not provide a stack, and there are multiple ambiguous buildpacks, we assume that they intended to rename the one with an empty stack.
		When("the user does not provide a stack", func() {
			var session *Session

			JustBeforeEach(func() {
				session = helpers.CF("rename-buildpack", oldBuildpackName, newBuildpackName)
			})

			When("no buildpacks with the old name exist", func() {
				It("returns a buildpack not found error", func() {
					Eventually(session).Should(Say(`Renaming buildpack %s to %s as %s\.\.\.`, oldBuildpackName, newBuildpackName, username))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("Buildpack '%s' not found", oldBuildpackName))
					Eventually(session).Should(Exit(1))
				})
			})

			When("one buildpack with the old name exists with a stack association", func() {
				BeforeEach(func() {
					helpers.SetupBuildpackWithStack(oldBuildpackName, stacks[0])
				})

				When("renaming to unique name", func() {
					It("successfully renames the buildpack", func() {
						Eventually(session).Should(Say(`Renaming buildpack %s to %s as %s\.\.\.`, oldBuildpackName, newBuildpackName, username))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Exit(0))
					})
				})

				When("renaming to an invalid name", func() {
					BeforeEach(func() {
						newBuildpackName += ",hey.dots!"
					})
					It("complains about the new name", func() {
						Eventually(session).Should(Say(`Renaming buildpack %s to %s as %s\.\.\.`, oldBuildpackName, newBuildpackName, username))
						Eventually(session.Err).Should(Say("Buildpack is invalid: (?:name ){1,2}can only contain alphanumeric characters"))
						Eventually(session).Should(Say("FAILED"))
						Eventually(session).Should(Exit(1))
					})
				})

				When("renaming to the same name as an existing buildpack with no stack association", func() {
					BeforeEach(func() {
						helpers.SetupBuildpackWithoutStack(newBuildpackName)
					})

					It("successfully renames the buildpack", func() {
						Eventually(session).Should(Say(`Renaming buildpack %s to %s as %s\.\.\.`, oldBuildpackName, newBuildpackName, username))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Exit(0))
					})

				})

				When("renaming to the same name as an existing buildpack with a different stack association", func() {
					BeforeEach(func() {
						helpers.SetupBuildpackWithStack(newBuildpackName, stacks[1])
					})

					It("successfully renames the buildpack", func() {
						Eventually(session).Should(Say(`Renaming buildpack %s to %s as %s\.\.\.`, oldBuildpackName, newBuildpackName, username))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Exit(0))
					})

				})

				When("renaming to the same name as an existing buildpack with the same stack assocation", func() {
					BeforeEach(func() {
						helpers.SetupBuildpackWithStack(newBuildpackName, stacks[0])
					})

					It("returns a buildpack name/stack taken error", func() {
						Eventually(session).Should(Say(`Renaming buildpack %s to %s as %s\.\.\.`, oldBuildpackName, newBuildpackName, username))
						Eventually(session).Should(Say("FAILED"))
						Eventually(session.Err).Should(Say("The buildpack name %s is already in use for the stack %s", newBuildpackName, stacks[0]))
						Eventually(session).Should(Exit(1))
					})
				})
			})

			When("there are multiple existing buildpacks with the old name", func() {
				When("none of the buildpacks has an empty stack", func() {
					BeforeEach(func() {
						helpers.SetupBuildpackWithStack(oldBuildpackName, stacks[0])
						helpers.SetupBuildpackWithStack(oldBuildpackName, stacks[1])
					})

					It("returns a buildpack not found error", func() {
						Eventually(session).Should(Say(`Renaming buildpack %s to %s as %s\.\.\.`, oldBuildpackName, newBuildpackName, username))
						Eventually(session).Should(Say("FAILED"))
						Eventually(session.Err).Should(Say(`Multiple buildpacks named %s found\. Specify a stack name by using a '-s' flag\.`, oldBuildpackName))
						Eventually(session).Should(Exit(1))
					})
				})

				When("one of the existing buildpacks with the old name has an empty stack", func() {
					BeforeEach(func() {
						helpers.SetupBuildpackWithStack(oldBuildpackName, stacks[0])
						helpers.SetupBuildpackWithoutStack(oldBuildpackName)
					})

					When("renaming to unique name", func() {
						It("successfully renames the buildpack", func() {
							Eventually(session).Should(Say(`Renaming buildpack %s to %s as %s\.\.\.`, oldBuildpackName, newBuildpackName, username))
							Eventually(session).Should(Say("OK"))
							Eventually(session).Should(Exit(0))
						})
					})

					When("renaming to the same name as another buildpack", func() {
						When("the existing buildpack with the new name has a non-empty stack", func() {
							BeforeEach(func() {
								helpers.SetupBuildpackWithStack(newBuildpackName, stacks[1])
							})

							It("successfully renames the buildpack", func() {
								Eventually(session).Should(Say(`Renaming buildpack %s to %s as %s\.\.\.`, oldBuildpackName, newBuildpackName, username))
								Eventually(session).Should(Say("OK"))
								Eventually(session).Should(Exit(0))
							})
						})

						When("the existing buildpack with the new name has an empty stack", func() {
							BeforeEach(func() {
								helpers.SetupBuildpackWithoutStack(newBuildpackName)
							})

							It("returns a buildpack name/stack taken error", func() {
								Eventually(session).Should(Say(`Renaming buildpack %s to %s as %s\.\.\.`, oldBuildpackName, newBuildpackName, username))
								Eventually(session).Should(Say("FAILED"))
								Eventually(session.Err).Should(Say("Buildpack %s already exists without a stack", newBuildpackName))
								Eventually(session).Should(Exit(1))
							})
						})
					})
				})
			})
		})
	})
})
