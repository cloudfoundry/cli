package plugininstaller

import (
	"github.com/cloudfoundry/cli/cf/actors/pluginrepo"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/utils"
	"github.com/cloudfoundry/cli/utils/downloader"
)

//go:generate counterfeiter . PluginInstaller

type PluginInstaller interface {
	Install(inputSourceFilepath string) string
}

type Context struct {
	Checksummer    utils.Sha1Checksum
	FileDownloader downloader.Downloader
	GetPluginRepos pluginReposFetcher
	PluginRepo     pluginrepo.PluginRepo
	RepoName       string
	UI             terminal.UI
}

type pluginReposFetcher func() []models.PluginRepo

func NewPluginInstaller(context *Context) PluginInstaller {
	var installer PluginInstaller

	pluginDownloader := &PluginDownloader{UI: context.UI, FileDownloader: context.FileDownloader}
	if context.RepoName == "" {
		installer = &pluginInstallerWithoutRepo{
			UI:               context.UI,
			PluginDownloader: pluginDownloader,
			RepoName:         context.RepoName,
		}
	} else {
		installer = &pluginInstallerWithRepo{
			UI:               context.UI,
			PluginDownloader: pluginDownloader,
			RepoName:         context.RepoName,
			Checksummer:      context.Checksummer,
			PluginRepo:       context.PluginRepo,
			GetPluginRepos:   context.GetPluginRepos,
		}
	}
	return installer
}
