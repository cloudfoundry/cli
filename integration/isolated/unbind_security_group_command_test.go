package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("unbind-security-group command", func() {
	BeforeEach(func() {
		helpers.LoginCF()
	})

	Describe("help", func() {
		Context("when --help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("unbind-security-group", "--help")
				Eventually(session).Should(Exit(0))
				Expect(session.Out).To(Say("NAME:"))
				Expect(session.Out).To(Say("\\s+unbind-security-group - Unbind a security group from a space"))
				Expect(session.Out).To(Say("USAGE:"))
				Expect(session.Out).To(Say("\\s+cf unbind-security-group SECURITY_GROUP ORG SPACE"))
				Expect(session.Out).To(Say("TIP: Changes will not apply to existing running applications until they are restarted\\."))
				Expect(session.Out).To(Say("SEE ALSO:"))
				Expect(session.Out).To(Say("\\s+apps, restart, security-groups"))
			})
		})
	})

	Context("when the environment is not setup correctly", func() {
		Context("when no API endpoint is set", func() {
			BeforeEach(func() {
				helpers.UnsetAPI()
			})

			It("fails with no API endpoint set message", func() {
				session := helpers.CF("unbind-security-group", "some-security-group")
				Eventually(session).Should(Exit(1))
				Expect(session.Out).To(Say("FAILED"))
				Expect(session.Err).To(Say("No API endpoint set. Use 'cf login' or 'cf api' to target an endpoint."))
			})
		})

		Context("when not logged in", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
			})

			It("fails with not logged in message", func() {
				session := helpers.CF("unbind-security-group", "some-security-group")
				Eventually(session).Should(Exit(1))
				Expect(session.Out).To(Say("FAILED"))
				Expect(session.Err).To(Say("Not logged in. Use 'cf login' to log in."))
			})
		})

		Context("when no org is targeted and no org is specified on the command line", func() {
			BeforeEach(func() {
				helpers.ClearTarget()
			})

			It("fails with no org targeted error", func() {
				session := helpers.CF("unbind-security-group", "some-security-group")
				Eventually(session).Should(Exit(1))
				Expect(session.Out).To(Say("FAILED"))
				Expect(session.Err).To(Say("No org targeted, use 'cf target -o ORG' to target an org."))
			})
		})

		Context("when no space is targeted and no org is specified on the command line", func() {
			BeforeEach(func() {
				helpers.ClearTarget()
				orgName := helpers.NewOrgName()
				helpers.CreateOrg(orgName)
				helpers.TargetOrg(orgName)
			})

			It("fails with no org targeted error", func() {
				session := helpers.CF("unbind-security-group", "some-security-group")
				Eventually(session).Should(Exit(1))
				Expect(session.Out).To(Say("FAILED"))
				Expect(session.Err).To(Say("No space targeted, use 'cf target -s SPACE' to target a space."))
			})
		})
	})

	Context("when the command line arguments are invalid", func() {
		Context("when the security group is not provided", func() {
			It("fails with an incorrect usage message and displays help", func() {
				session := helpers.CF("unbind-security-group")
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Incorrect Usage: the required argument `SECURITY_GROUP` was not provided"))
				Expect(session.Out).To(Say("USAGE:"))
			})
		})

		Context("when the space is not provided", func() {
			It("fails with an incorrect usage message and displays help", func() {
				session := helpers.CF("unbind-security-group", "some-security-group", "some-org")
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Incorrect Usage: the required arguments `SECURITY_GROUP`, `ORG`, and `SPACE` were not provided"))
				Expect(session.Out).To(Say("USAGE:"))
			})
		})

		Context("with some setup", func() {
			var (
				orgName      string
				spaceName    string
				secGroup     helpers.SecurityGroup
				secGroupName string
			)

			BeforeEach(func() {
				orgName = helpers.NewOrgName()
				spaceName = helpers.NewSpaceName()
				helpers.CreateOrgAndSpace(orgName, spaceName)

				secGroupName = helpers.NewSecGroupName()
				secGroup = helpers.NewSecurityGroup(secGroupName, "tcp", "127.0.0.1", "8443", "some-description")
				secGroup.Create()
			})

			AfterEach(func() {
				secGroup.Delete()
			})

			Context("when the provided org doesn't exist", func() {
				It("fails with an 'org not found' message", func() {
					session := helpers.CF("unbind-security-group", secGroupName, "some-other-org", "some-other-space")
					Eventually(session).Should(Exit(1))
					Expect(session.Out).To(Say("FAILED"))
					Expect(session.Err).To(Say("Organization 'some-other-org' not found\\."))
				})
			})

			Context("when the provided space doesn't exist", func() {
				It("fails with a 'space not found' message", func() {
					session := helpers.CF("unbind-security-group", secGroupName, orgName, "some-other-space")
					Eventually(session).Should(Exit(1))
					Expect(session.Out).To(Say("FAILED"))
					Expect(session.Err).To(Say("Space 'some-other-space' not found\\."))
				})
			})

			Context("when the provided security group doesn't exist", func() {
				It("fails with a 'security group not found' message", func() {
					session := helpers.CF("unbind-security-group", "some-other-security-group", orgName, spaceName)
					Eventually(session).Should(Exit(1))
					Expect(session.Out).To(Say("FAILED"))
					Expect(session.Err).To(Say("Security group 'some-other-security-group' not found\\."))
				})
			})
		})
	})

	Context("when the security group, org, and space exist", func() {
		var (
			orgName      string
			spaceName    string
			secGroup     helpers.SecurityGroup
			secGroupName string
			username     string
		)

		BeforeEach(func() {
			username, _ = helpers.GetCredentials()

			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()
			helpers.CreateOrgAndSpace(orgName, spaceName)

			secGroupName = helpers.NewSecGroupName()
			secGroup = helpers.NewSecurityGroup(secGroupName, "tcp", "127.0.0.1", "8443", "some-description")
			secGroup.Create()

		})

		AfterEach(func() {
			secGroup.Delete()
		})

		Context("when the space isn't bound to the security group", func() {
			It("successfully runs the command", func() {
				session := helpers.CF("unbind-security-group", secGroupName, orgName, spaceName)
				Eventually(session.Out).Should(Say("Unbinding security group %s from %s/%s as %s", secGroupName, orgName, spaceName, username))
				Eventually(session.Out).Should(Say("OK\n\n"))
				Eventually(session.Out).Should(Say("TIP: Changes will not apply to existing running applications until they are restarted\\."))
				Eventually(session).Should(Exit(0))
			})
		})

		Context("when a space is bound to a security group", func() {
			BeforeEach(func() {
				Eventually(helpers.CF("bind-security-group", secGroupName, orgName, spaceName)).Should(Exit(0))
			})

			Context("when the org and space are not provided", func() {
				BeforeEach(func() {
					helpers.TargetOrgAndSpace(orgName, spaceName)
				})

				It("successfully unbinds the space from the security group", func() {
					session := helpers.CF("unbind-security-group", secGroupName)
					Eventually(session.Out).Should(Say("Unbinding security group %s from %s/%s as %s", secGroupName, orgName, spaceName, username))
					Eventually(session.Out).Should(Say("OK\n\n"))
					Eventually(session.Out).Should(Say("TIP: Changes will not apply to existing running applications until they are restarted\\."))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when the org and space are provided", func() {
				BeforeEach(func() {
					helpers.ClearTarget()
				})

				It("successfully unbinds the space from the security group", func() {
					session := helpers.CF("unbind-security-group", secGroupName, orgName, spaceName)
					Eventually(session.Out).Should(Say("Unbinding security group %s from %s/%s as %s", secGroupName, orgName, spaceName, username))
					Eventually(session.Out).Should(Say("OK\n\n"))
					Eventually(session.Out).Should(Say("TIP: Changes will not apply to existing running applications until they are restarted\\."))
					Eventually(session).Should(Exit(0))
				})
			})
		})
	})
})
