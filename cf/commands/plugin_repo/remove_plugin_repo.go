package plugin_repo

import (
	"strings"

	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"

	. "github.com/cloudfoundry/cli/cf/i18n"
)

type RemovePluginRepo struct {
	ui     terminal.UI
	config core_config.ReadWriter
}

func init() {
	command_registry.Register(&RemovePluginRepo{})
}

func (cmd *RemovePluginRepo) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{
		Name:        "remove-plugin-repo",
		Description: T("Remove a plugin repository"),
		Usage: T(`CF_NAME remove-plugin-repo [REPO_NAME] [URL]

EXAMPLE:
   cf remove-plugin-repo PrivateRepo
`),
		TotalArgs: 1,
	}
}

func (cmd *RemovePluginRepo) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + command_registry.Commands.CommandUsage("remove-plugin-repo"))
	}

	return
}

func (cmd *RemovePluginRepo) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	return cmd
}

func (cmd *RemovePluginRepo) Execute(c flags.FlagContext) {
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
