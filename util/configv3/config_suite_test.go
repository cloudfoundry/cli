package configv3_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Config Suite")
}

func setup() string {
	homeDir, err := ioutil.TempDir("", "cli-config-tests")
	Expect(err).NotTo(HaveOccurred())
	err = os.Setenv("CF_HOME", homeDir)
	Expect(err).NotTo(HaveOccurred())
	return homeDir
}

func teardown(homeDir string) {
	if homeDir != "" {
		err := os.RemoveAll(homeDir)
		Expect(err).ToNot(HaveOccurred())
		err = os.Unsetenv("CF_HOME")
		Expect(err).ToNot(HaveOccurred())
	}
}

func setConfig(homeDir string, rawConfig string) {
	helpers.SetConfigContent(filepath.Join(homeDir, ".cf"), rawConfig)
}

func setPluginConfig(pluginDir string, rawConfig string) {
	helpers.SetConfigContent(pluginDir, rawConfig)
}
