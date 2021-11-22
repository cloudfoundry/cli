package shared

import (
	"code.cloudfoundry.org/cli/command"
)

type PackageDisplayer struct {
	ui     command.UI
	config command.Config
}

func NewPackageDisplayer(ui command.UI, config command.Config) PackageDisplayer {
	return PackageDisplayer{
		ui:     ui,
		config: config,
	}
}

func (display PackageDisplayer) DisplaySetupMessage(appName, currentUser string, isDockerImage bool) error {
	var flavorTextTemplate string
	if isDockerImage {
		flavorTextTemplate = "Creating docker package for app {{.AppName}} in org {{.CurrentOrg}} / space {{.CurrentSpace}} as {{.CurrentUser}}..."
	} else {
		flavorTextTemplate = "Creating and uploading bits package for app {{.AppName}} in org {{.CurrentOrg}} / space {{.CurrentSpace}} as {{.CurrentUser}}..."
	}

	display.ui.DisplayTextWithFlavor(flavorTextTemplate, map[string]interface{}{
		"AppName":      appName,
		"CurrentSpace": display.config.TargetedSpace().Name,
		"CurrentOrg":   display.config.TargetedOrganization().Name,
		"CurrentUser":  currentUser,
	})

	return nil
}
