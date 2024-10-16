package plugininstaller

import (
	"code.cloudfoundry.org/cli/v8/cf/actors/pluginrepo"
	"code.cloudfoundry.org/cli/v8/cf/models"
	"code.cloudfoundry.org/cli/v8/cf/terminal"
	"code.cloudfoundry.org/cli/v8/cf/util/downloader"
	"code.cloudfoundry.org/cli/v8/util"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . PluginInstaller

type PluginInstaller interface {
	Install(inputSourceFilepath string) string
}

type Context struct {
	Checksummer    util.Sha1Checksum
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
