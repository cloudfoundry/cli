package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("target command", func() {
	var (
		orgName   string
		spaceName string
	)

	BeforeEach(func() {
		helpers.LoginCF()

		orgName = helpers.NewOrgName()
		spaceName = helpers.RandomName()
		helpers.RunIfExperimental("target command refactor is still experimental")
	})

	Context("help", func() {
		It("displays help", func() {
			session := helpers.CF("target", "--help")
			Eventually(session.Out).Should(Say("NAME:"))
			Eventually(session.Out).Should(Say("   target - Set or view the targeted org or space"))
			Eventually(session.Out).Should(Say("USAGE:"))
			Eventually(session.Out).Should(Say("   cf target \\[-o ORG\\] \\[-s SPACE\\]"))
			Eventually(session.Out).Should(Say("ALIAS:"))
			Eventually(session.Out).Should(Say("   t"))
			Eventually(session.Out).Should(Say("OPTIONS:"))
			Eventually(session.Out).Should(Say("   -o      Organization"))
			Eventually(session.Out).Should(Say("   -s      Space"))
			Eventually(session.Out).Should(Say("SEE ALSO:"))
			Eventually(session.Out).Should(Say("   create-org, create-space, login, orgs, spaces"))
			Eventually(session).Should(Exit(0))
		})
	})

	Context("when the environment is not setup correctly", func() {
		Context("when no API endpoint is set", func() {
			BeforeEach(func() {
				helpers.UnsetAPI()
			})

			It("fails with no API endpoint set message", func() {
				session := helpers.CF("target", "-o", "some-org", "-s", "some-space")
				Eventually(session.Err).Should(Say("No API endpoint set. Use 'cf login' or 'cf api' to target an endpoint."))
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when not logged in", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
			})

			Context("when trying to target an org", func() {
				It("fails with not logged in message", func() {
					session := helpers.CF("target", "-o", "some-org")
					Eventually(session.Err).Should(Say("Not logged in. Use 'cf login' to log in."))
					Eventually(session.Out).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				})
			})

			Context("when trying to target a space", func() {
				It("fails with not logged in message", func() {
					session := helpers.CF("target", "-s", "some-space")
					Eventually(session.Err).Should(Say("Not logged in. Use 'cf login' to log in."))
					Eventually(session.Out).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				})
			})

			Context("when trying to target an org and space", func() {
				It("fails with not logged in message", func() {
					session := helpers.CF("target", "-o", "some-org", "-s", "some-space")
					Eventually(session.Out).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("Not logged in. Use 'cf login' to log in."))
					Eventually(session).Should(Exit(1))
				})
			})

			Context("when trying to get the target", func() {
				It("fails with not logged in message", func() {
					session := helpers.CF("target")
					Eventually(session.Out).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("Not logged in. Use 'cf login' to log in."))
					Eventually(session).Should(Exit(1))
				})
			})
		})
	})

	Context("when no arguments are provided", func() {
		BeforeEach(func() {
			setupCF(ReadOnlyOrg, ReadOnlySpace)
		})

		It("displays current target information", func() {
			username, _ := helpers.GetCredentials()
			session := helpers.CF("target")
			Eventually(session.Out).Should(Say("API endpoint:   %s", apiURL))
			Eventually(session.Out).Should(Say(`API version:    [\d.]+`))
			Eventually(session.Out).Should(Say("User:           %s", username))
			Eventually(session.Out).Should(Say("Org:            %s", ReadOnlyOrg))
			Eventually(session.Out).Should(Say("Space:          %s", ReadOnlySpace))
			Eventually(session).Should(Exit(0))
		})
	})

	Context("when only an org argument is provided", func() {
		Context("when the org does not exist", func() {
			// We set targets to verify that the target command
			// preserves existing targets in failure
			BeforeEach(func() {
				setupCF(ReadOnlyOrg, ReadOnlySpace)
			})

			It("displays org not found, exits 1, and preserves existing target", func() {
				session := helpers.CF("target", "-o", orgName)
				Eventually(session.Err).Should(Say("Organization '%s' not found", orgName))
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))

				session = helpers.CF("target")
				Eventually(session.Out).Should(Say("Org:            %s", ReadOnlyOrg))
				Eventually(session.Out).Should(Say("Space:          %s", ReadOnlySpace))
				Eventually(session).Should(Exit(0))
			})
		})

		Context("when the org exists", func() {
			BeforeEach(func() {
				helpers.CreateOrg(orgName)
				helpers.TargetOrg(orgName)
			})

			Context("when there are no spaces in the org", func() {
				BeforeEach(func() {
					helpers.ClearTarget()
				})

				It("only targets the org and exits 0", func() {
					username, _ := helpers.GetCredentials()
					session := helpers.CF("target", "-o", orgName)
					Eventually(session.Out).Should(Say("API endpoint:   %s", apiURL))
					Eventually(session.Out).Should(Say(`API version:    [\d.]+`))
					Eventually(session.Out).Should(Say("User:           %s", username))
					Eventually(session.Out).Should(Say("Org:            %s", orgName))
					Eventually(session.Out).Should(Say("No space targeted, use 'cf target -s SPACE"))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when there is only one space in the org", func() {
				BeforeEach(func() {
					helpers.CreateSpace(spaceName)
					helpers.ClearTarget()
				})

				It("targets the org and space and exits 0", func() {
					username, _ := helpers.GetCredentials()
					session := helpers.CF("target", "-o", orgName)
					Eventually(session.Out).Should(Say("API endpoint:   %s", apiURL))
					Eventually(session.Out).Should(Say(`API version:    [\d.]+`))
					Eventually(session.Out).Should(Say("User:           %s", username))
					Eventually(session.Out).Should(Say("Org:            %s", orgName))
					Eventually(session.Out).Should(Say("Space:          %s", spaceName))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when there are multiple spaces in the org", func() {
				BeforeEach(func() {
					helpers.CreateSpace(spaceName)
					helpers.CreateSpace(helpers.RandomName())
					helpers.ClearTarget()
				})

				It("targets the org only and exits 0", func() {
					username, _ := helpers.GetCredentials()
					session := helpers.CF("target", "-o", orgName)
					Eventually(session.Out).Should(Say("API endpoint:   %s", apiURL))
					Eventually(session.Out).Should(Say(`API version:    [\d.]+`))
					Eventually(session.Out).Should(Say("User:           %s", username))
					Eventually(session.Out).Should(Say("Org:            %s", orgName))
					Eventually(session.Out).Should(Say("No space targeted, use 'cf target -s SPACE"))
					Eventually(session).Should(Exit(0))
				})

				Context("when there is an existing targeted space", func() {
					BeforeEach(func() {
						session := helpers.CF("target", "-o", orgName, "-s", spaceName)
						Eventually(session).Should(Exit(0))
					})

					It("unsets the targeted space", func() {
						session := helpers.CF("target", "-o", orgName)
						Eventually(session.Out).Should(Say("No space targeted, use 'cf target -s SPACE"))
						Eventually(session).Should(Exit(0))
					})
				})
			})
		})
	})

	Context("when only a space argument is provided", func() {
		Context("when there is an existing targeted org", func() {
			existingSpace := helpers.RandomName()

			BeforeEach(func() {
				helpers.LoginCF()
				// We create and set a space to verify that the target command
				// preserves existing targets in failure
				helpers.CreateOrgAndSpace(orgName, existingSpace)
				helpers.TargetOrgAndSpace(orgName, existingSpace)
			})

			Context("when the space exists", func() {
				BeforeEach(func() {
					helpers.CreateSpace(spaceName)
				})

				It("targets the space and exits 0", func() {
					username, _ := helpers.GetCredentials()
					session := helpers.CF("target", "-s", spaceName)
					Eventually(session.Out).Should(Say("API endpoint:   %s", apiURL))
					Eventually(session.Out).Should(Say(`API version:    [\d.]+`))
					Eventually(session.Out).Should(Say("User:           %s", username))
					Eventually(session.Out).Should(Say("Org:            %s", orgName))
					Eventually(session.Out).Should(Say("Space:          %s", spaceName))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when the space does not exist", func() {
				It("displays space not found, exits 1, and preserves existing target", func() {
					session := helpers.CF("target", "-s", spaceName)
					Eventually(session.Err).Should(Say("Space '%s' not found.", spaceName))
					Eventually(session.Out).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))

					session = helpers.CF("target")
					Eventually(session.Out).Should(Say("Org:            %s", orgName))
					Eventually(session.Out).Should(Say("Space:          %s", existingSpace))
					Eventually(session).Should(Exit(0))
				})
			})
		})

		Context("when there is not an existing targeted org", func() {
			It("displays org must be targeted first and exits 1", func() {
				session := helpers.CF("target", "-s", spaceName)
				Eventually(session.Err).Should(Say("An org must be targeted before targeting a space"))
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})
	})

	Context("when both org and space arguments are provided", func() {
		// We set the targets to verify that the target command preserves existing targets
		// in failure
		BeforeEach(func() {
			setupCF(ReadOnlyOrg, ReadOnlySpace)
		})

		Context("when the org does not exist", func() {
			It("displays org not found, exits 1, and preserves existing targets", func() {
				session := helpers.CF("target", "-o", orgName, "-s", spaceName)
				Eventually(session.Err).Should(Say("Organization '%s' not found", orgName))
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))

				session = helpers.CF("target")
				Eventually(session.Out).Should(Say("Org:            %s", ReadOnlyOrg))
				Eventually(session.Out).Should(Say("Space:          %s", ReadOnlySpace))
				Eventually(session).Should(Exit(0))
			})
		})

		Context("when the org exists", func() {
			BeforeEach(func() {
				helpers.CreateOrg(orgName)
			})

			Context("when the space exists", func() {
				BeforeEach(func() {
					helpers.TargetOrg(orgName)
					helpers.CreateSpace(spaceName)
					helpers.ClearTarget()
				})

				It("targets the org and space and exits 0", func() {
					username, _ := helpers.GetCredentials()
					session := helpers.CF("target", "-o", orgName, "-s", spaceName)
					Eventually(session.Out).Should(Say("API endpoint:   %s", apiURL))
					Eventually(session.Out).Should(Say(`API version:    [\d.]+`))
					Eventually(session.Out).Should(Say("User:           %s", username))
					Eventually(session.Out).Should(Say("Org:            %s", orgName))
					Eventually(session.Out).Should(Say("Space:          %s", spaceName))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when the space does not exist", func() {
				It("displays space not found, exits 1, and preserves existing targets", func() {
					session := helpers.CF("target", "-o", orgName, "-s", spaceName)
					Eventually(session.Err).Should(Say("Space '%s' not found.", spaceName))
					Eventually(session.Out).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))

					session = helpers.CF("target")
					Eventually(session.Out).Should(Say("Org:            %s", ReadOnlyOrg))
					Eventually(session.Out).Should(Say("Space:          %s", ReadOnlySpace))
					Eventually(session).Should(Exit(0))
				})
			})
		})
	})
})
