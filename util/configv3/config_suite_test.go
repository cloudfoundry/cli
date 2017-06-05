package configv3_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	. "code.cloudfoundry.org/cli/util/configv3"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

var (
	oldLang       string
	oldLCAll      string
	oldExpVal     string
	oldPluginHome string
)

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Config Suite")
}

var _ = BeforeSuite(func() {
	// specifically for when we run unit tests locally
	// we save and unset these variables in case they're present
	// since we want to load a default config
	oldLang = os.Getenv("LANG")
	oldLCAll = os.Getenv("LC_ALL")
	oldExpVal = os.Getenv("CF_CLI_EXPERIMENTAL")
	oldPluginHome = os.Getenv("CF_PLUGIN_HOME")

	err := os.Unsetenv("LANG")
	Expect(err).NotTo(HaveOccurred())
	err = os.Unsetenv("LC_ALL")
	Expect(err).NotTo(HaveOccurred())
	err = os.Unsetenv("CF_CLI_EXPERIMENTAL")
	Expect(err).NotTo(HaveOccurred())
	err = os.Unsetenv("CF_PLUGIN_HOME")
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	err := os.Setenv("LANG", oldLang)
	Expect(err).NotTo(HaveOccurred())
	err = os.Setenv("LC_ALL", oldLCAll)
	Expect(err).NotTo(HaveOccurred())
	err = os.Setenv("CF_CLI_EXPERIMENTAL", oldExpVal)
	Expect(err).NotTo(HaveOccurred())
	err = os.Setenv("CF_PLUGIN_HOME", oldPluginHome)
	Expect(err).NotTo(HaveOccurred())
})

func createAndSetHomeDir() string {
	homeDir, err := ioutil.TempDir("", "cli-config-tests")
	Expect(err).NotTo(HaveOccurred())

	// check platform
	if runtime.GOOS == "windows" {
		err = os.Setenv("USERPROFILE", homeDir)
	} else {
		err = os.Setenv("HOME", homeDir)
	}

	Expect(err).NotTo(HaveOccurred())
	return homeDir
}

func removeAndUnsetHomeDir(homeDir string) {
	if homeDir != "" {
		err := os.RemoveAll(homeDir)
		Expect(err).ToNot(HaveOccurred())

		// check platform
		if runtime.GOOS == "windows" {
			err = os.Unsetenv("USERPROFILE")
		} else {
			err = os.Unsetenv("HOME")
		}

		Expect(err).ToNot(HaveOccurred())
	}
}

func writeConfig(homeDir string, rawConfig string) {
	err := os.MkdirAll(filepath.Join(homeDir, ".cf"), 0777)
	Expect(err).ToNot(HaveOccurred())
	err = ioutil.WriteFile(filepath.Join(homeDir, ".cf", "config.json"), []byte(rawConfig), 0644)
	Expect(err).ToNot(HaveOccurred())
}

func writePluginConfig(pluginDir string, rawConfig string) {
	err := os.MkdirAll(filepath.Join(pluginDir), 0777)
	Expect(err).ToNot(HaveOccurred())
	err = ioutil.WriteFile(filepath.Join(pluginDir, "config.json"), []byte(rawConfig), 0644)
	Expect(err).ToNot(HaveOccurred())
}

func getPluginsHome() string {
	var pluginsRoot string

	switch {
	case os.Getenv("CF_PLUGIN_HOME") != "":
		pluginsRoot = os.Getenv("CF_PLUGIN_HOME")
	default:
		pluginsRoot = HomeDirectory(false)
	}

	return filepath.Join(pluginsRoot, ".cf", "plugins")
}
