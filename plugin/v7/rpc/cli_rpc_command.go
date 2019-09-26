// +build V7

package rpc

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"sync"

	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cli/cf/trace"
	plugin "code.cloudfoundry.org/cli/plugin/v7"
	plugin_models "code.cloudfoundry.org/cli/plugin/v7/models"
	"code.cloudfoundry.org/cli/version"
	"github.com/blang/semver"
)

type CliRpcCmd struct {
	PluginMetadata       *plugin.PluginMetadata
	MetadataMutex        *sync.RWMutex
	outputCapture        OutputCapture
	terminalOutputSwitch TerminalOutputSwitch
	cliConfig            coreconfig.Repository
	repoLocator          api.RepositoryLocator
	newCmdRunner         CommandRunner
	outputBucket         *bytes.Buffer
	logger               trace.Printer
	stdout               io.Writer
}

func (cmd *CliRpcCmd) IsMinCliVersion(passedVersion string, retVal *bool) error {
	if version.VersionString() == version.DefaultVersion {
		*retVal = true
		return nil
	}

	actualVersion, err := semver.Make(version.VersionString())
	if err != nil {
		return err
	}

	requiredVersion, err := semver.Make(passedVersion)
	if err != nil {
		return err
	}

	*retVal = actualVersion.GTE(requiredVersion)

	return nil
}

func (cmd *CliRpcCmd) SetPluginMetadata(pluginMetadata plugin.PluginMetadata, retVal *bool) error {
	cmd.MetadataMutex.Lock()
	defer cmd.MetadataMutex.Unlock()

	cmd.PluginMetadata = &pluginMetadata
	*retVal = true
	return nil
}

func (cmd *CliRpcCmd) DisableTerminalOutput(disable bool, retVal *bool) error {
	cmd.terminalOutputSwitch.DisableTerminalOutput(disable)
	*retVal = true
	return nil
}

func (cmd *CliRpcCmd) CallCoreCommand(args []string, retVal *bool) error {
	var err error
	cmdRegistry := commandregistry.Commands

	cmd.outputBucket = &bytes.Buffer{}
	cmd.outputCapture.SetOutputBucket(cmd.outputBucket)

	if cmdRegistry.CommandExists(args[0]) {
		deps := commandregistry.NewDependency(cmd.stdout, cmd.logger, dialTimeout)

		//set deps objs to be the one used by all other commands
		//once all commands are converted, we can make fresh deps for each command run
		deps.Config = cmd.cliConfig
		deps.RepoLocator = cmd.repoLocator

		//set command ui's TeePrinter to be the one used by RpcService, for output to be captured
		deps.UI = terminal.NewUI(os.Stdin, cmd.stdout, cmd.outputCapture.(*terminal.TeePrinter), cmd.logger)

		err = cmd.newCmdRunner.Command(args, deps, false)
	} else {
		*retVal = false
		return nil
	}

	if err != nil {
		*retVal = false
		return err
	}

	*retVal = true
	return nil
}

func (cmd *CliRpcCmd) GetOutputAndReset(args bool, retVal *[]string) error {
	v := strings.TrimSuffix(cmd.outputBucket.String(), "\n")
	*retVal = strings.Split(v, "\n")
	return nil
}

func (cmd *CliRpcCmd) GetApp(appName string, retVal *plugin_models.GetAppModel) error {
	return errors.New("unimplemented")
}
