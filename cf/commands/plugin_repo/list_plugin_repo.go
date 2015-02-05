package plugin_repo

import (
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"

	. "github.com/cloudfoundry/cli/cf/i18n"
)

type ListPluginRepo struct {
	ui     terminal.UI
	config core_config.Reader
}

func NewListPluginRepo(ui terminal.UI, config core_config.Reader) ListPluginRepo {
	return ListPluginRepo{
		ui:     ui,
		config: config,
	}
}

func (cmd ListPluginRepo) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "list-plugin-repo",
		Description: T("list all the added plugin repository"),
		Usage:       T("CF_NAME list-plugin-repo"),
	}
}

func (cmd ListPluginRepo) GetRequirements(_ requirements.Factory, c *cli.Context) (req []requirements.Requirement, err error) {
	return
}

func (cmd ListPluginRepo) Run(c *cli.Context) {
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
