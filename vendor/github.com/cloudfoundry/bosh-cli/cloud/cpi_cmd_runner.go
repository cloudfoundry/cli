package cloud

import (
	"bytes"
	"encoding/json"
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type CmdInput struct {
	Method    string        `json:"method"`
	Arguments []interface{} `json:"arguments"`
	Context   CmdContext    `json:"context"`
}

type CmdContext struct {
	DirectorID string `json:"director_uuid"`
}

func (c CmdContext) String() string {
	bytes, err := json.Marshal(c)
	if err != nil {
		panic(fmt.Sprintf("Error stringifying CmdContext %#v: %s", c, err.Error()))
	}
	return fmt.Sprintf("CmdContext%s", string(bytes))
}

type CmdError struct {
	Type      string `json:"type"`
	Message   string `json:"message"`
	OkToRetry bool   `json:"ok_to_retry"`
}

func (e CmdError) String() string {
	bytes, err := json.Marshal(e)
	if err != nil {
		panic(fmt.Sprintf("Error stringifying CmdError %#v: %s", e, err.Error()))
	}
	return fmt.Sprintf("CmdError%s", string(bytes))
}

type CmdOutput struct {
	Result interface{} `json:"result"`
	Error  *CmdError   `json:"error,omitempty"`
	Log    string      `json:"log"`
}

type CPICmdRunner interface {
	Run(context CmdContext, method string, args ...interface{}) (CmdOutput, error)
}

type cpiCmdRunner struct {
	cmdRunner boshsys.CmdRunner
	cpi       CPI
	logger    boshlog.Logger
	logTag    string
}

func NewCPICmdRunner(
	cmdRunner boshsys.CmdRunner,
	cpi CPI,
	logger boshlog.Logger,
) CPICmdRunner {
	return &cpiCmdRunner{
		cmdRunner: cmdRunner,
		cpi:       cpi,
		logger:    logger,
		logTag:    "cpiCmdRunner",
	}
}

func (r *cpiCmdRunner) Run(context CmdContext, method string, args ...interface{}) (CmdOutput, error) {
	cmdInput := CmdInput{
		Method:    method,
		Arguments: args,
		Context:   context,
	}
	inputBytes, err := json.Marshal(cmdInput)
	if err != nil {
		return CmdOutput{}, bosherr.WrapErrorf(err, "Marshalling external CPI command input %#v", cmdInput)
	}

	cmdPath := r.cpi.ExecutablePath()
	cmd := boshsys.Command{
		Name: cmdPath,
		Env: map[string]string{
			"BOSH_PACKAGES_DIR": r.cpi.PackagesDir,
			"BOSH_JOBS_DIR":     r.cpi.JobsDir,
			"PATH":              "/usr/local/bin:/usr/bin:/bin:/sbin",
		},
		UseIsolatedEnv: true,
		Stdin:          bytes.NewReader(inputBytes),
	}
	stdout, stderr, exitCode, err := r.cmdRunner.RunComplexCommand(cmd)
	r.logger.Debug(r.logTag, "Exit Code %d when executing external CPI command '%s'\nSTDIN: '%s'\nSTDOUT: '%s'\nSTDERR: '%s'", exitCode, cmdPath, string(inputBytes), stdout, stderr)
	if err != nil {
		return CmdOutput{}, bosherr.WrapErrorf(err, "Executing external CPI command: '%s'", cmdPath)
	}

	cmdOutput := CmdOutput{}
	err = json.Unmarshal([]byte(stdout), &cmdOutput)
	if err != nil {
		return CmdOutput{}, bosherr.WrapErrorf(err, "Unmarshalling external CPI command output: STDOUT: '%s', STDERR: '%s'", stdout, stderr)
	}

	r.logger.Debug(r.logTag, cmdOutput.Log)

	return cmdOutput, err
}
