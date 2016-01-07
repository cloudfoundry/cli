package plugin_repo

import (
	"strings"

	"github.com/cloudfoundry/cli/cf/actors/plugin_repo"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
	"github.com/cloudfoundry/cli/flags/flag"

	clipr "github.com/cloudfoundry-incubator/cli-plugin-repo/models"

	. "github.com/cloudfoundry/cli/cf/i18n"
)

type RepoPlugins struct {
	ui         terminal.UI
	config     core_config.Reader
	pluginRepo plugin_repo.PluginRepo
}

func init() {
	command_registry.Register(&RepoPlugins{})
}

func (cmd *RepoPlugins) MetaData() command_registry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["r"] = &cliFlags.StringFlag{ShortName: "r", Usage: T("Repo Name - List plugins from just this repository")}

	return command_registry.CommandMetadata{
		Name:        T("repo-plugins"),
		Description: T("List all available plugins in all added repositories"),
		Usage: T(`CF_NAME repo-plugins

EXAMPLE:
   cf repo-plugins [-r REPO_NAME]
`),
		Flags: fs,
	}
}

func (cmd *RepoPlugins) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	return
}

func (cmd *RepoPlugins) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.pluginRepo = deps.PluginRepo
	return cmd
}

func (cmd *RepoPlugins) Execute(c flags.FlagContext) {
	var repos []models.PluginRepo
	repoName := c.String("r")

	repos = cmd.config.PluginRepos()

	if repoName == "" {
		cmd.ui.Say(T("Getting plugins from all repositories ... "))
	} else {
		index := cmd.findRepoIndex(repoName)
		if index != -1 {
			cmd.ui.Say(T("Getting plugins from repository '") + repoName + "' ...")
			repos = []models.PluginRepo{repos[index]}
		} else {
			cmd.ui.Failed(repoName + T(" does not exist as an available plugin repo."+"\nTip: use `add-plugin-repo` command to add repos."))
		}
	}

	cmd.ui.Say("")

	repoPlugins, repoError := cmd.pluginRepo.GetPlugins(repos)

	cmd.printTable(repoPlugins)

	cmd.printErrors(repoError)
}

func (cmd RepoPlugins) printTable(repoPlugins map[string][]clipr.Plugin) {
	for k, plugins := range repoPlugins {
		cmd.ui.Say(terminal.ColorizeBold(T("Repository: ")+k, 33))
		table := cmd.ui.Table([]string{T("name"), T("version"), T("description")})
		for _, p := range plugins {
			table.Add(p.Name, p.Version, p.Description)
		}
		table.Print()
		cmd.ui.Say("")
	}
}

func (cmd RepoPlugins) printErrors(repoError []string) {
	if len(repoError) > 0 {
		cmd.ui.Say(terminal.ColorizeBold(T("Logged errors:"), 31))
		for _, e := range repoError {
			cmd.ui.Say(terminal.Colorize(e, 31))
		}
		cmd.ui.Say("")
	}
}

func (cmd RepoPlugins) findRepoIndex(repoName string) int {
	repos := cmd.config.PluginRepos()
	for i, repo := range repos {
		if strings.ToLower(repo.Name) == strings.ToLower(repoName) {
			return i
		}
	}
	return -1
}
