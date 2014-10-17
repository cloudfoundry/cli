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

	cmd := exec.Command("go", "build", "-o", filepath.Join(dir, "..", "fixtures", "plugins", "test_1.exe"), filepath.Join(dir, "..", "fixtures", "plugins", "test_1.go"))
	err = cmd.Run()
	if err != nil {
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	cmd = exec.Command("go", "build", "-o", filepath.Join(dir, "..", "fixtures", "plugins", "test_2.exe"), filepath.Join(dir, "..", "fixtures", "plugins", "test_2.go"))
	err = cmd.Run()
	if err != nil {
		fmt.Println(err)
	}

	cmd = exec.Command("go", "build", "-o", filepath.Join(dir, "..", "fixtures", "plugins", "test_with_push.exe"), filepath.Join(dir, "..", "fixtures", "plugins", "test_with_push.go"))
	err = cmd.Run()
	if err != nil {
		fmt.Println(err)
	}

	cmd = exec.Command("go", "build", "-o", filepath.Join(dir, "..", "fixtures", "plugins", "test_with_help.exe"), filepath.Join(dir, "..", "fixtures", "plugins", "test_with_help.go"))
	err = cmd.Run()
	if err != nil {
		fmt.Println(err)
	}
	Expect(err).NotTo(HaveOccurred())

	RunSpecs(t, "Main Suite")
}
