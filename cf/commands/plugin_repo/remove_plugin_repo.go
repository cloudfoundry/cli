package plugin_repo

import (
	"strings"

	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"

	. "github.com/cloudfoundry/cli/cf/i18n"
)

type RemovePluginRepo struct {
	ui     terminal.UI
	config core_config.ReadWriter
}

func NewRemovePluginRepo(ui terminal.UI, config core_config.ReadWriter) RemovePluginRepo {
	return RemovePluginRepo{
		ui:     ui,
		config: config,
	}
}

func (cmd RemovePluginRepo) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "remove-plugin-repo",
		Description: T("Remove a plugin repository"),
		Usage: T(`CF_NAME remove-plugin-repo [REPO_NAME] [URL]

EXAMPLE:
   cf remove-plugin-repo PrivateRepo
`),
		TotalArgs: 1,
	}
}

func (cmd RemovePluginRepo) GetRequirements(_ requirements.Factory, c *cli.Context) (req []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		cmd.ui.FailWithUsage(c)
	}
	return
}

func (cmd RemovePluginRepo) Run(c *cli.Context) {

	cmd.ui.Say("")
	repoName := strings.Trim(c.Args()[0], " ")

	if i := cmd.findRepoIndex(repoName); i != -1 {
		cmd.config.UnSetPluginRepo(i)
		cmd.ui.Ok()
		cmd.ui.Say(repoName + T(" removed from list of repositories"))
		cmd.ui.Say("")
	} else {
		cmd.ui.Failed(repoName + T(" does not exist as a repo"))
	}
}

func (cmd RemovePluginRepo) findRepoIndex(repoName string) int {
	repos := cmd.config.PluginRepos()
	for i, repo := range repos {
		if strings.ToLower(repo.Name) == strings.ToLower(repoName) {
			return i
		}
	}
	return -1
}
