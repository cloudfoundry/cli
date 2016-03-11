package commands

import (
	"fmt"
	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/command_registry"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
)

type Version struct {
	ui terminal.UI
}

func init() {
	command_registry.Register(&Version{})
}

func (cmd *Version) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{
		Name:        "version",
		Description: T("Print the version"),
		Usage: []string{
			"CF_NAME version",
			"\n\n   ",
			T("'{{.VersionShort}}' and '{{.VersionLong}}' are also accepted.", map[string]string{
				"VersionShort": "cf -v",
				"VersionLong":  "cf --version",
			}),
		},
	}
}

func (cmd *Version) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	return cmd
}

func (cmd *Version) Requirements(requirementsFactory requirements.Factory, context flags.FlagContext) []requirements.Requirement {
	reqs := []requirements.Requirement{}
	return reqs
}

func (cmd *Version) Execute(context flags.FlagContext) {
	cmd.ui.Say(fmt.Sprintf("%s version %s-%s", cf.Name, cf.Version, cf.BuiltOnDate))
}
