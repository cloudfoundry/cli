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

func (cmd *TestCommand) GetRequirements(factory requirements.Factory, c *cli.Context) (reqs []Requirement) {
	return cmd.Reqs
}

func (cmd *TestCommand) Run(c *cli.Context) {
	cmd.WasRunWith = c
}

type TestRequirement struct {
	Passes          bool
	WasExecutedWith *cli.Context
}

func (r *TestRequirement) Execute(c *cli.Context) (err error) {
	r.WasExecutedWith = c

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

	assert.Equal(t, passingReq.WasExecutedWith, ctxt)
	assert.Equal(t, failingReq.WasExecutedWith, ctxt)

	assert.Nil(t, lastReq.WasExecutedWith)
	assert.Nil(t, cmd.WasRunWith)

	assert.Error(t, err)
}
