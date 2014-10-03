package main_test

import (
	"fmt"
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
	if err != nil {
		panic(err)
	}

	cmd := exec.Command("go", "build", "-o", path.Join(dir, "..", "fixtures", "config", "main-plugin-test-config", ".cf", "plugins", "test_1.exe"), path.Join(dir, "..", "fixtures", "plugins", "test_1.go"))
	err = cmd.Run()
	if err != nil {
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	cmd = exec.Command("go", "build", "-o", path.Join(dir, "..", "fixtures", "config", "main-plugin-test-config", ".cf", "plugins", "test_2.exe"), path.Join(dir, "..", "fixtures", "plugins", "test_2.go"))
	err = cmd.Run()
	if err != nil {
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	RunSpecs(t, "Main Suite")
}
