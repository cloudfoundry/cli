package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/resources"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("bind-security-group command", func() {
	var (
		orgName      string
		secGroupName string
		someOrgName  string
		spaceName1   string
		spaceName2   string
	)

	BeforeEach(func() {
		orgName = helpers.NewOrgName()
		secGroupName = helpers.NewSecurityGroupName()
		someOrgName = helpers.NewOrgName()
		spaceName1 = helpers.NewSpaceName()
		spaceName2 = helpers.NewSpaceName()

		helpers.LoginCF()
	})

	Describe("help", func() {
		When("--help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("bind-security-group", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say(`\s+bind-security-group - Bind a security group to a particular space, or all existing spaces of an org`))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say(`\s+cf bind-security-group SECURITY_GROUP ORG \[--lifecycle \(running \| staging\)\] \[--space SPACE\]`))
				Eventually(session).Should(Say(`TIP: Changes require an app restart \(for running\) or restage \(for staging\) to apply to existing applications\.`))
				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say(`\s+--lifecycle      Lifecycle phase the group applies to\. \(Default: running\)`))
				Eventually(session).Should(Say(`\s+--space          Space to bind the security group to\. \(Default: all existing spaces in org\)`))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say(`\s+apps, bind-running-security-group, bind-staging-security-group, restart, security-groups`))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the lifecycle flag is invalid", func() {
		It("outputs a message and usage", func() {
			session := helpers.CF("bind-security-group", secGroupName, someOrgName, "--lifecycle", "invalid")
			Eventually(session.Err).Should(Say("Incorrect Usage: Invalid value `invalid' for option `--lifecycle'. Allowed values are: running or staging"))
			Eventually(session).Should(Say("USAGE:"))
			Eventually(session).Should(Exit(1))
		})
	})

	When("the lifecycle flag has no argument", func() {
		It("outputs a message and usage", func() {
			session := helpers.CF("bind-security-group", secGroupName, someOrgName, "--lifecycle")
			Eventually(session.Err).Should(Say("Incorrect Usage: expected argument for flag `--lifecycle'"))
			Eventually(session).Should(Say("USAGE:"))
			Eventually(session).Should(Exit(1))
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(false, false, ReadOnlyOrg, "bind-security-group", "security-group-name", "org-name", "--space", "space-name")
		})
	})

	When("the input is invalid", func() {
		When("the security group is not provided", func() {
			It("fails with an incorrect usage message and displays help", func() {
				session := helpers.CF("bind-security-group")
				Eventually(session.Err).Should(Say("Incorrect Usage: the required arguments `SECURITY_GROUP` and `ORG` were not provided"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the org is not provided", func() {
			It("fails with an incorrect usage message and displays help", func() {
				session := helpers.CF("bind-security-group", secGroupName)
				Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `ORG` was not provided"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Exit(1))
			})
		})
	})

	When("the security group doesn't exist", func() {
		It("fails with a security group not found message", func() {
			session := helpers.CF("bind-security-group", "some-security-group-that-doesn't-exist", someOrgName)
			Eventually(session.Err).Should(Say("Security group 'some-security-group-that-doesn't-exist' not found."))
			Eventually(session).Should(Say("FAILED"))
			Eventually(session).Should(Exit(1))
		})
	})

	When("the security group exists", func() {
		var (
			someSecurityGroup resources.SecurityGroup
			ports             string
			description       string
		)

		BeforeEach(func() {
			ports = "53"
			description = "SG"
			someSecurityGroup = helpers.NewSecurityGroup(secGroupName, "tcp", "0.0.0.0/0", &ports, &description)
			helpers.CreateSecurityGroup(someSecurityGroup)
		})

		When("the org doesn't exist", func() {
			It("fails with an org not found message", func() {
				session := helpers.CF("bind-security-group", secGroupName, someOrgName)
				Eventually(session.Err).Should(Say("Organization '%s' not found.", someOrgName))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the org exists", func() {
			BeforeEach(func() {
				helpers.CreateOrg(orgName)
				helpers.TargetOrg(orgName)
			})

			AfterEach(func() {
				helpers.QuickDeleteOrg(orgName)
			})

			When("the space doesn't exist", func() {
				It("fails with a space not found message", func() {
					session := helpers.CF("bind-security-group", secGroupName, orgName, "--space", "space-doesnt-exist")
					userName, _ := helpers.GetCredentials()
					Eventually(session).Should(Say("Assigning running security group %s to space space-doesnt-exist in org %s as %s...", secGroupName, orgName, userName))
					Eventually(session.Err).Should(Say("Space 'space-doesnt-exist' not found."))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("there are no spaces in this org", func() {
				It("does not bind the security group to any space", func() {
					session := helpers.CF("bind-security-group", secGroupName, orgName)
					userName, _ := helpers.GetCredentials()
					Eventually(session).Should(Say("Assigning running security group %s to all spaces in org %s as %s...", secGroupName, orgName, userName))
					Eventually(session).Should(Say("No spaces in org %s.", orgName))
					Consistently(session).ShouldNot(Say("OK"))
					Eventually(session).Should(Say(`TIP: Changes require an app restart \(for running\) or restage \(for staging\) to apply to existing applications\.`))
					Eventually(session).Should(Exit(0))
				})
			})

			When("there are spaces in this org", func() {
				BeforeEach(func() {
					helpers.CreateSpace(spaceName1)
					helpers.CreateSpace(spaceName2)
				})

				When("the lifecycle flag is not set", func() {
					When("binding to all spaces in an org", func() {
						It("binds the security group to each space", func() {
							session := helpers.CF("bind-security-group", secGroupName, orgName)
							userName, _ := helpers.GetCredentials()
							Eventually(session).Should(Say(`Assigning running security group %s to space INTEGRATION-SPACE.* in org %s as %s\.\.\.`, secGroupName, orgName, userName))
							Eventually(session).Should(Say("OK"))
							Eventually(session).Should(Say(`Assigning running security group %s to space INTEGRATION-SPACE.* in org %s as %s\.\.\.`, secGroupName, orgName, userName))
							Eventually(session).Should(Say("OK"))
							Eventually(session).Should(Say(`TIP: Changes require an app restart \(for running\) or restage \(for staging\) to apply to existing applications\.`))
							Eventually(session).Should(Exit(0))
						})
					})
				})

				When("binding to a particular space", func() {
					It("binds the security group to the space", func() {
						session := helpers.CF("bind-security-group", secGroupName, orgName, "--space", spaceName1)
						userName, _ := helpers.GetCredentials()
						Eventually(session).Should(Say(`Assigning running security group %s to space %s in org %s as %s\.\.\.`, secGroupName, spaceName1, orgName, userName))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Say(`TIP: Changes require an app restart \(for running\) or restage \(for staging\) to apply to existing applications\.`))
						Eventually(session).Should(Exit(0))
					})
				})

				When("the lifecycle flag is running", func() {
					When("binding to a particular space", func() {
						It("binds the security group to the space", func() {
							session := helpers.CF("bind-security-group", secGroupName, orgName, "--space", spaceName1, "--lifecycle", "running")
							userName, _ := helpers.GetCredentials()
							Eventually(session).Should(Say(`Assigning running security group %s to space %s in org %s as %s\.\.\.`, secGroupName, spaceName1, orgName, userName))
							Eventually(session).Should(Say("OK"))
							Eventually(session).Should(Say(`TIP: Changes require an app restart \(for running\) or restage \(for staging\) to apply to existing applications\.`))
							Eventually(session).Should(Exit(0))
						})
					})
				})

				When("the lifecycle flag is staging", func() {
					When("binding to all spaces in an org", func() {
						It("binds the security group to each space", func() {
							session := helpers.CF("bind-security-group", secGroupName, orgName, "--lifecycle", "staging")
							userName, _ := helpers.GetCredentials()
							Eventually(session).Should(Say(`Assigning staging security group %s to space INTEGRATION-SPACE.* in org %s as %s\.\.\.`, secGroupName, orgName, userName))
							Eventually(session).Should(Say("OK"))
							Eventually(session).Should(Say(`Assigning staging security group %s to space INTEGRATION-SPACE.* in org %s as %s\.\.\.`, secGroupName, orgName, userName))
							Eventually(session).Should(Say("OK"))
							Eventually(session).Should(Say(`TIP: Changes require an app restart \(for running\) or restage \(for staging\) to apply to existing applications\.`))
							Eventually(session).Should(Exit(0))
						})
					})

					When("binding to a particular space", func() {
						It("binds the security group to the space", func() {
							session := helpers.CF("bind-security-group", secGroupName, orgName, "--space", spaceName1, "--lifecycle", "staging")
							userName, _ := helpers.GetCredentials()
							Eventually(session).Should(Say(`Assigning staging security group %s to space %s in org %s as %s\.\.\.`, secGroupName, spaceName1, orgName, userName))
							Eventually(session).Should(Say("OK"))
							Eventually(session).Should(Say(`TIP: Changes require an app restart \(for running\) or restage \(for staging\) to apply to existing applications\.`))
							Eventually(session).Should(Exit(0))
						})
					})
				})
			})
		})
	})
})
