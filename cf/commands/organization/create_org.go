package organization

import (
	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/api/organizations"
	"github.com/cloudfoundry/cli/cf/api/quotas"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/flag_helpers"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type CreateOrg struct {
	ui        terminal.UI
	config    configuration.Reader
	orgRepo   organizations.OrganizationRepository
	quotaRepo quotas.QuotaRepository
}

func NewCreateOrg(ui terminal.UI, config configuration.Reader, orgRepo organizations.OrganizationRepository, quotaRepo quotas.QuotaRepository) (cmd CreateOrg) {
	cmd.ui = ui
	cmd.config = config
	cmd.orgRepo = orgRepo
	cmd.quotaRepo = quotaRepo
	return
}

func (cmd CreateOrg) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "create-org",
		ShortName:   "co",
		Description: T("Create an org"),
		Usage:       T("CF_NAME create-org ORG"),
		Flags: []cli.Flag{
			flag_helpers.NewStringFlag("q", T("Quota to assign to the newly created org (excluding this option results in assignment of default quota)")),
		},
	}
}

func (cmd CreateOrg) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		cmd.ui.FailWithUsage(c)
	}

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}
	return
}

func (cmd CreateOrg) Run(c *cli.Context) {
	name := c.Args()[0]
	cmd.ui.Say(T("Creating org {{.OrgName}} as {{.Username}}...",
		map[string]interface{}{
			"OrgName":  terminal.EntityNameColor(name),
			"Username": terminal.EntityNameColor(cmd.config.Username())}))

	org := models.Organization{OrganizationFields: models.OrganizationFields{Name: name}}

	quotaName := c.String("q")
	if quotaName != "" {
		quota, err := cmd.quotaRepo.FindByName(quotaName)
		if err != nil {
			cmd.ui.Failed(err.Error())
		}

		org.QuotaDefinition.Guid = quota.Guid
	}

	err := cmd.orgRepo.Create(org)
	if err != nil {
		if apiErr, ok := err.(errors.HttpError); ok && apiErr.ErrorCode() == errors.ORG_EXISTS {
			cmd.ui.Ok()
			cmd.ui.Warn(T("Org {{.OrgName}} already exists",
				map[string]interface{}{"OrgName": name}))
			return
		} else {
			cmd.ui.Failed(err.Error())
		}
	}

	cmd.ui.Ok()
	cmd.ui.Say(T("\nTIP: Use '{{.Command}}' to target new org",
		map[string]interface{}{"Command": terminal.CommandColor(cf.Name() + " target -o " + name)}))
}
