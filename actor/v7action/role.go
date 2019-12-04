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
		roleToCreate.Username = userNameOrGUID
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
		roleToCreate.Username = userNameOrGUID
		roleToCreate.Origin = userOrigin
	}

	warnings, err := actor.CreateOrgRole(constant.OrgUserRole, orgGUID, userNameOrGUID, userOrigin, isClient)
	if err != nil {
		_, isIdempotentError := err.(ccerror.RoleAlreadyExistsError)
		_, isForbiddenError := err.(ccerror.ForbiddenError)

		if !isIdempotentError && !isForbiddenError {
			return warnings, err
		}
	}

	_, ccv3Warnings, err := actor.CloudControllerClient.CreateRole(roleToCreate)
	warnings = append(warnings, ccv3Warnings...)

	return warnings, err
}

func (actor Actor) DeleteSpaceRole(roleType constant.RoleType, spaceGUID string, userNameOrGUID string, userOrigin string, isClient bool) (Warnings, error) {
	var userGUID string
	var allWarnings Warnings
	if isClient {
		user, warnings, err := actor.CloudControllerClient.GetUser(userNameOrGUID)
		allWarnings = append(allWarnings, warnings...)
		if err != nil {
			_, ok := err.(ccerror.UserNotFoundError)
			if ok {
				err = ccerror.UserNotFoundError{Username: userNameOrGUID, IsClient: isClient}
			}
			return Warnings(allWarnings), err
		}
		userGUID = user.GUID
	} else {
		ccv3Users, warnings, err := actor.CloudControllerClient.GetUsers(
			ccv3.Query{
				Key:    ccv3.UsernamesFilter,
				Values: []string{userNameOrGUID},
			},
			ccv3.Query{
				Key:    ccv3.OriginsFilter,
				Values: []string{userOrigin},
			},
		)
		allWarnings = append(allWarnings, warnings...)
		if err != nil {
			return Warnings(allWarnings), err
		}
		if len(ccv3Users) == 0 {
			return allWarnings, ccerror.UserNotFoundError{Username: userNameOrGUID, Origin: userOrigin}
		}
		userGUID = ccv3Users[0].GUID
	}

	roleGUID, warnings, err := actor.GetRoleGUID(spaceGUID, userGUID, roleType)
	print("ROLE GUID: ", roleGUID)
	allWarnings = append(allWarnings, warnings...)
	if err != nil || roleGUID == "" {
		return allWarnings, err
	}

	jobURL, deleteRoleWarnings, err := actor.CloudControllerClient.DeleteRole(roleGUID)
	allWarnings = append(allWarnings, deleteRoleWarnings...)
	if err != nil {
		return allWarnings, err
	}

	pollJobWarnings, err := actor.CloudControllerClient.PollJob(jobURL)
	allWarnings = append(allWarnings, pollJobWarnings...)
	if err != nil {
		return allWarnings, err
	}

	return allWarnings, nil
}

func (actor Actor) GetRoleGUID(spaceGUID string, userGUID string, roleType constant.RoleType) (string, Warnings, error) {
	ccv3Roles, _, warnings, err := actor.CloudControllerClient.GetRoles(
		ccv3.Query{
			Key:    ccv3.UserGUIDFilter,
			Values: []string{userGUID},
		},
		ccv3.Query{
			Key:    ccv3.RoleTypesFilter,
			Values: []string{string(roleType)},
		},
		ccv3.Query{
			Key:    ccv3.SpaceGUIDFilter,
			Values: []string{spaceGUID},
		},
	)

	if err != nil {
		return "", Warnings(warnings), err
	}

	if len(ccv3Roles) == 0 {
		return "", Warnings(warnings), nil
	}

	return ccv3Roles[0].GUID, Warnings(warnings), nil
}

func (actor Actor) GetOrgUsersByRoleType(orgGuid string) (map[constant.RoleType][]User, Warnings, error) {
	return actor.getUsersByRoleType(orgGuid, ccv3.OrganizationGUIDFilter)
}

func (actor Actor) GetSpaceUsersByRoleType(spaceGuid string) (map[constant.RoleType][]User, Warnings, error) {
	return actor.getUsersByRoleType(spaceGuid, ccv3.SpaceGUIDFilter)
}

func (actor Actor) getUsersByRoleType(guid string, filterKey ccv3.QueryKey) (map[constant.RoleType][]User, Warnings, error) {
	ccv3Roles, includes, ccWarnings, err := actor.CloudControllerClient.GetRoles(
		ccv3.Query{
			Key:    filterKey,
			Values: []string{guid},
		},
		ccv3.Query{
			Key:    ccv3.Include,
			Values: []string{"user"},
		},
	)
	if err != nil {
		return nil, Warnings(ccWarnings), err
	}
	usersByGuids := make(map[string]ccv3.User)
	for _, user := range includes.Users {
		usersByGuids[user.GUID] = user
	}
	usersByRoleType := make(map[constant.RoleType][]User)
	for _, role := range ccv3Roles {
		user := User(usersByGuids[role.UserGUID])
		usersByRoleType[role.Type] = append(usersByRoleType[role.Type], user)
	}
	return usersByRoleType, Warnings(ccWarnings), nil
}
