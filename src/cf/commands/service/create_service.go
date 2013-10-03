package service

import (
	"cf"
	"cf/api"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"fmt"
	"github.com/codegangsta/cli"
)

type CreateService struct {
	ui          terminal.UI
	serviceRepo api.ServiceRepository
}

func NewCreateService(ui terminal.UI, sR api.ServiceRepository) (cmd CreateService) {
	cmd.ui = ui
	cmd.serviceRepo = sR
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

	offerings, apiStatus := cmd.serviceRepo.GetServiceOfferings()
	if apiStatus.NotSuccessful() {
		cmd.ui.Failed(apiStatus.Message)
		return
	}

	offering, err := findOffering(offerings, offeringName)
	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}

	plan, err := findPlan(offering.Plans, planName)
	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}

	cmd.ui.Say("Creating service %s", terminal.EntityNameColor(name))

	var alreadyExists bool
	alreadyExists, apiStatus = cmd.serviceRepo.CreateServiceInstance(name, plan)
	if apiStatus.NotSuccessful() {
		cmd.ui.Failed(apiStatus.Message)
		return
	}

	cmd.ui.Ok()

	if alreadyExists {
		cmd.ui.Warn("Service %s already exists", name)
	}
}

func findOffering(offerings []cf.ServiceOffering, name string) (offering cf.ServiceOffering, err error) {
	for _, offering := range offerings {
		if name == offering.Label {
			return offering, nil
		}
	}

	err = errors.New(fmt.Sprintf("Could not find offering with name %s", name))
	return
}

func findPlan(plans []cf.ServicePlan, name string) (plan cf.ServicePlan, err error) {
	for _, plan := range plans {
		if name == plan.Name {
			return plan, nil
		}
	}

	err = errors.New(fmt.Sprintf("Could not find plan with name %s", name))
	return
}
