package v7action

import (
	"code.cloudfoundry.org/cli/resources"
)

// GetRevisionsByApplicationNameAndSpace returns revisions for application.
func (actor *Actor) GetRevisionsByApplicationNameAndSpace(appName string, spaceGUID string) ([]resources.Revision, Warnings, error) {
	app, warnings, appErr := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	if appErr != nil {
		return []resources.Revision{}, warnings, appErr
	}

	revisions, v3Warnings, apiErr := actor.CloudControllerClient.GetApplicationRevisions(app.GUID)
	warnings = append(warnings, v3Warnings...)

	return revisions, warnings, apiErr
}

func (actor Actor) GetRevisionByApplicationAndVersion(appGUID string, revisionVersion int) (resources.Revision, Warnings, error) {
	revisions, warnings, _ := actor.CloudControllerClient.GetApplicationRevisions(appGUID)

	for _, revision := range revisions {
		if revision.Version == revisionVersion {
			return revision, Warnings(warnings), nil
		}
	}
	return resources.Revision{}, Warnings(warnings), nil
}
