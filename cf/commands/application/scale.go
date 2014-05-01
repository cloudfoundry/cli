package application

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/flag_helpers"
	"github.com/cloudfoundry/cli/cf/formatters"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type Scale struct {
	ui        terminal.UI
	config    configuration.Reader
	restarter ApplicationRestarter
	appReq    requirements.ApplicationRequirement
	appRepo   api.ApplicationRepository
}

func NewScale(ui terminal.UI, config configuration.Reader, restarter ApplicationRestarter, appRepo api.ApplicationRepository) (cmd *Scale) {
	cmd = new(Scale)
	cmd.ui = ui
	cmd.config = config
	cmd.restarter = restarter
	cmd.appRepo = appRepo
	return
}

func (command *Scale) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "scale",
		Description: "Change or view the instance count, disk space limit, and memory limit for an app",
		Usage:       "CF_NAME scale APP [-i INSTANCES] [-k DISK] [-m MEMORY] [-f]",
		Flags: []cli.Flag{
			flag_helpers.NewIntFlag("i", "Number of instances"),
			flag_helpers.NewStringFlag("k", "Disk limit (e.g. 256M, 1024M, 1G)"),
			flag_helpers.NewStringFlag("m", "Memory limit (e.g. 256M, 1024M, 1G)"),
			cli.BoolFlag{Name: "f", Usage: "Force restart of app without prompt"},
		},
	}
}

func (cmd *Scale) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "scale")
		return
	}

	cmd.appReq = requirementsFactory.NewApplicationRequirement(c.Args()[0])

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedSpaceRequirement(),
		cmd.appReq,
	}
	return
}

var bytesInAMegabyte uint64 = 1024 * 1024

func (cmd *Scale) Run(c *cli.Context) {
	currentApp := cmd.appReq.GetApplication()
	if !anyFlagsSet(c) {
		cmd.ui.Say("Showing current scale of app %s in org %s / space %s as %s...",
			terminal.EntityNameColor(currentApp.Name),
			terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			terminal.EntityNameColor(cmd.config.SpaceFields().Name),
			terminal.EntityNameColor(cmd.config.Username()),
		)
		cmd.ui.Ok()
		cmd.ui.Say("")

		cmd.ui.Say("%s %s", terminal.HeaderColor("memory:"), formatters.ByteSize(currentApp.Memory*bytesInAMegabyte))
		cmd.ui.Say("%s %s", terminal.HeaderColor("disk:"), formatters.ByteSize(currentApp.DiskQuota*bytesInAMegabyte))
		cmd.ui.Say("%s %d", terminal.HeaderColor("instances:"), currentApp.InstanceCount)

		return
	}

	params := models.AppParams{}
	shouldRestart := false

	if c.String("m") != "" {
		memory, err := formatters.ToMegabytes(c.String("m"))
		if err != nil {
			cmd.ui.Failed("Invalid memory limit: %s\n%s", c.String("m"), err)
		}
		params.Memory = &memory
		shouldRestart = true
	}

	if c.String("k") != "" {
		diskQuota, err := formatters.ToMegabytes(c.String("k"))
		if err != nil {
			cmd.ui.Failed("Invalid disk quota: %s\n%s", c.String("k"), err)
		}
		params.DiskQuota = &diskQuota
		shouldRestart = true
	}

	if c.IsSet("i") {
		instances := c.Int("i")
		if instances > 0 {
			params.InstanceCount = &instances
		} else {
			cmd.ui.Failed("Invalid instance count: %d\nInstance count must be a positive integer", instances)
		}
	}

	if shouldRestart && !cmd.confirmRestart(c, currentApp.Name) {
		return
	}

	cmd.ui.Say("Scaling app %s in org %s / space %s as %s...",
		terminal.EntityNameColor(currentApp.Name),
		terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
		terminal.EntityNameColor(cmd.config.SpaceFields().Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	updatedApp, apiErr := cmd.appRepo.Update(currentApp.Guid, params)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()

	if shouldRestart {
		cmd.restarter.ApplicationRestart(updatedApp)
	}
}

func (cmd *Scale) confirmRestart(context *cli.Context, appName string) bool {
	if context.Bool("f") {
		return true
	} else {
		result := cmd.ui.Confirm("This will cause the app to restart. Are you sure you want to scale %s?", terminal.EntityNameColor(appName))
		cmd.ui.Say("")
		return result
	}
}

func anyFlagsSet(context *cli.Context) bool {
	return context.IsSet("m") || context.IsSet("k") || context.IsSet("i")
}
