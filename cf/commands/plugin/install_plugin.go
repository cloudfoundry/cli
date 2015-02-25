package plugin

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/cloudfoundry/cli/cf/actors/plugin_repo"
	"github.com/cloudfoundry/cli/cf/command"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/configuration/plugin_config"
	"github.com/cloudfoundry/cli/cf/flag_helpers"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/fileutils"
	"github.com/cloudfoundry/cli/plugin"
	"github.com/cloudfoundry/cli/plugin/rpc"
	"github.com/cloudfoundry/cli/utils"
	"github.com/codegangsta/cli"

	clipr "github.com/cloudfoundry-incubator/cli-plugin-repo/models"
)

type PluginInstall struct {
	ui           terminal.UI
	config       core_config.Reader
	pluginConfig plugin_config.PluginConfiguration
	coreCmds     map[string]command.Command
	pluginRepo   plugin_repo.PluginRepo
	checksum     utils.Sha1Checksum
	rpcService   *rpc.CliRpcService
}

func NewPluginInstall(ui terminal.UI, config core_config.Reader, pluginConfig plugin_config.PluginConfiguration, coreCmds map[string]command.Command, pluginRepo plugin_repo.PluginRepo, checksum utils.Sha1Checksum, rpcService *rpc.CliRpcService) *PluginInstall {
	return &PluginInstall{
		ui:           ui,
		config:       config,
		pluginConfig: pluginConfig,
		coreCmds:     coreCmds,
		pluginRepo:   pluginRepo,
		checksum:     checksum,
		rpcService:   rpcService,
	}
}

func (cmd *PluginInstall) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "install-plugin",
		Description: T("Install the plugin defined in command argument"),
		Usage: T(`CF_NAME install-plugin URL or LOCAL-PATH/TO/PLUGIN [-r REPO_NAME]

The command will download the plugin binary from repository if '-r' is provided

EXAMPLE:
   cf install-plugin https://github.com/cf-experimental/plugin-foobar
   cf install-plugin ~/Downloads/plugin-foobar
   cf install-plugin plugin-echo -r My-Repo 
`),
		Flags: []cli.Flag{
			flag_helpers.NewStringFlag("r", T("repo name where the plugin binary is located")),
		},
		TotalArgs: 1,
	}
}

func (cmd *PluginInstall) GetRequirements(_ requirements.Factory, c *cli.Context) (req []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		cmd.ui.FailWithUsage(c)
	}

	return
}

func (cmd *PluginInstall) Run(c *cli.Context) {
	downloader := fileutils.NewDownloader(os.TempDir())

	removeTmpFile := func() {
		err := downloader.RemoveFile()
		if err != nil {
			cmd.ui.Say(T("Problem removing downloaded binary in temp directory: ") + err.Error())
		}
	}
	defer removeTmpFile()

	pluginSourceFilepath := c.Args()[0]

	repoName := c.String("r")

	if repoName != "" {
		targetPluginName := strings.ToLower(c.Args()[0])

		cmd.ui.Say(T("Looking up '{{.filePath}}' from repository '{{.repoName}}'", map[string]interface{}{"filePath": pluginSourceFilepath, "repoName": repoName}))

		repoModel, err := cmd.getRepoFromConfig(repoName)
		if err != nil {
			cmd.ui.Failed(err.Error() + "\n" + T("Tip: use 'add-plugin-repo' to register the repo"))
		}

		pluginList, repoAry := cmd.pluginRepo.GetPlugins([]models.PluginRepo{repoModel})
		if len(repoAry) != 0 {
			cmd.ui.Failed(T("Error getting plugin metadata from repo: ") + repoAry[0])
		}

		found := false
		sha1 := ""
		for _, plugin := range findRepoCaseInsensity(pluginList, repoName) {
			if strings.ToLower(plugin.Name) == targetPluginName {
				found = true
				pluginSourceFilepath, sha1 = cmd.downloadBinary(plugin, downloader)

				cmd.checksum.SetFilePath(pluginSourceFilepath)
				if !cmd.checksum.CheckSha1(sha1) {
					cmd.ui.Failed(T("Downloaded plugin binary's checksum does not match repo metadata"))
				}
			}

		}
		if !found {
			cmd.ui.Failed(pluginSourceFilepath + T(" is not available in repo '") + repoName + "'")
		}
	} else {
		if filepath.Dir(pluginSourceFilepath) == "." {
			pluginSourceFilepath = "./" + filepath.Clean(pluginSourceFilepath)
		}

		cmd.ui.Say("")
		if strings.HasPrefix(pluginSourceFilepath, "https://") || strings.HasPrefix(pluginSourceFilepath, "http://") ||
			strings.HasPrefix(pluginSourceFilepath, "ftp://") || strings.HasPrefix(pluginSourceFilepath, "ftps://") {
			cmd.ui.Say(T("Attempting to download binary file from internet address..."))
			pluginSourceFilepath = cmd.tryDownloadPluginBinaryfromGivenPath(pluginSourceFilepath, downloader)
		} else if !cmd.ensureCandidatePluginBinaryExistsAtGivenPath(pluginSourceFilepath) {
			cmd.ui.Failed(T("File not found locally, make sure the file exists at given path ..."))
		}

	}

	cmd.ui.Say(fmt.Sprintf(T("Installing plugin {{.PluginPath}}...", map[string]interface{}{"PluginPath": pluginSourceFilepath})))

	_, pluginExecutableName := filepath.Split(pluginSourceFilepath)

	pluginDestinationFilepath := filepath.Join(cmd.pluginConfig.GetPluginPath(), pluginExecutableName)

	cmd.ensurePluginBinaryWithSameFileNameDoesNotAlreadyExist(pluginDestinationFilepath, pluginExecutableName)

	pluginMetadata := cmd.runBinaryAndObtainPluginMetadata(pluginSourceFilepath)

	cmd.ensurePluginIsSafeForInstallation(pluginMetadata, pluginDestinationFilepath, pluginSourceFilepath)

	cmd.installPlugin(pluginMetadata, pluginDestinationFilepath, pluginSourceFilepath)

	cmd.ui.Ok()
	cmd.ui.Say(fmt.Sprintf(T("Plugin {{.PluginName}} v{{.Version}} successfully installed.", map[string]interface{}{"PluginName": pluginMetadata.Name, "Version": fmt.Sprintf("%d.%d.%d", pluginMetadata.Version.Major, pluginMetadata.Version.Minor, pluginMetadata.Version.Build)})))
}

func (cmd *PluginInstall) ensurePluginBinaryWithSameFileNameDoesNotAlreadyExist(pluginDestinationFilepath, pluginExecutableName string) {
	_, err := os.Stat(pluginDestinationFilepath)
	if err == nil || os.IsExist(err) {
		cmd.ui.Failed(fmt.Sprintf(T("The file {{.PluginExecutableName}} already exists under the plugin directory.\n",
			map[string]interface{}{
				"PluginExecutableName": pluginExecutableName,
			})))
	} else if !os.IsNotExist(err) {
		cmd.ui.Failed(fmt.Sprintf(T("Unexpected error has occurred:\n{{.Error}}", map[string]interface{}{"Error": err.Error()})))
	}
}

func (cmd *PluginInstall) ensurePluginIsSafeForInstallation(pluginMetadata *plugin.PluginMetadata, pluginDestinationFilepath string, pluginSourceFilepath string) {
	plugins := cmd.pluginConfig.Plugins()
	if pluginMetadata.Name == "" {
		cmd.ui.Failed(fmt.Sprintf(T("Unable to obtain plugin name for executable {{.Executable}}", map[string]interface{}{"Executable": pluginSourceFilepath})))
	}

	if _, ok := plugins[pluginMetadata.Name]; ok {
		cmd.ui.Failed(fmt.Sprintf(T("Plugin name {{.PluginName}} is already taken", map[string]interface{}{"PluginName": pluginMetadata.Name})))
	}

	if pluginMetadata.Commands == nil {
		cmd.ui.Failed(fmt.Sprintf(T("Error getting command list from plugin {{.FilePath}}", map[string]interface{}{"FilePath": pluginSourceFilepath})))
	}

	shortNames := cmd.getShortNames()

	for _, pluginCmd := range pluginMetadata.Commands {
		//check for command conflicting core commands/alias
		if _, exists := cmd.coreCmds[pluginCmd.Name]; exists || shortNames[pluginCmd.Name] || pluginCmd.Name == "help" {
			cmd.ui.Failed(fmt.Sprintf(T("Command `{{.Command}}` in the plugin being installed is a native CF command/alias.  Rename the `{{.Command}}` command in the plugin being installed in order to enable its installation and use.",
				map[string]interface{}{"Command": pluginCmd.Name})))
		}

		//check for alias conflicting core command/alias
		if _, exists := cmd.coreCmds[pluginCmd.Alias]; exists || shortNames[pluginCmd.Alias] || pluginCmd.Alias == "help" {
			cmd.ui.Failed(fmt.Sprintf(T("Alias `{{.Command}}` in the plugin being installed is a native CF command/alias.  Rename the `{{.Command}}` command in the plugin being installed in order to enable its installation and use.",
				map[string]interface{}{"Command": pluginCmd.Alias})))
		}

		for installedPluginName, installedPlugin := range plugins {
			for _, installedPluginCmd := range installedPlugin.Commands {

				//check for command conflicting other plugin commands/alias
				if installedPluginCmd.Name == pluginCmd.Name || installedPluginCmd.Alias == pluginCmd.Name {
					cmd.ui.Failed(fmt.Sprintf(T("Command `{{.Command}}` is a command/alias in plugin '{{.PluginName}}'.  You could try uninstalling plugin '{{.PluginName}}' and then install this plugin in order to invoke the `{{.Command}}` command.  However, you should first fully understand the impact of uninstalling the existing '{{.PluginName}}' plugin.",
						map[string]interface{}{"Command": pluginCmd.Name, "PluginName": installedPluginName})))
				}

				//check for alias conflicting other plugin commands/alias
				if pluginCmd.Alias != "" && (installedPluginCmd.Name == pluginCmd.Alias || installedPluginCmd.Alias == pluginCmd.Alias) {
					cmd.ui.Failed(fmt.Sprintf(T("Alias `{{.Command}}` is a command/alias in plugin '{{.PluginName}}'.  You could try uninstalling plugin '{{.PluginName}}' and then install this plugin in order to invoke the `{{.Command}}` command.  However, you should first fully understand the impact of uninstalling the existing '{{.PluginName}}' plugin.",
						map[string]interface{}{"Command": pluginCmd.Alias, "PluginName": installedPluginName})))
				}
			}
		}
	}

}

func (cmd *PluginInstall) installPlugin(pluginMetadata *plugin.PluginMetadata, pluginDestinationFilepath, pluginSourceFilepath string) {
	err := fileutils.CopyFile(pluginDestinationFilepath, pluginSourceFilepath)
	if err != nil {
		cmd.ui.Failed(fmt.Sprintf(T("Could not copy plugin binary: \n{{.Error}}", map[string]interface{}{"Error": err.Error()})))
	}

	configMetadata := plugin_config.PluginMetadata{
		Location: pluginDestinationFilepath,
		Version:  pluginMetadata.Version,
		Commands: pluginMetadata.Commands,
	}

	cmd.pluginConfig.SetPlugin(pluginMetadata.Name, configMetadata)
}

func (cmd *PluginInstall) runBinaryAndObtainPluginMetadata(pluginSourceFilepath string) *plugin.PluginMetadata {
	err := cmd.rpcService.Start()
	if err != nil {
		cmd.ui.Failed(err.Error())
	}
	defer cmd.rpcService.Stop()

	cmd.runPluginBinary(pluginSourceFilepath, cmd.rpcService.Port())

	return cmd.rpcService.RpcCmd.PluginMetadata
}

func (cmd *PluginInstall) ensureCandidatePluginBinaryExistsAtGivenPath(pluginSourceFilepath string) bool {
	_, err := os.Stat(pluginSourceFilepath)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

func (cmd *PluginInstall) tryDownloadPluginBinaryfromGivenPath(pluginSourceFilepath string, downloader fileutils.Downloader) string {
	size, filename, err := downloader.DownloadFile(pluginSourceFilepath)

	if err != nil {
		cmd.ui.Failed(fmt.Sprintf(T("Download attempt failed: {{.Error}}\n\nUnable to install, plugin is not available from the given url.", map[string]interface{}{"Error": err.Error()})))
	}

	cmd.ui.Say(fmt.Sprintf("%d "+T("bytes downloaded")+"...", size))

	executablePath := filepath.Join(downloader.SavePath(), filename)
	os.Chmod(executablePath, 0700)

	return executablePath
}

func (cmd *PluginInstall) getShortNames() map[string]bool {
	shortNames := make(map[string]bool)
	for _, singleCmd := range cmd.coreCmds {
		metaData := singleCmd.Metadata()
		if metaData.ShortName != "" {
			shortNames[metaData.ShortName] = true
		}
	}
	return shortNames
}

func (cmd *PluginInstall) runPluginBinary(location string, servicePort string) {
	pluginInvocation := exec.Command(location, servicePort, "SendMetadata")

	err := pluginInvocation.Run()
	if err != nil {
		cmd.ui.Failed(err.Error())
	}
}

func (cmd *PluginInstall) getRepoFromConfig(repoName string) (models.PluginRepo, error) {
	targetRepo := strings.ToLower(repoName)
	list := cmd.config.PluginRepos()

	for i, repo := range list {
		if strings.ToLower(repo.Name) == targetRepo {
			return list[i], nil
		}
	}

	return models.PluginRepo{}, errors.New(repoName + T(" not found"))
}

func (cmd *PluginInstall) downloadBinary(plugin clipr.Plugin, downloader fileutils.Downloader) (string, string) {
	arch := runtime.GOARCH

	switch runtime.GOOS {
	case "darwin":
		return cmd.tryDownloadPluginBinaryfromGivenPath(cmd.getBinaryUrl(plugin, "osx"), downloader), cmd.getBinaryChecksum(plugin, "osx")
	case "linux":
		if arch == "386" {
			return cmd.tryDownloadPluginBinaryfromGivenPath(cmd.getBinaryUrl(plugin, "linux32"), downloader), cmd.getBinaryChecksum(plugin, "linux32")
		} else {
			return cmd.tryDownloadPluginBinaryfromGivenPath(cmd.getBinaryUrl(plugin, "linux64"), downloader), cmd.getBinaryChecksum(plugin, "linux64")
		}
	case "windows":
		if arch == "386" {
			return cmd.tryDownloadPluginBinaryfromGivenPath(cmd.getBinaryUrl(plugin, "win32"), downloader), cmd.getBinaryChecksum(plugin, "win32")
		} else {
			return cmd.tryDownloadPluginBinaryfromGivenPath(cmd.getBinaryUrl(plugin, "win64"), downloader), cmd.getBinaryChecksum(plugin, "win64")
		}
	default:
		cmd.binaryNotAvailable()
	}
	return "", ""
}

func (cmd *PluginInstall) getBinaryUrl(plugin clipr.Plugin, os string) string {
	for _, binary := range plugin.Binaries {
		if binary.Platform == os {
			return binary.Url
		}
	}
	cmd.binaryNotAvailable()
	return ""
}

func (cmd *PluginInstall) getBinaryChecksum(plugin clipr.Plugin, os string) string {
	for _, binary := range plugin.Binaries {
		if binary.Platform == os {
			return binary.Checksum
		}
	}
	return ""
}

func (cmd *PluginInstall) binaryNotAvailable() {
	cmd.ui.Failed(T("Plugin requested has no binary available for your OS: ") + runtime.GOOS + ", " + runtime.GOARCH)
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
