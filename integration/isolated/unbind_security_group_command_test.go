package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
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
		Context("when --help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("unbind-security-group", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("\\s+unbind-security-group - Unbind a security group from a space"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("\\s+cf unbind-security-group SECURITY_GROUP ORG SPACE \\[--lifecycle \\(running \\| staging\\)\\]"))
				Eventually(session).Should(Say("TIP: Changes require an app restart \\(for running\\) or restage \\(for staging\\) to apply to existing applications\\."))
				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say("\\s+--lifecycle      Lifecycle phase the group applies to \\(Default: running\\)"))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("\\s+apps, restart, security-groups"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("when the lifecycle flag is invalid", func() {
		It("outputs a message and usage", func() {
			session := helpers.CF("unbind-security-group", securityGroupName, "some-org", "--lifecycle", "invalid")
			Eventually(session.Err).Should(Say("Incorrect Usage: Invalid value `invalid' for option `--lifecycle'. Allowed values are: running or staging"))
			Eventually(session).Should(Say("USAGE:"))
			Eventually(session).Should(Exit(1))
		})
	})

	Context("when the lifecycle flag has no argument", func() {
		It("outputs a message and usage", func() {
			session := helpers.CF("unbind-security-group", securityGroupName, "some-org", "--lifecycle")
			Eventually(session.Err).Should(Say("Incorrect Usage: expected argument for flag `--lifecycle'"))
			Eventually(session).Should(Say("USAGE:"))
			Eventually(session).Should(Exit(1))
		})
	})

	Context("when the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "unbind-security-group", securityGroupName)
		})
	})

	Context("when the input is invalid", func() {
		Context("when the security group is not provided", func() {
			It("fails with an incorrect usage message and displays help", func() {
				session := helpers.CF("unbind-security-group")
				Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `SECURITY_GROUP` was not provided"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the space is not provided", func() {
			It("fails with an incorrect usage message and displays help", func() {
				session := helpers.CF("unbind-security-group", securityGroupName, "some-org")
				Eventually(session.Err).Should(Say("Incorrect Usage: the required arguments `SECURITY_GROUP`, `ORG`, and `SPACE` were not provided"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Exit(1))
			})
		})
	})

	Context("when the security group doesn't exist", func() {
		BeforeEach(func() {
			helpers.CreateOrgAndSpace(orgName, spaceName)
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		It("fails with a 'security group not found' message", func() {
			session := helpers.CF("unbind-security-group", "some-other-security-group", orgName, spaceName)
			Eventually(session).Should(Say("FAILED"))
			Eventually(session.Err).Should(Say("Security group 'some-other-security-group' not found\\."))
			Eventually(session).Should(Exit(1))
		})
	})

	Context("when the security group exists", func() {
		BeforeEach(func() {
			someSecurityGroup := helpers.NewSecurityGroup(securityGroupName, "tcp", "127.0.0.1", "8443", "some-description")
			someSecurityGroup.Create()
		})

		Context("when the org doesn't exist", func() {
			It("fails with an 'org not found' message", func() {
				session := helpers.CF("unbind-security-group", securityGroupName, "some-other-org", "some-other-space")
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("Organization 'some-other-org' not found\\."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the org exists", func() {
			var username string

			BeforeEach(func() {
				username, _ = helpers.GetCredentials()

				helpers.CreateOrg(orgName)
				helpers.TargetOrg(orgName)
			})

			AfterEach(func() {
				helpers.QuickDeleteOrg(orgName)
			})

			Context("when the space doesn't exist", func() {
				It("fails with a 'space not found' message", func() {
					session := helpers.CF("unbind-security-group", securityGroupName, orgName, "some-other-space")
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("Space 'some-other-space' not found\\."))
					Eventually(session).Should(Exit(1))
				})
			})

			Context("when the space exists", func() {
				BeforeEach(func() {
					helpers.CreateSpace(spaceName)
				})

				Context("when the space isn't bound to the security group in any lifecycle", func() {
					It("successfully runs the command", func() {
						session := helpers.CF("unbind-security-group", securityGroupName, orgName, spaceName)
						Eventually(session).Should(Say("Unbinding security group %s from org %s / space %s as %s\\.\\.\\.", securityGroupName, orgName, spaceName, username))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Say("TIP: Changes require an app restart \\(for running\\) or restage \\(for staging\\) to apply to existing applications\\."))
						Eventually(session).Should(Exit(0))
					})
				})

				Context("when a space is bound to a security group in the running lifecycle", func() {
					BeforeEach(func() {
						Eventually(helpers.CF("bind-security-group", securityGroupName, orgName, spaceName)).Should(Exit(0))
					})

					Context("when the lifecycle flag is not set", func() {
						Context("when the org and space are not provided", func() {
							BeforeEach(func() {
								helpers.TargetOrgAndSpace(orgName, spaceName)
							})

							It("successfully unbinds the space from the security group", func() {
								session := helpers.CF("unbind-security-group", securityGroupName)
								Eventually(session).Should(Say("Unbinding security group %s from org %s / space %s as %s\\.\\.\\.", securityGroupName, orgName, spaceName, username))
								Eventually(session).Should(Say("OK"))
								Eventually(session).Should(Say("TIP: Changes require an app restart \\(for running\\) or restage \\(for staging\\) to apply to existing applications\\."))
								Eventually(session).Should(Exit(0))
							})
						})

						Context("when the org and space are provided", func() {
							BeforeEach(func() {
								helpers.ClearTarget()
							})

							It("successfully unbinds the space from the security group", func() {
								session := helpers.CF("unbind-security-group", securityGroupName, orgName, spaceName)
								Eventually(session).Should(Say("Unbinding security group %s from org %s / space %s as %s\\.\\.\\.", securityGroupName, orgName, spaceName, username))
								Eventually(session).Should(Say("OK"))
								Eventually(session).Should(Say("TIP: Changes require an app restart \\(for running\\) or restage \\(for staging\\) to apply to existing applications\\."))
								Eventually(session).Should(Exit(0))
							})
						})
					})

					Context("when the lifecycle flag is running", func() {
						Context("when the org and space are not provided", func() {
							BeforeEach(func() {
								helpers.TargetOrgAndSpace(orgName, spaceName)
							})

							It("successfully unbinds the space from the security group", func() {
								session := helpers.CF("unbind-security-group", securityGroupName, "--lifecycle", "running")
								Eventually(session).Should(Say("Unbinding security group %s from org %s / space %s as %s\\.\\.\\.", securityGroupName, orgName, spaceName, username))
								Eventually(session).Should(Say("OK"))
								Eventually(session).Should(Say("TIP: Changes require an app restart \\(for running\\) or restage \\(for staging\\) to apply to existing applications\\."))
								Eventually(session).Should(Exit(0))
							})
						})

						Context("when the org and space are provided", func() {
							BeforeEach(func() {
								helpers.ClearTarget()
							})

							It("successfully unbinds the space from the security group", func() {
								session := helpers.CF("unbind-security-group", securityGroupName, orgName, spaceName, "--lifecycle", "running")
								Eventually(session).Should(Say("Unbinding security group %s from org %s / space %s as %s\\.\\.\\.", securityGroupName, orgName, spaceName, username))
								Eventually(session).Should(Say("OK"))
								Eventually(session).Should(Say("TIP: Changes require an app restart \\(for running\\) or restage \\(for staging\\) to apply to existing applications\\."))
								Eventually(session).Should(Exit(0))
							})
						})
					})

					Context("when the lifecycle flag is staging", func() {
						Context("when the org and space are not provided", func() {
							BeforeEach(func() {
								helpers.TargetOrgAndSpace(orgName, spaceName)
							})

							It("displays an error and exits 1", func() {
								session := helpers.CF("unbind-security-group", securityGroupName, "--lifecycle", "staging")
								Eventually(session).Should(Say("Unbinding security group %s from org %s / space %s as %s\\.\\.\\.", securityGroupName, orgName, spaceName, username))
								Eventually(session).Should(Say("OK"))
								Eventually(session.Err).Should(Say("Security group %s not bound to this space for lifecycle phase 'staging'\\.", securityGroupName))
								Eventually(session).Should(Exit(0))
							})
						})

						Context("when the org and space are provided", func() {
							BeforeEach(func() {
								helpers.ClearTarget()
							})

							It("displays an error and exits 1", func() {
								session := helpers.CF("unbind-security-group", securityGroupName, orgName, spaceName, "--lifecycle", "staging")
								Eventually(session).Should(Say("Unbinding security group %s from org %s / space %s as %s\\.\\.\\.", securityGroupName, orgName, spaceName, username))
								Eventually(session).Should(Say("OK"))
								Eventually(session.Err).Should(Say("Security group %s not bound to this space for lifecycle phase 'staging'\\.", securityGroupName))
								Eventually(session).Should(Exit(0))
							})
						})
					})
				})

				Context("when a space is bound to a security group in the staging lifecycle", func() {
					BeforeEach(func() {
						Eventually(helpers.CF("bind-security-group", securityGroupName, orgName, spaceName, "--lifecycle", "staging")).Should(Exit(0))
					})

					Context("when the lifecycle flag is not set", func() {
						Context("when the org and space are not provided", func() {
							BeforeEach(func() {
								helpers.TargetOrgAndSpace(orgName, spaceName)
							})

							It("displays an error and exits 1", func() {
								session := helpers.CF("unbind-security-group", securityGroupName)
								Eventually(session).Should(Say("Unbinding security group %s from org %s / space %s as %s\\.\\.\\.", securityGroupName, orgName, spaceName, username))
								Eventually(session).Should(Say("OK"))
								Eventually(session.Err).Should(Say("Security group %s not bound to this space for lifecycle phase 'running'\\.", securityGroupName))
								Eventually(session).Should(Exit(0))
							})
						})

						Context("when the org and space are provided", func() {
							BeforeEach(func() {
								helpers.ClearTarget()
							})

							It("displays an error and exits 1", func() {
								session := helpers.CF("unbind-security-group", securityGroupName, orgName, spaceName)
								Eventually(session).Should(Say("Unbinding security group %s from org %s / space %s as %s\\.\\.\\.", securityGroupName, orgName, spaceName, username))
								Eventually(session).Should(Say("OK"))
								Eventually(session.Err).Should(Say("Security group %s not bound to this space for lifecycle phase 'running'\\.", securityGroupName))
								Eventually(session).Should(Exit(0))
							})
						})
					})

					Context("when the lifecycle flag is running", func() {
						Context("when the org and space are not provided", func() {
							BeforeEach(func() {
								helpers.TargetOrgAndSpace(orgName, spaceName)
							})

							It("displays an error and exits 1", func() {
								session := helpers.CF("unbind-security-group", securityGroupName, "--lifecycle", "running")
								Eventually(session).Should(Say("Unbinding security group %s from org %s / space %s as %s\\.\\.\\.", securityGroupName, orgName, spaceName, username))
								Eventually(session).Should(Say("OK"))
								Eventually(session.Err).Should(Say("Security group %s not bound to this space for lifecycle phase 'running'\\.", securityGroupName))
								Eventually(session).Should(Exit(0))
							})
						})

						Context("when the org and space are provided", func() {
							BeforeEach(func() {
								helpers.ClearTarget()
							})

							It("displays an error and exits 1", func() {
								session := helpers.CF("unbind-security-group", securityGroupName, orgName, spaceName, "--lifecycle", "running")
								Eventually(session).Should(Say("Unbinding security group %s from org %s / space %s as %s\\.\\.\\.", securityGroupName, orgName, spaceName, username))
								Eventually(session).Should(Say("OK"))
								Eventually(session.Err).Should(Say("Security group %s not bound to this space for lifecycle phase 'running'\\.", securityGroupName))
								Eventually(session).Should(Exit(0))
							})
						})
					})

					Context("when the lifecycle flag is staging", func() {
						Context("when the org and space are not provided", func() {
							BeforeEach(func() {
								helpers.TargetOrgAndSpace(orgName, spaceName)
							})

							It("successfully unbinds the space from the security group", func() {
								session := helpers.CF("unbind-security-group", securityGroupName, "--lifecycle", "staging")
								Eventually(session).Should(Say("Unbinding security group %s from org %s / space %s as %s\\.\\.\\.", securityGroupName, orgName, spaceName, username))
								Eventually(session).Should(Say("OK"))
								Eventually(session).Should(Say("TIP: Changes require an app restart \\(for running\\) or restage \\(for staging\\) to apply to existing applications\\."))
								Eventually(session).Should(Exit(0))
							})
						})

						Context("when the org and space are provided", func() {
							BeforeEach(func() {
								helpers.ClearTarget()
							})

							It("successfully unbinds the space from the security group", func() {
								session := helpers.CF("unbind-security-group", securityGroupName, orgName, spaceName, "--lifecycle", "staging")
								Eventually(session).Should(Say("Unbinding security group %s from org %s / space %s as %s\\.\\.\\.", securityGroupName, orgName, spaceName, username))
								Eventually(session).Should(Say("OK"))
								Eventually(session).Should(Say("TIP: Changes require an app restart \\(for running\\) or restage \\(for staging\\) to apply to existing applications\\."))
								Eventually(session).Should(Exit(0))
							})
						})
					})
				})
			})
		})
	})
})
