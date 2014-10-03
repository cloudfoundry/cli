package plugin_test

import (
	"fmt"
	"os"
	"os/exec"
	"path"

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
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	cmd := exec.Command("go", "build", "-o", path.Join(dir, "..", "..", "..", "fixtures", "config", "plugin-config", ".cf", "plugins", "test_1"), path.Join(dir, "..", "..", "..", "fixtures", "plugins", "test_1.go"))
	err = cmd.Run()
	if err != nil {
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	cmd = exec.Command("go", "build", "-o", path.Join(dir, "..", "..", "..", "fixtures", "config", "plugin-config", ".cf", "plugins", "test_2"), path.Join(dir, "..", "..", "..", "fixtures", "plugins", "test_2.go"))
	err = cmd.Run()
	if err != nil {
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	cmd = exec.Command("go", "build", "-o", path.Join(dir, "..", "..", "..", "fixtures", "config", "plugin-config", ".cf", "plugins", "empty_plugin"), path.Join(dir, "..", "..", "..", "fixtures", "plugins", "empty_plugin.go"))
	err = cmd.Run()
	if err != nil {
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	RunSpecs(t, "Plugin Suite")
}
