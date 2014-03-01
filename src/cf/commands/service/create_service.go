package service

import (
	"cf/api"
	"cf/configuration"
	"cf/models"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"fmt"
	"github.com/codegangsta/cli"
)

type CreateService struct {
	ui          terminal.UI
	config      configuration.Reader
	serviceRepo api.ServiceRepository
}

func NewCreateService(ui terminal.UI, config configuration.Reader, serviceRepo api.ServiceRepository) (cmd CreateService) {
	cmd.ui = ui
	cmd.config = config
	cmd.serviceRepo = serviceRepo
	return
}

func (cmd CreateService) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 3 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "create-service")
		return
	}

	return
}

func (cmd CreateService) Run(c *cli.Context) {
	offeringName := c.Args()[0]
	planName := c.Args()[1]
	name := c.Args()[2]

	cmd.ui.Say("Creating service %s in org %s / space %s as %s...",
		terminal.EntityNameColor(name),
		terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
		terminal.EntityNameColor(cmd.config.SpaceFields().Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	offerings, apiResponse := cmd.serviceRepo.GetServiceOfferingsForSpace(cmd.config.SpaceFields().Guid)
	if apiResponse != nil {
		cmd.ui.Failed(apiResponse.Error())
		return
	}

	offerings, err := findOfferings(offerings, offeringName)
	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}

	plan, err := findPlanFromOfferings(offerings, planName)
	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}

	var identicalAlreadyExists bool
	identicalAlreadyExists, apiResponse = cmd.serviceRepo.CreateServiceInstance(name, plan.Guid)
	if apiResponse != nil {
		cmd.ui.Failed(apiResponse.Error())
		return
	}

	cmd.ui.Ok()

	if identicalAlreadyExists {
		cmd.ui.Warn("Service %s already exists", name)
	}
}

func findOfferings(offerings []models.ServiceOffering, name string) (matchingOfferings models.ServiceOfferings, err error) {
	for _, offering := range offerings {
		if name == offering.Label {
			matchingOfferings = append(matchingOfferings, offering)
		}
	}

	if len(matchingOfferings) == 0 {
		err = errors.New(fmt.Sprintf("Could not find any offerings with name %s", name))
	}
	return
}

func findPlanFromOfferings(offerings models.ServiceOfferings, name string) (plan models.ServicePlanFields, err error) {
	for _, offering := range offerings {
		for _, plan := range offering.Plans {
			if name == plan.Name {
				return plan, nil
			}
		}
	}

	err = errors.New(fmt.Sprintf("Could not find plan with name %s", name))
	return
}
