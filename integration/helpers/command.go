package helpers

import (
	"io"
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

func CF(args ...string) *Session {
	session, err := Start(
		exec.Command("cf", args...),
		NewPrefixedWriter("OUT: ", GinkgoWriter),
		NewPrefixedWriter("ERR: ", GinkgoWriter))
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

	session, err := Start(
		command,
		NewPrefixedWriter("OUT: ", GinkgoWriter),
		NewPrefixedWriter("ERR: ", GinkgoWriter))
	Expect(err).NotTo(HaveOccurred())
	return session
}

func CFWithStdin(stdin io.Reader, args ...string) *Session {
	command := exec.Command("cf", args...)
	command.Stdin = stdin
	session, err := Start(
		command,
		NewPrefixedWriter("OUT: ", GinkgoWriter),
		NewPrefixedWriter("ERR: ", GinkgoWriter))
	Expect(err).NotTo(HaveOccurred())
	return session
}

func CFWithEnv(envVars map[string]string, args ...string) *Session {
	return CustomCF(CFEnv{EnvVars: envVars}, args...)
}
