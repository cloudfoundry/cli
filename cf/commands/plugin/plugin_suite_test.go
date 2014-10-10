package plugin_test

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/testhelpers/configuration"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestPlugin(t *testing.T) {
	config := configuration.NewRepositoryWithDefaults()
	i18n.T = i18n.Init(config)

	RegisterFailHandler(Fail)

	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	cmd := exec.Command("go", "build", "-o", filepath.Join(dir, "..", "..", "..", "fixtures", "plugins", "test_1.exe"), filepath.Join(dir, "..", "..", "..", "fixtures", "plugins", "test_1.go"))
	err = cmd.Run()
	if err != nil {
		panic(err)
	}

	cmd = exec.Command("go", "build", "-o", filepath.Join(dir, "..", "..", "..", "fixtures", "plugins", "test_2.exe"), filepath.Join(dir, "..", "..", "..", "fixtures", "plugins", "test_2.go"))
	err = cmd.Run()
	if err != nil {
		panic(err)
	}

	cmd = exec.Command("go", "build", "-o", filepath.Join(dir, "..", "..", "..", "fixtures", "plugins", "existedName.exe"), filepath.Join(dir, "..", "..", "..", "fixtures", "plugins", "test_1.go"))
	err = cmd.Run()
	if err != nil {
		panic(err)
	}

	RunSpecs(t, "Plugin Suite")
}
