package plugin_installer

import (
	"github.com/cloudfoundry/cli/cf/actors/plugin_repo"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/utils"
)

type PluginInstaller interface {
	Install(inputSourceFilepath string) string
}

type PluginInstallerContext struct {
	PluginDownloader *PluginDownloader
	RepoName         string
	Checksummer      utils.Sha1Checksum
	PluginRepo       plugin_repo.PluginRepo
	Ui               terminal.UI
	GetPluginRepos   pluginReposFetcher
}

type pluginReposFetcher func() []models.PluginRepo

func NewPluginInstaller(context *PluginInstallerContext) (installer PluginInstaller) {
	if context.RepoName == "" {
		installer = &PluginInstallerWithoutRepo{
			Ui:               context.Ui,
			PluginDownloader: context.PluginDownloader,
			RepoName:         context.RepoName,
		}
	} else {
		installer = &PluginInstallerWithRepo{
			Ui:               context.Ui,
			PluginDownloader: context.PluginDownloader,
			RepoName:         context.RepoName,
			Checksummer:      context.Checksummer,
			PluginRepo:       context.PluginRepo,
			GetPluginRepos:   context.GetPluginRepos,
		}
	}
	return installer
}
