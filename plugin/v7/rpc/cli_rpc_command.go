// +build V7

package rpc

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"sync"

	"code.cloudfoundry.org/cli/command"
	v7command "code.cloudfoundry.org/cli/command/v7"
	plugin "code.cloudfoundry.org/cli/plugin/v7"
	plugin_models "code.cloudfoundry.org/cli/plugin/v7/models"
	"code.cloudfoundry.org/cli/version"
	"github.com/blang/semver"
)

type CliRpcCmd struct {
	PluginMetadata       *plugin.PluginMetadata
	MetadataMutex        *sync.RWMutex
	Config               command.Config
	AppActor             v7command.AppActor
	outputCapture        OutputCapture
	terminalOutputSwitch TerminalOutputSwitch
	outputBucket         *bytes.Buffer
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
	return errors.New("unimplemented")
}

func (cmd *CliRpcCmd) GetOutputAndReset(args bool, retVal *[]string) error {
	return errors.New("unimplemented")
}

func (cmd *CliRpcCmd) GetApp(appName string, retVal *plugin_models.GetAppModel) error {
	spaceGUID := cmd.Config.TargetedSpace().GUID
	app, _, err := cmd.AppActor.GetDetailedAppSummary(appName, spaceGUID, true)

	retVal.Guid = app.GUID
	retVal.Name = app.Name
	retVal.SpaceGuid = spaceGUID
	retVal.State = strings.ToLower(string(app.State))

	// BackpackUrl is no longer singular, the current droplet may have many buildpacks associated
	// TODO: Also expose current droplet as field on the model?
	if len(app.CurrentDroplet.Buildpacks) > 0 {
		retVal.BuildpackUrl = app.CurrentDroplet.Buildpacks[0].Name
	}

	// Command is no longer singular, the application may have multiple processes
	// TODO: Also expose all processes as a field on the model?
	if len(app.ProcessSummaries) > 0 {
		retVal.Command = app.ProcessSummaries[0].Command.String()
		retVal.DiskQuota = int64(app.ProcessSummaries[0].DiskInMB.Value)
		retVal.HealthCheckTimeout = int(app.ProcessSummaries[0].HealthCheckTimeout)
		retVal.InstanceCount = app.ProcessSummaries[0].Instances.Value
		retVal.Memory = int64(app.ProcessSummaries[0].MemoryInMB.Value)
		retVal.RunningInstances = app.ProcessSummaries[0].HealthyInstanceCount()
	}

	// In v7 this requires calling actor.GetEnvironmentVariablesByApplicationNameAndSpace(appName, spaceGUID)
	// cmd.pluginAppModel.EnvironmentVars = getSummaryApp.EnvironmentVars

	// APP.DetectedStartCommand is v2-only
	// cmd.pluginAppModel.DetectedStartCommand = getSummaryApp.DetectedStartCommand

	// In v7 this requires calling actor.GetApplicationPackages
	// cmd.pluginAppModel.PackageState = getSummaryApp.PackageState

	// We don't currently have this field in V7 (it does exist in CAPI v3)
	// cmd.pluginAppModel.PackageUpdatedAt = getSummaryApp.PackageUpdatedAt

	// The stack guid is not provided in the normal app response, we would need to call actor.GetStackByName to get this
	retVal.Stack = &plugin_models.GetApp_Stack{
		Name: app.StackName,
	}

	return err
}
