package fakes

import (
	bicloud "github.com/cloudfoundry/bosh-cli/cloud"
)

type FakeCPICmdRunner struct {
	RunInputs    []RunInput
	RunCmdOutput bicloud.CmdOutput
	RunErr       error
}

type RunInput struct {
	Context   bicloud.CmdContext
	Method    string
	Arguments []interface{}
}

func NewFakeCPICmdRunner() *FakeCPICmdRunner {
	return &FakeCPICmdRunner{}
}

func (r *FakeCPICmdRunner) Run(context bicloud.CmdContext, method string, args ...interface{}) (bicloud.CmdOutput, error) {
	r.RunInputs = append(r.RunInputs, RunInput{
		Context:   context,
		Method:    method,
		Arguments: args,
	})
	return r.RunCmdOutput, r.RunErr
}
