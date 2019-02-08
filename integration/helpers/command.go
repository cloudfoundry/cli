package helpers

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
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
	WriteCommand(nil, args)
	args = removeEmptyArgs(args)
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

	WriteCommand(cfEnv.EnvVars, args)
	session, err := Start(
		command,
		NewPrefixedWriter(DebugOutPrefix, GinkgoWriter),
		NewPrefixedWriter(DebugErrPrefix, GinkgoWriter))
	Expect(err).NotTo(HaveOccurred())
	return session
}

func DebugCustomCF(cfEnv CFEnv, args ...string) *Session {
	if cfEnv.EnvVars == nil {
		cfEnv.EnvVars = map[string]string{}
	}
	cfEnv.EnvVars["CF_LOG_LEVEL"] = "debug"

	return CustomCF(cfEnv, args...)
}

func CFWithStdin(stdin io.Reader, args ...string) *Session {
	WriteCommand(nil, args)
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

func WriteCommand(env map[string]string, args []string) {
	display := []string{
		DebugCommandPrefix,
	}

	isPass := regexp.MustCompile("(?i)password|token")
	for key, val := range env {
		if isPass.MatchString(key) {
			val = "*****"
		}
		display = append(display, fmt.Sprintf("%s=%s", key, val))
	}

	display = append(display, "cf")
	display = append(display, args...)
	GinkgoWriter.Write([]byte(strings.Join(append(display, "\n"), " ")))
}

func removeEmptyArgs(args []string) []string {
	returnArgs := make([]string, 0, len(args))

	for _, arg := range args {
		if arg != "" {
			returnArgs = append(returnArgs, arg)
		}
	}
	return returnArgs
}
