// Package v7pushaction contains the business logic for orchestrating a V2 app
// push.
package v7pushaction

import (
	"regexp"

	"code.cloudfoundry.org/cli/util/manifestparser"
)

// UpdatePushPlanFunc is a function that is used by CreatePushPlans to setup
// push plans for the push command.
type UpdatePushPlanFunc func(pushPlan PushPlan, manifestApp manifestparser.Application) (PushPlan, error)

// Warnings is a list of warnings returned back from the cloud controller
type Warnings []string

// Actor handles all business logic for Cloud Controller v2 operations.
type Actor struct {
	SharedActor   SharedActor
	V2Actor       V2Actor
	V7Actor       V7Actor
	PushPlanFuncs []UpdatePushPlanFunc

	startWithProtocol *regexp.Regexp
	urlValidator      *regexp.Regexp
}

const ProtocolRegexp = "^https?://|^tcp://"
const URLRegexp = "^(?:https?://|tcp://)?(?:(?:[\\w-]+\\.)|(?:[*]\\.))+\\w+(?:\\:\\d+)?(?:/.*)*(?:\\.\\w+)?$"

// NewActor returns a new actor.
func NewActor(v2Actor V2Actor, v3Actor V7Actor, sharedActor SharedActor) *Actor {
	actor := &Actor{
		SharedActor: sharedActor,
		V2Actor:     v2Actor,
		V7Actor:     v3Actor,

		startWithProtocol: regexp.MustCompile(ProtocolRegexp),
		urlValidator:      regexp.MustCompile(URLRegexp),
	}

	actor.PushPlanFuncs = []UpdatePushPlanFunc{
		SetupApplicationForPushPlan,
		SetupDockerImageCredentialsForPushPlan,
		SetupBitsPathForPushPlan,
		actor.SetupAllResourcesForPushPlan,
		SetupNoStartForPushPlan,
		SetupSkipRouteCreationForPushPlan,
	}

	return actor
}
