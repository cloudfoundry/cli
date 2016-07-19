package application

import (
	"errors"
	"fmt"

	"code.cloudfoundry.org/cli/cf/api/appfiles"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
)

type Files struct {
	ui           terminal.UI
	config       coreconfig.Reader
	appFilesRepo appfiles.Repository
	appReq       requirements.DEAApplicationRequirement
}

func init() {
	commandregistry.Register(&Files{})
}

func (cmd *Files) MetaData() commandregistry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["i"] = &flags.IntFlag{ShortName: "i", Usage: T("Instance")}

	return commandregistry.CommandMetadata{
		Name:        "files",
		ShortName:   "f",
		Description: T("Print out a list of files in a directory or the contents of a specific file of an app running on the DEA backend"),
		Usage: []string{
			T(`CF_NAME files APP_NAME [PATH] [-i INSTANCE]
			
TIP:
  To list and inspect files of an app running on the Diego backend, use 'CF_NAME ssh'`),
		},
		Flags: fs,
	}
}

func (cmd *Files) Requirements(requirementsFactory requirements.Factory, c flags.FlagContext) ([]requirements.Requirement, error) {
	if len(c.Args()) != 1 && len(c.Args()) != 2 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + commandregistry.Commands.CommandUsage("files"))
		return nil, fmt.Errorf("Incorrect usage: %d arguments of %d required", len(c.Args()), 1)
	}

	cmd.appReq = requirementsFactory.NewDEAApplicationRequirement(c.Args()[0])

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
		cmd.appReq,
	}

	return reqs, nil
}

func (cmd *Files) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.appFilesRepo = deps.RepoLocator.GetAppFilesRepository()
	return cmd
}

func (cmd *Files) Execute(c flags.FlagContext) error {
	app := cmd.appReq.GetApplication()

	var instance int
	if c.IsSet("i") {
		instance = c.Int("i")
		if instance < 0 {
			return errors.New(T("Invalid instance: {{.Instance}}\nInstance must be a positive integer",
				map[string]interface{}{
					"Instance": instance,
				}))
		}
		if instance >= app.InstanceCount {
			return errors.New(T("Invalid instance: {{.Instance}}\nInstance must be less than {{.InstanceCount}}",
				map[string]interface{}{
					"Instance":      instance,
					"InstanceCount": app.InstanceCount,
				}))
		}
	}

	cmd.ui.Say(T("Getting files for app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...",
		map[string]interface{}{
			"AppName":   terminal.EntityNameColor(app.Name),
			"OrgName":   terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"SpaceName": terminal.EntityNameColor(cmd.config.SpaceFields().Name),
			"Username":  terminal.EntityNameColor(cmd.config.Username())}))

	path := "/"
	if len(c.Args()) > 1 {
		path = c.Args()[1]
	}

	list, err := cmd.appFilesRepo.ListFiles(app.GUID, instance, path)
	if err != nil {
		return err
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	if list == "" {
		cmd.ui.Say("Empty file or folder")
	} else {
		cmd.ui.Say("%s", list)
	}
	return nil
}
