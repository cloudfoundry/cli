package commands_test

import (
	. "cf/commands"
	"cf/requirements"
	"errors"
	"github.com/codegangsta/cli"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

type TestCommand struct {
	Reqs       []Requirement
	WasRunWith *cli.Context
}

func (cmd *TestCommand) GetRequirements(factory requirements.Factory, c *cli.Context) (reqs []Requirement, err error) {
	reqs = cmd.Reqs
	return
}

func (cmd *TestCommand) Run(c *cli.Context) {
	cmd.WasRunWith = c
}

type TestRequirement struct {
	Passes      bool
	WasExecuted bool
}

func (r *TestRequirement) Execute() (err error) {
	r.WasExecuted = true

	if !r.Passes {
		return errors.New("Error in requirement")
	}

	return
}

func TestRun(t *testing.T) {
	runner := Runner{}
	passingReq := TestRequirement{Passes: true}
	failingReq := TestRequirement{Passes: false}
	lastReq := TestRequirement{Passes: true}

	cmd := TestCommand{
		Reqs: []Requirement{&passingReq, &failingReq, &lastReq},
	}

	ctxt := testhelpers.NewContext("login", []string{})
	err := runner.Run(&cmd, ctxt)

	assert.True(t, passingReq.WasExecuted, ctxt)
	assert.True(t, failingReq.WasExecuted, ctxt)

	assert.False(t, lastReq.WasExecuted)
	assert.Nil(t, cmd.WasRunWith)

	assert.Error(t, err)
}
