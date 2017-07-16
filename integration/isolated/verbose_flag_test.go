package isolated

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/util/configv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Verbose", func() {
	Context("v2 legacy", func() {
		DescribeTable("displays verbose output",
			func(command func() *Session) {
				helpers.LoginCF()

				session := command()
				Eventually(session).Should(Say("REQUEST:"))
				Eventually(session).Should(Say("GET /v2/organizations"))
				Eventually(session).Should(Say("RESPONSE:"))
				Eventually(session).Should(Exit(0))
			},

			Entry("when the -v option is provided with additional command", func() *Session {
				return helpers.CF("-v", "orgs")
			}),

			Entry("when the CF_TRACE env variable is set", func() *Session {
				return helpers.CFWithEnv(map[string]string{"CF_TRACE": "true"}, "orgs")
			}),
		)
	})

	Context("v2 refactor", func() {
		DescribeTable("displays verbose output to terminal",
			func(env string, configTrace string, flag bool) {
				tmpDir, err := ioutil.TempDir("", "")
				defer os.RemoveAll(tmpDir)
				Expect(err).NotTo(HaveOccurred())

				setupCF(ReadOnlyOrg, ReadOnlySpace)

				var envMap map[string]string
				if env != "" {
					if string(env[0]) == "/" {
						env = filepath.Join(tmpDir, env)
					}
					envMap = map[string]string{"CF_TRACE": env}
				}

				// We use 'create-user' because it makes a request via the UAA client
				// and a request via the CC client, testing the logging wrapper in both
				// clients.
				randomUsername := helpers.RandomUsername()
				randomPassword := helpers.RandomPassword()
				command := []string{"create-user", randomUsername, randomPassword}

				if flag {
					command = append(command, "-v")
				}

				if configTrace != "" {
					if string(configTrace[0]) == "/" {
						configTrace = filepath.Join(tmpDir, configTrace)
					}
					session := helpers.CF("config", "--trace", configTrace)
					Eventually(session).Should(Exit(0))
				}

				session := helpers.CFWithEnv(envMap, command...)

				Eventually(session).Should(Say("REQUEST:"))
				Eventually(session).Should(Say("GET /v2/info"))
				Eventually(session).Should(Say("RESPONSE:"))
				Eventually(session).Should(Say(`"token_endpoint": "http.*"`))
				Eventually(session).Should(Say("REQUEST:"))
				Eventually(session).Should(Say("POST /Users"))
				Eventually(session).Should(Say("User-Agent: cf/[\\w.+-]+ \\(go\\d+\\.\\d+\\.\\d+; %s %s\\)", runtime.GOARCH, runtime.GOOS))
				Eventually(session).Should(Say("RESPONSE:"))
				Eventually(session).Should(Say("REQUEST:"))
				Eventually(session).Should(Say("POST /v2/users"))
				Eventually(session).Should(Say("User-Agent: cf/[\\w.+-]+ \\(go\\d+\\.\\d+\\.\\d+; %s %s\\)", runtime.GOARCH, runtime.GOOS))
				Eventually(session).Should(Say("RESPONSE:"))
				Eventually(session).Should(Exit(0))
			},

			Entry("CF_TRACE true: enables verbose", "true", "", false),
			Entry("CF_TRACE true, config trace false: enables verbose", "true", "false", false),
			Entry("CF_TRACE true, config trace file path: enables verbose AND logging to file", "true", "/foo", false),

			Entry("CF_TRACE false, '-v': enables verbose", "false", "", true),
			Entry("CF_TRACE false, config trace file path, '-v': enables verbose AND logging to file", "false", "/foo", true),

			Entry("CF_TRACE empty:, '-v': enables verbose", "", "", true),
			Entry("CF_TRACE empty, config trace true: enables verbose", "", "true", false),
			Entry("CF_TRACE empty, config trace file path, '-v': enables verbose AND logging to file", "", "/foo", true),

			Entry("CF_TRACE filepath, '-v': enables logging to file", "/foo", "", true),
			Entry("CF_TRACE filepath, config trace true: enables verbose AND logging to file", "/foo", "true", false),
			Entry("CF_TRACE filepath, config trace filepath, '-v': enables verbose AND logging to file for BOTH paths", "/foo", "/bar", true),
		)

		DescribeTable("displays verbose output to multiple files",
			func(env string, configTrace string, flag bool, location []string) {
				tmpDir, err := ioutil.TempDir("", "")
				defer os.RemoveAll(tmpDir)
				Expect(err).NotTo(HaveOccurred())

				setupCF(ReadOnlyOrg, ReadOnlySpace)

				var envMap map[string]string
				if env != "" {
					if string(env[0]) == "/" {
						env = filepath.Join(tmpDir, env)
					}
					envMap = map[string]string{"CF_TRACE": env}
				}

				// We use 'create-user' because it makes a request via the UAA client
				// and a request via the CC client, testing the logging wrapper in both
				// clients.
				randomUsername := helpers.RandomUsername()
				randomPassword := helpers.RandomPassword()
				command := []string{"create-user", randomUsername, randomPassword}

				if flag {
					command = append(command, "-v")
				}

				if configTrace != "" {
					if string(configTrace[0]) == "/" {
						configTrace = filepath.Join(tmpDir, configTrace)
					}
					session := helpers.CF("config", "--trace", configTrace)
					Eventually(session).Should(Exit(0))
				}

				session := helpers.CFWithEnv(envMap, command...)
				Eventually(session).Should(Exit(0))

				for _, filePath := range location {
					contents, err := ioutil.ReadFile(tmpDir + filePath)
					Expect(err).ToNot(HaveOccurred())

					Expect(string(contents)).To(MatchRegexp("REQUEST:"))
					Expect(string(contents)).To(MatchRegexp("POST /Users"))
					Expect(string(contents)).To(MatchRegexp("RESPONSE:"))
					Expect(string(contents)).To(MatchRegexp("REQUEST:"))
					Expect(string(contents)).To(MatchRegexp("POST /v2/users"))
					Expect(string(contents)).To(MatchRegexp("RESPONSE:"))

					stat, err := os.Stat(tmpDir + filePath)
					Expect(err).ToNot(HaveOccurred())

					if runtime.GOOS == "windows" {
						Expect(stat.Mode().String()).To(Equal(os.FileMode(0666).String()))
					} else {
						Expect(stat.Mode().String()).To(Equal(os.FileMode(0600).String()))
					}
				}
			},

			Entry("CF_Trace true, config trace file path: enables verbose AND logging to file", "true", "/foo", false, []string{"/foo"}),

			Entry("CF_TRACE false, config trace file path: enables logging to file", "false", "/foo", false, []string{"/foo"}),
			Entry("CF_TRACE false, config trace file path, '-v': enables verbose AND logging to file", "false", "/foo", true, []string{"/foo"}),

			Entry("CF_TRACE empty, config trace file path: enables logging to file", "", "/foo", false, []string{"/foo"}),
			Entry("CF_TRACE empty, config trace file path, '-v': enables verbose AND logging to file", "", "/foo", true, []string{"/foo"}),

			Entry("CF_TRACE filepath: enables logging to file", "/foo", "", false, []string{"/foo"}),
			Entry("CF_TRACE filepath, '-v': enables logging to file", "/foo", "", true, []string{"/foo"}),
			Entry("CF_TRACE filepath, config trace true: enables verbose AND logging to file", "/foo", "true", false, []string{"/foo"}),
			Entry("CF_TRACE filepath, config trace filepath: enables logging to file for BOTH paths", "/foo", "/bar", false, []string{"/foo", "/bar"}),
			Entry("CF_TRACE filepath, config trace filepath, '-v': enables verbose AND logging to file for BOTH paths", "/foo", "/bar", true, []string{"/foo", "/bar"}),
		)
	})

	Context("v3", func() {
		DescribeTable("displays verbose output to terminal",
			func(env string, configTrace string, flag bool) {
				tmpDir, err := ioutil.TempDir("", "")
				defer os.RemoveAll(tmpDir)
				Expect(err).NotTo(HaveOccurred())

				setupCF(ReadOnlyOrg, ReadOnlySpace)

				// Invalidate the access token to cause a token refresh in order to
				// test the call to the UAA.
				helpers.SetConfig(func(config *configv3.Config) {
					config.ConfigFile.AccessToken = helpers.InvalidAccessToken()
				})

				var envMap map[string]string
				if env != "" {
					if string(env[0]) == "/" {
						env = filepath.Join(tmpDir, env)
					}
					envMap = map[string]string{"CF_TRACE": env}
				}

				command := []string{"run-task", "app", "echo"}

				if flag {
					command = append(command, "-v")
				}

				if configTrace != "" {
					if string(configTrace[0]) == "/" {
						configTrace = filepath.Join(tmpDir, configTrace)
					}
					session := helpers.CF("config", "--trace", configTrace)
					Eventually(session).Should(Exit(0))
				}

				session := helpers.CFWithEnv(envMap, command...)

				Eventually(session).Should(Say("REQUEST:"))
				Eventually(session).Should(Say("GET /v3/apps"))
				Eventually(session).Should(Say("User-Agent: cf/[\\w.+-]+ \\(go\\d+\\.\\d+\\.\\d+; %s %s\\)", runtime.GOARCH, runtime.GOOS))
				Eventually(session).Should(Say("RESPONSE:"))
				Eventually(session).Should(Say("REQUEST:"))
				Eventually(session).Should(Say("POST /oauth/token"))
				Eventually(session).Should(Say("User-Agent: cf/[\\w.+-]+ \\(go\\d+\\.\\d+\\.\\d+; %s %s\\)", runtime.GOARCH, runtime.GOOS))
				Eventually(session).Should(Say("\\[PRIVATE DATA HIDDEN\\]")) //This is required to test the previous line. If it fails, the previous matcher went too far.
				Eventually(session).Should(Say("RESPONSE:"))
				Eventually(session).Should(Exit(1))
			},

			Entry("CF_TRACE true: enables verbose", "true", "", false),
			Entry("CF_Trace true, config trace false: enables verbose", "true", "false", false),
			Entry("CF_Trace true, config trace file path: enables verbose AND logging to file", "true", "/foo", false),

			Entry("CF_TRACE false, '-v': enables verbose", "false", "", true),
			Entry("CF_TRACE false, config trace file path, '-v': enables verbose AND logging to file", "false", "/foo", true),

			Entry("CF_TRACE empty:, '-v': enables verbose", "", "", true),
			Entry("CF_TRACE empty, config trace true: enables verbose", "", "true", false),
			Entry("CF_TRACE empty, config trace file path, '-v': enables verbose AND logging to file", "", "/foo", true),

			Entry("CF_TRACE filepath, '-v': enables logging to file", "/foo", "", true),
			Entry("CF_TRACE filepath, config trace true: enables verbose AND logging to file", "/foo", "true", false),
			Entry("CF_TRACE filepath, config trace filepath, '-v': enables verbose AND logging to file for BOTH paths", "/foo", "/bar", true),
		)

		DescribeTable("displays verbose output to multiple files",
			func(env string, configTrace string, flag bool, location []string) {
				tmpDir, err := ioutil.TempDir("", "")
				defer os.RemoveAll(tmpDir)
				Expect(err).NotTo(HaveOccurred())

				setupCF(ReadOnlyOrg, ReadOnlySpace)

				// Invalidate the access token to cause a token refresh in order to
				// test the call to the UAA.
				helpers.SetConfig(func(config *configv3.Config) {
					config.ConfigFile.AccessToken = helpers.InvalidAccessToken()
				})

				var envMap map[string]string
				if env != "" {
					if string(env[0]) == "/" {
						env = filepath.Join(tmpDir, env)
					}
					envMap = map[string]string{"CF_TRACE": env}
				}

				command := []string{"run-task", "app", "echo"}

				if flag {
					command = append(command, "-v")
				}

				if configTrace != "" {
					if string(configTrace[0]) == "/" {
						configTrace = filepath.Join(tmpDir, configTrace)
					}
					session := helpers.CF("config", "--trace", configTrace)
					Eventually(session).Should(Exit(0))
				}

				session := helpers.CFWithEnv(envMap, command...)
				Eventually(session).Should(Exit(1))

				for _, filePath := range location {
					contents, err := ioutil.ReadFile(tmpDir + filePath)
					Expect(err).ToNot(HaveOccurred())

					Expect(string(contents)).To(MatchRegexp("REQUEST:"))
					Expect(string(contents)).To(MatchRegexp("GET /v3/apps"))
					Expect(string(contents)).To(MatchRegexp("RESPONSE:"))
					Expect(string(contents)).To(MatchRegexp("REQUEST:"))
					Expect(string(contents)).To(MatchRegexp("POST /oauth/token"))
					Expect(string(contents)).To(MatchRegexp("RESPONSE:"))

					stat, err := os.Stat(tmpDir + filePath)
					Expect(err).ToNot(HaveOccurred())

					if runtime.GOOS == "windows" {
						Expect(stat.Mode().String()).To(Equal(os.FileMode(0666).String()))
					} else {
						Expect(stat.Mode().String()).To(Equal(os.FileMode(0600).String()))
					}
				}
			},

			Entry("CF_Trace true, config trace file path: enables verbose AND logging to file", "true", "/foo", false, []string{"/foo"}),

			Entry("CF_TRACE false, config trace file path: enables logging to file", "false", "/foo", false, []string{"/foo"}),
			Entry("CF_TRACE false, config trace file path, '-v': enables verbose AND logging to file", "false", "/foo", true, []string{"/foo"}),

			Entry("CF_TRACE empty, config trace file path: enables logging to file", "", "/foo", false, []string{"/foo"}),
			Entry("CF_TRACE empty, config trace file path, '-v': enables verbose AND logging to file", "", "/foo", true, []string{"/foo"}),

			Entry("CF_TRACE filepath: enables logging to file", "/foo", "", false, []string{"/foo"}),
			Entry("CF_TRACE filepath, '-v': enables logging to file", "/foo", "", true, []string{"/foo"}),
			Entry("CF_TRACE filepath, config trace true: enables verbose AND logging to file", "/foo", "true", false, []string{"/foo"}),
			Entry("CF_TRACE filepath, config trace filepath: enables logging to file for BOTH paths", "/foo", "/bar", false, []string{"/foo", "/bar"}),
			Entry("CF_TRACE filepath, config trace filepath, '-v': enables verbose AND logging to file for BOTH paths", "/foo", "/bar", true, []string{"/foo", "/bar"}),
		)
	})

	Describe("NOAA", func() {
		DescribeTable("displays verbose output to terminal",
			func(env string, configTrace string, flag bool) {
				tmpDir, err := ioutil.TempDir("", "")
				defer os.RemoveAll(tmpDir)
				Expect(err).NotTo(HaveOccurred())

				orgName := helpers.NewOrgName()
				spaceName := helpers.NewSpaceName()

				appName := helpers.PrefixedRandomName("app")

				setupCF(orgName, spaceName)

				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CF("push", appName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
				})

				var envMap map[string]string
				if env != "" {
					if string(env[0]) == "/" {
						env = filepath.Join(tmpDir, env)
					}
					envMap = map[string]string{"CF_TRACE": env}
				}

				command := []string{"logs", appName}

				if flag {
					command = append(command, "-v")
				}

				if configTrace != "" {
					if string(configTrace[0]) == "/" {
						configTrace = filepath.Join(tmpDir, configTrace)
					}
					session := helpers.CF("config", "--trace", configTrace)
					Eventually(session).Should(Exit(0))
				}

				session := helpers.CFWithEnv(envMap, command...)

				Eventually(session).Should(Say("REQUEST:"))
				Eventually(session).Should(Say("POST /oauth/token"))
				Eventually(session).Should(Say("\\[PRIVATE DATA HIDDEN\\]"))
				Eventually(session).Should(Say("WEBSOCKET REQUEST:"))
				Eventually(session).Should(Say("Authorization: \\[PRIVATE DATA HIDDEN\\]"))
				Eventually(session.Kill()).Should(Exit())
			},

			Entry("CF_TRACE true: enables verbose", "true", "", false),
			Entry("CF_Trace true, config trace false: enables verbose", "true", "false", false),
			Entry("CF_Trace true, config trace file path: enables verbose AND logging to file", "true", "/foo", false),

			Entry("CF_TRACE false, '-v': enables verbose", "false", "", true),
			Entry("CF_TRACE false, config trace file path, '-v': enables verbose AND logging to file", "false", "/foo", true),

			Entry("CF_TRACE empty:, '-v': enables verbose", "", "", true),
			Entry("CF_TRACE empty, config trace true: enables verbose", "", "true", false),
			Entry("CF_TRACE empty, config trace file path, '-v': enables verbose AND logging to file", "", "/foo", true),

			Entry("CF_TRACE filepath, '-v': enables logging to file", "/foo", "", true),
			Entry("CF_TRACE filepath, config trace true: enables verbose AND logging to file", "/foo", "true", false),
			Entry("CF_TRACE filepath, config trace filepath, '-v': enables verbose AND logging to file for BOTH paths", "/foo", "/bar", true),
		)

		DescribeTable("displays verbose output to multiple files",
			func(env string, configTrace string, location []string) {
				tmpDir, err := ioutil.TempDir("", "")
				defer os.RemoveAll(tmpDir)
				Expect(err).NotTo(HaveOccurred())

				orgName := helpers.NewOrgName()
				spaceName := helpers.NewSpaceName()

				appName := helpers.PrefixedRandomName("app")

				setupCF(orgName, spaceName)

				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CF("push", appName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
				})

				var envMap map[string]string
				if env != "" {
					if string(env[0]) == "/" {
						env = filepath.Join(tmpDir, env)
					}
					envMap = map[string]string{"CF_TRACE": env}
				}

				if configTrace != "" {
					if strings.HasPrefix(configTrace, "/") {
						configTrace = filepath.Join(tmpDir, configTrace)
					}
					session := helpers.CF("config", "--trace", configTrace)
					Eventually(session).Should(Exit(0))
				}

				session := helpers.CFWithEnv(envMap, "logs", "-v", appName)

				Eventually(session).Should(Say("WEBSOCKET RESPONSE"))
				Eventually(session.Kill()).Should(Exit())

				for _, filePath := range location {
					contents, err := ioutil.ReadFile(tmpDir + filePath)
					Expect(err).ToNot(HaveOccurred())

					Expect(string(contents)).To(MatchRegexp("REQUEST:"))
					Expect(string(contents)).To(MatchRegexp("POST /oauth/token"))
					Expect(string(contents)).To(MatchRegexp("\\[PRIVATE DATA HIDDEN\\]"))
					Expect(string(contents)).To(MatchRegexp("WEBSOCKET REQUEST:"))
					Expect(string(contents)).To(MatchRegexp("Authorization: \\[PRIVATE DATA HIDDEN\\]"))

					stat, err := os.Stat(tmpDir + filePath)
					Expect(err).ToNot(HaveOccurred())

					if runtime.GOOS == "windows" {
						Expect(stat.Mode().String()).To(Equal(os.FileMode(0666).String()))
					} else {
						Expect(stat.Mode().String()).To(Equal(os.FileMode(0600).String()))
					}
				}
			},

			Entry("CF_Trace true, config trace file path: enables verbose AND logging to file", "true", "/foo", []string{"/foo"}),

			Entry("CF_TRACE false, config trace file path: enables logging to file", "false", "/foo", []string{"/foo"}),
			Entry("CF_TRACE false, config trace file path, '-v': enables verbose AND logging to file", "false", "/foo", []string{"/foo"}),

			Entry("CF_TRACE empty, config trace file path: enables logging to file", "", "/foo", []string{"/foo"}),
			Entry("CF_TRACE empty, config trace file path, '-v': enables verbose AND logging to file", "", "/foo", []string{"/foo"}),

			Entry("CF_TRACE filepath: enables logging to file", "/foo", "", []string{"/foo"}),
			Entry("CF_TRACE filepath, '-v': enables logging to file", "/foo", "", []string{"/foo"}),
			Entry("CF_TRACE filepath, config trace true: enables verbose AND logging to file", "/foo", "true", []string{"/foo"}),
			Entry("CF_TRACE filepath, config trace filepath: enables logging to file for BOTH paths", "/foo", "/bar", []string{"/foo", "/bar"}),
			Entry("CF_TRACE filepath, config trace filepath, '-v': enables verbose AND logging to file for BOTH paths", "/foo", "/bar", []string{"/foo", "/bar"}),
		)
	})
})
