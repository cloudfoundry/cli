package main_test

import (
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestMain(t *testing.T) {
	RegisterFailHandler(Fail)
	dir, err := os.Getwd()
	Expect(err).NotTo(HaveOccurred())

	//Named test binaries with .exe so they will work correctly on windows
	//sad panda :(
	cmd := exec.Command("go", "build", "-o", filepath.Join(dir, "..", "fixtures", "plugins", "test.exe"), filepath.Join(dir, "..", "fixtures", "plugins", "test.go"))
	err = cmd.Run()
	Expect(err).NotTo(HaveOccurred())

	cmd = exec.Command("go", "build", "-o", filepath.Join(dir, "..", "fixtures", "plugins", "plugin2.exe"), filepath.Join(dir, "..", "fixtures", "plugins", "plugin2.go"))
	err = cmd.Run()
	Expect(err).NotTo(HaveOccurred())

	RunSpecs(t, "Main Suite")
}
