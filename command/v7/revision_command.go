package v7

import (
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

type RevisionCommand struct {
	RequiredArgs flag.AppName `positional-args:"yes"`

	OptionalArgs flag.RevisionNumber `positional-args:"yes"`

	UI     command.UI
	Config command.Config
}

func (cmd *RevisionCommand) Setup(config command.Config, ui command.UI) error {
	cmd.Config = config
	cmd.UI = ui

	return nil
}

func (cmd *RevisionCommand) Execute(args []string) error {

	cmd.displayTargetTable()

	return nil
}

// displayTargetTable neatly displays target information.
func (cmd *RevisionCommand) displayTargetTable() {
	cmd.UI.DisplayTextWithFlavor("Getting revision " + cmd.OptionalArgs.RevisionNumber + " for app {{.AppName}} ...", map[string]interface{}{
		"AppName": cmd.RequiredArgs.AppName,
	})
	if(cmd.OptionalArgs.RevisionNumber == "v2"){
		cmd.UI.DisplayText("Revision  App    Status    Description           Created By   Created On")
	cmd.UI.DisplayText("2         dora   deployed  new droplet deployed  Cody.A       Tue 22 Jan 11:45:30 PST 2019")

	cmd.UI.DisplayText("")
	cmd.UI.DisplayText("Droplet:")
	cmd.UI.DisplayText("staging_droplet")
	cmd.UI.DisplayText("")
	cmd.UI.DisplayText("Labels:")
	cmd.UI.DisplayText("environment: production")
	cmd.UI.DisplayText("tmc.cloud.vmware.com/creator:myvmware306_vmware-hol.com")
	cmd.UI.DisplayText("type: data")
	cmd.UI.DisplayText("")
	cmd.UI.DisplayText("Annotations:")
	cmd.UI.DisplayText("cody: `app dev`, `415-300-3000`, `data services`")
	cmd.UI.DisplayText("")
	cmd.UI.DisplayText("Environment Variables:")
	cmd.UI.DisplayText("potato: fried")
	cmd.UI.DisplayText("karen: o")
	cmd.UI.DisplayText("ASPNETCORE_ENVIRONMENT: Development")

	}
	if(cmd.OptionalArgs.RevisionNumber == "v1"){
		cmd.UI.DisplayText("Revision  App    Status    Description           Created By   Created On")
	cmd.UI.DisplayText("Revision   Last deployed                   Description")
	cmd.UI.DisplayText("1         dora   deployed  new droplet deployed  Cody.A       Tue 22 Jan 11:45:30 PST 2019")

	cmd.UI.DisplayText("")
	cmd.UI.DisplayText("Droplet:")
	cmd.UI.DisplayText("staging_droplet")
	cmd.UI.DisplayText("")
	cmd.UI.DisplayText("Labels:")
	cmd.UI.DisplayText("environment: production")
	cmd.UI.DisplayText("tmc.cloud.vmware.com/creator:myvmware306_vmware-hol.com")
	cmd.UI.DisplayText("type: data")
	cmd.UI.DisplayText("")
	cmd.UI.DisplayText("Annotations:")
	cmd.UI.DisplayText("cody: `app dev`, `415-300-3000`, `data services`")
	cmd.UI.DisplayText("")
	cmd.UI.DisplayText("Environment Variables:")
	cmd.UI.DisplayText("potato: fried")
	cmd.UI.DisplayText("karen: o")
	cmd.UI.DisplayText("ASPNETCORE_ENVIRONMENT: Development")

	}
	}
