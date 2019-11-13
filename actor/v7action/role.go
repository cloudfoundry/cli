package v7action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
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

func (actor Actor) GetOrgRole(roleType constant.RoleType, orgGUID string, userGUID string) (Role, Warnings, error) {
	roles, warnings, err := actor.CloudControllerClient.GetRoles(
		ccv3.Query{Key: ccv3.TypeFilter, Values: []string{string(roleType)}},
		ccv3.Query{Key: ccv3.OrganizationGUIDFilter, Values: []string{orgGUID}},
		ccv3.Query{Key: ccv3.UserGUIDFilter, Values: []string{userGUID}},
	)

	if err != nil {
		return Role{}, Warnings(warnings), err
	}

	if len(roles) == 0 {
		return Role{}, Warnings(warnings), actionerror.RoleNotFoundError{}
	}

	return Role(roles[0]), Warnings(warnings), err
}

func (actor Actor) GetSpaceRole(roleType constant.RoleType, spaceGUID string, userGUID string) (Role, Warnings, error) {
	roles, warnings, err := actor.CloudControllerClient.GetRoles(
		ccv3.Query{Key: ccv3.TypeFilter, Values: []string{string(roleType)}},
		ccv3.Query{Key: ccv3.SpaceGUIDFilter, Values: []string{spaceGUID}},
		ccv3.Query{Key: ccv3.UserGUIDFilter, Values: []string{userGUID}},
	)

	if err != nil {
		return Role{}, Warnings(warnings), err
	}

	if len(roles) == 0 {
		return Role{}, Warnings(warnings), actionerror.RoleNotFoundError{}
	}

	return Role(roles[0]), Warnings(warnings), err
}
