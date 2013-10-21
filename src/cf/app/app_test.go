package app

import (
	"cf/api"
	"cf/commands"
	"cf/configuration"
	"cf/net"
	"github.com/codegangsta/cli"
	"github.com/stretchr/testify/assert"
	"strings"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

type FakeRunner struct {
	cmdFactory commands.Factory
	t          *testing.T
	cmdName    string
}

func (runner *FakeRunner) RunCmdByName(cmdName string, c *cli.Context) (err error) {
	_, err = runner.cmdFactory.GetByCmdName(cmdName)
	if err != nil {
		runner.t.Fatal("Error instantiating command with name", cmdName)
		return
	}
	runner.cmdName = cmdName
	return
}

func TestCommands(t *testing.T) {
	availableCmds := []string{
		"api",
		"app",
		"apps",
		"auth",
		"bind-service",
		"buildpacks",
		"create-buildpack",
		"create-org",
		"create-service",
		"create-service-auth-token",
		"create-service-broker",
		"create-space",
		"create-user",
		"create-user-provided-service",
		"delete",
		"delete-buildpack",
		"delete-org",
		"delete-route",
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
		"org-users",
		"orgs",
		"passwd",
		"push",
		"quotas",
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
		"set-space-role",
		"share-domain",
		"space",
		"space-users",
		"spaces",
		"stacks",
		"start",
		"stop",
		"target",
		"unbind-service",
		"unmap-domain",
		"unmap-route",
		"unset-env",
		"unset-org-role",
		"unset-space-role",
		"update-buildpack",
		"update-service-broker",
		"update-user-provided-service",
	}

	for _, cmdName := range availableCmds {
		ui := &testterm.FakeUI{}
		config := &configuration.Configuration{}
		configRepo := testconfig.FakeConfigRepository{}

		repoLocator := api.NewRepositoryLocator(config, configRepo, map[string]net.Gateway{
			"auth":             net.NewUAAGateway(),
			"cloud-controller": net.NewCloudControllerGateway(),
			"uaa":              net.NewUAAGateway(),
		})

		cmdFactory := commands.NewFactory(ui, config, configRepo, repoLocator)
		cmdRunner := &FakeRunner{cmdFactory: cmdFactory, t: t}
		app, _ := NewApp(cmdRunner)
		app.Run([]string{"", cmdName})

		assert.Equal(t, cmdRunner.cmdName, cmdName)
	}
}

func TestUsageIncludesCommandName(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	cmdRunner := commands.NewRunner(nil, reqFactory)
	app, _ := NewApp(cmdRunner)
	for _, cmd := range app.Commands {
		assert.Contains(t, strings.Split(cmd.Usage, "\n")[0], cmd.Name)
	}
}
