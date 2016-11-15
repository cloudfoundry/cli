package integration

import (
	"io/ioutil"
	"os"
	"os/exec"

	helpers "code.cloudfoundry.org/cli/integration/helpers"
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
				login := exec.Command("cf", "auth", "admin", "admin")
				loginSession, err := Start(login, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(loginSession).Should(Exit(0))

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
				orgName := helpers.PrefixedRandomName("testorg")
				spaceName := helpers.PrefixedRandomName("testspace")
				setupCF(orgName, spaceName)

				defer helpers.CF("delete-org", "-f", orgName)

				var envMap map[string]string
				if env != "" {
					envMap = map[string]string{"CF_TRACE": env}
				}
				command := []string{"delete-orphaned-routes", "-f"}
				if flag {
					command = append(command, "-v")
				}
				if configTrace != "" {
					session := helpers.CF("config", "--trace", configTrace)
					Eventually(session).Should(Exit(0))
				}
				session := helpers.CFWithEnv(envMap, command...)

				Eventually(session).Should(Say("REQUEST:"))
				Eventually(session).Should(Say("GET /v2/spaces"))
				Eventually(session).Should(Say("RESPONSE:"))
				Eventually(session).Should(Exit(0))
			},

			Entry("CF_TRACE true: enables verbose", "true", "", false),
			Entry("CF_Trace true, config trace false: enables verbose", "true", "false", false),
			Entry("CF_Trace true, config trace file path: enables verbose AND logging to file", "true", "/foo/bar", false),

			Entry("CF_TRACE false, '-v': enables verbose", "false", "", true),
			Entry("CF_TRACE false, config trace true: enables verbose", "false", "true", false),
			Entry("CF_TRACE false, config trace file path, '-v': enables verbose AND logging to file", "false", "/foo/bar", true),

			Entry("CF_TRACE empty:, '-v': enables verbose", "", "", true),
			Entry("CF_TRACE empty, config trace true: enables verbose", "", "true", false),
			Entry("CF_TRACE empty, config trace file path, '-v': enables verbose AND logging to file", "", "/foo/bar", true),

			Entry("CF_TRACE filepath, '-v': enables logging to file", "/foo/bar", "", true),
			Entry("CF_TRACE filepath, config trace true: enables verbose AND logging to file", "/foo/bar", "true", false),
			Entry("CF_TRACE filepath, config trace filepath, '-v': enables verbose AND logging to file for BOTH paths", "/foo/bar", "/baz", true),
		)

		DescribeTable("displays verbose output to file",
			func(env string, configTrace string, flag bool, location []string) {
				tmpDir, err := ioutil.TempDir("", "")
				Expect(err).NotTo(HaveOccurred())

				defer os.RemoveAll(tmpDir)

				orgName := helpers.PrefixedRandomName("testorg")
				spaceName := helpers.PrefixedRandomName("testspace")
				setupCF(orgName, spaceName)

				defer helpers.CF("delete-org", "-f", orgName)

				var envMap map[string]string
				if env != "" {
					if string(env[0]) == "/" {
						env = tmpDir + env
					}
					envMap = map[string]string{"CF_TRACE": env}
				}
				command := []string{"delete-orphaned-routes", "-f"}
				if flag {
					command = append(command, "-v")
				}

				if configTrace != "" {
					session := helpers.CF("config", "--trace", tmpDir+configTrace)
					Eventually(session).Should(Exit(0))
				}
				session := helpers.CFWithEnv(envMap, command...)
				Eventually(session).Should(Exit(0))

				for _, filePath := range location {
					contents, err := ioutil.ReadFile(tmpDir + filePath)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(contents)).To(MatchRegexp("REQUEST:"))
					Expect(string(contents)).To(MatchRegexp("GET /v2/spaces"))
					Expect(string(contents)).To(MatchRegexp("RESPONSE:"))
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
				orgName := helpers.PrefixedRandomName("testorg")
				spaceName := helpers.PrefixedRandomName("testspace")
				setupCF(orgName, spaceName)

				defer helpers.CF("delete-org", "-f", orgName)

				var envMap map[string]string
				if env != "" {
					envMap = map[string]string{"CF_TRACE": env}
				}
				command := []string{"run-task", "app", "echo"}
				if flag {
					command = append(command, "-v")
				}
				if configTrace != "" {
					session := helpers.CF("config", "--trace", configTrace)
					Eventually(session).Should(Exit(0))
				}
				session := helpers.CFWithEnv(envMap, command...)

				Eventually(session).Should(Say("REQUEST:"))
				Eventually(session).Should(Say("GET /v3/apps"))
				Eventually(session).Should(Say("RESPONSE:"))
				Eventually(session).Should(Exit(1))
			},

			Entry("CF_TRACE true: enables verbose", "true", "", false),
			Entry("CF_Trace true, config trace false: enables verbose", "true", "false", false),
			Entry("CF_Trace true, config trace file path: enables verbose AND logging to file", "true", "/foo/bar", false),

			Entry("CF_TRACE false, '-v': enables verbose", "false", "", true),
			Entry("CF_TRACE false, config trace true: enables verbose", "false", "true", false),
			Entry("CF_TRACE false, config trace file path, '-v': enables verbose AND logging to file", "false", "/foo/bar", true),

			Entry("CF_TRACE empty:, '-v': enables verbose", "", "", true),
			Entry("CF_TRACE empty, config trace true: enables verbose", "", "true", false),
			Entry("CF_TRACE empty, config trace file path, '-v': enables verbose AND logging to file", "", "/foo/bar", true),

			Entry("CF_TRACE filepath, '-v': enables logging to file", "/foo/bar", "", true),
			Entry("CF_TRACE filepath, config trace true: enables verbose AND logging to file", "/foo/bar", "true", false),
			Entry("CF_TRACE filepath, config trace filepath, '-v': enables verbose AND logging to file for BOTH paths", "/foo/bar", "/baz", true),
		)

		DescribeTable("displays verbose output to file",
			func(env string, configTrace string, flag bool, location []string) {
				tmpDir, err := ioutil.TempDir("", "")
				Expect(err).NotTo(HaveOccurred())

				defer os.RemoveAll(tmpDir)

				orgName := helpers.PrefixedRandomName("testorg")
				spaceName := helpers.PrefixedRandomName("testspace")
				setupCF(orgName, spaceName)

				defer helpers.CF("delete-org", "-f", orgName)

				var envMap map[string]string
				if env != "" {
					if string(env[0]) == "/" {
						env = tmpDir + env
					}
					envMap = map[string]string{"CF_TRACE": env}
				}
				command := []string{"run-task", "app", "echo"}
				if flag {
					command = append(command, "-v")
				}

				if configTrace != "" {
					session := helpers.CF("config", "--trace", tmpDir+configTrace)
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
})
