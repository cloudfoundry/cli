package app_test

import (
	"cf/app"
	"cf/commands"
	"cf/requirements"
	"github.com/codegangsta/cli"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

type FakeCmd struct {
	factory *FakeCmdFactory
}

func (cmd FakeCmd) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	return
}

func (cmd FakeCmd) Run(c *cli.Context) {
	cmd.factory.CmdCompleted = true
}

type FakeCmdFactory struct {
	CmdName      string
	CmdCompleted bool
}

func (f *FakeCmdFactory) GetByCmdName(cmdName string) (cmd commands.Command, err error) {
	f.CmdName = cmdName
	cmd = FakeCmd{f}
	return
}

func TestCommands(t *testing.T) {
	availableCmds := []string{
		"api",
		"app",
		"apps",
		"bind-service",
		"create-org",
		"create-service",
		"create-space",
		"create-user-provided-service",
		"delete",
		"delete-org",
		"delete-service",
		"delete-space",
		"env",
		"files",
		"login",
		"logout",
		"logs",
		"map-route",
		"marketplace",
		"org",
		"orgs",
		"passwd",
		"push",
		"rename",
		"rename-org",
		"rename-service",
		"rename-space",
		"reserve-domain",
		"reserve-route",
		"restart",
		"routes",
		"scale",
		"service",
		"services",
		"set-env",
		"space",
		"spaces",
		"stacks",
		"start",
		"stop",
		"target",
		"unbind-service",
		"unmap-route",
		"unset-env",
	}

	for _, cmdName := range availableCmds {
		cmdFactory := &FakeCmdFactory{}
		reqFactory := &testhelpers.FakeReqFactory{}
		app, _ := app.NewApp(cmdFactory, reqFactory)
		app.Run([]string{"", cmdName})

		assert.Equal(t, cmdFactory.CmdName, cmdName)
		assert.True(t, cmdFactory.CmdCompleted)
	}
}
