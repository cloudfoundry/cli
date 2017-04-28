package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("bind-security-group command", func() {
	var (
		orgName    string
		spaceName1 string
		spaceName2 string
	)

	BeforeEach(func() {
		orgName = helpers.NewOrgName()
		spaceName1 = helpers.NewSpaceName()
		spaceName2 = helpers.NewSpaceName()

		helpers.LoginCF()
	})

	Describe("help", func() {
		Context("when --help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("bind-security-group", "--help")
				Eventually(session.Out).Should(Say("NAME:"))
				Eventually(session.Out).Should(Say("   bind-security-group - Bind a security group to a particular space, or all existing spaces of an org"))
				Eventually(session.Out).Should(Say("USAGE:"))
				Eventually(session.Out).Should(Say("   cf bind-security-group SECURITY_GROUP ORG \\[SPACE\\]"))
				Eventually(session.Out).Should(Say("TIP: Changes will not apply to existing running applications until they are restarted."))
				Eventually(session.Out).Should(Say("SEE ALSO:"))
				Eventually(session.Out).Should(Say("   apps, bind-running-security-group, bind-staging-security-group, restart, security-groups"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("when the environment is not setup correctly", func() {
		Context("when no API endpoint is set", func() {
			BeforeEach(func() {
				helpers.UnsetAPI()
			})

			It("fails with no API endpoint set message", func() {
				session := helpers.CF("bind-security-group", "some-security-group", "some-org")
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("No API endpoint set. Use 'cf login' or 'cf api' to target an endpoint."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when not logged in", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
			})

			It("fails with not logged in message", func() {
				session := helpers.CF("bind-security-group", "some-security-group", "some-org")
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("Not logged in. Use 'cf login' to log in."))
				Eventually(session).Should(Exit(1))
			})
		})
	})

	Context("when the input is invalid", func() {
		Context("when the security group is not provided", func() {
			It("fails with an incorrect usage message and displays help", func() {
				session := helpers.CF("bind-security-group")
				Eventually(session.Err).Should(Say("Incorrect Usage: the required arguments `SECURITY_GROUP` and `ORG` were not provided"))
				Eventually(session.Out).Should(Say("USAGE:"))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the org is not provided", func() {
			It("fails with an incorrect usage message and displays help", func() {
				session := helpers.CF("bind-security-group", "some-security-group")
				Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `ORG` was not provided"))
				Eventually(session.Out).Should(Say("USAGE:"))
				Eventually(session).Should(Exit(1))
			})
		})
	})

	Context("when the security group doesn't exist", func() {
		It("fails with a security group not found message", func() {
			session := helpers.CF("bind-security-group", "some-security-group-that-doesn't-exist", "some-org")
			Eventually(session.Err).Should(Say("Security group 'some-security-group-that-doesn't-exist' not found."))
			Eventually(session.Out).Should(Say("FAILED"))
			Eventually(session).Should(Exit(1))
		})
	})

	Context("when the security group exists", func() {
		var someSecurityGroup helpers.SecurityGroup

		BeforeEach(func() {
			someSecurityGroup = helpers.NewSecurityGroup("some-security-group", "tcp", "0.0.0.0/0", "53", "")
			someSecurityGroup.Create()
		})

		AfterEach(func() {
			someSecurityGroup.Delete()
		})

		Context("when the org doesn't exist", func() {
			It("fails with a org not found message", func() {
				session := helpers.CF("bind-security-group", "some-security-group", "some-org")
				Eventually(session.Err).Should(Say("Organization 'some-org' not found."))
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the org exists", func() {
			BeforeEach(func() {
				helpers.CreateOrg(orgName)
				helpers.TargetOrg(orgName)
			})

			Context("when binding to all spaces in an org", func() {
				Context("when there are spaces in this org", func() {
					BeforeEach(func() {
						helpers.CreateSpace(spaceName1)
						helpers.CreateSpace(spaceName2)
					})
					It("binds the security group to each space", func() {
						session := helpers.CF("bind-security-group", "some-security-group", orgName)
						userName, _ := helpers.GetCredentials()
						Eventually(session.Out).Should(Say("Assigning security group some-security-group to space INTEGRATION-SPACE.* in org %s as %s...", orgName, userName))
						Eventually(session.Out).Should(Say("OK"))
						Eventually(session.Out).Should(Say("Assigning security group some-security-group to space INTEGRATION-SPACE.* in org %s as %s...", orgName, userName))
						Eventually(session.Out).Should(Say("OK"))
						Eventually(session.Out).Should(Say("TIP: Changes will not apply to existing running applications until they are restarted."))
						Eventually(session).Should(Exit(0))
					})
				})

				Context("when there are no spaces in this org", func() {
					It("does not bind the security group to any space", func() {
						session := helpers.CF("bind-security-group", "some-security-group", orgName)
						Consistently(session.Out).ShouldNot(Say("Assigning security group"))
						Consistently(session.Out).ShouldNot(Say("OK"))
						Eventually(session.Out).Should(Say("TIP: Changes will not apply to existing running applications until they are restarted."))
						Eventually(session).Should(Exit(0))
					})
				})
			})

			Context("when binding to a particular space", func() {
				Context("when the space exists", func() {
					BeforeEach(func() {
						helpers.CreateSpace(spaceName1)
					})
					It("binds the security group to the space", func() {
						session := helpers.CF("bind-security-group", "some-security-group", orgName, spaceName1)
						userName, _ := helpers.GetCredentials()
						Eventually(session.Out).Should(Say("Assigning security group some-security-group to space %s in org %s as %s...", spaceName1, orgName, userName))
						Eventually(session.Out).Should(Say("OK"))
						Eventually(session.Out).Should(Say("TIP: Changes will not apply to existing running applications until they are restarted."))
						Eventually(session).Should(Exit(0))
					})
				})
				Context("when the space doesn't exist", func() {
					It("fails with a space not found message", func() {
						session := helpers.CF("bind-security-group", "some-security-group", orgName, "space-doesnt-exist")
						Eventually(session.Err).Should(Say("Space 'space-doesnt-exist' not found."))
						Eventually(session.Out).Should(Say("FAILED"))
						Eventually(session).Should(Exit(1))
					})
				})
			})
		})
	})
})
