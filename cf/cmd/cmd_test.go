package cmd_test

import (
	"bufio"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var buildPath string

var _ = SynchronizedBeforeSuite(func() []byte {
	path, buildErr := Build("code.cloudfoundry.org/cli")
	Expect(buildErr).NotTo(HaveOccurred())
	return []byte(path)
}, func(data []byte) {
	buildPath = string(data)
})

// gexec.Build leaves a compiled binary behind in /tmp.
var _ = SynchronizedAfterSuite(func() {}, func() {
	CleanupBuildArtifacts()
})

var _ = Describe("main", func() {
	var (
		old_PLUGINS_HOME string
	)

	BeforeEach(func() {
		old_PLUGINS_HOME = os.Getenv("CF_PLUGIN_HOME")

		dir, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())

		fullDir := filepath.Join(dir, "..", "..", "fixtures", "config", "main-plugin-test-config")
		err = os.Setenv("CF_PLUGIN_HOME", fullDir)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		err := os.Setenv("CF_PLUGIN_HOME", old_PLUGINS_HOME)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Help menu with -h/--help", func() {
		It("prints the help output with our custom template when run with 'cf -h'", func() {
			output := Cf("-h", "-a")
			Eventually(output.Out.Contents).Should(ContainSubstring("A command line tool to interact with Cloud Foundry"))
			Eventually(output.Out.Contents).Should(ContainSubstring("CF_TRACE=true"))
		})

		It("prints the help output with our custom template when run with 'cf --help'", func() {
			output := Cf("--help", "-a")
			Eventually(output.Out.Contents).Should(ContainSubstring("A command line tool to interact with Cloud Foundry"))
			Eventually(output.Out.Contents).Should(ContainSubstring("CF_TRACE=true"))
		})

		It("accepts -h and --h flags for all commands", func() {
			result := Cf("push", "-h")
			Consistently(result.Out).ShouldNot(Say("Incorrect Usage"))
			Eventually(result.Out.Contents).Should(ContainSubstring("USAGE"))

			result = Cf("push", "--no-route", "-h")
			Consistently(result.Out).ShouldNot(Say("Incorrect Usage"))
			Eventually(result.Out.Contents).Should(ContainSubstring("USAGE"))

			result = Cf("target", "--h")
			Consistently(result.Out).ShouldNot(Say("Incorrect Usage"))
			Eventually(result.Out.Contents).Should(ContainSubstring("USAGE"))
		})

		It("accepts -h before the command name", func() {
			result := Cf("-h", "push", "--no-route")
			Consistently(result.Out).ShouldNot(Say("Incorrect Usage"))
			Consistently(result.Out).ShouldNot(Say("Start an app"))
			Eventually(result.Out.Contents).Should(ContainSubstring("USAGE"))
			Eventually(result.Out.Contents).Should(ContainSubstring("push"))
		})

		It("accepts -h before the command alias", func() {
			result := Cf("-h", "p", "--no-route")
			Consistently(result.Out).ShouldNot(Say("Incorrect Usage"))
			Consistently(result.Out).ShouldNot(Say("Start an app"))
			Eventually(result.Out.Contents).Should(ContainSubstring("USAGE"))
			Eventually(result.Out.Contents).Should(ContainSubstring("push"))
		})
	})

	Describe("Shows version with -v or --version", func() {
		It("prints the cf version if '-v' flag is provided", func() {
			output := Cf("-v")
			Eventually(output.Out.Contents).Should(ContainSubstring("cf version"))
			Eventually(output).Should(Exit(0))
		})

		It("prints the cf version if '--version' flag is provided", func() {
			output := Cf("--version")
			Eventually(output.Out.Contents).Should(ContainSubstring("cf version"))
			Eventually(output).Should(Exit(0))
		})
	})

	Describe("Enables verbose output with -v", func() {
		BeforeEach(func() {
			client := http.Client{Timeout: 3 * time.Second}
			_, err := client.Get("http://api.bosh-lite.com/v2/info")
			if err != nil {
				Skip("unable to communicate with bosh-lite, skipping")
			}
			setApiOutput := Cf("api", "http://api.bosh-lite.com", "--skip-ssl-validation")
			Eventually(setApiOutput.Out.Contents).Should(ContainSubstring("OK"))
		})

		// Normally cf curl only shows the output of the response
		// When using trace, it also shows the request/response information
		It("enables verbose output when -v is provided before a command", func() {
			output := Cf("-v", "curl", "/v2/info")
			Consistently(output.Out.Contents).ShouldNot(ContainSubstring("Invalid flag: -v"))
			Eventually(output.Out.Contents).Should(ContainSubstring("GET /v2/info HTTP/1.1"))
		})

		It("enables verbose output when -v is provided after a command", func() {
			output := Cf("curl", "/v2/info", "-v")
			Consistently(output.Out.Contents).ShouldNot(ContainSubstring("Invalid flag: -v"))
			Eventually(output.Out.Contents).Should(ContainSubstring("GET /v2/info HTTP/1.1"))
		})
	})

	Describe("Commands with new command structure", func() {
		It("prints usage help for all commands by providing `help` flag", func() {
			output := Cf("api", "-h")
			Eventually(output.Out.Contents).Should(ContainSubstring("USAGE"))
			Eventually(output.Out.Contents).Should(ContainSubstring("OPTIONS"))
		})

		It("accepts -h and --h flags for commands", func() {
			result := Cf("api", "-h")
			Consistently(result.Out).ShouldNot(Say("Invalid flag: -h"))
			Eventually(result.Out.Contents).Should(ContainSubstring("api - Set or view target api url"))

			result = Cf("api", "--h")
			Consistently(result.Out).ShouldNot(Say("Invalid flag: --h"))
			Eventually(result.Out.Contents).Should(ContainSubstring("api - Set or view target api url"))
		})
	})

	Describe("exit codes", func() {
		It("exits non-zero when an unknown command is invoked", func() {
			result := Cf("some-command-that-should-never-actually-be-a-real-thing-i-can-use")

			Eventually(result, 3*time.Second).Should(Say("not a registered command"))
			Eventually(result).Should(Exit(1))
		})

		It("exits non-zero when known command is invoked with invalid option", func() {
			result := Cf("push", "--crazy")
			Eventually(result).Should(Exit(1))
		})
	})

	It("can print help menu by executing only the command `cf`", func() {
		output := Cf()
		Eventually(output.Out.Contents).Should(ContainSubstring("Cloud Foundry command line tool"))
	})

	It("show user suggested commands for typos", func() {
		output := Cf("hlp")
		Eventually(output.Out, 3*time.Second).Should(Say("'hlp' is not a registered command. See 'cf help'"))
		Eventually(output.Out, 3*time.Second).Should(Say("Did you mean?"))
	})

	It("does not display requirement errors twice", func() {
		output := Cf("space")
		Eventually(output).Should(Exit(1))
		Expect(output.Err).To(Say("the required argument `SPACE` was not provided"))
		Expect(output.Err).NotTo(Say("the required argument `SPACE` was not provided"))
		Expect(output.Out).NotTo(Say("the required argument `SPACE` was not provided"))
	})

	Describe("Plugins", func() {
		It("Can call a plugin command from the Plugins configuration if it does not exist as a cf command", func() {
			output := Cf("test_1_cmd1")
			Eventually(output.Out).Should(Say("You called cmd1 in test_1"))
		})

		It("Can call a plugin command via alias if it does not exist as a cf command", func() {
			output := Cf("test_1_cmd1_alias")
			Eventually(output.Out).Should(Say("You called cmd1 in test_1"))
		})

		It("Can call another plugin command when more than one plugin is installed", func() {
			output := Cf("test_2_cmd1")
			Eventually(output.Out).Should(Say("You called cmd1 in test_2"))
		})

		It("hide suggetsions for commands that aren't close to anything", func() {
			output := Cf("this-does-not-match-any-command")
			Eventually(output.Out, 3*time.Second).Should(Say("'this-does-not-match-any-command' is not a registered command. See 'cf help'"))
			Consistently(output.Out, 3*time.Second).ShouldNot(Say("Did you mean?\n help"))
		})

		It("informs user for any invalid commands", func() {
			output := Cf("foo-bar")
			Eventually(output.Out, 3*time.Second).Should(Say("'foo-bar' is not a registered command"))
		})

		It("Calls help if the plugin shares the same name", func() {
			output := Cf("help")
			Consistently(output.Out, 1).ShouldNot(Say("You called help in test_with_help"))
		})

		It("shows help with a '-h' or '--help' flag in plugin command", func() {
			output := Cf("test_1_cmd1", "-h")
			Consistently(output.Out).ShouldNot(Say("You called cmd1 in test_1"))
			Eventually(output.Out.Contents).Should(ContainSubstring("USAGE:"))
			Eventually(output.Out.Contents).Should(ContainSubstring("OPTIONS:"))
		})

		It("Calls the core push command if the plugin shares the same name", func() {
			output := Cf("push")
			Consistently(output.Out, 1).ShouldNot(Say("You called push in test_with_push"))
		})

		It("Calls the core short name if a plugin shares the same name", func() {
			output := Cf("p")
			Consistently(output.Out, 1).ShouldNot(Say("You called p within the plugin"))
		})

		It("Passes all arguments to a plugin", func() {
			output := Cf("my-say", "foo")
			Eventually(output.Out).Should(Say("foo"))
		})

		It("Passes all arguments and flags to a plugin", func() {
			output := Cf("my-say", "foo", "--loud")
			Eventually(output.Out).Should(Say("FOO"))
		})

		It("Calls a plugin that calls core commands", func() {
			output := Cf("awesomeness")
			Eventually(output.Out).Should(Say("my-say")) //look for another plugin
		})

		It("Sends stdoutput to the plugin to echo", func() {
			output := Cf("core-command", "plugins")
			Eventually(output.Out.Contents).Should(MatchRegexp("Command output from the plugin(.*\\W)*awesomeness(.*\\W)*FIN"))
		})

		It("Can call a core commmand from a plugin without terminal output", func() {
			output := Cf("core-command-quiet", "plugins")
			Eventually(output.Out.Contents).Should(MatchRegexp("^\n---------- Command output from the plugin"))
		})

		It("Can call a plugin that requires stdin (interactive)", func() {
			session := CfWithIo("input", "silly\n")
			Eventually(session.Out).Should(Say("silly"))
		})

		It("exits 1 when a plugin panics", func() {
			session := Cf("panic")
			Eventually(session).Should(Exit(1))
		})

		It("exits 1 when a plugin exits 1", func() {
			session := Cf("exit1")
			Eventually(session).Should(Exit(1))
		})

		It("show user suggested plugin commands for typos", func() {
			output := Cf("test_1_cmd")
			Eventually(output.Out, 3*time.Second).Should(Say("'test_1_cmd' is not a registered command. See 'cf help'"))
			Eventually(output.Out, 3*time.Second).Should(Say("Did you mean?"))
		})
	})
})

func Cf(args ...string) *Session {
	session, err := Start(exec.Command(buildPath, args...), GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())

	return session
}

func CfWithIo(command string, args string) *Session {
	cmd := exec.Command(buildPath, command)

	stdin, err := cmd.StdinPipe()
	Expect(err).ToNot(HaveOccurred())

	buffer := bufio.NewWriter(stdin)
	buffer.WriteString(args)
	buffer.Flush()

	session, err := Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())

	return session
}

func CfWith_CF_HOME(cfHome string, args ...string) *Session {
	cmd := exec.Command(buildPath, args...)
	cmd.Env = append(cmd.Env, "CF_HOME="+cfHome)
	session, err := Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())

	return session
}
