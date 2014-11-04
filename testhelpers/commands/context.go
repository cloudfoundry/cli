package commands

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/app"
	"github.com/cloudfoundry/cli/cf/command_factory"
	"github.com/cloudfoundry/cli/cf/command_runner"
	testPluginConfig "github.com/cloudfoundry/cli/cf/configuration/plugin_config/fakes"
	"github.com/cloudfoundry/cli/cf/manifest"
	"github.com/cloudfoundry/cli/cf/net"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	"github.com/codegangsta/cli"
)

func NewContext(cmdName string, args []string) *cli.Context {
	targetCommand := findCommand(cmdName)

	flagSet := new(flag.FlagSet)
	for i, _ := range targetCommand.Flags {
		targetCommand.Flags[i].Apply(flagSet)
	}

	// move all flag args to the beginning of the list, go requires them all upfront
	if targetCommand.SkipFlagParsing == false {
		firstFlagIndex := -1
		for index, arg := range args {
			if strings.HasPrefix(arg, "-") {
				firstFlagIndex = index
				break
			}
		}
		if firstFlagIndex > 0 {
			args := args[0:firstFlagIndex]
			flags := args[firstFlagIndex:]
			flagSet.Parse(append(flags, args...))
		} else {
			flagSet.Parse(args[0:])
		}
	} else {
		flagSet.Parse(args[0:])
	}

	return cli.NewContext(cli.NewApp(), flagSet, new(flag.FlagSet))
}

func findCommand(cmdName string) (cmd cli.Command) {
	fakeUI := &testterm.FakeUI{}
	configRepo := testconfig.NewRepository()
	pluginConfig := &testPluginConfig.FakePluginConfiguration{}
	manifestRepo := manifest.NewManifestDiskRepository()
	apiRepoLocator := api.NewRepositoryLocator(configRepo, map[string]net.Gateway{
		"auth":             net.NewUAAGateway(configRepo, fakeUI),
		"cloud-controller": net.NewCloudControllerGateway(configRepo, time.Now, fakeUI),
		"uaa":              net.NewUAAGateway(configRepo, fakeUI),
	})

	cmdFactory := command_factory.NewFactory(fakeUI, configRepo, manifestRepo, apiRepoLocator, pluginConfig)
	requirementsFactory := &testreq.FakeReqFactory{}
	cmdRunner := command_runner.NewRunner(cmdFactory, requirementsFactory, fakeUI)
	myApp := app.NewApp(cmdRunner, cmdFactory.CommandMetadatas()...)

	for _, cmd := range myApp.Commands {
		if cmd.Name == cmdName {
			return cmd
		}
	}
	panic(fmt.Sprintf("command %s does not exist", cmdName))
	return
}
