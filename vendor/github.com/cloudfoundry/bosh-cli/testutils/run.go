package testutils

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega/gexec"
)

func BuildExecutable() error {
	return BuildExecutableForArch("")
}

func BuildExecutableForArch(arch string) error {
	buildArg := "./../bin/build"
	if arch != "" {
		buildArg = buildArg + "-" + arch
	}

	session, err := RunCommand(buildArg)
	if session.ExitCode() != 0 {
		return fmt.Errorf("Failed to build bosh:\nstdout:\n%s\nstderr:\n%s", session.Out.Contents(), session.Err.Contents())
	}

	return err
}

func RunCommand(cmd string, args ...string) (*gexec.Session, error) {
	return RunComplexCommand(exec.Command(cmd, args...))
}

func RunComplexCommand(cmd *exec.Cmd) (*gexec.Session, error) {
	session, err := gexec.Start(cmd, ginkgo.GinkgoWriter, ginkgo.GinkgoWriter)
	if err != nil {
		return nil, err
	}

	session.Wait(120 * time.Second)

	return session, nil
}
