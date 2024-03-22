package helpers

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cli/util/configv3"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// TurnOffColors sets CF_COLOR to 'false'.
func TurnOffColors() {
	Expect(os.Setenv("CF_COLOR", "false")).To(Succeed())
}

// TurnOnExperimental sets CF_CLI_EXPERIMENTAL to 'true'.
func TurnOnExperimental() {
	Expect(os.Setenv("CF_CLI_EXPERIMENTAL", "true")).To(Succeed())
}

// TurnOffExperimental unsets CF_CLI_EXPERIMENTAL.
func TurnOffExperimental() {
	Expect(os.Unsetenv("CF_CLI_EXPERIMENTAL")).To(Succeed())
}

// SetHomeDir sets CF_HOME and CF_PLUGIN_HOME to a temp directory and outputs
// the created directory through GinkgoWriter.
func SetHomeDir() string {
	var err error
	homeDir, err := ioutil.TempDir("", "cli-integration-test")
	Expect(err).NotTo(HaveOccurred())

	setHomeDirsTo(homeDir, homeDir)
	return homeDir
}

// WithRandomHomeDir sets CF_HOME and CF_PLUGIN_HOME to a temp directory and outputs
// the created directory through GinkgoWriter. Then it executes the provided function
// 'action'. Finally, it’s restoring the previous CF_HOME and CF_PLUGIN_HOME.
func WithRandomHomeDir(action func()) {
	oldHomeDir, oldPluginHomeDir := getHomeDirs()
	homeDir := SetHomeDir()
	action()
	setHomeDirsTo(oldHomeDir, oldPluginHomeDir)
	DestroyHomeDir(homeDir)
}

func getHomeDirs() (string, string) {
	homeDir := os.Getenv("CF_HOME")
	pluginHomeDir := os.Getenv("CF_PLUGIN_HOME")
	return homeDir, pluginHomeDir
}

func setHomeDirsTo(homeDir string, pluginHomeDir string) {
	GinkgoWriter.Write([]byte(fmt.Sprintln("\nHOME DIR>", homeDir)))

	Expect(os.Setenv("CF_HOME", homeDir)).To(Succeed())
	Expect(os.Setenv("CF_PLUGIN_HOME", pluginHomeDir)).To(Succeed())
}

// SetupSynchronizedSuite runs a setup function in its own CF context, creating
// and destroying a home directory around it.
func SetupSynchronizedSuite(setup func()) {
	homeDir := SetHomeDir()
	SetAPI()
	LoginCF()
	setup()
	DestroyHomeDir(homeDir)
}

// DestroyHomeDir safely removes the given directory checking for errors.
func DestroyHomeDir(homeDir string) {
	if homeDir != "" {
		Eventually(func() error { return os.RemoveAll(homeDir) }).Should(Succeed())
	}
}

// GetConfig loads a CF config JSON file and returns the parsed struct.
func GetConfig() *configv3.Config {
	c, err := configv3.LoadConfig()
	Expect(err).ToNot(HaveOccurred())
	return c
}

// SetConfig allows for a given function to modify a CF config JSON and writes
// the result back down to the filesystem.
func SetConfig(cb func(conf *configv3.Config)) {
	config, err := configv3.LoadConfig()
	Expect(err).ToNot(HaveOccurred())

	cb(config)

	err = config.WriteConfig()
	Expect(err).ToNot(HaveOccurred())
}

// SetConfigContent writes given raw config into given directory as "config.json".
func SetConfigContent(dir string, rawConfig string) {
	err := os.MkdirAll(filepath.Join(dir), 0777)
	Expect(err).ToNot(HaveOccurred())
	err = ioutil.WriteFile(filepath.Join(dir, "config.json"), []byte(rawConfig), 0644)
	Expect(err).ToNot(HaveOccurred())
}

// ExpiredAccessToken returns an example expired bearer token.
func ExpiredAccessToken() string {
	return "bearer eyJhbGciOiJSUzI1NiIsImtpZCI6ImxlZ2FjeS10b2tlbi1rZXkiLCJ0eXAiOiJKV1QifQ.eyJqdGkiOiJiMjZmMTFjMWNhYmI0ZmY0ODhlN2RhYTJkZTQxMTA4NiIsInN1YiI6IjBjZWMwY2E4LTA5MmYtNDkzYy1hYmExLWM4ZTZiMTRiODM3NiIsInNjb3BlIjpbIm9wZW5pZCIsInJvdXRpbmcucm91dGVyX2dyb3Vwcy53cml0ZSIsInNjaW0ucmVhZCIsImNsb3VkX2NvbnRyb2xsZXIuYWRtaW4iLCJ1YWEudXNlciIsInJvdXRpbmcucm91dGVyX2dyb3Vwcy5yZWFkIiwiY2xvdWRfY29udHJvbGxlci5yZWFkIiwicGFzc3dvcmQud3JpdGUiLCJjbG91ZF9jb250cm9sbGVyLndyaXRlIiwiZG9wcGxlci5maXJlaG9zZSIsInNjaW0ud3JpdGUiXSwiY2xpZW50X2lkIjoiY2YiLCJjaWQiOiJjZiIsImF6cCI6ImNmIiwiZ3JhbnRfdHlwZSI6InBhc3N3b3JkIiwidXNlcl9pZCI6IjBjZWMwY2E4LTA5MmYtNDkzYy1hYmExLWM4ZTZiMTRiODM3NiIsIm9yaWdpbiI6InVhYSIsInVzZXJfbmFtZSI6ImFkbWluIiwiZW1haWwiOiJhZG1pbiIsImF1dGhfdGltZSI6MTQ4OTY4Njg0OCwicmV2X3NpZyI6IjgzM2I4N2Q0IiwiaWF0IjoxNDg5Njg2ODQ4LCJleHAiOjE0ODk2ODc0NDgsImlzcyI6Imh0dHBzOi8vdWFhLmJvc2gtbGl0ZS5jb20vb2F1dGgvdG9rZW4iLCJ6aWQiOiJ1YWEiLCJhdWQiOlsic2NpbSIsImNsb3VkX2NvbnRyb2xsZXIiLCJwYXNzd29yZCIsImNmIiwidWFhIiwib3BlbmlkIiwiZG9wcGxlciIsInJvdXRpbmcucm91dGVyX2dyb3VwcyJdfQ.UeWpPsI5GEvhiQ0HzcCno7u80KbceMmnKHxO89saPrnsDOsbC4zwtz9AeEIvuqClXJCzS4WiOfkx7za0yFkR6z4LZlQc6t_9oq9KYMNCavQSsscYvuUXQH0zarvgptqzLU8miO30uFVVfYbRsLnJVu_5A8C1H29Gedky-70irPc1fZm__nFd8UaUyD2aj50B2M_t1lTkZbdzRn-gORhYAMcVUQNc9Mezj04uT9BAA8oKPzkt2yPN4JZddLvetJXjnp6Ug9x9GL1mfQTP7NVAPIVXSV84p8q_3WPOxjNb28dYGYqEfDNZMgu_nV0JSTXCq3l23jDA8ty8tJ_eYYjDBg"
}
