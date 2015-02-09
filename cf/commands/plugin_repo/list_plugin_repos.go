package plugin_repo

import (
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"

	. "github.com/cloudfoundry/cli/cf/i18n"
)

type ListPluginRepos struct {
	ui     terminal.UI
	config core_config.Reader
}

func NewListPluginRepos(ui terminal.UI, config core_config.Reader) ListPluginRepos {
	return ListPluginRepos{
		ui:     ui,
		config: config,
	}
}

func (cmd ListPluginRepos) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "list-plugin-repos",
		Description: T("list all the added plugin repository"),
		Usage:       T("CF_NAME list-plugin-repos"),
	}
}

func (cmd ListPluginRepos) GetRequirements(_ requirements.Factory, c *cli.Context) (req []requirements.Requirement, err error) {
	return
}

func (cmd ListPluginRepos) Run(c *cli.Context) {
	repos := cmd.config.PluginRepos()

	table := terminal.NewTable(cmd.ui, []string{T("Repo Name"), T("Url")})

	for _, repo := range repos {
		table.Add(repo.Name, repo.Url)
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	table.Print()

	cmd.ui.Say("")
}
