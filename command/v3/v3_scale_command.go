package v3

import (
	"fmt"
	"os"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v3/shared"
)

//go:generate counterfeiter . V3ScaleActor

type V3ScaleActor interface {
	GetAppScaleSummaryByNameAndSpace(name string, spaceGUID string) (v3action.AppScaleSummary, v3action.Warnings, error)
	UpdateAppScale(name string, spaceGUID string, numInstances int, memUsage int, diskUsage int) (v3action.AppScaleSummary, v3action.Warnings, error)
}

type V3ScaleCommand struct {
	RequiredArgs    flag.AppName `positional-args:"yes"`
	NumInstances    int          `short:"i" description:"Number of instances"`
	DiskLimit       string       `short:"k" description:"Disk limit (e.g. 256M, 1024M, 1G)"`
	MemoryLimit     string       `short:"m" description:"Memory limit (e.g. 256M, 1024M, 1G)"`
	usage           interface{}  `usage:"CF_NAME v3-scale APP_NAME [-i INSTANCES] [-k DISK] [-m MEMORY]"`
	relatedCommands interface{}  `related_commands:"v3-push"`

	UI          command.UI
	Config      command.Config
	Actor       V3ScaleActor
	SharedActor command.SharedActor
}

func (cmd *V3ScaleCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor()

	client, _, err := shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}
	cmd.Actor = v3action.NewActor(client, config)

	return nil
}

func (cmd V3ScaleCommand) Execute(args []string) error {
	cmd.UI.DisplayText(command.ExperimentalWarning)
	cmd.UI.DisplayNewline()

	var (
		//		app v3action.Application
		err error
	)

	err = cmd.SharedActor.CheckTarget(cmd.Config, true, true)
	if err != nil {
		return shared.HandleError(err)
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "User: %s\n", user)

	// app, err = cmd.getApplication()
	//	if _, ok := err.(v3action.ApplicationNotFoundError); ok {
	//		app, err = cmd.createApplication(user.Name)
	//		if err != nil {
	//			return shared.HandleError(err)
	//		}
	//	} else if err != nil {
	//		return shared.HandleError(err)
	//	} else {
	//		app, err = cmd.updateApplication(user.Name, app.GUID)
	//		if err != nil {
	//			return shared.HandleError(err)
	//		}
	//	}
	return nil
}
