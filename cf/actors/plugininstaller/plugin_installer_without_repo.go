package plugininstaller

import (
	"os"
	"path/filepath"
	"strings"

	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/terminal"
)

type pluginInstallerWithoutRepo struct {
	UI               terminal.UI
	PluginDownloader *PluginDownloader
	DownloadFromPath downloadFromPath
	RepoName         string
}

func (installer *pluginInstallerWithoutRepo) Install(inputSourceFilepath string) (outputSourceFilepath string) {
	if filepath.Dir(inputSourceFilepath) == "." {
		outputSourceFilepath = "./" + filepath.Clean(inputSourceFilepath)
	} else {
		outputSourceFilepath = inputSourceFilepath
	}

	installer.UI.Say("")
	if strings.HasPrefix(outputSourceFilepath, "https://") || strings.HasPrefix(outputSourceFilepath, "http://") ||
		strings.HasPrefix(outputSourceFilepath, "ftp://") || strings.HasPrefix(outputSourceFilepath, "ftps://") {
		installer.UI.Say(T("Attempting to download binary file from internet address..."))
		return installer.PluginDownloader.downloadFromPath(outputSourceFilepath)
	} else if !installer.ensureCandidatePluginBinaryExistsAtGivenPath(outputSourceFilepath) {
		installer.UI.Failed(T("File not found locally, make sure the file exists at given path {{.filepath}}", map[string]interface{}{"filepath": outputSourceFilepath}))
	}

	return outputSourceFilepath
}

func (installer *pluginInstallerWithoutRepo) ensureCandidatePluginBinaryExistsAtGivenPath(pluginSourceFilepath string) bool {
	_, err := os.Stat(pluginSourceFilepath)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}
