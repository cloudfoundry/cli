package main_test

import (
	"os"
	"os/exec"
	"path"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestMain(t *testing.T) {
	RegisterFailHandler(Fail)
	dir, err := os.Getwd()
	Expect(err).NotTo(HaveOccurred())

	cmd := exec.Command("go", "build", "-o", path.Join(dir, "..", "fixtures", "plugins", "test"), path.Join(dir, "..", "fixtures", "plugins", "test.go"))
	err = cmd.Run()
	Expect(err).NotTo(HaveOccurred())

	cmd = exec.Command("go", "build", "-o", path.Join(dir, "..", "fixtures", "plugins", "plugin2"), path.Join(dir, "..", "fixtures", "plugins", "plugin2.go"))
	err = cmd.Run()
	Expect(err).NotTo(HaveOccurred())

	RunSpecs(t, "Main Suite")
}
