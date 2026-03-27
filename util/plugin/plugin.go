package plugin

import (
	"errors"
	"fmt"
	"os"

	plugin_transition "code.cloudfoundry.org/cli/v8/plugin/transition"
	"code.cloudfoundry.org/cli/v8/util/configv3"
	"code.cloudfoundry.org/cli/v8/util/ui"

	"code.cloudfoundry.org/cli/v8/command/common"
	"code.cloudfoundry.org/cli/v8/command/translatableerror"
)

var ErrFailed = errors.New("command failed")
var ParseErr = errors.New("incorrect type for arg")

type DisplayUsage interface {
	DisplayUsage()
}

type UI interface {
	DisplayError(err error)
	DisplayWarning(template string, templateValues ...map[string]interface{})
	DisplayText(template string, templateValues ...map[string]interface{})
	FlushDeferred()
}

func PluginCommandNames() []string {
	var names []string

	config, configErr := configv3.LoadConfig()
	if configErr != nil {
		return names
	}

	for _, plugin := range config.Plugins() {
		for _, pluginCommand := range plugin.Commands {
			names = append(names, pluginCommand.Name)
		}
	}

	return names
}

func RunPlugin(plugin configv3.Plugin) error {
	_, commandUI, err := getCFConfigAndCommandUIObjects()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		return err
	}
	defer commandUI.FlushDeferred()
	pluginErr := plugin_transition.RunPlugin(plugin, commandUI)
	if pluginErr != nil {
		return handleError(pluginErr, commandUI)
	}
	return nil
}

func getCFConfigAndCommandUIObjects() (*configv3.Config, *ui.UI, error) {
	cfConfig, configErr := configv3.LoadConfig(configv3.FlagOverride{
		Verbose: common.Commands.VerboseOrVersion,
	})
	if configErr != nil {
		if _, ok := configErr.(translatableerror.EmptyConfigError); !ok {
			return nil, nil, configErr
		}
	}
	commandUI, err := ui.NewUI(cfConfig)
	return cfConfig, commandUI, err
}

func handleError(passedErr error, commandUI UI) error {
	if passedErr == nil {
		return nil
	}

	translatedErr := translatableerror.ConvertToTranslatableError(passedErr)
	commandUI.DisplayError(translatedErr)

	if _, ok := translatedErr.(DisplayUsage); ok {
		return ParseErr
	}

	return ErrFailed
}
