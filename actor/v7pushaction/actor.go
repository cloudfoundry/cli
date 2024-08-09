// Package v7pushaction contains the business logic for orchestrating a V2 app
// push.
package v7pushaction

import (
	"regexp"
)

// Warnings is a list of warnings returned back from the cloud controller
type Warnings []string

// Actor handles all business logic for Cloud Controller operations.
type Actor struct {
	SharedActor SharedActor
	V7Actor     V7Actor

	PreparePushPlanSequence   []UpdatePushPlanFunc
	ChangeApplicationSequence func(plan PushPlan) []ChangeApplicationFunc
	TransformManifestSequence []HandleFlagOverrideFunc

	startWithProtocol *regexp.Regexp
	urlValidator      *regexp.Regexp
}

const ProtocolRegexp = "^https?://|^tcp://"
const URLRegexp = "^(?:https?://|tcp://)?(?:(?:[\\w-]+\\.)|(?:[*]\\.))+\\w+(?:\\:\\d+)?(?:/.*)*(?:\\.\\w+)?$"

// NewActor returns a new actor.
func NewActor(v3Actor V7Actor, sharedActor SharedActor) *Actor {
	actor := &Actor{
		SharedActor: sharedActor,
		V7Actor:     v3Actor,

		startWithProtocol: regexp.MustCompile(ProtocolRegexp),
		urlValidator:      regexp.MustCompile(URLRegexp),
	}

	actor.TransformManifestSequence = []HandleFlagOverrideFunc{
		// app name override must come first, so it can trim the manifest
		// from multiple apps down to just one
		HandleAppNameOverride,

		HandleInstancesOverride,
		HandleStartCommandOverride,

		// Type must come before endpoint because endpoint validates against type
		HandleHealthCheckTypeOverride,
		HandleHealthCheckEndpointOverride,

		HandleHealthCheckTimeoutOverride,
		HandleMemoryOverride,
		HandleDiskOverride,
		HandleLogRateLimitOverride,
		HandleNoRouteOverride,
		HandleRandomRouteOverride,
		HandleTaskOverride,

		// this must come after all routing related transforms
		HandleDefaultRouteOverride,

		HandleDockerImageOverride,
		HandleDockerUsernameOverride,
		HandleStackOverride,
		HandleBuildpacksOverride,
		HandleStrategyOverride,
		HandleAppPathOverride,
		HandleDropletPathOverride,
	}

	actor.PreparePushPlanSequence = []UpdatePushPlanFunc{
		SetDefaultBitsPathForPushPlan,
		SetupDropletPathForPushPlan,
		actor.SetupAllResourcesForPushPlan,
		SetupDeploymentInformationForPushPlan,
		SetupNoStartForPushPlan,
		SetupNoWaitForPushPlan,
		SetupTaskAppForPushPlan,
	}

	actor.ChangeApplicationSequence = func(plan PushPlan) []ChangeApplicationFunc {
		var sequence []ChangeApplicationFunc
		sequence = append(sequence, actor.GetPrepareApplicationSourceSequence(plan)...)
		sequence = append(sequence, actor.GetRuntimeSequence(plan)...)
		return sequence
	}

	return actor
}
