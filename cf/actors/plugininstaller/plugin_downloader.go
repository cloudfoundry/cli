package plugininstaller

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	clipr "github.com/cloudfoundry-incubator/cli-plugin-repo/web"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/utils/downloader"
)

type PluginDownloader struct {
	UI             terminal.UI
	FileDownloader downloader.Downloader
}
type downloadFromPath func(string, downloader.Downloader) string

func (downloader *PluginDownloader) downloadFromPath(pluginSourceFilepath string) string {
	size, filename, err := downloader.FileDownloader.DownloadFile(pluginSourceFilepath)

	if err != nil {
		downloader.UI.Failed(fmt.Sprintf(T("Download attempt failed: {{.Error}}\n\nUnable to install, plugin is not available from the given url.", map[string]interface{}{"Error": err.Error()})))
	}

	downloader.UI.Say(fmt.Sprintf("%d "+T("bytes downloaded")+"...", size))

	executablePath := filepath.Join(downloader.FileDownloader.SavePath(), filename)
	err = os.Chmod(executablePath, 0700)
	if err != nil {
		downloader.UI.Failed(fmt.Sprintf(T("Failed to make plugin executable: {{.Error}}", map[string]interface{}{"Error": err.Error()})))
	}

	return executablePath
}

func (downloader *PluginDownloader) downloadFromPlugin(plugin clipr.Plugin) (string, string) {
	arch := runtime.GOARCH

	switch runtime.GOOS {
	case "darwin":
		return downloader.downloadFromPath(downloader.getBinaryURL(plugin, "osx")), downloader.getBinaryChecksum(plugin, "osx")
	case "linux":
		if arch == "386" {
			return downloader.downloadFromPath(downloader.getBinaryURL(plugin, "linux32")), downloader.getBinaryChecksum(plugin, "linux32")
		}
		return downloader.downloadFromPath(downloader.getBinaryURL(plugin, "linux64")), downloader.getBinaryChecksum(plugin, "linux64")
	case "windows":
		if arch == "386" {
			return downloader.downloadFromPath(downloader.getBinaryURL(plugin, "win32")), downloader.getBinaryChecksum(plugin, "win32")
		}
		return downloader.downloadFromPath(downloader.getBinaryURL(plugin, "win64")), downloader.getBinaryChecksum(plugin, "win64")
	default:
		downloader.binaryNotAvailable()
	}
	return "", ""
}

func (downloader *PluginDownloader) getBinaryURL(plugin clipr.Plugin, os string) string {
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
	downloader.UI.Failed(T("Plugin requested has no binary available for your OS: ") + runtime.GOOS + ", " + runtime.GOARCH)
}
