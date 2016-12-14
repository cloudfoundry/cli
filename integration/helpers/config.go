package helpers

import (
	"io/ioutil"
	"os"

	"code.cloudfoundry.org/cli/util/configv3"

	. "github.com/onsi/gomega"
)

func TurnOffColors() {
	os.Setenv("CF_COLOR", "false")
}

func SetHomeDir() string {
	var err error
	homeDir, err := ioutil.TempDir("", "cli-gats-test")
	Expect(err).NotTo(HaveOccurred())

	os.Setenv("CF_HOME", homeDir)
	os.Setenv("CF_PLUGIN_HOME", homeDir)
	return homeDir
}

func DestroyHomeDir(homeDir string) {
	if homeDir != "" {
		os.RemoveAll(homeDir)
	}
}

func SetConfig(cb func(conf *configv3.Config)) {
	config, err := configv3.LoadConfig()
	Expect(err).ToNot(HaveOccurred())

	cb(config)

	err = configv3.WriteConfig(config)
	Expect(err).ToNot(HaveOccurred())
}
