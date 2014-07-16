package application

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/flag_helpers"
	"github.com/cloudfoundry/cli/cf/formatters"
	. "github.com/cloudfoundry/cli/cf/i18n"
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

func (cmd *Scale) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "scale",
		Description: T("Change or view the instance count, disk space limit, and memory limit for an app"),
		Usage:       T("CF_NAME scale APP [-i INSTANCES] [-k DISK] [-m MEMORY] [-f]"),
		Flags: []cli.Flag{
			flag_helpers.NewIntFlag("i", T("Number of instances")),
			flag_helpers.NewStringFlag("k", T("Disk limit (e.g. 256M, 1024M, 1G)")),
			flag_helpers.NewStringFlag("m", T("Memory limit (e.g. 256M, 1024M, 1G)")),
			cli.BoolFlag{Name: "f", Usage: T("Force restart of app without prompt")},
		},
	}
}

func (cmd *Scale) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		cmd.ui.FailWithUsage(c)
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
		cmd.ui.Say(T("Showing current scale of app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUser}}...",
			map[string]interface{}{
				"AppName":     terminal.EntityNameColor(currentApp.Name),
				"OrgName":     terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
				"SpaceName":   terminal.EntityNameColor(cmd.config.SpaceFields().Name),
				"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
			}))
		cmd.ui.Ok()
		cmd.ui.Say("")

		cmd.ui.Say("%s %s", terminal.HeaderColor(T("memory:")), formatters.ByteSize(currentApp.Memory*bytesInAMegabyte))
		cmd.ui.Say("%s %s", terminal.HeaderColor(T("disk:")), formatters.ByteSize(currentApp.DiskQuota*bytesInAMegabyte))
		cmd.ui.Say("%s %d", terminal.HeaderColor(T("instances:")), currentApp.InstanceCount)

		return
	}

	params := models.AppParams{}
	shouldRestart := false

	if c.String("m") != "" {
		memory, err := formatters.ToMegabytes(c.String("m"))
		if err != nil {
			cmd.ui.Failed(T("Invalid memory limit: {{.Memory}}\n{{.ErrorDescription}}",
				map[string]interface{}{
					"Memory":           c.String("m"),
					"ErrorDescription": err,
				}))
		}
		params.Memory = &memory
		shouldRestart = true
	}

	if c.String("k") != "" {
		diskQuota, err := formatters.ToMegabytes(c.String("k"))
		if err != nil {
			cmd.ui.Failed(T("Invalid disk quota: {{.DiskQuota}}\n{{.ErrorDescription}}",
				map[string]interface{}{
					"DiskQuota":        c.String("k"),
					"ErrorDescription": err,
				}))
		}
		params.DiskQuota = &diskQuota
		shouldRestart = true
	}

	if c.IsSet("i") {
		instances := c.Int("i")
		if instances > 0 {
			params.InstanceCount = &instances
		} else {
			cmd.ui.Failed(T("Invalid instance count: {{.InstanceCount}}\nInstance count must be a positive integer",
				map[string]interface{}{
					"InstanceCount": instances,
				}))
		}
	}

	if shouldRestart && !cmd.confirmRestart(c, currentApp.Name) {
		return
	}

	cmd.ui.Say(T("Scaling app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"AppName":     terminal.EntityNameColor(currentApp.Name),
			"OrgName":     terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"SpaceName":   terminal.EntityNameColor(cmd.config.SpaceFields().Name),
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
		}))

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
		result := cmd.ui.Confirm(T("This will cause the app to restart. Are you sure you want to scale {{.AppName}}?",
			map[string]interface{}{"AppName": terminal.EntityNameColor(appName)}))
		cmd.ui.Say("")
		return result
	}
}

func anyFlagsSet(context *cli.Context) bool {
	return context.IsSet("m") || context.IsSet("k") || context.IsSet("i")
}
