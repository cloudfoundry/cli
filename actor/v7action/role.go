package v7action

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
)

type Role ccv3.Role

func (actor Actor) CreateOrgRole(roleType constant.RoleType, orgGUID string, userNameOrGUID string, userOrigin string, isClient bool) (Warnings, error) {
	roleToCreate := ccv3.Role{
		Type:    roleType,
		OrgGUID: orgGUID,
	}

	if isClient {
		roleToCreate.UserGUID = userNameOrGUID
	} else {
		roleToCreate.UserName = userNameOrGUID
		roleToCreate.Origin = userOrigin
	}

	_, warnings, err := actor.CloudControllerClient.CreateRole(roleToCreate)

	return Warnings(warnings), err
}

func (actor Actor) CreateSpaceRole(roleType constant.RoleType, orgGUID string, spaceGUID string, userNameOrGUID string, userOrigin string, isClient bool) (Warnings, error) {
	roleToCreate := ccv3.Role{
		Type:      roleType,
		SpaceGUID: spaceGUID,
	}

	if isClient {
		roleToCreate.UserGUID = userNameOrGUID
	} else {
		roleToCreate.UserName = userNameOrGUID
		roleToCreate.Origin = userOrigin
	}

	warnings, err := actor.CreateOrgRole(constant.OrgUserRole, orgGUID, userNameOrGUID, userOrigin, isClient)
	if err != nil {
		if _, isIdempotentError := err.(ccerror.RoleAlreadyExistsError); !isIdempotentError {
			return warnings, err
		}
	}

	_, ccv3Warnings, err := actor.CloudControllerClient.CreateRole(roleToCreate)
	warnings = append(warnings, ccv3Warnings...)

	return warnings, err
}

func (actor Actor) GetRolesByOrg(orgGuid string) ([]Role, Warnings, error) {
	_, ccWarnings, err := actor.CloudControllerClient.GetRoles(ccv3.Query{
		Key:    ccv3.OrganizationGUIDFilter,
		Values: []string{orgGuid},
	})
	if err != nil {
		return []Role{}, Warnings(ccWarnings), err
	}

	return []Role{}, Warnings{}, nil
}
