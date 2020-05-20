package v7action

import (
	"io"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/resources"
)

// CreateApplicationDroplet creates a new droplet without a package for the app with
// guid appGUID.
func (actor Actor) CreateApplicationDroplet(appGUID string) (resources.Droplet, Warnings, error) {
	ccDroplet, warnings, err := actor.CloudControllerClient.CreateDroplet(appGUID)

	return ccDroplet, Warnings(warnings), err
}

// SetApplicationDropletByApplicationNameAndSpace sets the droplet for an application.
func (actor Actor) SetApplicationDropletByApplicationNameAndSpace(appName string, spaceGUID string, dropletGUID string) (Warnings, error) {
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

func (actor Actor) SetApplicationDroplet(appGUID string, dropletGUID string) (Warnings, error) {
	_, warnings, err := actor.CloudControllerClient.SetApplicationDroplet(appGUID, dropletGUID)

	if newErr, ok := err.(ccerror.UnprocessableEntityError); ok {
		return Warnings(warnings), actionerror.AssignDropletError{Message: newErr.Message}
	}

	return Warnings(warnings), err
}

// GetApplicationDroplets returns the list of droplets that belong to application.
func (actor Actor) GetApplicationDroplets(appName string, spaceGUID string) ([]resources.Droplet, Warnings, error) {
	allWarnings := Warnings{}
	application, warnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return nil, allWarnings, err
	}

	droplets, apiWarnings, err := actor.CloudControllerClient.GetDroplets(
		ccv3.Query{Key: ccv3.AppGUIDFilter, Values: []string{application.GUID}},
		ccv3.Query{Key: ccv3.OrderBy, Values: []string{ccv3.CreatedAtDescendingOrder}},
	)
	actorWarnings := Warnings(apiWarnings)
	allWarnings = append(allWarnings, actorWarnings...)
	if err != nil {
		return nil, allWarnings, err
	}

	if len(droplets) == 0 {
		return []resources.Droplet{}, allWarnings, nil
	}

	currentDroplet, apiWarnings, err := actor.CloudControllerClient.GetApplicationDropletCurrent(application.GUID)
	allWarnings = append(allWarnings, apiWarnings...)
	if err != nil {
		if _, ok := err.(ccerror.DropletNotFoundError); ok {
			return droplets, allWarnings, nil
		}
		return []resources.Droplet{}, allWarnings, err
	}

	for i, droplet := range droplets {
		if droplet.GUID == currentDroplet.GUID {
			droplets[i].IsCurrent = true
		}
	}

	return droplets, allWarnings, err
}

func (actor Actor) GetCurrentDropletByApplication(appGUID string) (resources.Droplet, Warnings, error) {
	droplet, warnings, err := actor.CloudControllerClient.GetApplicationDropletCurrent(appGUID)
	switch err.(type) {
	case ccerror.ApplicationNotFoundError:
		return resources.Droplet{}, Warnings(warnings), actionerror.ApplicationNotFoundError{GUID: appGUID}
	case ccerror.DropletNotFoundError:
		return resources.Droplet{}, Warnings(warnings), actionerror.DropletNotFoundError{AppGUID: appGUID}
	}
	return droplet, Warnings(warnings), err
}

func (actor Actor) UploadDroplet(dropletGUID string, dropletPath string, progressReader io.Reader, size int64) (Warnings, error) {
	var allWarnings Warnings

	jobURL, uploadWarnings, err := actor.CloudControllerClient.UploadDropletBits(dropletGUID, dropletPath, progressReader, size)
	allWarnings = append(allWarnings, uploadWarnings...)
	if err != nil {
		return allWarnings, err
	}

	jobWarnings, jobErr := actor.CloudControllerClient.PollJob(jobURL)
	allWarnings = append(allWarnings, jobWarnings...)
	if jobErr != nil {
		return allWarnings, jobErr
	}

	return allWarnings, nil
}
