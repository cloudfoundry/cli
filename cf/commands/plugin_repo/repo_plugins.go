package plugin_repo

import (
	"strings"

	"github.com/cloudfoundry/cli/cf/actors/plugin_repo"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/flag_helpers"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"

	clipr "github.com/cloudfoundry-incubator/cli-plugin-repo/models"

	. "github.com/cloudfoundry/cli/cf/i18n"
)

type RepoPlugins struct {
	ui         terminal.UI
	config     core_config.Reader
	pluginRepo plugin_repo.PluginRepo
}

func NewRepoPlugins(ui terminal.UI, config core_config.Reader, pluginRepo plugin_repo.PluginRepo) RepoPlugins {
	return RepoPlugins{
		ui:         ui,
		config:     config,
		pluginRepo: pluginRepo,
	}
}

func (cmd RepoPlugins) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        T("repo-plugins"),
		Description: T("List all available plugins in all added repositories"),
		Usage: T(`CF_NAME repo-plugins

EXAMPLE:
   cf repo-plugins [-r REPO_NAME]
`),
		Flags: []cli.Flag{
			flag_helpers.NewStringFlag("r", T("Repo Name - List plugins from just this repository")),
		},
	}
}

func (cmd RepoPlugins) GetRequirements(_ requirements.Factory, c *cli.Context) (req []requirements.Requirement, err error) {
	return
}

func (cmd RepoPlugins) Run(c *cli.Context) {
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

func getListEndpoint(url string) string {
	if strings.HasSuffix(url, "/") {
		return url + "list"
	}
	return url + "/list"
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
