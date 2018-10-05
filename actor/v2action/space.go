package v2action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	uaaconst "code.cloudfoundry.org/cli/api/uaa/constant"
)

// Space represents a CLI Space
type Space ccv2.Space

func (actor Actor) CreateSpace(spaceName, orgName, quotaName string) (Space, Warnings, error) {
	var allWarnings Warnings
	org, getOrgWarnings, err := actor.GetOrganizationByName(orgName)
	allWarnings = append(allWarnings, Warnings(getOrgWarnings)...)
	if err != nil {
		return Space{}, allWarnings, err
	}

	var spaceQuota SpaceQuota
	if quotaName != "" {
		var getQuotaWarnings Warnings
		spaceQuota, getQuotaWarnings, err = actor.GetSpaceQuotaByName(quotaName, org.GUID)
		allWarnings = append(allWarnings, Warnings(getQuotaWarnings)...)
		if err != nil {
			return Space{}, allWarnings, err
		}
	}

	space, spaceWarnings, err := actor.CloudControllerClient.CreateSpace(spaceName, org.GUID)
	allWarnings = append(allWarnings, Warnings(spaceWarnings)...)
	if err != nil {
		if _, ok := err.(ccerror.SpaceNameTakenError); ok {
			return Space{}, allWarnings, actionerror.SpaceNameTakenError{Name: spaceName}
		}
		return Space{}, allWarnings, err
	}

	if quotaName != "" {
		var setQuotaWarnings Warnings
		setQuotaWarnings, err = actor.SetSpaceQuota(space.GUID, spaceQuota.GUID)
		allWarnings = append(allWarnings, Warnings(setQuotaWarnings)...)

		if err != nil {
			return Space{}, allWarnings, err
		}
	}

	return Space(space), allWarnings, err
}

func (actor Actor) DeleteSpaceByNameAndOrganizationName(spaceName string, orgName string) (Warnings, error) {
	var allWarnings Warnings

	org, warnings, err := actor.GetOrganizationByName(orgName)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	space, warnings, err := actor.GetSpaceByOrganizationAndName(org.GUID, spaceName)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	job, deleteWarnings, err := actor.CloudControllerClient.DeleteSpaceJob(space.GUID)
	allWarnings = append(allWarnings, Warnings(deleteWarnings)...)
	if err != nil {
		return allWarnings, err
	}

	warnings, err = actor.PollJob(Job(job))
	allWarnings = append(allWarnings, Warnings(warnings)...)

	return allWarnings, err
}

// GetOrganizationSpaces returns a list of spaces in the specified org
func (actor Actor) GetOrganizationSpaces(orgGUID string) ([]Space, Warnings, error) {
	ccv2Spaces, warnings, err := actor.CloudControllerClient.GetSpaces(ccv2.Filter{
		Type:     constant.OrganizationGUIDFilter,
		Operator: constant.EqualOperator,
		Values:   []string{orgGUID},
	})
	if err != nil {
		return []Space{}, Warnings(warnings), err
	}

	spaces := make([]Space, len(ccv2Spaces))
	for i, ccv2Space := range ccv2Spaces {
		spaces[i] = Space(ccv2Space)
	}

	return spaces, Warnings(warnings), nil
}

// GetSpaceByOrganizationAndName returns an Space based on the org and name.
func (actor Actor) GetSpaceByOrganizationAndName(orgGUID string, spaceName string) (Space, Warnings, error) {
	ccv2Spaces, warnings, err := actor.CloudControllerClient.GetSpaces(
		ccv2.Filter{
			Type:     constant.NameFilter,
			Operator: constant.EqualOperator,
			Values:   []string{spaceName},
		},
		ccv2.Filter{
			Type:     constant.OrganizationGUIDFilter,
			Operator: constant.EqualOperator,
			Values:   []string{orgGUID},
		},
	)
	if err != nil {
		return Space{}, Warnings(warnings), err
	}

	if len(ccv2Spaces) == 0 {
		return Space{}, Warnings(warnings), actionerror.SpaceNotFoundError{Name: spaceName}
	}

	if len(ccv2Spaces) > 1 {
		return Space{}, Warnings(warnings), actionerror.MultipleSpacesFoundError{OrgGUID: orgGUID, Name: spaceName}
	}

	return Space(ccv2Spaces[0]), Warnings(warnings), nil
}

// GrantSpaceManagerByUsername makes the provided user a Space Manager in the
// space with the provided guid.
func (actor Actor) GrantSpaceManagerByUsername(orgGUID string, spaceGUID string, username string) (Warnings, error) {
	if actor.Config.UAAGrantType() == string(uaaconst.GrantTypeClientCredentials) {
		return actor.grantSpaceManagerByClientCredentials(orgGUID, spaceGUID, username)
	}

	return actor.grantSpaceManagerByUsername(orgGUID, spaceGUID, username)
}

func (actor Actor) grantSpaceManagerByClientCredentials(orgGUID string, spaceGUID string, clientID string) (Warnings, error) {
	ccv2Warnings, err := actor.CloudControllerClient.UpdateOrganizationUser(orgGUID, clientID)
	warnings := Warnings(ccv2Warnings)
	if err != nil {
		return warnings, err
	}

	ccv2Warnings, err = actor.CloudControllerClient.UpdateSpaceManager(spaceGUID, clientID)
	warnings = append(warnings, Warnings(ccv2Warnings)...)
	if err != nil {
		return warnings, err
	}

	return warnings, err
}

func (actor Actor) grantSpaceManagerByUsername(orgGUID string, spaceGUID string, username string) (Warnings, error) {
	ccv2Warnings, err := actor.CloudControllerClient.UpdateOrganizationUserByUsername(orgGUID, username)
	warnings := Warnings(ccv2Warnings)
	if err != nil {
		return warnings, err
	}

	ccv2Warnings, err = actor.CloudControllerClient.UpdateSpaceManagerByUsername(spaceGUID, username)
	warnings = append(warnings, Warnings(ccv2Warnings)...)

	return warnings, err
}

// GrantSpaceDeveloperByUsername makes the provided user a Space Developer in the
// space with the provided guid.
func (actor Actor) GrantSpaceDeveloperByUsername(spaceGUID string, username string) (Warnings, error) {
	if actor.Config.UAAGrantType() == string(uaaconst.GrantTypeClientCredentials) {
		warnings, err := actor.CloudControllerClient.UpdateSpaceDeveloper(spaceGUID, username)

		return Warnings(warnings), err
	}

	warnings, err := actor.CloudControllerClient.UpdateSpaceDeveloperByUsername(spaceGUID, username)
	return Warnings(warnings), err
}
