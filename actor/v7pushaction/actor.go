// Package v7pushaction contains the business logic for orchestrating a V2 app
// push.
package v7pushaction

import (
	"regexp"

	"code.cloudfoundry.org/cli/util/randomword"
)

// Warnings is a list of warnings returned back from the cloud controller
type Warnings []string

// Actor handles all business logic for Cloud Controller v2 operations.
type Actor struct {
	SharedActor SharedActor
	V7Actor     V7Actor

	PreparePushPlanSequence   []UpdatePushPlanFunc
	ChangeApplicationSequence func(plan PushPlan) []ChangeApplicationFunc
	RandomWordGenerator       RandomWordGenerator

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

		RandomWordGenerator: new(randomword.Generator),
		startWithProtocol:   regexp.MustCompile(ProtocolRegexp),
		urlValidator:        regexp.MustCompile(URLRegexp),
	}

	actor.PreparePushPlanSequence = []UpdatePushPlanFunc{
		SetupApplicationForPushPlan,
		SetupDockerImageCredentialsForPushPlan,
		SetupBitsPathForPushPlan,
		SetupDropletPathForPushPlan,
		actor.SetupAllResourcesForPushPlan,
		SetupDeploymentStrategyForPushPlan,
		SetupNoStartForPushPlan,
		SetupNoWaitForPushPlan,
		SetupSkipRouteCreationForPushPlan,
		SetupScaleWebProcessForPushPlan,
		SetupUpdateWebProcessForPushPlan,
	}

	actor.ChangeApplicationSequence = func(plan PushPlan) []ChangeApplicationFunc {
		var sequence []ChangeApplicationFunc
		sequence = append(sequence, actor.GetUpdateSequence(plan)...)
		sequence = append(sequence, actor.GetPrepareApplicationSourceSequence(plan)...)
		sequence = append(sequence, actor.GetRuntimeSequence(plan)...)
		return sequence
	}

	return actor
}
