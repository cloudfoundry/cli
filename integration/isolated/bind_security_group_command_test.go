package isolated

import (
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	. "github.com/onsi/gomega/ghttp"
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
		Context("when --help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("bind-security-group", "--help")
				Eventually(session.Out).Should(Say("NAME:"))
				Eventually(session.Out).Should(Say("\\s+bind-security-group - Bind a security group to a particular space, or all existing spaces of an org"))
				Eventually(session.Out).Should(Say("USAGE:"))
				Eventually(session.Out).Should(Say("\\s+cf bind-security-group SECURITY_GROUP ORG \\[SPACE\\] \\[--lifecycle \\(running \\| staging\\)\\]"))
				Eventually(session.Out).Should(Say("TIP: Changes require an app restart \\(for running\\) or restage \\(for staging\\) to apply to existing applications\\."))
				Eventually(session.Out).Should(Say("OPTIONS:"))
				Eventually(session.Out).Should(Say("\\s+--lifecycle      Lifecycle phase the group applies to \\(Default: running\\)"))
				Eventually(session.Out).Should(Say("SEE ALSO:"))
				Eventually(session.Out).Should(Say("\\s+apps, bind-running-security-group, bind-staging-security-group, restart, security-groups"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("when the lifecycle flag is invalid", func() {
		It("outputs a message and usage", func() {
			session := helpers.CF("bind-security-group", secGroupName, someOrgName, "--lifecycle", "invalid")
			Eventually(session.Err).Should(Say("Incorrect Usage: Invalid value `invalid' for option `--lifecycle'. Allowed values are: running or staging"))
			Eventually(session.Out).Should(Say("USAGE:"))
			Eventually(session).Should(Exit(1))
		})
	})

	Context("when the lifecycle flag has no argument", func() {
		It("outputs a message and usage", func() {
			session := helpers.CF("bind-security-group", secGroupName, someOrgName, "--lifecycle")
			Eventually(session.Err).Should(Say("Incorrect Usage: expected argument for flag `--lifecycle'"))
			Eventually(session.Out).Should(Say("USAGE:"))
			Eventually(session).Should(Exit(1))
		})
	})

	Context("when the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(false, false, ReadOnlyOrg, "bind-security-group", "security-group-name", "org-name", "space-name")
		})
	})

	Context("when the server's API version is too low", func() {
		var server *Server

		BeforeEach(func() {
			server = NewTLSServer()
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/v2/info"),
					RespondWith(http.StatusOK, `{"api_version":"2.34.0"}`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/v2/info"),
					RespondWith(http.StatusOK, fmt.Sprintf(`{"api_version":"2.34.0", "authorization_endpoint": "%s"}`, server.URL())),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/login"),
					RespondWith(http.StatusOK, `{}`),
				),
			)
			Eventually(helpers.CF("api", server.URL(), "--skip-ssl-validation")).Should(Exit(0))
		})

		AfterEach(func() {
			server.Close()
		})

		It("reports an error with a minimum-version message", func() {
			session := helpers.CF("bind-security-group", secGroupName, orgName, spaceName1, "--lifecycle", "staging")

			Eventually(session.Err).Should(Say("Lifecycle value 'staging' requires CF API version 2\\.68\\.0\\ or higher. Your target is 2\\.34\\.0\\."))
			Eventually(session).Should(Exit(1))
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
				session := helpers.CF("bind-security-group", secGroupName)
				Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `ORG` was not provided"))
				Eventually(session.Out).Should(Say("USAGE:"))
				Eventually(session).Should(Exit(1))
			})
		})
	})

	Context("when the security group doesn't exist", func() {
		It("fails with a security group not found message", func() {
			session := helpers.CF("bind-security-group", "some-security-group-that-doesn't-exist", someOrgName)
			Eventually(session.Err).Should(Say("Security group 'some-security-group-that-doesn't-exist' not found."))
			Eventually(session.Out).Should(Say("FAILED"))
			Eventually(session).Should(Exit(1))
		})
	})

	Context("when the security group exists", func() {
		var someSecurityGroup helpers.SecurityGroup

		BeforeEach(func() {
			someSecurityGroup = helpers.NewSecurityGroup(secGroupName, "tcp", "0.0.0.0/0", "53", "")
			someSecurityGroup.Create()
		})

		Context("when the org doesn't exist", func() {
			It("fails with an org not found message", func() {
				session := helpers.CF("bind-security-group", secGroupName, someOrgName)
				Eventually(session.Err).Should(Say("Organization '%s' not found.", someOrgName))
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the org exists", func() {
			BeforeEach(func() {
				helpers.CreateOrg(orgName)
				helpers.TargetOrg(orgName)
			})

			AfterEach(func() {
				helpers.QuickDeleteOrg(orgName)
			})

			Context("when the space doesn't exist", func() {
				It("fails with a space not found message", func() {
					session := helpers.CF("bind-security-group", secGroupName, orgName, "space-doesnt-exist")
					Eventually(session.Err).Should(Say("Space 'space-doesnt-exist' not found."))
					Eventually(session.Out).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				})
			})

			Context("when there are no spaces in this org", func() {
				It("does not bind the security group to any space", func() {
					session := helpers.CF("bind-security-group", secGroupName, orgName)
					Consistently(session.Out).ShouldNot(Say("Assigning security group"))
					Consistently(session.Out).ShouldNot(Say("OK"))
					Eventually(session.Out).Should(Say("TIP: Changes require an app restart \\(for running\\) or restage \\(for staging\\) to apply to existing applications\\."))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when there are spaces in this org", func() {
				BeforeEach(func() {
					helpers.CreateSpace(spaceName1)
					helpers.CreateSpace(spaceName2)
				})

				Context("when the lifecycle flag is not set", func() {
					Context("when binding to all spaces in an org", func() {
						It("binds the security group to each space", func() {
							session := helpers.CF("bind-security-group", secGroupName, orgName)
							userName, _ := helpers.GetCredentials()
							Eventually(session.Out).Should(Say("Assigning security group %s to space INTEGRATION-SPACE.* in org %s as %s\\.\\.\\.", secGroupName, orgName, userName))
							Eventually(session.Out).Should(Say("OK"))
							Eventually(session.Out).Should(Say("Assigning security group %s to space INTEGRATION-SPACE.* in org %s as %s\\.\\.\\.", secGroupName, orgName, userName))
							Eventually(session.Out).Should(Say("OK"))
							Eventually(session.Out).Should(Say("TIP: Changes require an app restart \\(for running\\) or restage \\(for staging\\) to apply to existing applications\\."))
							Eventually(session).Should(Exit(0))
						})
					})
				})

				Context("when binding to a particular space", func() {
					It("binds the security group to the space", func() {
						session := helpers.CF("bind-security-group", secGroupName, orgName, spaceName1)
						userName, _ := helpers.GetCredentials()
						Eventually(session.Out).Should(Say("Assigning security group %s to space %s in org %s as %s\\.\\.\\.", secGroupName, spaceName1, orgName, userName))
						Eventually(session.Out).Should(Say("OK"))
						Eventually(session.Out).Should(Say("TIP: Changes require an app restart \\(for running\\) or restage \\(for staging\\) to apply to existing applications\\."))
						Eventually(session).Should(Exit(0))
					})
				})

				Context("when the lifecycle flag is running", func() {
					Context("when binding to a particular space", func() {
						It("binds the security group to the space", func() {
							session := helpers.CF("bind-security-group", secGroupName, orgName, spaceName1, "--lifecycle", "running")
							userName, _ := helpers.GetCredentials()
							Eventually(session.Out).Should(Say("Assigning security group %s to space %s in org %s as %s\\.\\.\\.", secGroupName, spaceName1, orgName, userName))
							Eventually(session.Out).Should(Say("OK"))
							Eventually(session.Out).Should(Say("TIP: Changes require an app restart \\(for running\\) or restage \\(for staging\\) to apply to existing applications\\."))
							Eventually(session).Should(Exit(0))
						})
					})
				})

				Context("when the lifecycle flag is staging", func() {
					Context("when binding to all spaces in an org", func() {
						It("binds the security group to each space", func() {
							session := helpers.CF("bind-security-group", secGroupName, orgName, "--lifecycle", "staging")
							userName, _ := helpers.GetCredentials()
							Eventually(session.Out).Should(Say("Assigning security group %s to space INTEGRATION-SPACE.* in org %s as %s\\.\\.\\.", secGroupName, orgName, userName))
							Eventually(session.Out).Should(Say("OK"))
							Eventually(session.Out).Should(Say("Assigning security group %s to space INTEGRATION-SPACE.* in org %s as %s\\.\\.\\.", secGroupName, orgName, userName))
							Eventually(session.Out).Should(Say("OK"))
							Eventually(session.Out).Should(Say("TIP: Changes require an app restart \\(for running\\) or restage \\(for staging\\) to apply to existing applications\\."))
							Eventually(session).Should(Exit(0))
						})
					})

					Context("when binding to a particular space", func() {
						It("binds the security group to the space", func() {
							session := helpers.CF("bind-security-group", secGroupName, orgName, spaceName1, "--lifecycle", "staging")
							userName, _ := helpers.GetCredentials()
							Eventually(session.Out).Should(Say("Assigning security group %s to space %s in org %s as %s\\.\\.\\.", secGroupName, spaceName1, orgName, userName))
							Eventually(session.Out).Should(Say("OK"))
							Eventually(session.Out).Should(Say("TIP: Changes require an app restart \\(for running\\) or restage \\(for staging\\) to apply to existing applications\\."))
							Eventually(session).Should(Exit(0))
						})
					})
				})
			})
		})
	})
})
