package v7action

import (
	"strconv"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/versioncheck"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/resources"
)

const MinimumCCAPIVersionForDeployable = "3.86.0"

// GetRevisionsByApplicationNameAndSpace returns revisions for application.
func (actor *Actor) GetRevisionsByApplicationNameAndSpace(appName string, spaceGUID string) ([]resources.Revision, Warnings, error) {
	var warnings Warnings
	app, appWarnings, appErr := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	warnings = append(warnings, appWarnings...)
	if appErr != nil {
		return []resources.Revision{}, warnings, appErr
	}
	revisions, apiWarnings, apiErr := actor.CloudControllerClient.GetApplicationRevisions(
		app.GUID,
		ccv3.Query{Key: ccv3.OrderBy, Values: []string{"-created_at"}},
	)
	warnings = append(warnings, apiWarnings...)
	if apiErr != nil {
		return []resources.Revision{}, warnings, apiErr
	}
	versionRequirementMet, versionErr := versioncheck.IsMinimumAPIVersionMet(actor.Config.APIVersion(), MinimumCCAPIVersionForDeployable)
	if versionErr != nil {
		return []resources.Revision{}, warnings, versionErr
	}
	if versionRequirementMet == false {
		_, deployableWarnings, deployableErr := actor.setRevisionsDeployableByDropletStateForApp(app.GUID, revisions)
		warnings = append(warnings, deployableWarnings...)
		if deployableErr != nil {
			return []resources.Revision{}, warnings, deployableErr
		}
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

	versionRequirementMet, versionErr := versioncheck.IsMinimumAPIVersionMet(actor.Config.APIVersion(), MinimumCCAPIVersionForDeployable)
	if versionErr != nil {
		return resources.Revision{}, Warnings(warnings), versionErr
	}
	if versionRequirementMet == false {
		_, deployableWarnings, deployableErr := actor.setRevisionsDeployableByDropletStateForApp(appGUID, revisions)
		warnings = append(warnings, deployableWarnings...)
		if deployableErr != nil {
			return resources.Revision{}, Warnings(warnings), deployableErr
		}
	}

	return revisions[0], Warnings(warnings), nil
}

func (actor Actor) setRevisionsDeployableByDropletStateForApp(appGUID string, revisions []resources.Revision) ([]resources.Revision, Warnings, error) {
	droplets, warnings, err := actor.CloudControllerClient.GetDroplets(
		ccv3.Query{Key: ccv3.AppGUIDFilter, Values: []string{appGUID}},
	)
	if err != nil {
		return []resources.Revision{}, Warnings(warnings), err
	}
	for i := range revisions {
		for _, droplet := range droplets {
			if revisions[i].Droplet.GUID == droplet.GUID {
				revisions[i].Deployable = (droplet.State == constant.DropletStaged)
			}
		}
	}
	return revisions, Warnings(warnings), nil
}

func (actor Actor) GetApplicationRevisionsDeployed(appGUID string) ([]resources.Revision, Warnings, error) {

	revisions, warnings, apiErr := actor.CloudControllerClient.GetApplicationRevisionsDeployed(appGUID)

	if apiErr != nil {
		return []resources.Revision{}, Warnings(warnings), apiErr
	}

	return revisions, Warnings(warnings), nil
}
