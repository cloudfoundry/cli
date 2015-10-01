package plugin_installer

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	. "github.com/cloudfoundry/cli/cf/i18n"

	clipr "github.com/cloudfoundry-incubator/cli-plugin-repo/models"
	"github.com/cloudfoundry/cli/cf/actors/plugin_repo"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/fileutils"
	"github.com/cloudfoundry/cli/utils"
)

type PluginInstaller interface {
	Install(inputSourceFilepath string) string
}

type PluginDownloader struct {
	Ui             terminal.UI
	FileDownloader fileutils.Downloader
}

type InstallerContext struct {
	PluginDownloader *PluginDownloader
	RepoName         string
	Checksummer      utils.Sha1Checksum
	PluginRepo       plugin_repo.PluginRepo
	Ui               terminal.UI
	GetPluginRepos   pluginReposFetcher
}

type pluginReposFetcher func() []models.PluginRepo
type downloadFromPath func(pluginSourceFilepath string, downloader fileutils.Downloader) string

type PluginInstallerWithRepo struct {
	Ui               terminal.UI
	PluginDownloader *PluginDownloader
	DownloadFromPath downloadFromPath
	RepoName         string
	Checksummer      utils.Sha1Checksum
	PluginRepo       plugin_repo.PluginRepo
	GetPluginRepos   pluginReposFetcher
}

type PluginInstallerWithoutRepo struct {
	Ui               terminal.UI
	PluginDownloader *PluginDownloader
	DownloadFromPath downloadFromPath
	RepoName         string
}

func NewPluginInstaller(context *InstallerContext) (installer PluginInstaller) {
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

func (installer *PluginInstallerWithoutRepo) Install(inputSourceFilepath string) (outputSourceFilepath string) {
	if filepath.Dir(inputSourceFilepath) == "." {
		outputSourceFilepath = "./" + filepath.Clean(inputSourceFilepath)
	} else {
		outputSourceFilepath = inputSourceFilepath
	}

	installer.Ui.Say("")
	if strings.HasPrefix(outputSourceFilepath, "https://") || strings.HasPrefix(outputSourceFilepath, "http://") ||
		strings.HasPrefix(outputSourceFilepath, "ftp://") || strings.HasPrefix(outputSourceFilepath, "ftps://") {
		installer.Ui.Say(T("Attempting to download binary file from internet address..."))
		return installer.PluginDownloader.downloadFromPath(outputSourceFilepath)
	} else if !installer.ensureCandidatePluginBinaryExistsAtGivenPath(outputSourceFilepath) {
		installer.Ui.Failed(T("File not found locally, make sure the file exists at given path {{.filepath}}", map[string]interface{}{"filepath": outputSourceFilepath}))
	}

	return outputSourceFilepath
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

func (installer *PluginInstallerWithoutRepo) ensureCandidatePluginBinaryExistsAtGivenPath(pluginSourceFilepath string) bool {
	_, err := os.Stat(pluginSourceFilepath)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

func (downloader *PluginDownloader) downloadFromPath(pluginSourceFilepath string) string {
	size, filename, err := downloader.FileDownloader.DownloadFile(pluginSourceFilepath)

	if err != nil {
		downloader.Ui.Failed(fmt.Sprintf(T("Download attempt failed: {{.Error}}\n\nUnable to install, plugin is not available from the given url.", map[string]interface{}{"Error": err.Error()})))
	}

	downloader.Ui.Say(fmt.Sprintf("%d "+T("bytes downloaded")+"...", size))

	executablePath := filepath.Join(downloader.FileDownloader.SavePath(), filename)
	os.Chmod(executablePath, 0700)

	return executablePath
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

func (downloader *PluginDownloader) downloadFromPlugin(plugin clipr.Plugin) (string, string) {
	arch := runtime.GOARCH

	switch runtime.GOOS {
	case "darwin":
		return downloader.downloadFromPath(downloader.getBinaryUrl(plugin, "osx")), downloader.getBinaryChecksum(plugin, "osx")
	case "linux":
		if arch == "386" {
			return downloader.downloadFromPath(downloader.getBinaryUrl(plugin, "linux32")), downloader.getBinaryChecksum(plugin, "linux32")
		} else {
			return downloader.downloadFromPath(downloader.getBinaryUrl(plugin, "linux64")), downloader.getBinaryChecksum(plugin, "linux64")
		}
	case "windows":
		if arch == "386" {
			return downloader.downloadFromPath(downloader.getBinaryUrl(plugin, "win32")), downloader.getBinaryChecksum(plugin, "win32")
		} else {
			return downloader.downloadFromPath(downloader.getBinaryUrl(plugin, "win64")), downloader.getBinaryChecksum(plugin, "win64")
		}
	default:
		downloader.binaryNotAvailable()
	}
	return "", ""
}

func (downloader *PluginDownloader) getBinaryUrl(plugin clipr.Plugin, os string) string {
	for _, binary := range plugin.Binaries {
		if binary.Platform == os {
			return binary.Url
		}
	}
	downloader.binaryNotAvailable()
	return ""
}

func (downloader *PluginDownloader) getBinaryChecksum(plugin clipr.Plugin, os string) string {
	for _, binary := range plugin.Binaries {
		if binary.Platform == os {
			return binary.Checksum
		}
	}
	return ""
}

func (downloader *PluginDownloader) binaryNotAvailable() {
	downloader.Ui.Failed(T("Plugin requested has no binary available for your OS: ") + runtime.GOOS + ", " + runtime.GOARCH)
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
