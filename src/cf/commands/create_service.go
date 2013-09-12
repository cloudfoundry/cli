package commands

import (
	"cf"
	"cf/api"
	"cf/requirements"
	term "cf/terminal"
	"errors"
	"fmt"
	"github.com/codegangsta/cli"
	"strings"
)

type CreateService struct {
	ui          term.UI
	serviceRepo api.ServiceRepository
}

func NewCreateService(ui term.UI, sR api.ServiceRepository) (cmd CreateService) {
	cmd.ui = ui
	cmd.serviceRepo = sR
	return
}

func (cmd CreateService) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
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
		paramsMap[param] = cmd.ui.Ask("%s%s", param, term.PromptColor(">"))
	}

	cmd.ui.Say("Creating service...")
	err := cmd.serviceRepo.CreateUserProvidedServiceInstance(name, paramsMap)
	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}

	cmd.ui.Ok()
	return
}

func (cmd CreateService) createService(name string, offeringName string, planName string) {
	offerings, apiErr := cmd.serviceRepo.GetServiceOfferings()
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	offering, err := findOffering(offerings, offeringName)
	if err != nil {
		cmd.ui.Failed("Offering not found")
		return
	}

	plan, err := findPlan(offering.Plans, planName)
	if err != nil {
		cmd.ui.Failed("Plan not found")
		return
	}

	cmd.ui.Say("Creating service %s", term.EntityNameColor(name))
	apiErr = cmd.serviceRepo.CreateServiceInstance(name, plan)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
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
