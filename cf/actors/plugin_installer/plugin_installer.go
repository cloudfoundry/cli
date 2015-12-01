package plugin_installer

import (
	"github.com/cloudfoundry/cli/cf/actors/plugin_repo"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/downloader"
	"github.com/cloudfoundry/cli/utils"
)

type PluginInstaller interface {
	Install(inputSourceFilepath string) string
}

type PluginInstallerContext struct {
	Checksummer    utils.Sha1Checksum
	FileDownloader downloader.Downloader
	GetPluginRepos pluginReposFetcher
	PluginRepo     plugin_repo.PluginRepo
	RepoName       string
	Ui             terminal.UI
}

type pluginReposFetcher func() []models.PluginRepo

func NewPluginInstaller(context *PluginInstallerContext) (installer PluginInstaller) {
	pluginDownloader := &PluginDownloader{Ui: context.Ui, FileDownloader: context.FileDownloader}
	if context.RepoName == "" {
		installer = &PluginInstallerWithoutRepo{
			Ui:               context.Ui,
			PluginDownloader: pluginDownloader,
			RepoName:         context.RepoName,
		}
	} else {
		installer = &PluginInstallerWithRepo{
			Ui:               context.Ui,
			PluginDownloader: pluginDownloader,
			RepoName:         context.RepoName,
			Checksummer:      context.Checksummer,
			PluginRepo:       context.PluginRepo,
			GetPluginRepos:   context.GetPluginRepos,
		}
	}
	return installer
}
