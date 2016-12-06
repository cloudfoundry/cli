package common

import "code.cloudfoundry.org/cli/command"

type VersionCommand struct {
	usage  interface{} `usage:"CF_NAME version\n\n   'cf -v' and 'cf --version' are also accepted."`
	UI     command.UI
	Config command.Config
}

func (cmd *VersionCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	return nil
}

func (cmd VersionCommand) Execute(args []string) error {
	cmd.UI.DisplayText("{{.BinaryName}} version {{.BinaryVersion}}-{{.BinaryBuildDate}}",
		map[string]interface{}{
			"BinaryName":      cmd.Config.BinaryName(),
			"BinaryVersion":   cmd.Config.BinaryVersion(),
			"BinaryBuildDate": cmd.Config.BinaryBuildDate(),
		})
	return nil
}
