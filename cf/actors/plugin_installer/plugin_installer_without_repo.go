package plugin_installer

import (
	"os"
	"path/filepath"
	"strings"

	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type PluginInstallerWithoutRepo struct {
	Ui               terminal.UI
	PluginDownloader *PluginDownloader
	DownloadFromPath downloadFromPath
	RepoName         string
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

func (installer *PluginInstallerWithoutRepo) ensureCandidatePluginBinaryExistsAtGivenPath(pluginSourceFilepath string) bool {
	_, err := os.Stat(pluginSourceFilepath)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}
