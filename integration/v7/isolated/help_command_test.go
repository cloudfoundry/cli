package isolated

import (
	"os/exec"
	"strings"

	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	"code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("help command", func() {
	It("appears in cf help -a", func() {
		session := helpers.CF("help", "-a")
		Eventually(session).Should(Exit(0))
		Expect(session).To(HaveCommandInCategoryWithDescription("help", "GETTING STARTED", "Show help"))
	})

	DescribeTable("displays help for common commands",
		func(setup func() *exec.Cmd) {
			cmd := setup()
			session, err := Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(Say("Cloud Foundry command line tool"))
			Eventually(session).Should(Say(`\[global options\] command \[arguments...\] \[command options\]`))
			Eventually(session).Should(Say("Before getting started:"))
			Eventually(session).Should(Say(`  config\s+login,l\s+target,t`))
			Eventually(session).Should(Say("Application lifecycle:"))
			Eventually(session).Should(Say(`  apps,a\s+run-task,rt\s+events`))
			Eventually(session).Should(Say(`  restage,rg\s+scale`))

			Eventually(session).Should(Say("Services integration:"))
			Eventually(session).Should(Say(`  marketplace,m\s+create-user-provided-service,cups`))
			Eventually(session).Should(Say(`  services,s\s+update-user-provided-service,uups`))

			Eventually(session).Should(Say("Route and domain management:"))
			Eventually(session).Should(Say(`  routes,r\s+delete-route\s+create-private-domain`))
			Eventually(session).Should(Say(`  domains\s+map-route`))

			Eventually(session).Should(Say("Space management:"))
			Eventually(session).Should(Say(`  spaces\s+create-space,csp\s+set-space-role`))

			Eventually(session).Should(Say("Org management:"))
			Eventually(session).Should(Say(`  orgs,o\s+set-org-role`))

			Eventually(session).Should(Say("CLI plugin management:"))
			Eventually(session).Should(Say("  install-plugin    list-plugin-repos"))
			Eventually(session).Should(Say("Global options:"))
			Eventually(session).Should(Say("  --help, -h                         Show help"))
			Eventually(session).Should(Say("  -v                                 Print API request diagnostics to stdout"))

			Eventually(session).Should(Say(`TIP: Use 'cf help -a' to see all commands\.`))
			Eventually(session).Should(Exit(0))
		},

		Entry("when cf is run without providing a command or a flag", func() *exec.Cmd {
			return exec.Command("cf")
		}),

		Entry("when cf help is run", func() *exec.Cmd {
			return exec.Command("cf", "help")
		}),

		Entry("when cf is run with -h flag alone", func() *exec.Cmd {
			return exec.Command("cf", "-h")
		}),

		Entry("when cf is run with --help flag alone", func() *exec.Cmd {
			return exec.Command("cf", "--help")
		}),
	)

	DescribeTable("displays help for all commands",
		func(setup func() *exec.Cmd) {
			cmd := setup()
			session, err := Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Say("USAGE:"))
			Eventually(session).Should(Say("VERSION:"))
			Eventually(session).Should(Say("GETTING STARTED:"))
			Eventually(session).Should(Say("ENVIRONMENT VARIABLES:"))
			Eventually(session).Should(Say(`CF_DIAL_TIMEOUT=6\s+Max wait time to establish a connection, including name resolution, in seconds`))
			Eventually(session).Should(Say("GLOBAL OPTIONS:"))
			Eventually(session).Should(Exit(0))
		},

		Entry("when cf help is run", func() *exec.Cmd {
			return exec.Command("cf", "help", "-a")
		}),

		Entry("when cf is run with -h -a flag", func() *exec.Cmd {
			return exec.Command("cf", "-h", "-a")
		}),

		Entry("when cf is run with --help -a flag", func() *exec.Cmd {
			return exec.Command("cf", "--help", "-a")
		}),
	)

	Describe("commands that appear in cf help -a", func() {
		It("includes run-task", func() {
			session := helpers.CF("help", "-a")
			Eventually(session).Should(Say(`run-task\s+Run a one-off task on an app`))
			Eventually(session).Should(Exit(0))
		})

		It("includes list-task", func() {
			session := helpers.CF("help", "-a")
			Eventually(session).Should(Say(`tasks\s+List tasks of an app`))
			Eventually(session).Should(Exit(0))
		})

		It("includes terminate-task", func() {
			session := helpers.CF("help", "-a")
			Eventually(session).Should(Say(`terminate-task\s+Terminate a running task of an app`))
			Eventually(session).Should(Exit(0))
		})
	})

	Context("displays the help text for a given command", func() {
		DescribeTable("displays the help",
			func(setup func() (*exec.Cmd, int)) {
				cmd, exitCode := setup()
				session, err := Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("create-user-provided-service - Make a user-provided service instance available to CF apps"))
				Eventually(session).Should(Say(`cf create-user-provided-service SERVICE_INSTANCE \[-p CREDENTIALS\] \[-l SYSLOG_DRAIN_URL\] \[-r ROUTE_SERVICE_URL\]`))
				Eventually(session).Should(Say(`-l\s+URL to which logs for bound applications will be streamed`))
				Eventually(session).Should(Exit(exitCode))
			},

			Entry("when a command is called with the --help flag", func() (*exec.Cmd, int) {
				return exec.Command("cf", "create-user-provided-service", "--help"), 0
			}),

			Entry("when a command is called with the --help flag and command arguments", func() (*exec.Cmd, int) {
				return exec.Command("cf", "create-user-provided-service", "-l", "http://example.com", "--help"), 0
			}),

			Entry("when a command is called with the --help flag and command arguments prior to the command", func() (*exec.Cmd, int) {
				return exec.Command("cf", "-l", "create-user-provided-service", "--help"), 1
			}),

			Entry("when the help command is passed a command name", func() (*exec.Cmd, int) {
				return exec.Command("cf", "help", "create-user-provided-service"), 0
			}),

			Entry("when the --help flag is passed with a command name", func() (*exec.Cmd, int) {
				return exec.Command("cf", "--help", "create-user-provided-service"), 0
			}),

			Entry("when the -h flag is passed with a command name", func() (*exec.Cmd, int) {
				return exec.Command("cf", "-h", "create-user-provided-service"), 0
			}),

			Entry("when the help command is passed a command alias", func() (*exec.Cmd, int) {
				return exec.Command("cf", "help", "cups"), 0
			}),

			Entry("when the --help flag is passed with a command alias", func() (*exec.Cmd, int) {
				return exec.Command("cf", "--help", "cups"), 0
			}),

			Entry("when the --help flag is passed after a command alias", func() (*exec.Cmd, int) {
				return exec.Command("cf", "cups", "--help"), 0
			}),

			Entry("when an invalid flag is passed", func() (*exec.Cmd, int) {
				return exec.Command("cf", "create-user-provided-service", "--invalid-flag"), 1
			}),

			Entry("when missing required arguments", func() (*exec.Cmd, int) {
				return exec.Command("cf", "create-user-provided-service"), 1
			}),

			Entry("when missing arguments to flags", func() (*exec.Cmd, int) {
				return exec.Command("cf", "create-user-provided-service", "foo", "-l"), 1
			}),
		)

		When("the command uses timeout environment variables", func() {
			DescribeTable("shows the CF_STAGING_TIMEOUT and CF_STARTUP_TIMEOUT environment variables",
				func(setup func() (*exec.Cmd, int)) {
					cmd, exitCode := setup()
					session, err := Start(cmd, GinkgoWriter, GinkgoWriter)
					Expect(err).NotTo(HaveOccurred())

					Eventually(session).Should(Say("ENVIRONMENT:"))
					Eventually(session).Should(Say("CF_STAGING_TIMEOUT=15\\s+Max wait time for staging, in minutes"))
					Eventually(session).Should(Say("CF_STARTUP_TIMEOUT=5\\s+Max wait time for app instance startup, in minutes"))
					Eventually(session).Should(Exit(exitCode))
				},

				Entry("cf push", func() (*exec.Cmd, int) {
					return exec.Command("cf", "h", "push"), 0
				}),

				Entry("cf start", func() (*exec.Cmd, int) {
					return exec.Command("cf", "h", "start"), 0
				}),

				Entry("cf restart", func() (*exec.Cmd, int) {
					return exec.Command("cf", "h", "restart"), 0
				}),

				Entry("cf copy-source", func() (*exec.Cmd, int) {
					return exec.Command("cf", "h", "copy-source"), 0
				}),

				Entry("cf restage", func() (*exec.Cmd, int) {
					return exec.Command("cf", "h", "restage"), 0
				}),
			)
		})
	})

	When("the command does not exist", func() {
		DescribeTable("help displays an error message",
			func(command func() *exec.Cmd) {
				session, err := Start(command(), GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session.Err).Should(Say("'rock' is not a registered command. See 'cf help -a'"))
				Eventually(session).Should(Exit(1))
			},

			Entry("passing --help into rock (cf rock --help)", func() *exec.Cmd {
				return exec.Command("cf", "rock", "--help")
			}),

			Entry("passing the --help flag (cf --help rock)", func() *exec.Cmd {
				return exec.Command("cf", "--help", "rock")
			}),

			Entry("calling the help command directly", func() *exec.Cmd {
				return exec.Command("cf", "help", "rock")
			}),
		)

	})

	When("the option does not exist", func() {
		DescribeTable("help display an error message as well as help for common commands",

			func(command func() *exec.Cmd) {
				session, err := Start(command(), GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(session).Should(Exit(1))
				Eventually(session).Should(Say("Before getting started:")) // common help
				Expect(strings.Count(string(session.Err.Contents()), "unknown flag")).To(Equal(1))
			},

			Entry("passing invalid option", func() *exec.Cmd {
				return exec.Command("cf", "-c")
			}),

			Entry("passing -a option", func() *exec.Cmd {
				return exec.Command("cf", "-a")
			}),
		)
	})
})
