package v3action

import "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"

// Droplet represents a Cloud Controller droplet.
type Droplet struct {
	GUID       string
	Stack      string
	Buildpacks []Buildpack
}

type Buildpack ccv3.Buildpack

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

	return allWarnings, err
}
