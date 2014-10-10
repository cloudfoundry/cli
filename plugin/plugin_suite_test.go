package plugin_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestPlugin(t *testing.T) {
	RegisterFailHandler(Fail)
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	cmd := exec.Command("go", "build", "-o", filepath.Join(dir, "..", "fixtures", "config", "plugin-config", ".cf", "plugins", "test_1.exe"), filepath.Join(dir, "..", "fixtures", "plugins", "test_1.go"))
	err = cmd.Run()
	if err != nil {
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	cmd = exec.Command("go", "build", "-o", filepath.Join(dir, "..", "fixtures", "config", "plugin-config", ".cf", "plugins", "noRpc.exe"), filepath.Join(dir, "..", "fixtures", "plugins", "noRpc.go"))
	err = cmd.Run()
	if err != nil {
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	RunSpecs(t, "Plugin Suite")
}
