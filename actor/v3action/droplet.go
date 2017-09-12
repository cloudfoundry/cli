package v3action

import (
	"net/url"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
)

type DropletState string

const (
	DropletStateStaged  DropletState = "STAGED"
	DropletStateFailed  DropletState = "FAILED"
	DropletStateCopying DropletState = "COPYING"
	DropletStateExpired DropletState = "EXPIRED"
)

// Droplet represents a Cloud Controller droplet.
type Droplet struct {
	GUID       string
	State      DropletState
	CreatedAt  string
	Stack      string
	Image      string
	Buildpacks []Buildpack
}

type Buildpack ccv3.DropletBuildpack

// AssignDropletError is returned when assigning the current droplet of an app
// fails
type AssignDropletError struct {
	Message string
}

func (a AssignDropletError) Error() string {
	return a.Message
}

// SetApplicationDroplet sets the droplet for an application.
func (actor Actor) SetApplicationDroplet(appName string, spaceGUID string, dropletGUID string) (Warnings, error) {
	allWarnings := Warnings{}
	application, warnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}
	_, apiWarnings, err := actor.CloudControllerClient.SetApplicationDroplet(application.GUID, dropletGUID)
	actorWarnings := Warnings(apiWarnings)
	allWarnings = append(allWarnings, actorWarnings...)

	if newErr, ok := err.(ccerror.UnprocessableEntityError); ok {
		return allWarnings, AssignDropletError{Message: newErr.Message}
	}

	return allWarnings, err
}

// GetApplicationDroplets returns the list of droplets that belong to applicaiton.
func (actor Actor) GetApplicationDroplets(appName string, spaceGUID string) ([]Droplet, Warnings, error) {
	allWarnings := Warnings{}
	application, warnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return nil, allWarnings, err
	}

	ccv3Droplets, apiWarnings, err := actor.CloudControllerClient.GetApplicationDroplets(application.GUID, url.Values{})
	actorWarnings := Warnings(apiWarnings)
	allWarnings = append(allWarnings, actorWarnings...)
	if err != nil {
		return nil, allWarnings, err
	}

	var droplets []Droplet
	for _, ccv3Droplet := range ccv3Droplets {
		droplets = append(droplets, actor.convertCCToActorDroplet(ccv3Droplet))
	}

	return droplets, allWarnings, err
}

func (actor Actor) convertCCToActorDroplet(ccv3Droplet ccv3.Droplet) Droplet {
	var buildpacks []Buildpack
	for _, ccv3Buildpack := range ccv3Droplet.Buildpacks {
		buildpacks = append(buildpacks, Buildpack(ccv3Buildpack))
	}

	return Droplet{
		GUID:       ccv3Droplet.GUID,
		State:      DropletState(ccv3Droplet.State),
		CreatedAt:  ccv3Droplet.CreatedAt,
		Stack:      ccv3Droplet.Stack,
		Buildpacks: buildpacks,
		Image:      ccv3Droplet.Image,
	}
}
