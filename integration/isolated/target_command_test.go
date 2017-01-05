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
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Out).Should(Say("No API endpoint set. Use 'cf login' or 'cf api' to target an endpoint."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when not logged in", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
			})

			Context("when trying to set the target", func() {
				It("fails with not logged in message", func() {
					session := helpers.CF("target", "-o", "some-org", "-s", "some-space")
					Eventually(session.Out).Should(Say("FAILED"))
					Eventually(session.Out).Should(Say("Not logged in. Use 'cf login' to log in."))
					Eventually(session).Should(Exit(1))
				})
			})

			Context("when trying to get the target", func() {
				It("fails with not logged in message", func() {
					session := helpers.CF("target")
					Eventually(session.Out).Should(Say("Not logged in. Use 'cf login' to log in."))
					Eventually(session.Out).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				})
			})
		})
	})

	Context("when no arguments are given", func() {
		BeforeEach(func() {
			setupCF(ReadOnlyOrg, ReadOnlySpace)
		})

		It("displays the currently targeted org and space", func() {
			username, _ := helpers.GetCredentials()
			session := helpers.CF("target")
			Eventually(session.Out).Should(Say(`API endpoint:   %s \(API version: [\d.]+\)`, apiURL))
			Eventually(session.Out).Should(Say("User:           %s", username))
			Eventually(session.Out).Should(Say("Org:            %s", ReadOnlyOrg))
			Eventually(session.Out).Should(Say("Space:          %s", ReadOnlySpace))
			Eventually(session).Should(Exit(0))
		})
	})

	Context("when an org argument is given", func() {
		Context("when the org does not exist", func() {
			It("displays org not found and exits 1", func() {
				session := helpers.CF("target", "-o", orgName)
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Out).Should(Say("Could not target org."))
				Eventually(session.Out).Should(Say("Organization %s not found", orgName))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the org exists", func() {
			BeforeEach(func() {
				helpers.CreateOrg(orgName)
				helpers.TargetOrg(orgName)
			})

			Context("when there is only one space in the org", func() {
				BeforeEach(func() {
					helpers.CreateSpace(spaceName)
					helpers.ClearTarget()
				})

				It("targets the org and space and exits 0", func() {
					username, _ := helpers.GetCredentials()
					session := helpers.CF("target", "-o", orgName)
					Eventually(session.Out).Should(Say(`API endpoint:   %s \(API version: [\d.]+\)`, apiURL))
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
					Eventually(session.Out).Should(Say(`API endpoint:   %s \(API version: [\d.]+\)`, apiURL))
					Eventually(session.Out).Should(Say("User:           %s", username))
					Eventually(session.Out).Should(Say("Org:            %s", orgName))
					Eventually(session.Out).Should(Say("Space:          No space targeted, use 'cf target -s SPACE"))
					Eventually(session).Should(Exit(0))
				})

				Context("when another space in the org is already targetted", func() {
					BeforeEach(func() {
						session := helpers.CF("target", "-o", orgName, "-s", spaceName)
						Eventually(session).Should(Exit(0))
					})

					It("untargets the targetted space", func() {
						session := helpers.CF("target", "-o", orgName)
						Eventually(session.Out).Should(Say("Space:          No space targeted, use 'cf target -s SPACE"))
						Eventually(session).Should(Exit(0))
					})
				})
			})
		})
	})

	Context("when a space argument is given without an org", func() {
		Context("when org has already been set", func() {
			BeforeEach(func() {
				helpers.LoginCF()
				helpers.CreateOrg(orgName)
				helpers.TargetOrg(orgName)
			})

			Context("when space exists", func() {
				BeforeEach(func() {
					helpers.CreateSpace(spaceName)
				})

				It("targets the space and exits 0", func() {
					username, _ := helpers.GetCredentials()
					session := helpers.CF("target", "-s", spaceName)
					Eventually(session.Out).Should(Say(`API endpoint:   %s \(API version: [\d.]+\)`, apiURL))
					Eventually(session.Out).Should(Say("User:           %s", username))
					Eventually(session.Out).Should(Say("Org:            %s", orgName))
					Eventually(session.Out).Should(Say("Space:          %s", spaceName))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when space does not exist", func() {
				It("displays space not found and exits 1", func() {
					session := helpers.CF("target", "-s", spaceName)
					Eventually(session.Out).Should(Say("FAILED"))
					Eventually(session.Out).Should(Say("Unable to access space %s", spaceName))
					Eventually(session.Out).Should(Say("Space %s not found", spaceName))
					Eventually(session).Should(Exit(1))
				})
			})
		})

		Context("when org has not been set", func() {
			It("displays org must be targetted first and exits 1", func() {
				session := helpers.CF("target", "-s", spaceName)
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Out).Should(Say("An org must be targeted before targeting a space"))
				Eventually(session).Should(Exit(1))
			})
		})
	})

	Context("when both org and space arguments are given", func() {
		Context("when the org does not exist", func() {
			It("displays org not found and exits 1", func() {
				session := helpers.CF("target", "-o", orgName, "-s", spaceName)
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Out).Should(Say("Could not target org."))
				Eventually(session.Out).Should(Say("Organization %s not found", orgName))
				Eventually(session).Should(Exit(1))
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

				It("targets the org and exits 0", func() {
					username, _ := helpers.GetCredentials()
					session := helpers.CF("target", "-o", orgName, "-s", spaceName)
					Eventually(session.Out).Should(Say(`API endpoint:   %s \(API version: [\d.]+\)`, apiURL))
					Eventually(session.Out).Should(Say("User:           %s", username))
					Eventually(session.Out).Should(Say("Org:            %s", orgName))
					Eventually(session.Out).Should(Say("Space:          %s", spaceName))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when the space does not exist", func() {
				It("displays space not found and exits 1", func() {
					session := helpers.CF("target", "-o", orgName, "-s", spaceName)
					Eventually(session.Out).Should(Say("FAILED"))
					Eventually(session.Out).Should(Say("Unable to access space %s", spaceName))
					Eventually(session.Out).Should(Say("Space %s not found", spaceName))
					Eventually(session).Should(Exit(1))
				})
			})
		})
	})
})
