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
	DebugCommandPrefix        = "\nCMD>"
	DebugCommandPrefixWithDir = "\nCMD %s>"
	DebugOutPrefix            = "OUT: "
	DebugErrPrefix            = "ERR: "
)

var isPass = regexp.MustCompile("(?i)password|token")

// CF runs a 'cf' command with given arguments.
func CF(args ...string) *Session {
	WriteCommand("", nil, args)
	session, err := Start(
		exec.Command("cf", args...),
		NewPrefixedWriter(DebugOutPrefix, GinkgoWriter),
		NewPrefixedWriter(DebugErrPrefix, GinkgoWriter))
	Expect(err).NotTo(HaveOccurred())
	return session
}

// CFEnv represents configuration for running a 'cf' command. It allows us to
// run a 'cf' command with a custom working directory, specific environment
// variables, and stdin.
type CFEnv struct {
	WorkingDirectory string
	EnvVars          map[string]string
	stdin            io.Reader
}

// CustomCF runs a 'cf' command with a custom environment and given arguments.
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

	WriteCommand("", cfEnv.EnvVars, args)
	session, err := Start(
		command,
		NewPrefixedWriter(DebugOutPrefix, GinkgoWriter),
		NewPrefixedWriter(DebugErrPrefix, GinkgoWriter))
	Expect(err).NotTo(HaveOccurred())
	return session
}

// DebugCustomCF runs a 'cf' command with a custom environment and given
// arguments, with CF_LOG_LEVEL set to 'debug'.
func DebugCustomCF(cfEnv CFEnv, args ...string) *Session {
	if cfEnv.EnvVars == nil {
		cfEnv.EnvVars = map[string]string{}
	}
	cfEnv.EnvVars["CF_LOG_LEVEL"] = "debug"

	return CustomCF(cfEnv, args...)
}

// CFWithStdin runs a 'cf' command with a custom stdin and given arguments.
func CFWithStdin(stdin io.Reader, args ...string) *Session {
	WriteCommand("", nil, args)
	command := exec.Command("cf", args...)
	command.Stdin = stdin
	session, err := Start(
		command,
		NewPrefixedWriter(DebugOutPrefix, GinkgoWriter),
		NewPrefixedWriter(DebugErrPrefix, GinkgoWriter))
	Expect(err).NotTo(HaveOccurred())
	return session
}

// CFWithEnv runs a 'cf' command with specified environment variables and given arguments.
func CFWithEnv(envVars map[string]string, args ...string) *Session {
	return CustomCF(CFEnv{EnvVars: envVars}, args...)
}

// WriteCommand prints the working directory, the environment variables, and
// 'cf' with the given arguments. Environment variables that are passwords will
// be redacted.
func WriteCommand(workingDir string, env map[string]string, args []string) {
	start := DebugCommandPrefix
	if workingDir != "" {
		start = fmt.Sprintf(DebugCommandPrefixWithDir, workingDir)
	}

	display := []string{
		start,
	}

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
