package isolated

import (
	"os"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("create-user-provided-service command", func() {
	Describe("help", func() {
		expectHelpMessage := func(session *Session) {
			Expect(session).To(SatisfyAll(
				Say(`NAME:`),
				Say(`create-user-provided-service - Make a user-provided service instance available to CF apps`),
				Say(`USAGE:`),
				Say(`cf create-user-provided-service SERVICE_INSTANCE \[-p CREDENTIALS\] \[-l SYSLOG_DRAIN_URL\] \[-r ROUTE_SERVICE_URL\] \[-t TAGS\]`),
				Say(`Pass comma separated credential parameter names to enable interactive mode:`),
				Say(`cf create-user-provided-service SERVICE_INSTANCE -p "comma, separated, parameter, names"`),
				Say(`Pass credential parameters as JSON to create a service non-interactively:`),
				Say(`cf create-user-provided-service SERVICE_INSTANCE -p '\{"key1":"value1","key2":"value2"\}'`),
				Say(`Specify a path to a file containing JSON:`),
				Say(`cf create-user-provided-service SERVICE_INSTANCE -p PATH_TO_FILE`),
				Say(`EXAMPLES:`),
				Say(`cf create-user-provided-service my-db-mine -p "username, password"`),
				Say(`cf create-user-provided-service my-db-mine -p /path/to/credentials.json`),
				Say(`cf create-user-provided-service my-db-mine -t "list, of, tags"`),
				Say(`cf create-user-provided-service my-drain-service -l syslog://example.com`),
				Say(`cf create-user-provided-service my-route-service -r https://example.com`),
				Say(`Linux/Mac:`),
				Say(`cf create-user-provided-service my-db-mine -p '\{"username":"admin","password":"pa55woRD"\}'`),
				Say(`Windows Command Line:`),
				Say(`cf create-user-provided-service my-db-mine -p "\{\\"username\\":\\"admin\\",\\"password\\":\\"pa55woRD\\"\}"`),
				Say(`Windows PowerShell:`),
				Say(`cf create-user-provided-service my-db-mine -p '\{\\"username\\":\\"admin\\",\\"password\\":\\"pa55woRD\\"\}'`),
				Say(`ALIAS:`),
				Say(`cups`),
				Say(`OPTIONS:`),
				Say(`-l      URL to which logs for bound applications will be streamed`),
				Say(`-p      Credentials, provided inline or in a file, to be exposed in the VCAP_SERVICES environment variable for bound applications`),
				Say(`-r      URL to which requests for bound routes will be forwarded. Scheme for this URL must be https`),
				Say(`-t      User provided tags`),
				Say(`SEE ALSO:`),
				Say(`bind-service, services`),
			))
		}

		When("--help flag is set", func() {
			It("displays command usage to output", func() {
				session := helpers.CF("create-user-provided-service", "--help")
				Eventually(session).Should(Exit(0))

				expectHelpMessage(session)
			})
		})

		When("no arguments provided", func() {
			It("fails and displays command usage", func() {
				session := helpers.CF("create-user-provided-service")
				Eventually(session).Should(Exit(1))

				Expect(session.Err).To(Say("Incorrect Usage: the required argument `SERVICE_INSTANCE` was not provided"))
				expectHelpMessage(session)
			})
		})

		When("an superfluous argument is provided", func() {
			It("fails and displays command usage", func() {
				session := helpers.CF("create-user-provided-service", "name", "extraparam")
				Eventually(session).Should(Exit(1))

				Expect(session.Err).To(Say(`Incorrect Usage: unexpected argument "extraparam"`))
				expectHelpMessage(session)
			})
		})

		When("an unsupported flag is provided", func() {
			It("fails and displays command usage", func() {
				session := helpers.CF("create-user-provided-service", "name", "--do-magic")
				Eventually(session).Should(Exit(1))

				Expect(session.Err).To(Say("Incorrect Usage: unknown flag `do-magic"))
				expectHelpMessage(session)
			})
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "create-user-provided-service", "foo")
		})
	})

	When("targeting a space", func() {
		var (
			userName    string
			orgName     string
			spaceName   string
			serviceName string
		)

		expectOKMessage := func(session *Session, serviceName, orgName, spaceName, userName string) {
			Expect(session.Out).To(SatisfyAll(
				Say("Creating user provided service %s in org %s / space %s as %s...", serviceName, orgName, spaceName, userName),
				Say("OK"),
			))
		}

		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()
			helpers.SetupCF(orgName, spaceName)
			userName, _ = helpers.GetCredentials()
			serviceName = helpers.PrefixedRandomName("ups")
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		When("a name is provided", func() {
			It("displays success message, exits 0, and creates the service", func() {
				session := helpers.CF(`create-user-provided-service`, serviceName)

				Eventually(session).Should(Exit(0))
				expectOKMessage(session, serviceName, orgName, spaceName, userName)

				session = helpers.CF("service", serviceName)
				Eventually(session).Should(Exit(0))
				Expect(session).To(Say(`name:\s+%s`, serviceName))
			})
		})

		When("all parameters are provided", func() {
			It("displays success message, exits 0, and creates the service", func() {
				session := helpers.CF(
					`create-user-provided-service`, serviceName,
					`-p`, `'{"username":"password"}'`,
					`-t`, `"list, of, tags"`,
					`-l`, `syslog://example-syslog.com`,
					`-r`, `https://example-route.com`,
				)

				Eventually(session).Should(Exit(0))
				expectOKMessage(session, serviceName, orgName, spaceName, userName)

				session = helpers.CF("service", serviceName)
				Eventually(session).Should(Exit(0))
				Expect(session).To(SatisfyAll(
					Say(`name:\s+%s`, serviceName),
					Say(`tags:\s+list,\s*of,\s*tags`),
					Say(`route service url:\s+https://example-route.com`),
				))
			})
		})

		When("requesting interactive credentials", func() {
			var buffer *Buffer

			BeforeEach(func() {
				buffer = NewBuffer()
				_, err := buffer.Write([]byte("fake-username\nfake-password\n"))
				Expect(err).ToNot(HaveOccurred())
			})

			It("requests the credentials at a prompt", func() {
				session := helpers.CFWithStdin(buffer, "create-user-provided-service", serviceName, "-p", `"username,password"`)

				Eventually(session).Should(Say("username: "))
				Eventually(session).Should(Say("password: "))
				Consistently(session).ShouldNot(Say("fake-username"), "credentials should not be echoed to the user")
				Consistently(session).ShouldNot(Say("fake-password"), "credentials should not be echoed to the user")
				Eventually(session).Should(Exit(0))

				expectOKMessage(session, serviceName, orgName, spaceName, userName)
			})
		})

		When("reading JSON credentials from a file", func() {
			var path string

			BeforeEach(func() {
				path = helpers.TempFileWithContent(`{"some": "credentials"}`)
			})

			AfterEach(func() {
				Expect(os.Remove(path)).To(Succeed())
			})

			It("accepts a file path", func() {
				session := helpers.CF("create-user-provided-service", serviceName, "-p", path)

				By("checking that it does not interpret the file name as request for an interactive credential prompt")
				Consistently(session.Out.Contents()).ShouldNot(ContainSubstring(path))

				By("succeeding")
				Eventually(session).Should(Exit(0))
				expectOKMessage(session, serviceName, orgName, spaceName, userName)
			})
		})
	})
})
