package runner_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"sync/atomic"

	"code.cloudfoundry.org/cli/v8/plugin/runner"
	"code.cloudfoundry.org/cli/v8/util/configv3"
	"code.cloudfoundry.org/cli/v8/util/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("PluginRunner", func() {
	var (
		config    *configv3.Config
		commandUI *ui.UI
		plugin    configv3.Plugin
		out       *Buffer
		cfHome    string
	)

	BeforeEach(func() {
		// Token persistence must never touch the developer's real ~/.cf.
		cfHome = GinkgoT().TempDir()
		GinkgoT().Setenv("CF_HOME", cfHome)

		out = NewBuffer()
		commandUI = ui.NewTestUI(nil, out, NewBuffer())

		config = &configv3.Config{
			ConfigFile: configv3.JSONConfig{
				ConfigVersion: 3,
				Target:        "https://api.example.com",
				PluginRepositories: []configv3.PluginRepository{
					{
						Name: "CF-Community",
						URL:  "https://plugins.cloudfoundry.org",
					},
				},
			},
		}

		plugin = configv3.Plugin{
			Name:     "test-plugin",
			Location: "/path/to/test-plugin",
			Version: configv3.PluginVersion{
				Major: 1,
				Minor: 0,
				Build: 0,
			},
			Commands: []configv3.PluginCommand{
				{
					Name:     "test-command",
					Alias:    "tc",
					HelpText: "A test command",
					UsageDetails: configv3.PluginUsageDetails{
						Usage: "cf test-command [options]",
					},
				},
				{
					Name:     "another-command",
					HelpText: "Another test command",
					UsageDetails: configv3.PluginUsageDetails{
						Usage: "cf another-command",
					},
				},
			},
		}
	})

	Describe("NewPluginRunner", func() {
		It("creates a new PluginRunner instance", func() {
			pluginRunner := runner.NewPluginRunner(config, commandUI, plugin, nil)
			Expect(pluginRunner).ToNot(BeNil())
		})

		It("accepts nil config", func() {
			pluginRunner := runner.NewPluginRunner(nil, commandUI, plugin, nil)
			Expect(pluginRunner).ToNot(BeNil())
		})

		It("accepts nil UI", func() {
			pluginRunner := runner.NewPluginRunner(config, nil, plugin, nil)
			Expect(pluginRunner).ToNot(BeNil())
		})

		It("accepts empty plugin", func() {
			emptyPlugin := configv3.Plugin{}
			pluginRunner := runner.NewPluginRunner(config, commandUI, emptyPlugin, nil)
			Expect(pluginRunner).ToNot(BeNil())
		})
	})

	Describe("Run", func() {
		var pluginRunner runner.PluginRunner

		BeforeEach(func() {
			pluginRunner = runner.NewPluginRunner(config, commandUI, plugin, nil)
		})

		Context("when the plugin binary does not exist", func() {
			It("returns ErrFailed", func() {
				err := pluginRunner.Run([]string{"test-command"})
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(runner.ErrFailed))
			})
		})

		Context("when the plugin binary path is empty", func() {
			BeforeEach(func() {
				plugin.Location = ""
				pluginRunner = runner.NewPluginRunner(config, commandUI, plugin, nil)
			})

			It("returns an error", func() {
				err := pluginRunner.Run([]string{"test-command"})
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when args are empty", func() {
			It("can still execute (plugin determines behavior)", func() {
				// This will fail because the binary doesn't exist, but it tests
				// that empty args don't cause a panic
				err := pluginRunner.Run([]string{})
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(runner.ErrFailed))
			})
		})

		Context("when no API endpoint is targeted", func() {
			BeforeEach(func() {
				config.ConfigFile.Target = ""
				pluginRunner = runner.NewPluginRunner(config, commandUI, plugin, nil)
			})

			It("still launches the plugin binary instead of failing on the missing target", func() {
				err := pluginRunner.Run([]string{"test-command"})
				Expect(err).To(Equal(runner.ErrFailed))
			})
		})

		Context("when the plugin exits", func() {
			BeforeEach(func() {
				config.ConfigFile.ConfigVersion = configv3.CurrentConfigVersion
				config.ConfigFile.Target = ""
				config.ConfigFile.Locale = "en-US"
				Expect(config.WriteConfig()).To(Succeed())
				var err error
				plugin.Location, err = os.Executable()
				Expect(err).ToNot(HaveOccurred())
			})

			DescribeTable("preserves disk edits when tokens have not changed",
				func(fail bool) {
					GinkgoT().Setenv("CF_CLI_RUNNER_TEST_HELPER", "write-locale")
					pluginRunner = runner.NewPluginRunner(config, commandUI, plugin, nil)
					args := []string{"test-command"}
					if fail {
						args = append(args, "fail")
					}
					err := pluginRunner.Run(args)
					if fail {
						Expect(err).To(BeAssignableToTypeOf(&exec.ExitError{}))
						Expect(err.(*exec.ExitError).ExitCode()).To(Equal(3))
					} else {
						Expect(err).ToNot(HaveOccurred())
					}
					persisted, err := configv3.LoadConfig()
					Expect(err).ToNot(HaveOccurred())
					Expect(persisted.ConfigFile.Locale).To(Equal("fr-FR"))
				},
				Entry("after success", false),
				Entry("after failure", true),
			)

			DescribeTable("persists tokens refreshed by a Get call's request wrapper",
				func(accessToken, refreshToken string, fail bool) {
					GinkgoT().Setenv("CF_CLI_RUNNER_TEST_HELPER", "get-orgs")
					var refreshes atomic.Int32
					server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.Header().Set("Content-Type", "application/json")
						switch r.URL.Path {
						case "/oauth/token":
							refreshes.Add(1)
							_ = json.NewEncoder(w).Encode(map[string]string{
								"access_token": accessToken, "refresh_token": refreshToken, "token_type": "bearer",
							})
						case "/v3/organizations":
							if fail {
								w.WriteHeader(http.StatusForbidden)
								_, _ = w.Write([]byte(`{"errors":[{"code":10003,"title":"CF-NotAuthorized","detail":"request failed after refresh"}]}`))
							} else {
								_, _ = w.Write([]byte(`{"resources":[],"pagination":{"next":null}}`))
							}
						default:
							w.WriteHeader(http.StatusNotFound)
						}
					}))
					DeferCleanup(server.Close)
					config.ConfigFile.Target = server.URL
					config.ConfigFile.UAAEndpoint = server.URL
					config.ConfigFile.AuthorizationEndpoint = server.URL
					config.ConfigFile.AccessToken = "bearer old-access-token"
					config.ConfigFile.RefreshToken = "old-refresh-token"
					Expect(config.WriteConfig()).To(Succeed())
					pluginRunner = runner.NewPluginRunner(config, commandUI, plugin, nil)
					err := pluginRunner.Run([]string{"test-command"})
					if fail {
						Expect(err).To(BeAssignableToTypeOf(&exec.ExitError{}))
						Expect(err.(*exec.ExitError).ExitCode()).To(Equal(3))
					} else {
						Expect(err).ToNot(HaveOccurred())
					}
					Expect(refreshes.Load()).To(Equal(int32(1)))
					persisted, err := configv3.LoadConfig()
					Expect(err).ToNot(HaveOccurred())
					Expect(persisted.ConfigFile.AccessToken).To(Equal("bearer " + accessToken))
					Expect(persisted.ConfigFile.RefreshToken).To(Equal(refreshToken))
				},
				Entry("when both tokens change", "new-access-token", "new-refresh-token", false),
				Entry("when only the access token changes", "new-access-token", "old-refresh-token", false),
				Entry("when only the refresh token changes", "old-access-token", "new-refresh-token", false),
				Entry("when the request fails after refreshing", "new-access-token", "new-refresh-token", true),
			)

			It("does not create config when plugin launch fails without changing tokens", func() {
				plugin.Location = "/path/to/nonexistent-plugin"
				Expect(os.Remove(configv3.ConfigFilePath())).To(Succeed())
				pluginRunner = runner.NewPluginRunner(config, commandUI, plugin, nil)
				Expect(pluginRunner.Run([]string{"test-command"})).To(Equal(runner.ErrFailed))
				Expect(configv3.ConfigFilePath()).ToNot(BeAnExistingFile())
			})
		})

		// The subprocess tests use this test executable as a plugin and local HTTP
		// servers. CF-backed integration coverage remains in the integration suites.
	})

	DescribeTable("CommandArgs",
		func(args []string, expected []string) {
			Expect(runner.CommandArgs(args)).To(Equal(expected))
		},
		Entry("command first", []string{"my-cmd", "-f", "x"}, []string{"my-cmd", "-f", "x"}),
		Entry("short global flag before the command", []string{"-v", "my-cmd", "arg"}, []string{"my-cmd", "arg"}),
		Entry("long global flags before the command", []string{"--verbose", "-v", "my-cmd", "--opt", "val"}, []string{"my-cmd", "--opt", "val"}),
		Entry("flags only", []string{"-v", "--help"}, nil),
		Entry("no arguments", []string{}, nil),
	)
})
