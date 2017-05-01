package plugin

import (
	"fmt"
	"sort"

	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/pluginconfig"
	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cli/util"
	"code.cloudfoundry.org/cli/util/sorting"
)

type Plugins struct {
	ui     terminal.UI
	config pluginconfig.PluginConfiguration
}

func init() {
	commandregistry.Register(&Plugins{})
}

func (cmd *Plugins) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["checksum"] = &flags.BoolFlag{Name: "checksum", Usage: T("Compute and show the sha1 value of the plugin binary file")}

	return commandregistry.CommandMetadata{
		Name:        "plugins",
		Description: T("List all available plugin commands"),
		Usage: []string{
			T("CF_NAME plugins"),
		},
		Flags: fs,
	}
}

func (cmd *Plugins) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	usageReq := requirements.NewUsageRequirement(commandregistry.CLICommandUsagePresenter(cmd),
		T("No argument required"),
		func() bool {
			return len(fc.Args()) != 0
		},
	)

	reqs := []requirements.Requirement{
		usageReq,
	}
	return reqs, nil
}

func (cmd *Plugins) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.PluginConfig
	return cmd
}

func (cmd *Plugins) Execute(c flags.FlagContext) error {
	var version string

	cmd.ui.Say(T("Listing Installed Plugins..."))

	plugins := cmd.config.Plugins()

	var table *terminal.UITable
	if c.Bool("checksum") {
		cmd.ui.Say(T("Computing sha1 for installed plugins, this may take a while ..."))
		table = cmd.ui.Table([]string{T("Plugin Name"), T("Version"), T("Command Name"), "sha1", T("Command Help")})
	} else {
		table = cmd.ui.Table([]string{T("Plugin Name"), T("Version"), T("Command Name"), T("Command Help")})
	}

	sortedPluginNames := make([]string, 0, len(plugins))
	for k := range plugins {
		sortedPluginNames = append(sortedPluginNames, k)
	}
	sort.Slice(sortedPluginNames, sorting.SortAlphabeticFunc(sortedPluginNames))

	for _, pluginName := range sortedPluginNames {
		metadata := plugins[pluginName]
		if metadata.Version.Major == 0 && metadata.Version.Minor == 0 && metadata.Version.Build == 0 {
			version = "N/A"
		} else {
			version = fmt.Sprintf("%d.%d.%d", metadata.Version.Major, metadata.Version.Minor, metadata.Version.Build)
		}

		for _, command := range metadata.Commands {
			args := []string{pluginName, version}

			if command.Alias != "" {
				args = append(args, command.Name+", "+command.Alias)
			} else {
				args = append(args, command.Name)
			}

			if c.Bool("checksum") {
				checksum := util.NewSha1Checksum(metadata.Location)
				sha1, err := checksum.ComputeFileSha1()
				if err != nil {
					args = append(args, "n/a")
				} else {
					args = append(args, fmt.Sprintf("%x", sha1))
				}
			}

			args = append(args, command.HelpText)
			table.Add(args...)
		}
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	err := table.Print()
	if err != nil {
		return err
	}
	return nil
}
