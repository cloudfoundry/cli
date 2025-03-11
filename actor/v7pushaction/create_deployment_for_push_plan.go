package v7pushaction

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/resources"
)

func (actor Actor) CreateDeploymentForApplication(pushPlan PushPlan, eventStream chan<- *PushEvent, progressBar ProgressBar) (PushPlan, Warnings, error) {
	eventStream <- &PushEvent{Plan: pushPlan, Event: StartingDeployment}

	dep := resources.Deployment{
		Strategy:      pushPlan.Strategy,
		DropletGUID:   pushPlan.DropletGUID,
		Relationships: resources.Relationships{constant.RelationshipTypeApplication: resources.Relationship{GUID: pushPlan.Application.GUID}},
	}

	if pushPlan.MaxInFlight > 0 {
		dep.Options = resources.DeploymentOpts{MaxInFlight: pushPlan.MaxInFlight}
	}

	if len(pushPlan.InstanceSteps) > 0 {
		dep.Options.CanaryDeploymentOptions = &resources.CanaryDeploymentOptions{Steps: []resources.CanaryStep{}}
		for _, w := range pushPlan.InstanceSteps {
			dep.Options.CanaryDeploymentOptions.Steps = append(dep.Options.CanaryDeploymentOptions.Steps, resources.CanaryStep{InstanceWeight: w})
		}
	}

	deploymentGUID, warnings, err := actor.V7Actor.CreateDeployment(dep)

	if err != nil {
		return pushPlan, Warnings(warnings), err
	}

	eventStream <- &PushEvent{Plan: pushPlan, Event: WaitingForDeployment}

	handleInstanceDetails := func(instanceDetails string) {
		eventStream <- &PushEvent{
			Plan:     pushPlan,
			Warnings: Warnings{instanceDetails},
			Event:    InstanceDetails,
		}
	}

	pollWarnings, err := actor.V7Actor.PollStartForDeployment(pushPlan.Application, deploymentGUID, pushPlan.NoWait, handleInstanceDetails)
	warnings = append(warnings, pollWarnings...)

	return pushPlan, Warnings(warnings), err
}
