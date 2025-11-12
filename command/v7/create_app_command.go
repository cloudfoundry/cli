package v7

import (
	"errors"
	"fmt"

	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/v8/command"
	"code.cloudfoundry.org/cli/v8/command/flag"
	"code.cloudfoundry.org/cli/v8/resources"
)

type CreateAppCommand struct {
	BaseCommand

	RequiredArgs    flag.AppName `positional-args:"yes"`
	AppType         flag.AppType `long:"app-type" choice:"buildpack" choice:"docker" choice:"cnb" description:"App lifecycle type to stage and run the app" default:"buildpack"`
	Buildpacks      []string     `long:"buildpack" short:"b" description:"Custom buildpack by name (e.g. my-buildpack), Docker image (e.g. docker://registry/image:tag), Git URL (e.g. 'https://github.com/cloudfoundry/java-buildpack.git') or Git URL with a branch or tag (e.g. 'https://github.com/cloudfoundry/java-buildpack.git#v3.3.0' for 'v3.3.0' tag). To use built-in buildpacks only, specify 'default' or 'null'"`
	usage           interface{}  `usage:"CF_NAME create-app APP_NAME [--app-type (buildpack | docker | cnb)]"`
	relatedCommands interface{}  `related_commands:"app, apps, push"`
}

func (cmd CreateAppCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Creating app {{.AppName}} in org {{.CurrentOrg}} / space {{.CurrentSpace}} as {{.CurrentUser}}...", map[string]interface{}{
		"AppName":      cmd.RequiredArgs.AppName,
		"CurrentSpace": cmd.Config.TargetedSpace().Name,
		"CurrentOrg":   cmd.Config.TargetedOrganization().Name,
		"CurrentUser":  user.Name,
	})

	cmd.UI.DisplayText(fmt.Sprintf("Using app type %q", constant.AppLifecycleType(cmd.AppType)))

	app := resources.Application{
		Name:                cmd.RequiredArgs.AppName,
		LifecycleType:       constant.AppLifecycleType(cmd.AppType),
		LifecycleBuildpacks: cmd.Buildpacks,
	}

	if constant.AppLifecycleType(cmd.AppType) == constant.AppLifecycleTypeCNB {
		err := command.MinimumCCAPIVersionCheck(cmd.Config.APIVersion(), ccversion.MinVersionCNB)
		if err != nil {
			return err
		}

		if len(cmd.Buildpacks) == 0 {
			return errors.New("buildpack(s) must be provided when using --app-type cnb")
		}

		creds, err := cmd.Config.CNBCredentials()
		if err != nil {
			return err
		}

		app.Credentials = creds
	}

	_, warnings, err := cmd.Actor.CreateApplicationInSpace(
		app,
		cmd.Config.TargetedSpace().GUID,
	)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		switch err.(type) {
		case ccerror.NameNotUniqueInSpaceError:
			cmd.UI.DisplayText(err.Error())
		default:
			return err
		}
	}

	cmd.UI.DisplayOK()

	return nil
}
