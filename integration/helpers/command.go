package helpers

import (
	"io"
	"os"
	"os/exec"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

const (
	DebugCommandPrefix = "\nCMD>"
	DebugOutPrefix     = "OUT: "
	DebugErrPrefix     = "ERR: "
)

func CF(args ...string) *Session {
	WriteCommand(args)
	session, err := Start(
		exec.Command("cf", args...),
		NewPrefixedWriter(DebugOutPrefix, GinkgoWriter),
		NewPrefixedWriter(DebugErrPrefix, GinkgoWriter))
	Expect(err).NotTo(HaveOccurred())
	return session
}

type CFEnv struct {
	WorkingDirectory string
	EnvVars          map[string]string
	stdin            io.Reader
}

func CustomCF(cfEnv CFEnv, args ...string) *Session {
	command := exec.Command("cf", args...)
	if cfEnv.stdin != nil {
		command.Stdin = cfEnv.stdin
	}
	if cfEnv.WorkingDirectory != "" {
		command.Dir = cfEnv.WorkingDirectory
	}

	if cfEnv.EnvVars != nil {
		env := os.Environ()
		for key, val := range cfEnv.EnvVars {
			env = AddOrReplaceEnvironment(env, key, val)
		}
		command.Env = env
	}

	WriteCommand(args)
	session, err := Start(
		command,
		NewPrefixedWriter(DebugOutPrefix, GinkgoWriter),
		NewPrefixedWriter(DebugErrPrefix, GinkgoWriter))
	Expect(err).NotTo(HaveOccurred())
	return session
}

func CFWithStdin(stdin io.Reader, args ...string) *Session {
	WriteCommand(args)
	command := exec.Command("cf", args...)
	command.Stdin = stdin
	session, err := Start(
		command,
		NewPrefixedWriter(DebugOutPrefix, GinkgoWriter),
		NewPrefixedWriter(DebugErrPrefix, GinkgoWriter))
	Expect(err).NotTo(HaveOccurred())
	return session
}

func CFWithEnv(envVars map[string]string, args ...string) *Session {
	return CustomCF(CFEnv{EnvVars: envVars}, args...)
}

func WriteCommand(args []string) {
	display := append([]string{DebugCommandPrefix, "cf"}, args...)
	GinkgoWriter.Write([]byte(strings.Join(append(display, "\n"), " ")))
}
