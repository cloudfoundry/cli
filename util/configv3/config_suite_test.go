package configv3_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

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
	err := os.MkdirAll(filepath.Join(homeDir, ".cf"), 0777)
	Expect(err).ToNot(HaveOccurred())
	err = ioutil.WriteFile(filepath.Join(homeDir, ".cf", "config.json"), []byte(rawConfig), 0644)
	Expect(err).ToNot(HaveOccurred())
}

func setPluginConfig(pluginDir string, rawConfig string) {
	err := os.MkdirAll(filepath.Join(pluginDir), 0777)
	Expect(err).ToNot(HaveOccurred())
	err = ioutil.WriteFile(filepath.Join(pluginDir, "config.json"), []byte(rawConfig), 0644)
	Expect(err).ToNot(HaveOccurred())
}
