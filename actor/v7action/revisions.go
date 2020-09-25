package v7action

import (
	"strconv"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/resources"
)

// GetRevisionsByApplicationNameAndSpace returns revisions for application.
func (actor *Actor) GetRevisionsByApplicationNameAndSpace(appName string, spaceGUID string) ([]resources.Revision, Warnings, error) {
	app, warnings, appErr := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	if appErr != nil {
		return []resources.Revision{}, warnings, appErr
	}
	revisions, v3Warnings, apiErr := actor.CloudControllerClient.GetApplicationRevisions(
		app.GUID,
		ccv3.Query{Key: ccv3.OrderBy, Values: []string{"-created_at"}},
	)
	warnings = append(warnings, v3Warnings...)
	if apiErr != nil {
		return []resources.Revision{}, warnings, apiErr
	}

	return revisions, warnings, nil
}

func (actor Actor) GetRevisionByApplicationAndVersion(appGUID string, revisionVersion int) (resources.Revision, Warnings, error) {
	query := ccv3.Query{
		Key:    ccv3.VersionsFilter,
		Values: []string{strconv.Itoa(revisionVersion)},
	}
	revisions, warnings, apiErr := actor.CloudControllerClient.GetApplicationRevisions(appGUID, query)
	if apiErr != nil {
		return resources.Revision{}, Warnings(warnings), apiErr
	}

	if len(revisions) > 1 {
		return resources.Revision{}, Warnings(warnings), actionerror.RevisionAmbiguousError{Version: revisionVersion}
	}

	if len(revisions) == 0 {
		return resources.Revision{}, Warnings(warnings), actionerror.RevisionNotFoundError{Version: revisionVersion}
	}

	return revisions[0], Warnings(warnings), nil
}

func (actor Actor) GetApplicationRevisionsDeployed(appGUID string) ([]resources.Revision, Warnings, error) {

	revisions, warnings, apiErr := actor.CloudControllerClient.GetApplicationRevisionsDeployed(appGUID)

	if apiErr != nil {
		return []resources.Revision{}, Warnings(warnings), apiErr
	}

	return revisions, Warnings(warnings), nil
}
