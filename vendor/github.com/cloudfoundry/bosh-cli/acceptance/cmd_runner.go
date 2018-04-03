package acceptance

import (
	"fmt"
	"io"
	"strings"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type CmdRunner interface {
	RunCommand(env map[string]string, args ...string) (string, string, int, error)
	RunStreamingCommand(out io.Writer, env map[string]string, args ...string) (string, string, int, error)
}

type cmdRunner struct {
	runner boshsys.CmdRunner
}

func NewCmdRunner(logger boshlog.Logger) CmdRunner {
	return &cmdRunner{
		runner: boshsys.NewExecCmdRunner(logger),
	}
}

func (r *cmdRunner) RunCommand(env map[string]string, args ...string) (string, string, int, error) {
	exports := make([]string, len(env))
	for k, v := range env {
		exports = append(exports, fmt.Sprintf("%s=%s", k, v))
	}

	argsWithEnv := append(exports, args...)
	return r.runner.RunCommand(
		"bash",
		"-c",
		strings.Join(argsWithEnv, " "),
	)
}

func (r *cmdRunner) RunStreamingCommand(out io.Writer, env map[string]string, args ...string) (string, string, int, error) {
	exports := make([]string, len(env))
	for k, v := range env {
		exports = append(exports, fmt.Sprintf("%s=%s", k, v))
	}

	argsWithEnv := append(exports, args...)

	cmd := boshsys.Command{
		Name: "bash",
		Args: []string{
			"-c",
			strings.Join(argsWithEnv, " "),
		},
		Stdout: out,
		Stderr: out,
	}

	// write command being run
	cmdString := fmt.Sprintf("> %s %s\n", cmd.Name, strings.Join(cmd.Args, " "))
	out.Write([]byte(cmdString))

	return r.runner.RunComplexCommand(cmd)
}
