package experimental

import (
	"net/http"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("login command", func() {
	BeforeEach(func() {
		helpers.SkipIfClientCredentialsTestMode()
		helpers.TurnOnExperimentalLogin()
	})

	AfterEach(func() {
		helpers.TurnOffExperimentalLogin()
	})

	Describe("Minimum Version Check", func() {
		When("the CLI version is lower than the minimum supported version by the CC", func() {
			var server *ghttp.Server

			BeforeEach(func() {
				server = helpers.StartServerWithMinimumCLIVersion("9000.0.0")

				fakeTokenResponse := map[string]string{
					"access_token": "",
					"token_type":   "bearer",
				}
				server.RouteToHandler(http.MethodPost, "/oauth/token",
					ghttp.RespondWithJSONEncoded(http.StatusOK, fakeTokenResponse))
				server.RouteToHandler(http.MethodGet, "/v3/organizations",
					ghttp.RespondWith(http.StatusOK, `{
					 "total_results": 0,
					 "total_pages": 1,
					 "resources": []}`))
			})

			AfterEach(func() {
				server.Close()
			})

			It("displays the warning and exits successfully", func() {
				session := helpers.CF("login", "-a", server.URL(), "--skip-ssl-validation")
				Eventually(session.Err).Should(Say(`Cloud Foundry API version .+ requires CLI version .+\. You are currently on version .+\. To upgrade your CLI, please visit: https://github.com/cloudfoundry/cli#downloads`))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Describe("Target Space", func() {
		var (
			orgName  string
			username string
			password string
		)

		BeforeEach(func() {
			helpers.LoginCF()
			orgName = helpers.NewOrgName()
			session := helpers.CF("create-org", orgName)
			Eventually(session).Should(Exit(0))
			username, password = helpers.CreateUserInOrgRole(orgName, "OrgManager")
		})

		When("only one space is available to the user", func() {
			var spaceName string

			BeforeEach(func() {
				spaceName = helpers.NewSpaceName()
				session := helpers.CF("create-space", "-o", orgName, spaceName)
				Eventually(session).Should(Exit(0))
				roleSession := helpers.CF("set-space-role", username, orgName, spaceName, "SpaceManager")
				Eventually(roleSession).Should(Exit(0))
			})

			It("logs the user in and targets the org and the space", func() {
				session := helpers.CF("login", "-u", username, "-p", password, "-a", apiURL, "--skip-ssl-validation")
				Eventually(session).Should(Exit(0))

				targetSession := helpers.CF("target")
				Eventually(targetSession).Should(Exit(0))
				Eventually(targetSession).Should(Say(`org:\s+%s`, orgName))
				Eventually(targetSession).Should(Say(`space:\s+%s`, spaceName))
			})

		})

	})

})
