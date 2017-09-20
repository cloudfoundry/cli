package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/util/configv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
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
		spaceName = helpers.NewSpaceName()
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

	Context("when both the access and refresh tokens are invalid", func() {
		BeforeEach(func() {
			helpers.SetConfig(func(conf *configv3.Config) {
				conf.SetAccessToken("bearer eyJhbGciOiJSUzI1NiIsImtpZCI6ImtleS0xIiwidHlwIjoiSldUIn0.eyJqdGkiOiJlNzQyMjg1NjNjZjc0ZGQ0YTU5YTA1NTUyMWVlYzlhNCIsInN1YiI6IjhkN2IxZjRlLTJhNGQtNGQwNy1hYWE0LTdjOTVlZDFhN2YzNCIsInNjb3BlIjpbInJvdXRpbmcucm91dGVyX2dyb3Vwcy5yZWFkIiwiY2xvdWRfY29udHJvbGxlci5yZWFkIiwicGFzc3dvcmQud3JpdGUiLCJjbG91ZF9jb250cm9sbGVyLndyaXRlIiwib3BlbmlkIiwicm91dGluZy5yb3V0ZXJfZ3JvdXBzLndyaXRlIiwiZG9wcGxlci5maXJlaG9zZSIsInNjaW0ud3JpdGUiLCJzY2ltLnJlYWQiLCJjbG91ZF9jb250cm9sbGVyLmFkbWluIiwidWFhLnVzZXIiXSwiY2xpZW50X2lkIjoiY2YiLCJjaWQiOiJjZiIsImF6cCI6ImNmIiwiZ3JhbnRfdHlwZSI6InBhc3N3b3JkIiwidXNlcl9pZCI6IjhkN2IxZjRlLTJhNGQtNGQwNy1hYWE0LTdjOTVlZDFhN2YzNCIsIm9yaWdpbiI6InVhYSIsInVzZXJfbmFtZSI6ImFkbWluIiwiZW1haWwiOiJhZG1pbiIsInJldl9zaWciOiI2ZjZkM2Y1YyIsImlhdCI6MTQ4Njc2NDQxNywiZXhwIjoxNDg2NzY1MDE3LCJpc3MiOiJodHRwczovL3VhYS5ib3NoLWxpdGUuY29tL29hdXRoL3Rva2VuIiwiemlkIjoidWFhIiwiYXVkIjpbImNsb3VkX2NvbnRyb2xsZXIiLCJzY2ltIiwicGFzc3dvcmQiLCJjZiIsInVhYSIsIm9wZW5pZCIsImRvcHBsZXIiLCJyb3V0aW5nLnJvdXRlcl9ncm91cHMiXX0.AhQI_-u9VzkQ1Z7yzibq7dBWbb5ucTDtwaXjeCf4rakl7hJvQYWI1meO9PSUI8oVbArBgOu0aOU6mfzDE8dSyZ1qAD0mhL5_c2iLGXdqUaPlXrX9vxuJZh_8vMTlxAnJ02c6ixbWaPWujvEIuiLb-QWa0NTbR9RDNyw1MbOQkdQ")

				conf.SetRefreshToken("bb8f7b209ff74409877974bce5752412-r")
			})
		})

		It("tells the user to login and exits with 1", func() {
			session := helpers.CF("target", "-o", "some-org", "-s", "some-space")
			Eventually(session.Err).Should(Say("The token expired, was revoked, or the token ID is incorrect. Please log back in to re-authenticate."))
			Eventually(session.Out).Should(Say("FAILED"))
			Eventually(session).Should(Exit(1))
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
			DescribeTable("fails with not logged in message",
				func(args ...string) {
					helpers.LogoutCF()
					cmd := append([]string{"target"}, args...)
					session := helpers.CF(cmd...)
					Eventually(session.Err).Should(Say("Not logged in. Use 'cf login' to log in."))
					Eventually(session.Out).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				},

				Entry("when trying to target an org", "-o", "some-org"),
				Entry("when trying to target a space", "-s", "some-space"),
				Entry("when trying to target an org and space", "-o", "some-org", "-s", "some-space"),
				Entry("when trying to get the target"),
			)
		})
	})

	Context("when no arguments are provided", func() {
		Context("when *no* org and space are targeted", func() {
			It("displays current target information", func() {
				username, _ := helpers.GetCredentials()
				session := helpers.CF("target")
				Eventually(session.Out).Should(Say("api endpoint:\\s+%s", apiURL))
				Eventually(session.Out).Should(Say("api version:\\s+[\\d.]+"))
				Eventually(session.Out).Should(Say("user:\\s+%s", username))
				Eventually(session.Out).Should(Say("No org or space targeted, use 'cf target -o ORG -s SPACE'"))
				Eventually(session).Should(Exit(0))
			})
		})

		Context("when targeted to an org and space", func() {
			BeforeEach(func() {
				setupCF(ReadOnlyOrg, ReadOnlySpace)
			})

			It("displays current target information", func() {
				username, _ := helpers.GetCredentials()
				session := helpers.CF("target")
				Eventually(session.Out).Should(Say("api endpoint:\\s+%s", apiURL))
				Eventually(session.Out).Should(Say("api version:\\s+[\\d.]+"))
				Eventually(session.Out).Should(Say("user:\\s+%s", username))
				Eventually(session.Out).Should(Say("org:\\s+%s", ReadOnlyOrg))
				Eventually(session.Out).Should(Say("space:\\s+%s", ReadOnlySpace))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("when only an org argument is provided", func() {
		Context("when the org does not exist", func() {
			// We set targets to verify that the target command
			// preserves existing targets in failure
			BeforeEach(func() {
				setupCF(ReadOnlyOrg, ReadOnlySpace)
			})

			It("displays org not found, exits 1, and clears existing targets", func() {
				session := helpers.CF("target", "-o", orgName)
				Eventually(session.Err).Should(Say("Organization '%s' not found", orgName))
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))

				session = helpers.CF("target")
				Eventually(session.Out).Should(Say("No org or space targeted, use 'cf target -o ORG -s SPACE'"))
				Eventually(session).Should(Exit(0))
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

			Context("when there are no spaces in the org", func() {
				BeforeEach(func() {
					helpers.ClearTarget()
				})

				It("only targets the org and exits 0", func() {
					username, _ := helpers.GetCredentials()
					session := helpers.CF("target", "-o", orgName)
					Eventually(session.Out).Should(Say("api endpoint:\\s+%s", apiURL))
					Eventually(session.Out).Should(Say("api version:\\s+[\\d.]+"))
					Eventually(session.Out).Should(Say("user:\\s+%s", username))
					Eventually(session.Out).Should(Say("org:\\s+%s", orgName))
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
					Eventually(session.Out).Should(Say("api endpoint:\\s+%s", apiURL))
					Eventually(session.Out).Should(Say("api version:\\s+[\\d.]+"))
					Eventually(session.Out).Should(Say("user:\\s+%s", username))
					Eventually(session.Out).Should(Say("org:\\s+%s", orgName))
					Eventually(session.Out).Should(Say("space:\\s+%s", spaceName))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when there are multiple spaces in the org", func() {
				BeforeEach(func() {
					helpers.CreateSpace(spaceName)
					helpers.CreateSpace(helpers.NewSpaceName())
					helpers.ClearTarget()
				})

				It("targets the org only and exits 0", func() {
					username, _ := helpers.GetCredentials()
					session := helpers.CF("target", "-o", orgName)
					Eventually(session.Out).Should(Say("api endpoint:\\s+%s", apiURL))
					Eventually(session.Out).Should(Say("api version:\\s+[\\d.]+"))
					Eventually(session.Out).Should(Say("user:\\s+%s", username))
					Eventually(session.Out).Should(Say("org:\\s+%s", orgName))
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
			BeforeEach(func() {
				helpers.LoginCF()
				Eventually(helpers.CF("target", "-o", ReadOnlyOrg)).Should(Exit(0))
			})

			Context("when the space exists", func() {
				It("targets the space and exits 0", func() {
					username, _ := helpers.GetCredentials()
					session := helpers.CF("target", "-s", ReadOnlySpace)
					Eventually(session.Out).Should(Say("api endpoint:\\s+%s", apiURL))
					Eventually(session.Out).Should(Say("api version:\\s+[\\d.]+"))
					Eventually(session.Out).Should(Say("user:\\s+%s", username))
					Eventually(session.Out).Should(Say("org:\\s+%s", ReadOnlyOrg))
					Eventually(session.Out).Should(Say("space:\\s+%s", ReadOnlySpace))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when the space does not exist", func() {
				It("displays space not found, exits 1, and clears existing targeted space", func() {
					session := helpers.CF("target", "-s", spaceName)
					Eventually(session.Err).Should(Say("Space '%s' not found.", spaceName))
					Eventually(session.Out).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))

					session = helpers.CF("target")
					Eventually(session.Out).Should(Say("org:\\s+%s", ReadOnlyOrg))
					Eventually(session.Out).Should(Say("No space targeted, use 'cf target -s SPACE'"))
					Eventually(session).Should(Exit(0))
				})
			})
		})

		Context("when there is not an existing targeted org", func() {
			It("displays org must be targeted first and exits 1", func() {
				session := helpers.CF("target", "-s", spaceName)
				Eventually(session.Err).Should(Say("No org targeted, use 'cf target -o ORG' to target an org."))
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
			It("displays org not found, exits 1, and clears existing targets", func() {
				session := helpers.CF("target", "-o", orgName, "-s", spaceName)
				Eventually(session.Err).Should(Say("Organization '%s' not found", orgName))
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))

				session = helpers.CF("target")
				Eventually(session.Out).Should(Say("No org or space targeted, use 'cf target -o ORG -s SPACE'"))
				Eventually(session).Should(Exit(0))
			})
		})

		Context("when the org exists", func() {
			BeforeEach(func() {
				helpers.CreateOrg(orgName)
			})

			AfterEach(func() {
				helpers.QuickDeleteOrg(orgName)
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
					Eventually(session.Out).Should(Say("api endpoint:\\s+%s", apiURL))
					Eventually(session.Out).Should(Say("api version:\\s+[\\d.]+"))
					Eventually(session.Out).Should(Say("user:\\s+%s", username))
					Eventually(session.Out).Should(Say("org:\\s+%s", orgName))
					Eventually(session.Out).Should(Say("space:\\s+%s", spaceName))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when the space does not exist", func() {
				It("displays space not found, exits 1, and clears the existing targets", func() {
					session := helpers.CF("target", "-o", orgName, "-s", spaceName)
					Eventually(session.Err).Should(Say("Space '%s' not found.", spaceName))
					Eventually(session.Out).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))

					session = helpers.CF("target")
					Eventually(session.Out).Should(Say("No org or space targeted, use 'cf target -o ORG -s SPACE'"))
					Eventually(session).Should(Exit(0))
				})
			})
		})
	})
})
