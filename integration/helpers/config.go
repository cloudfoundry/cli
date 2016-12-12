package helpers

import (
	"io/ioutil"
	"os"

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
