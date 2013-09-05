package commands

import (
	"cf"
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	term "cf/terminal"
	"errors"
	"fmt"
	"github.com/codegangsta/cli"
	"strings"
)

type CreateService struct {
	ui          term.UI
	config      *configuration.Configuration
	serviceRepo api.ServiceRepository
}

func NewCreateService(ui term.UI, config *configuration.Configuration, sR api.ServiceRepository) (cmd CreateService) {
	cmd.ui = ui
	cmd.config = config
	cmd.serviceRepo = sR
	return
}

func (cmd CreateService) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []Requirement, err error) {
	offeringName := c.String("offering")
	parameterList := c.String("parameters")

	if offeringName == "user-provided" && parameterList == "" {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "create-service")
		return
	}

	return
}

func (cmd CreateService) Run(c *cli.Context) {
	name := c.String("name")
	offeringName := c.String("offering")

	if offeringName == "user-provided" {
		params := c.String("parameters")
		cmd.createUserProvidedService(name, params)
	} else {
		planName := c.String("plan")
		cmd.createService(name, offeringName, planName)
	}
}

func (cmd CreateService) createUserProvidedService(name string, params string) {
	paramsMap := make(map[string]string)
	params = strings.Trim(params, `"`)

	for _, param := range strings.Split(params, ",") {
		param = strings.Trim(param, " ")
		paramsMap[param] = cmd.ui.Ask("%s%s", param, term.Cyan(">"))
	}

	cmd.ui.Say("Creating service...")
	err := cmd.serviceRepo.CreateUserProvidedServiceInstance(cmd.config, name, paramsMap)
	if err != nil {
		cmd.ui.Failed("Error creating user provided service instance", err)
		return
	}

	cmd.ui.Ok()
	return
}

func (cmd CreateService) createService(name string, offeringName string, planName string) {
	offerings, err := cmd.serviceRepo.GetServiceOfferings(cmd.config)
	if err != nil {
		cmd.ui.Failed("Error fetching offerings", err)
		return
	}

	offering, err := findOffering(offerings, offeringName)
	if err != nil {
		cmd.ui.Failed("Offering not found", nil)
		return
	}

	plan, err := findPlan(offering.Plans, planName)
	if err != nil {
		cmd.ui.Failed("Plan not found", nil)
		return
	}

	cmd.ui.Say("Creating service %s", term.Cyan(name))
	err = cmd.serviceRepo.CreateServiceInstance(cmd.config, name, plan)
	if err != nil {
		cmd.ui.Failed("Error creating plan", err)
		return
	}

	cmd.ui.Ok()
	return
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
