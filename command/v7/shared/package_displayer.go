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

func (display PackageDisplayer) DisplaySetupMessage(appName string, isDockerImage bool) error {
	var flavorTextTemplate string
	if isDockerImage {
		flavorTextTemplate = "Creating docker package for app {{.AppName}} in org {{.CurrentOrg}} / space {{.CurrentSpace}} as {{.CurrentUser}}..."
	} else {
		flavorTextTemplate = "Uploading and creating bits package for app {{.AppName}} in org {{.CurrentOrg}} / space {{.CurrentSpace}} as {{.CurrentUser}}..."
	}

	currentUser, err := display.config.CurrentUser()
	if err != nil {
		return err
	}
	display.ui.DisplayTextWithFlavor(flavorTextTemplate, map[string]interface{}{
		"AppName":      appName,
		"CurrentSpace": display.config.TargetedSpace().Name,
		"CurrentOrg":   display.config.TargetedOrganization().Name,
		"CurrentUser":  currentUser.Name,
	})

	return nil
}
