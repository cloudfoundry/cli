package commands

import (
	"cf"
	"cf/api"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type Scale struct {
	ui      terminal.UI
	starter ApplicationStarter
	stopper ApplicationStopper
	appReq  requirements.ApplicationRequirement
	appRepo api.ApplicationRepository
}

func NewScale(ui terminal.UI, starter ApplicationStarter, stopper ApplicationStopper, appRepo api.ApplicationRepository) (cmd *Scale) {
	cmd = new(Scale)
	cmd.ui = ui
	cmd.starter = starter
	cmd.stopper = stopper
	cmd.appRepo = appRepo
	return
}

func (cmd *Scale) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {

	if len(c.Args()) < 1 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "scale")
		return
	}

	cmd.appReq = reqFactory.NewApplicationRequirement(c.Args()[0])

	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
		reqFactory.NewTargetedSpaceRequirement(),
		cmd.appReq,
	}
	return
}

func (cmd *Scale) Run(c *cli.Context) {
	currentApp := cmd.appReq.GetApplication()
	cmd.ui.Say("Scaling app %s...", terminal.EntityNameColor(currentApp.Name))

	changedApp := cf.Application{
		Guid: currentApp.Guid,
	}

	diskQuota, err := extractMegaBytes(c.String("d"))
	if err != nil {
		cmd.ui.Say("Invalid value for disk quota.")
		cmd.ui.FailWithUsage(c, "scale")
		return
	}
	changedApp.DiskQuota = diskQuota

	memory, err := extractMegaBytes(c.String("m"))
	if err != nil {
		cmd.ui.Say("Invalid value for memory.")
		cmd.ui.FailWithUsage(c, "scale")
		return
	}
	changedApp.Memory = memory

	changedApp.Instances = c.Int("i")

	cmd.appRepo.Scale(changedApp)
	cmd.stopper.ApplicationStop(currentApp)
	cmd.starter.ApplicationStart(currentApp)
}

func extractMegaBytes(arg string) (megaBytes int, err error) {
	if  arg != "" {
		var byteSize int
		byteSize, err = bytesFromString(arg)
		if err != nil {
			return
		}
		megaBytes = byteSize / MEGABYTE
	}

	return
}
