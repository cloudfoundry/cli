package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("create-service command", func() {
	Describe("help", func() {

		matchHelpMessage := SatisfyAll(
			Say(`NAME:\n`),
			Say(`\s+create-service - Create a service instance\n`),
			Say(`\n`),
			Say(`USAGE:\n`),
			Say(`\s+cf create-service SERVICE PLAN SERVICE_INSTANCE \[-b SERVICE_BROKER\] \[-c JSON_PARAMS\] \[-t TAGS\]\n`),
			Say(`\s+Optionally provide service-specific configuration parameters in a valid JSON object in-line:\n`),
			Say(`\s+cf create-service SERVICE PLAN SERVICE_INSTANCE -c '{\"name\":\"value\",\"name\":\"value\"}'\n`),
			Say(`\s+Optionally provide a file containing service-specific configuration parameters in a valid JSON object\.\n`),
			Say(`\s+The path to the parameters file can be an absolute or relative path to a file:\n`),
			Say(`\s+cf create-service SERVICE PLAN SERVICE_INSTANCE -c PATH_TO_FILE\n`),
			Say(`\s+Example of valid JSON object:`),
			Say(`\s+{`),
			Say(`\s+\"cluster_nodes\": {`),
			Say(`\s+\"count\": 5,`),
			Say(`\s+\"memory_mb\": 1024`),
			Say(`\s+}`),
			Say(`\s+}`),
			Say(`TIP:`),
			Say(`\s+Use 'cf create-user-provided-service' to make user-provided service instances available to CF apps`),
			Say(`EXAMPLES:`),
			Say(`\s+Linux/Mac:\n`),
			Say(`\s+cf create-service db-service silver mydb -c '{\"ram_gb\":4}`),
			Say(`\s+Windows Command Line:`),
			Say(`\s+cf create-service db-service silver mydb -c \"{\\\"ram_gb\\\":4}\"`),
			Say(`\s+Windows PowerShell:`),
			Say(`\s+cf create-service db-service silver mydb -c '{\\\"ram_gb\\\":4}'`),
			Say(`\s+cf create-service db-service silver mydb -c ~/workspace/tmp/instance_config.json`),
			Say(`\s+cf create-service db-service silver mydb -t \"list, of, tags\"`),
			Say(`ALIAS:`),
			Say(`\s+cs`),
			Say(`OPTIONS:`),
			Say(`\s+-b      Create a service instance from a particular broker\. Required when service offering name is ambiguous`),
			Say(`\s+-c      Valid JSON object containing service-specific configuration parameters, provided either in-line or in a file\. For a list of supported configuration parameters, see documentation for the particular service offering\.`),
			Say(`\s+-t      User provided tags`),
			Say(`SEE ALSO:`),
			Say(`\s+bind-service, create-user-provided-service, marketplace, services`),
		)

		When("the -h flag is specified", func() {
			It("succeeds and prints help", func() {
				session := helpers.CF("create-service", "-h")
				Eventually(session).Should(Exit(0))
				Expect(session.Out).To(matchHelpMessage)
			})
		})

		When("the --help flag is specified", func() {
			It("succeeds and prints help", func() {
				session := helpers.CF("create-service", "--help")
				Eventually(session).Should(Exit(0))
				Expect(session.Out).To(matchHelpMessage)
			})
		})

		When("no arguments are provided", func() {
			It("displays a warning, the help text, and exits 1", func() {
				session := helpers.CF("create-service")
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Incorrect Usage: the required arguments `SERVICE`, `SERVICE_PLAN` and `SERVICE_INSTANCE` were not provided"))
				Expect(session.Out).To(matchHelpMessage)
			})
		})

		When("unknown flag is passed", func() {
			It("displays a warning, the help text, and exits 1", func() {
				session := helpers.CF("create-service", "-u")
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Incorrect Usage: unknown flag `u"))
				Expect(session.Out).To(matchHelpMessage)
			})
		})

		When("a flag is passed with no argument", func() {
			It("displays a warning, the help text, and exits 1", func() {
				session := helpers.CF("create-service", "-c")
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Incorrect Usage: expected argument for flag `-c'"))
				Expect(session.Out).To(matchHelpMessage)
			})
		})

	})
	Context("not logged in", func() {
		BeforeEach(func() {
			helpers.LogoutCF()
		})

		It("displays FAILED, an informative error message, and exits 1", func() {
			session := helpers.CF("create-service", "offering", "plan", "my-instance")
			Eventually(session).Should(Exit(1))
			Expect(session).To(Say("FAILED"))
			Expect(session.Err).To(Say("Not logged in. Use 'cf login' or 'cf login --sso' to log in."))
		})
	})
	Context("logged in", func() {
		BeforeEach(func() {
			helpers.LoginCF()
			helpers.TargetOrg(ReadOnlyOrg)
		})

		When("Space is not targeted", func() {
			It("Displays an error and exits", func() {
				session := helpers.CF("create-service", "offering", "plan", "my-instance")
				Eventually(session).Should(Exit(1))
				Expect(session).To(Say("FAILED"))
				Expect(session.Err).To(Say("No space targeted, use 'cf target -s SPACE' to target a space."))

			})
		})
	})

})
