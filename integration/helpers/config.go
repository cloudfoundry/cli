package helpers

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cli/util/configv3"

	. "github.com/onsi/gomega"
)

func TurnOffColors() {
	Expect(os.Setenv("CF_COLOR", "false")).To(Succeed())
}

func TurnOnExperimental() {
	Expect(os.Setenv("CF_CLI_EXPERIMENTAL", "true")).To(Succeed())
}

func SetHomeDir() string {
	var err error
	homeDir, err := ioutil.TempDir("", "cli-integration-test")
	Expect(err).NotTo(HaveOccurred())

	Expect(os.Setenv("CF_HOME", homeDir)).To(Succeed())
	Expect(os.Setenv("CF_PLUGIN_HOME", homeDir)).To(Succeed())
	return homeDir
}

func DestroyHomeDir(homeDir string) {
	if homeDir != "" {
		Expect(os.RemoveAll(homeDir)).To(Succeed())
	}
}

func SetConfig(cb func(conf *configv3.Config)) {
	config, err := configv3.LoadConfig()
	Expect(err).ToNot(HaveOccurred())

	cb(config)

	err = configv3.WriteConfig(config)
	Expect(err).ToNot(HaveOccurred())
}

func SetConfigContent(dir string, rawConfig string) {
	err := os.MkdirAll(filepath.Join(dir), 0777)
	Expect(err).ToNot(HaveOccurred())
	err = ioutil.WriteFile(filepath.Join(dir, "config.json"), []byte(rawConfig), 0644)
	Expect(err).ToNot(HaveOccurred())
}

func InvalidAccessToken() string {
	return "bearer eyJhbGciOiJSUzI1NiIsImtpZCI6ImxlZ2FjeS10b2tlbi1rZXkiLCJ0eXAiOiJKV1QifQ.eyJqdGkiOiJiMjZmMTFjMWNhYmI0ZmY0ODhlN2RhYTJkZTQxMTA4NiIsInN1YiI6IjBjZWMwY2E4LTA5MmYtNDkzYy1hYmExLWM4ZTZiMTRiODM3NiIsInNjb3BlIjpbIm9wZW5pZCIsInJvdXRpbmcucm91dGVyX2dyb3Vwcy53cml0ZSIsInNjaW0ucmVhZCIsImNsb3VkX2NvbnRyb2xsZXIuYWRtaW4iLCJ1YWEudXNlciIsInJvdXRpbmcucm91dGVyX2dyb3Vwcy5yZWFkIiwiY2xvdWRfY29udHJvbGxlci5yZWFkIiwicGFzc3dvcmQud3JpdGUiLCJjbG91ZF9jb250cm9sbGVyLndyaXRlIiwiZG9wcGxlci5maXJlaG9zZSIsInNjaW0ud3JpdGUiXSwiY2xpZW50X2lkIjoiY2YiLCJjaWQiOiJjZiIsImF6cCI6ImNmIiwiZ3JhbnRfdHlwZSI6InBhc3N3b3JkIiwidXNlcl9pZCI6IjBjZWMwY2E4LTA5MmYtNDkzYy1hYmExLWM4ZTZiMTRiODM3NiIsIm9yaWdpbiI6InVhYSIsInVzZXJfbmFtZSI6ImFkbWluIiwiZW1haWwiOiJhZG1pbiIsImF1dGhfdGltZSI6MTQ4OTY4Njg0OCwicmV2X3NpZyI6IjgzM2I4N2Q0IiwiaWF0IjoxNDg5Njg2ODQ4LCJleHAiOjE0ODk2ODc0NDgsImlzcyI6Imh0dHBzOi8vdWFhLmJvc2gtbGl0ZS5jb20vb2F1dGgvdG9rZW4iLCJ6aWQiOiJ1YWEiLCJhdWQiOlsic2NpbSIsImNsb3VkX2NvbnRyb2xsZXIiLCJwYXNzd29yZCIsImNmIiwidWFhIiwib3BlbmlkIiwiZG9wcGxlciIsInJvdXRpbmcucm91dGVyX2dyb3VwcyJdfQ.UeWpPsI5GEvhiQ0HzcCno7u80KbceMmnKHxO89saPrnsDOsbC4zwtz9AeEIvuqClXJCzS4WiOfkx7za0yFkR6z4LZlQc6t_9oq9KYMNCavQSsscYvuUXQH0zarvgptqzLU8miO30uFVVfYbRsLnJVu_5A8C1H29Gedky-70irPc1fZm__nFd8UaUyD2aj50B2M_t1lTkZbdzRn-gORhYAMcVUQNc9Mezj04uT9BAA8oKPzkt2yPN4JZddLvetJXjnp6Ug9x9GL1mfQTP7NVAPIVXSV84p8q_3WPOxjNb28dYGYqEfDNZMgu_nV0JSTXCq3l23jDA8ty8tJ_eYYjDBg"
}
