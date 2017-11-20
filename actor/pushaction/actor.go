// Package pushaction contains the business logic for orchestrating a V2 app
// push.
package pushaction

import (
	"regexp"

	"code.cloudfoundry.org/cli/util/randomword"
)

// Warnings is a list of warnings returned back from the cloud controller
type Warnings []string

// Actor handles all business logic for Cloud Controller v2 operations.
type Actor struct {
	SharedActor   SharedActor
	V2Actor       V2Actor
	WordGenerator RandomWordGenerator

	startWithProtocol *regexp.Regexp
	urlValidator      *regexp.Regexp
}

const ProtocolRegexp = "^https?://|^tcp://"
const URLRegexp = "^(?:https?://|tcp://)?(?:[\\w-]+\\.)+\\w+(?:\\:\\d+)?(?:/.*)*(?:\\.\\w+)?$"

// NewActor returns a new actor.
func NewActor(v2Actor V2Actor, sharedActor SharedActor) *Actor {
	return &Actor{
		SharedActor:   sharedActor,
		V2Actor:       v2Actor,
		WordGenerator: new(randomword.Generator),

		startWithProtocol: regexp.MustCompile(ProtocolRegexp),
		urlValidator:      regexp.MustCompile(URLRegexp),
	}
}
