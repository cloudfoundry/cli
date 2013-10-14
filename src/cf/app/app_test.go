package app

import (
	"cf/commands"
	"cf/requirements"
	"github.com/codegangsta/cli"
	"github.com/stretchr/testify/assert"
	"strings"
	testreq "testhelpers/requirements"
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
		"create-service-auth-token",
		"create-service-broker",
		"create-space",
		"create-user",
		"create-user-provided-service",
		"delete",
		"delete-org",
		"delete-service",
		"delete-service-auth-token",
		"delete-service-broker",
		"delete-space",
		"delete-user",
		"env",
		"events",
		"files",
		"login",
		"logout",
		"logs",
		"map-domain",
		"map-route",
		"marketplace",
		"org",
		"orgs",
		"passwd",
		"push",
		"rename",
		"rename-org",
		"rename-service",
		"rename-service-broker",
		"rename-space",
		"reserve-domain",
		"reserve-route",
		"restart",
		"routes",
		"scale",
		"service",
		"service-auth-tokens",
		"service-brokers",
		"services",
		"set-env",
		"set-org-role",
		"set-quota",
		//		"share-domain",
		"space",
		"spaces",
		"stacks",
		"start",
		"stop",
		"target",
		"unbind-service",
		"unmap-domain",
		"unmap-route",
		"unset-env",
		"update-service-broker",
		"update-user-provided-service",
	}

	for _, cmdName := range availableCmds {
		cmdFactory := &FakeCmdFactory{}
		reqFactory := &testreq.FakeReqFactory{}
		cmdRunner := commands.NewRunner(cmdFactory, reqFactory)
		app, _ := NewApp(cmdRunner)
		app.Run([]string{"", cmdName})

		assert.Equal(t, cmdFactory.CmdName, cmdName)
		assert.True(t, cmdFactory.CmdCompleted)
	}
}

func TestUsageIncludesCommandName(t *testing.T) {
	cmdFactory := &FakeCmdFactory{}
	reqFactory := &testreq.FakeReqFactory{}
	cmdRunner := commands.NewRunner(cmdFactory, reqFactory)
	app, _ := NewApp(cmdRunner)
	for _, cmd := range app.Commands {
		assert.Contains(t, strings.Split(cmd.Usage, "\n")[0], cmd.Name)
	}
}
