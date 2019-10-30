package v7pushaction

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
)

func ShouldCreateBitsPackage(plan PushPlan) bool {
	return plan.DropletPath == "" && plan.DockerImageCredentials.Path == ""
}

func ShouldCreateDockerPackage(plan PushPlan) bool {
	return plan.DropletPath == "" && plan.DockerImageCredentials.Path != ""
}

func ShouldCreateDroplet(plan PushPlan) bool {
	return plan.DropletPath != ""
}

func ShouldStagePackage(plan PushPlan) bool {
	return !plan.NoStart && plan.DropletPath == ""
}

func ShouldCreateDeployment(plan PushPlan) bool {
	return plan.Strategy == constant.DeploymentStrategyRolling
}

func ShouldStopApplication(plan PushPlan) bool {
	return plan.NoStart && plan.Application.State == constant.ApplicationStarted
}

func ShouldSetDroplet(plan PushPlan) bool {
	return !plan.NoStart || plan.DropletPath != ""
}

func ShouldRestart(plan PushPlan) bool {
	return !plan.NoStart
}

func (actor Actor) GetPrepareApplicationSourceSequence(plan PushPlan) []ChangeApplicationFunc {
	var prepareSourceSequence []ChangeApplicationFunc
	switch {
	case ShouldCreateBitsPackage(plan):
		prepareSourceSequence = append(prepareSourceSequence, actor.CreateBitsPackageForApplication)
	case ShouldCreateDockerPackage(plan):
		prepareSourceSequence = append(prepareSourceSequence, actor.CreateDockerPackageForApplication)
	case ShouldCreateDroplet(plan):
		prepareSourceSequence = append(prepareSourceSequence, actor.CreateDropletForApplication)
	}
	return prepareSourceSequence
}

func (actor Actor) GetRuntimeSequence(plan PushPlan) []ChangeApplicationFunc {

	if plan.TaskTypeApplication {
		return actor.getTaskAppRuntimeSequence(plan)
	} else {
		return actor.getDefaultRuntimeSequence(plan)
	}
}

func (actor Actor) getDefaultRuntimeSequence(plan PushPlan) []ChangeApplicationFunc {
	var runtimeSequence []ChangeApplicationFunc

	if ShouldStagePackage(plan) {
		runtimeSequence = append(runtimeSequence, actor.StagePackageForApplication)
	}

	if ShouldCreateDeployment(plan) {
		runtimeSequence = append(runtimeSequence, actor.CreateDeploymentForApplication)
	} else {
		if ShouldStopApplication(plan) {
			runtimeSequence = append(runtimeSequence, actor.StopApplication)
		}

		if ShouldSetDroplet(plan) {
			runtimeSequence = append(runtimeSequence, actor.SetDropletForApplication)
		}

		if ShouldRestart(plan) {
			runtimeSequence = append(runtimeSequence, actor.RestartApplication)
		}
	}

	return runtimeSequence
}

func (actor Actor) getTaskAppRuntimeSequence(plan PushPlan) []ChangeApplicationFunc {
	var runtimeSequence []ChangeApplicationFunc

	runtimeSequence = append(runtimeSequence, actor.StagePackageForApplication)
	runtimeSequence = append(runtimeSequence, actor.StopApplication)
	runtimeSequence = append(runtimeSequence, actor.SetDropletForApplication)

	return runtimeSequence
}
