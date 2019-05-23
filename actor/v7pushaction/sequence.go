package v7pushaction

import "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"

func ShouldUpdateApplication(plan PushPlan) bool {
	return plan.ApplicationNeedsUpdate
}

func ShouldUpdateRoutes(plan PushPlan) bool {
	return !(plan.SkipRouteCreation || plan.NoRouteFlag)
}

func ShouldScaleWebProcess(plan PushPlan) bool {
	return plan.ScaleWebProcessNeedsUpdate
}

func ShouldUpdateWebProcess(plan PushPlan) bool {
	return plan.UpdateWebProcessNeedsUpdate
}

func ShouldCreateBitsPackage(plan PushPlan) bool {
	return plan.DropletPath == "" && !plan.DockerImageCredentialsNeedsUpdate
}

func ShouldCreateDockerPackage(plan PushPlan) bool {
	return plan.DropletPath == "" && plan.DockerImageCredentialsNeedsUpdate
}

func ShouldCreateDroplet(plan PushPlan) bool {
	return plan.DropletPath != ""
}

func ShouldStagePackage(plan PushPlan) bool {
	return !plan.NoStart && plan.DropletPath == ""
}

func ShouldStopApplication(plan PushPlan) bool {
	return plan.NoStart && plan.Application.State == constant.ApplicationStarted
}

func ShouldSetDroplet(plan PushPlan) bool {
	return !plan.NoStart || plan.DropletPath != ""
}

func (actor Actor) GetUpdateSequence(plan PushPlan) []ChangeApplicationFunc {
	var updateSequence []ChangeApplicationFunc

	if ShouldUpdateApplication(plan) {
		updateSequence = append(updateSequence, actor.UpdateApplication)
	}

	if ShouldUpdateRoutes(plan) {
		updateSequence = append(updateSequence, actor.UpdateRoutesForApplication)
	}

	if ShouldScaleWebProcess(plan) {
		updateSequence = append(updateSequence, actor.ScaleWebProcessForApplication)
	}

	if ShouldUpdateWebProcess(plan) {
		updateSequence = append(updateSequence, actor.UpdateWebProcessForApplication)
	}

	return updateSequence
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
	var runtimeSequence []ChangeApplicationFunc

	if ShouldStagePackage(plan) {
		runtimeSequence = append(runtimeSequence, actor.StagePackageForApplication)
	}

	if ShouldStopApplication(plan) {
		runtimeSequence = append(runtimeSequence, actor.StopApplication)
	}

	if ShouldSetDroplet(plan) {
		runtimeSequence = append(runtimeSequence, actor.SetDropletForApplication)
	}

	return runtimeSequence
}
