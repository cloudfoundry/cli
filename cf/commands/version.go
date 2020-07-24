package commands

import (
	"fmt"

	"code.cloudfoundry.org/cli/cf"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cli/version"
)

type Version struct {
	ui terminal.UI
}

func init() {
	commandregistry.Register(&Version{})
}

func (cmd *Version) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
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

func (cmd *Version) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	return cmd
}

func (cmd *Version) Requirements(requirementsFactory requirements.Factory, context flags.FlagContext) ([]requirements.Requirement, error) {
	reqs := []requirements.Requirement{}
	return reqs, nil
}

func (cmd *Version) Execute(context flags.FlagContext) error {
	cmd.ui.Say(fmt.Sprintf("%s version %s", cf.Name, version.VersionString()))
	return nil
}
