package plugin_installer

import (
	"errors"
	"strings"

	clipr "github.com/cloudfoundry-incubator/cli-plugin-repo/models"
	"github.com/cloudfoundry/cli/cf/actors/plugin_repo"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/utils"
)

type PluginInstallerWithRepo struct {
	Ui               terminal.UI
	PluginDownloader *PluginDownloader
	DownloadFromPath downloadFromPath
	RepoName         string
	Checksummer      utils.Sha1Checksum
	PluginRepo       plugin_repo.PluginRepo
	GetPluginRepos   pluginReposFetcher
}

func (installer *PluginInstallerWithRepo) Install(inputSourceFilepath string) (outputSourceFilepath string) {
	targetPluginName := strings.ToLower(inputSourceFilepath)

	installer.Ui.Say(T("Looking up '{{.filePath}}' from repository '{{.repoName}}'", map[string]interface{}{"filePath": inputSourceFilepath, "repoName": installer.RepoName}))

	repoModel, err := installer.getRepoFromConfig(installer.RepoName)
	if err != nil {
		installer.Ui.Failed(err.Error() + "\n" + T("Tip: use 'add-plugin-repo' to register the repo"))
	}

	pluginList, repoAry := installer.PluginRepo.GetPlugins([]models.PluginRepo{repoModel})
	if len(repoAry) != 0 {
		installer.Ui.Failed(T("Error getting plugin metadata from repo: ") + repoAry[0])
	}

	found := false
	sha1 := ""
	for _, plugin := range findRepoCaseInsensity(pluginList, installer.RepoName) {
		if strings.ToLower(plugin.Name) == targetPluginName {
			found = true
			outputSourceFilepath, sha1 = installer.PluginDownloader.downloadFromPlugin(plugin)

			installer.Checksummer.SetFilePath(outputSourceFilepath)
			if !installer.Checksummer.CheckSha1(sha1) {
				installer.Ui.Failed(T("Downloaded plugin binary's checksum does not match repo metadata"))
			}
		}

	}
	if !found {
		installer.Ui.Failed(inputSourceFilepath + T(" is not available in repo '") + installer.RepoName + "'")
	}

	return outputSourceFilepath
}

func (installer *PluginInstallerWithRepo) getRepoFromConfig(repoName string) (models.PluginRepo, error) {
	targetRepo := strings.ToLower(repoName)
	list := installer.GetPluginRepos()

	for i, repo := range list {
		if strings.ToLower(repo.Name) == targetRepo {
			return list[i], nil
		}
	}

	return models.PluginRepo{}, errors.New(repoName + T(" not found"))
}

func findRepoCaseInsensity(repoList map[string][]clipr.Plugin, repoName string) []clipr.Plugin {
	target := strings.ToLower(repoName)
	for k, repo := range repoList {
		if strings.ToLower(k) == target {
			return repo
		}
	}
	return nil
}
