package isolated

import (
	"strings"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("curl command", func() {
	var ExpectHelpText = func(session *Session) {
		Eventually(session).Should(Say(`NAME:\n`))
		Eventually(session).Should(Say(`curl - Executes a request to the targeted API endpoint\n`))
		Eventually(session).Should(Say(`\n`))

		Eventually(session).Should(Say(`USAGE:\n`))
		Eventually(session).Should(Say(`\s+cf curl PATH \[-iv\] \[-X METHOD\] \[-H HEADER\] \[-d DATA\] \[--output FILE\]`))
		Eventually(session).Should(Say(`\s+By default 'cf curl' will perform a GET to the specified PATH. If data`))
		Eventually(session).Should(Say(`\s+is provided via -d, a POST will be performed instead, and the Content-Type\n`))
		Eventually(session).Should(Say(`\s+will be set to application/json. You may override headers with -H and the\n`))
		Eventually(session).Should(Say(`\s+request method with -X.\n`))
		Eventually(session).Should(Say(`\s+For API documentation, please visit http://apidocs.cloudfoundry.org.\n`))
		Eventually(session).Should(Say(`\n`))

		Eventually(session).Should(Say(`EXAMPLES:\n`))
		Eventually(session).Should(Say(`\s+cf curl \"/v2/apps\" -X GET -H \"Content-Type: application/x-www-form-urlencoded\" -d 'q=name:myapp'`))
		Eventually(session).Should(Say(`\s+cf curl \"/v2/apps\" -d @/path/to/file`))
		Eventually(session).Should(Say(`\n`))

		Eventually(session).Should(Say(`OPTIONS:\n`))
		Eventually(session).Should(Say(`\s+-H\s+Custom headers to include in the request, flag can be specified multiple times`))
		Eventually(session).Should(Say(`\s+-X\s+HTTP method \(GET,POST,PUT,DELETE,etc\)`))
		Eventually(session).Should(Say(`\s+-d\s+HTTP data to include in the request body, or '@' followed by a file name to read the data from`))
		Eventually(session).Should(Say(`\s+-i\s+Include response headers in the output`))
		Eventually(session).Should(Say(`\s+--output\s+Write curl body to FILE instead of stdout`))
	}

	var ExpectRequestHeaders = func(session *Session) {
		Eventually(session).Should(Say(`REQUEST: .*\n`))
		Eventually(session).Should(Say(`GET /v2/apps HTTP/1.1`))
		Eventually(session).Should(Say(`Host: .*\n`))
		Eventually(session).Should(Say(`Accept: .*\n`))
		Eventually(session).Should(Say(`Authorization:\s+\[PRIVATE DATA HIDDEN\]`))
		Eventually(session).Should(Say(`Content-Type: .*\n`))
		Eventually(session).Should(Say(`User-Agent: .*\n`))
	}

	var ExpectReponseHeaders = func(session *Session) {
		Eventually(session).Should(Say("HTTP/1.1 200 OK"))
		Eventually(session).Should(Say(`Connection:`))
		Eventually(session).Should(Say(`Content-Length:`))
		Eventually(session).Should(Say(`Content-Type:`))
		Eventually(session).Should(Say(`Date:`))
		Eventually(session).Should(Say(`Server:`))
		Eventually(session).Should(Say(`X-Content-Type-Options:`))
		Eventually(session).Should(Say(`X-Vcap-Request-Id:`))
	}

	Describe("Help Text", func() {
		When("--help flag is set", func() {
			It("Displays command usage to the output", func() {
				session := helpers.CF("curl", "--help")
				ExpectHelpText(session)
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Describe("Incorrect Usage", func() {
		When("no arguments are provided", func() {
			It("fails and displays the help text", func() {
				session := helpers.CF("curl")
				Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `PATH` was not provided"))
				ExpectHelpText(session)
				Eventually(session).Should(Exit(1))
			})
		})

		When("unkown flag is specified", func() {
			It("fails and displays the help text", func() {
				session := helpers.CF("curl", "--test")
				// TODO Legacy cf uses a weird quote arount test. This test needs be fixed for refactored command
				Eventually(session.Err).Should(Say("Incorrect Usage: unknown flag `test'"))
				ExpectHelpText(session)
				Eventually(session).Should(Exit(1))
			})
		})

		When("more than one path is specified", func() {
			It("fails and displays the help text", func() {
				session := helpers.CF("curl", "/v2/apps", "/v2/apps")
				Eventually(session).Should(Say("FAILED\n"))
				// TODO Legacy code uses Incorrect Usage.(dot) instead of Incorrect Usage: (colon). Fix this test after refactor
				Eventually(session).Should(Say("Incorrect Usage. An argument is missing or not correctly enclosed."))
				ExpectHelpText(session)
				Eventually(session).Should(Exit(1))
			})
		})
	})

	When("the user is not logged in", func() {
		It("makes the request and recieves an unauthenticated error", func() {
			session := helpers.CF("curl", "/v2/apps")
			expectedJSON := `{
				 "description": "Authentication error",
				 "error_code": "CF-NotAuthenticated",
				 "code": 10002
			}`
			Eventually(session).Should(Exit(0))
			Expect(session.Out.Contents()).To(MatchJSON(expectedJSON))
		})
	})

	When("the user is logged in", func() {
		BeforeEach(func() {
			orgName := helpers.NewOrgName()
			spaceName := helpers.NewSpaceName()

			helpers.SetupCF(orgName, spaceName)
			helpers.SwitchToOrgRole(orgName, "OrgManager")
		})

		When("PATH is valid", func() {
			var expectedJSON string

			BeforeEach(func() {
				expectedJSON = `{
            "total_results": 0,
            "total_pages": 1,
            "prev_url": null,
            "next_url": null,
            "resources": []
				}`
			})

			When("no flag is set", func() {
				It("makes the request and displays the json response", func() {
					session := helpers.CF("curl", "/v2/apps")
					Eventually(session).Should(Exit(0))
					Expect(session.Out.Contents()).To(MatchJSON(expectedJSON))
				})
			})

			When("-i flag is set", func() {
				It("displays the response headers", func() {
					session := helpers.CF("curl", "/v2/apps", "-i")
					Eventually(session).Should(Exit(0))

					ExpectReponseHeaders(session)
					contents := string(session.Out.Contents())
					jsonStartsAt := strings.Index(contents, "{")

					actualJSON := contents[jsonStartsAt:]
					Expect(actualJSON).To(MatchJSON(expectedJSON))
				})
			})

			When("-v flag is set", func() {
				It("displays the request headers and response headers", func() {
					session := helpers.CF("curl", "/v2/apps", "-v")
					Eventually(session).Should(Exit(0))

					ExpectRequestHeaders(session)
					ExpectReponseHeaders(session)

					contents := string(session.Out.Contents())
					jsonStartsAt := strings.Index(contents, "{")

					actualJSON := contents[jsonStartsAt:]
					Expect(actualJSON).To(MatchJSON(expectedJSON))
				})
			})
		})

		When("PATH is invalid", func() {
			It("makes the request and displays the unknown request json", func() {
				expectedJSON := `{
				 "description": "Unknown request",
				 "error_code": "CF-NotFound",
				 "code": 10000
				}`
				session := helpers.CF("curl", "/some-random-path")
				Eventually(session).Should(Exit(0))
				Expect(session.Out.Contents()).To(MatchJSON(expectedJSON))
			})
		})
	})
})
