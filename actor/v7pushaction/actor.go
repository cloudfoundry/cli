// Package v7pushaction contains the business logic for orchestrating a V2 app
// push.
package v7pushaction

import (
	"regexp"
)

// Warnings is a list of warnings returned back from the cloud controller
type Warnings []string

// Actor handles all business logic for Cloud Controller v2 operations.
type Actor struct {
	SharedActor SharedActor
	V2Actor     V2Actor
	V7Actor     V7Actor

	PushPlanFuncs []UpdatePushPlanFunc

	ChangeApplicationFuncs []ChangeApplicationFunc
	StartFuncs             []ChangeApplicationFunc
	NoStartFuncs           []ChangeApplicationFunc

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
		SetupDropletPathForPushPlan,
		actor.SetupAllResourcesForPushPlan,
		SetupNoStartForPushPlan,
		SetupSkipRouteCreationForPushPlan,
		SetupScaleWebProcessForPushPlan,
		SetupUpdateWebProcessForPushPlan,
	}

	actor.ChangeApplicationFuncs = []ChangeApplicationFunc{
		actor.UpdateApplication,
		actor.UpdateRoutesForApplication,
		actor.ScaleWebProcessForApplication,
		actor.UpdateWebProcessForApplication,
		actor.CreateBitsPackageForApplication,
		actor.CreateDockerPackageForApplication,
		actor.CreateDropletForApplication,
	}

	actor.StartFuncs = []ChangeApplicationFunc{
		actor.StagePackageForApplication,
		actor.SetDropletForApplication,
	}

	actor.NoStartFuncs = []ChangeApplicationFunc{
		actor.StopApplication,
		actor.ConditionallyRunFunc(actor.IsDropletPathSet, actor.SetDropletForApplication),
	}

	return actor
}
