package global

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("unbind-security-group command", func() {
	var (
		orgName           string
		securityGroupName string
		spaceName         string
	)

	BeforeEach(func() {
		orgName = helpers.NewOrgName()
		securityGroupName = helpers.NewSecurityGroupName()
		spaceName = helpers.NewSpaceName()

		helpers.LoginCF()
	})

	Describe("help", func() {
		When("--help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("unbind-security-group", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say(`\s+unbind-security-group - Unbind a security group from a space`))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say(`\s+cf unbind-security-group SECURITY_GROUP ORG SPACE \[--lifecycle \(running \| staging\)\]`))
				Eventually(session).Should(Say(`TIP: If Dynamic ASG's are enabled, changes will automatically apply for running and staging applications. Otherwise, changes will require an app restart \(for running\) or restage \(for staging\) to apply to existing applications\.`))
				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say(`\s+--lifecycle      Lifecycle phase the group applies to \(Default: running\)`))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say(`\s+apps, restart, security-groups`))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the lifecycle flag is invalid", func() {
		It("outputs a message and usage", func() {
			session := helpers.CF("unbind-security-group", securityGroupName, "some-org", "some-space", "--lifecycle", "invalid")
			Eventually(session.Err).Should(Say("Incorrect Usage: Invalid value `invalid' for option `--lifecycle'. Allowed values are: running or staging"))
			Eventually(session).Should(Say("USAGE:"))
			Eventually(session).Should(Exit(1))
		})
	})

	When("the lifecycle flag has no argument", func() {
		It("outputs a message and usage", func() {
			session := helpers.CF("unbind-security-group", securityGroupName, "some-org", "some-space", "--lifecycle")
			Eventually(session.Err).Should(Say("Incorrect Usage: expected argument for flag `--lifecycle'"))
			Eventually(session).Should(Say("USAGE:"))
			Eventually(session).Should(Exit(1))
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(false, false, ReadOnlyOrg, "unbind-security-group", securityGroupName, "some-org", "some-space")
		})
	})

	When("the input is invalid", func() {
		When("the security group is not provided", func() {
			It("fails with an incorrect usage message and displays help", func() {
				session := helpers.CF("unbind-security-group")
				Eventually(session.Err).Should(Say("Incorrect Usage: the required arguments `SECURITY_GROUP`, `ORG` and `SPACE` were not provided"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the space is not provided", func() {
			It("fails with an incorrect usage message and displays help", func() {
				session := helpers.CF("unbind-security-group", securityGroupName, "some-org")
				Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `SPACE` was not provided"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Exit(1))
			})
		})
	})

	When("the security group doesn't exist", func() {
		BeforeEach(func() {
			helpers.CreateOrgAndSpace(orgName, spaceName)
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		It("fails with a 'security group not found' message", func() {
			session := helpers.CF("unbind-security-group", "some-other-security-group", orgName, spaceName)
			Eventually(session).Should(Say("FAILED"))
			Eventually(session.Err).Should(Say(`Security group 'some-other-security-group' not found\.`))
			Eventually(session).Should(Exit(1))
		})
	})

	When("the security group exists", func() {
		BeforeEach(func() {
			port := "8443"
			description := "some-description"
			someSecurityGroup := helpers.NewSecurityGroup(securityGroupName, "tcp", "127.0.0.1", &port, &description)
			helpers.CreateSecurityGroup(someSecurityGroup)
		})

		When("the org doesn't exist", func() {
			It("fails with an 'org not found' message", func() {
				session := helpers.CF("unbind-security-group", securityGroupName, "some-other-org", "some-other-space")
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say(`Organization 'some-other-org' not found\.`))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the org exists", func() {
			var username string

			BeforeEach(func() {
				username, _ = helpers.GetCredentials()

				helpers.CreateOrg(orgName)
				helpers.TargetOrg(orgName)
			})

			AfterEach(func() {
				helpers.QuickDeleteOrg(orgName)
			})

			When("the space doesn't exist", func() {
				It("fails with a 'space not found' message", func() {
					session := helpers.CF("unbind-security-group", securityGroupName, orgName, "some-other-space")
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say(`Space 'some-other-space' not found\.`))
					Eventually(session).Should(Exit(1))
				})
			})

			When("the space exists", func() {
				BeforeEach(func() {
					helpers.CreateSpace(spaceName)
				})

				When("the space isn't bound to the security group in any lifecycle", func() {
					It("successfully runs the command", func() {
						session := helpers.CF("unbind-security-group", securityGroupName, orgName, spaceName)
						Eventually(session).Should(Say(`Unbinding running security group %s from org %s / space %s as %s\.\.\.`, securityGroupName, orgName, spaceName, username))
						Eventually(session.Err).Should(Say(`Security group %s not bound to space %s for lifecycle phase 'running'\.`, securityGroupName, spaceName))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Say(`TIP: If Dynamic ASG's are enabled, changes will automatically apply for running and staging applications. Otherwise, changes will require an app restart \(for running\) or restage \(for staging\) to apply to existing applications\.`))
						Eventually(session).Should(Exit(0))
					})
				})

				When("a space is bound to a security group in the running lifecycle", func() {
					BeforeEach(func() {
						Eventually(helpers.CF("bind-security-group", securityGroupName, orgName, "--space", spaceName)).Should(Exit(0))
					})

					When("the lifecycle flag is not set", func() {
						BeforeEach(func() {
							helpers.ClearTarget()
						})

						It("successfully unbinds the space from the security group", func() {
							session := helpers.CF("unbind-security-group", securityGroupName, orgName, spaceName)
							Eventually(session).Should(Say(`Unbinding running security group %s from org %s / space %s as %s\.\.\.`, securityGroupName, orgName, spaceName, username))
							Eventually(session).Should(Say("OK"))
							Eventually(session).Should(Say(`TIP: If Dynamic ASG's are enabled, changes will automatically apply for running and staging applications. Otherwise, changes will require an app restart \(for running\) or restage \(for staging\) to apply to existing applications\.`))
							Eventually(session).Should(Exit(0))
						})
					})

					When("the lifecycle flag is running", func() {
						BeforeEach(func() {
							helpers.ClearTarget()
						})

						It("successfully unbinds the space from the security group", func() {
							session := helpers.CF("unbind-security-group", securityGroupName, orgName, spaceName, "--lifecycle", "running")
							Eventually(session).Should(Say(`Unbinding running security group %s from org %s / space %s as %s\.\.\.`, securityGroupName, orgName, spaceName, username))
							Eventually(session).Should(Say("OK"))
							Eventually(session).Should(Say(`TIP: If Dynamic ASG's are enabled, changes will automatically apply for running and staging applications. Otherwise, changes will require an app restart \(for running\) or restage \(for staging\) to apply to existing applications\.`))
							Eventually(session).Should(Exit(0))
						})
					})

					When("the lifecycle flag is staging", func() {
						BeforeEach(func() {
							helpers.ClearTarget()
						})

						It("displays an error and exits 1", func() {
							session := helpers.CF("unbind-security-group", securityGroupName, orgName, spaceName, "--lifecycle", "staging")
							Eventually(session).Should(Say(`Unbinding staging security group %s from org %s / space %s as %s\.\.\.`, securityGroupName, orgName, spaceName, username))
							Eventually(session).Should(Say("OK"))
							Eventually(session.Err).Should(Say(`Security group %s not bound to space %s for lifecycle phase 'staging'\.`, securityGroupName, spaceName))
							Eventually(session).Should(Exit(0))
						})
					})
				})

				When("a space is bound to a security group in the staging lifecycle", func() {
					BeforeEach(func() {
						Eventually(helpers.CF("bind-security-group", securityGroupName, orgName, "--space", spaceName, "--lifecycle", "staging")).Should(Exit(0))
					})

					When("the lifecycle flag is not set", func() {
						BeforeEach(func() {
							helpers.ClearTarget()
						})

						It("displays an error and exits 1", func() {
							session := helpers.CF("unbind-security-group", securityGroupName, orgName, spaceName)
							Eventually(session).Should(Say(`Unbinding running security group %s from org %s / space %s as %s\.\.\.`, securityGroupName, orgName, spaceName, username))
							Eventually(session).Should(Say("OK"))
							Eventually(session.Err).Should(Say(`Security group %s not bound to space %s for lifecycle phase 'running'\.`, securityGroupName, spaceName))
							Eventually(session).Should(Exit(0))
						})
					})

					When("the lifecycle flag is running", func() {
						BeforeEach(func() {
							helpers.ClearTarget()
						})

						It("displays an error and exits 1", func() {
							session := helpers.CF("unbind-security-group", securityGroupName, orgName, spaceName, "--lifecycle", "running")
							Eventually(session).Should(Say(`Unbinding running security group %s from org %s / space %s as %s\.\.\.`, securityGroupName, orgName, spaceName, username))
							Eventually(session).Should(Say("OK"))
							Eventually(session.Err).Should(Say(`Security group %s not bound to space %s for lifecycle phase 'running'\.`, securityGroupName, spaceName))
							Eventually(session).Should(Exit(0))
						})
					})

					When("the lifecycle flag is staging", func() {
						BeforeEach(func() {
							helpers.ClearTarget()
						})

						It("successfully unbinds the space from the security group", func() {
							session := helpers.CF("unbind-security-group", securityGroupName, orgName, spaceName, "--lifecycle", "staging")
							Eventually(session).Should(Say(`Unbinding staging security group %s from org %s / space %s as %s\.\.\.`, securityGroupName, orgName, spaceName, username))
							Eventually(session).Should(Say("OK"))
							Eventually(session).Should(Say(`TIP: If Dynamic ASG's are enabled, changes will automatically apply for running and staging applications. Otherwise, changes will require an app restart \(for running\) or restage \(for staging\) to apply to existing applications\.`))
							Eventually(session).Should(Exit(0))
						})
					})
				})
			})
		})
	})
})
