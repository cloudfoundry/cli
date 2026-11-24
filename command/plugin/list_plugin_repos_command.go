package plugin

import (
	"code.cloudfoundry.org/cli/v8/command"
	"code.cloudfoundry.org/cli/v8/util/ui"
)

type ListPluginReposCommand struct {
	usage           interface{} `usage:"CF_NAME list-plugin-repos"`
	relatedCommands interface{} `related_commands:"add-plugin-repo, install-plugin"`
	UI              command.UI
	Config          command.Config
}

func (cmd *ListPluginReposCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	return nil
}

func (cmd ListPluginReposCommand) Execute(args []string) error {
	table := [][]string{{"Repo Name", "URL"}}
	for _, repository := range cmd.Config.PluginRepositories() {
		table = append(table, []string{repository.Name, repository.URL})
	}

	cmd.UI.DisplayOK()
	cmd.UI.DisplayTableWithHeader("", table, ui.DefaultTableSpacePadding)
	cmd.UI.DisplayNewline()
	return nil
}
