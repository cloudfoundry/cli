package v3action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
)

// Droplet represents a Cloud Controller droplet.
type Droplet struct {
	GUID       string
	State      constant.DropletState
	CreatedAt  string
	Stack      string
	Image      string
	Buildpacks []Buildpack
}

type Buildpack ccv3.DropletBuildpack

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
		return allWarnings, actionerror.AssignDropletError{Message: newErr.Message}
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

	ccv3Droplets, apiWarnings, err := actor.CloudControllerClient.GetDroplets(
		ccv3.Query{Key: ccv3.AppGUIDFilter, Values: []string{application.GUID}},
	)
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

func (actor Actor) GetCurrentDropletByApplication(appGUID string) (Droplet, Warnings, error) {
	droplet, warnings, err := actor.CloudControllerClient.GetApplicationDropletCurrent(appGUID)
	switch err.(type) {
	case ccerror.ApplicationNotFoundError:
		return Droplet{}, Warnings(warnings), actionerror.ApplicationNotFoundError{GUID: appGUID}
	case ccerror.DropletNotFoundError:
		return Droplet{}, Warnings(warnings), actionerror.DropletNotFoundError{AppGUID: appGUID}
	}
	return actor.convertCCToActorDroplet(droplet), Warnings(warnings), err
}

func (actor Actor) convertCCToActorDroplet(ccDroplet ccv3.Droplet) Droplet {
	var buildpacks []Buildpack
	for _, ccBuildpack := range ccDroplet.Buildpacks {
		buildpacks = append(buildpacks, Buildpack(ccBuildpack))
	}

	return Droplet{
		GUID:       ccDroplet.GUID,
		State:      constant.DropletState(ccDroplet.State),
		CreatedAt:  ccDroplet.CreatedAt,
		Stack:      ccDroplet.Stack,
		Buildpacks: buildpacks,
		Image:      ccDroplet.Image,
	}
}
