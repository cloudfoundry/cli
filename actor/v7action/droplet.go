package v7action

import (
	"io"

	"code.cloudfoundry.org/cli/v8/actor/actionerror"
	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/v8/resources"
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

func (actor Actor) DownloadCurrentDropletByAppName(appName string, spaceGUID string) ([]byte, string, Warnings, error) {
	var allWarnings Warnings

	app, warnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return []byte{}, "", allWarnings, err
	}

	droplet, ccWarnings, err := actor.CloudControllerClient.GetApplicationDropletCurrent(app.GUID)
	allWarnings = append(allWarnings, ccWarnings...)

	if err != nil {
		if _, ok := err.(ccerror.DropletNotFoundError); ok {
			return []byte{}, "", allWarnings, actionerror.DropletNotFoundError{}
		}
		return []byte{}, "", allWarnings, err
	}

	rawDropletBytes, ccWarnings, err := actor.CloudControllerClient.DownloadDroplet(droplet.GUID)
	allWarnings = append(allWarnings, ccWarnings...)
	if err != nil {
		return []byte{}, "", allWarnings, err
	}

	return rawDropletBytes, droplet.GUID, allWarnings, nil
}

func (actor Actor) DownloadDropletByGUIDAndAppName(dropletGUID string, appName string, spaceGUID string) ([]byte, Warnings, error) {
	var allWarnings Warnings

	app, warnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return []byte{}, allWarnings, err
	}

	droplets, getDropletWarnings, err := actor.CloudControllerClient.GetDroplets(
		ccv3.Query{Key: ccv3.GUIDFilter, Values: []string{dropletGUID}},
		ccv3.Query{Key: ccv3.AppGUIDFilter, Values: []string{app.GUID}},
		ccv3.Query{Key: ccv3.PerPage, Values: []string{"1"}},
		ccv3.Query{Key: ccv3.Page, Values: []string{"1"}},
	)
	allWarnings = append(allWarnings, getDropletWarnings...)
	if err != nil {
		return []byte{}, allWarnings, err
	}

	if len(droplets) == 0 {
		return []byte{}, allWarnings, actionerror.DropletNotFoundError{}
	}

	rawDropletBytes, ccWarnings, err := actor.CloudControllerClient.DownloadDroplet(dropletGUID)
	allWarnings = append(allWarnings, ccWarnings...)
	if err != nil {
		return []byte{}, allWarnings, err
	}

	return rawDropletBytes, allWarnings, nil
}
