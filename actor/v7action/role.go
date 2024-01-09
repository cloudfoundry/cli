package v7action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/resources"
)

func (actor Actor) CreateOrgRole(roleType constant.RoleType, orgGUID string, userNameOrGUID string, userOrigin string, isClient bool) (Warnings, error) {
	roleToCreate := resources.Role{
		Type:    roleType,
		OrgGUID: orgGUID,
	}

	if isClient {
		err := actor.UAAClient.ValidateClientUser(userNameOrGUID)
		if err != nil {
			return Warnings{}, err
		}

		roleToCreate.UserGUID = userNameOrGUID
	} else {
		roleToCreate.Username = userNameOrGUID
		roleToCreate.Origin = userOrigin
	}

	_, warnings, err := actor.CloudControllerClient.CreateRole(roleToCreate)

	return Warnings(warnings), err
}

func (actor Actor) CreateSpaceRole(roleType constant.RoleType, orgGUID string, spaceGUID string, userNameOrGUID string, userOrigin string, isClient bool) (Warnings, error) {
	roleToCreate := resources.Role{
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
		_, isUserNotFoundError := err.(actionerror.UserNotFoundError)

		if !isIdempotentError && !isForbiddenError && !isUserNotFoundError {
			return warnings, err
		}
	}

	_, ccv3Warnings, err := actor.CloudControllerClient.CreateRole(roleToCreate)
	warnings = append(warnings, ccv3Warnings...)

	return warnings, err
}

func (actor Actor) DeleteOrgRole(roleType constant.RoleType, orgGUID string, userNameOrGUID string, userOrigin string, isClient bool) (Warnings, error) {
	var userGUID string
	var allWarnings Warnings
	userGUID, warnings, err := actor.getUserGuidForDeleteRole(isClient, userNameOrGUID, userOrigin, allWarnings)
	allWarnings = append(allWarnings, warnings...)
	if err != nil {
		return allWarnings, err
	}

	roleGUID, warnings, err := actor.GetRoleGUID(ccv3.OrganizationGUIDFilter, orgGUID, userGUID, roleType)
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

func (actor Actor) DeleteSpaceRole(roleType constant.RoleType, spaceGUID string, userNameOrGUID string, userOrigin string, isClient bool) (Warnings, error) {
	var userGUID string
	var allWarnings Warnings
	userGUID, userWarnings, err := actor.getUserGuidForDeleteRole(isClient, userNameOrGUID, userOrigin, allWarnings)
	allWarnings = append(allWarnings, userWarnings...)
	if err != nil {
		return allWarnings, err
	}

	roleGUID, roleWarnings, err := actor.GetRoleGUID(ccv3.SpaceGUIDFilter, spaceGUID, userGUID, roleType)
	allWarnings = append(allWarnings, roleWarnings...)
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

func (actor Actor) getUserGuidForDeleteRole(isClient bool, userNameOrGUID string, userOrigin string, allWarnings Warnings) (string, Warnings, error) {
	var userGUID string
	if isClient {
		user, warnings, err := actor.CloudControllerClient.GetUser(userNameOrGUID)
		allWarnings = append(allWarnings, warnings...)
		if err != nil {
			if _, ok := err.(ccerror.UserNotFoundError); ok {
				err = actionerror.UserNotFoundError{Username: userNameOrGUID}
			}
			return "", allWarnings, err
		}
		userGUID = user.GUID
	} else {
		queries := []ccv3.Query{{
			Key:    ccv3.UsernamesFilter,
			Values: []string{userNameOrGUID},
		}}
		if userOrigin != "" {
			queries = append(queries, ccv3.Query{
				Key:    ccv3.OriginsFilter,
				Values: []string{userOrigin},
			})
		}

		ccv3Users, warnings, err := actor.CloudControllerClient.GetUsers(queries...)
		allWarnings = append(allWarnings, warnings...)
		if err != nil {
			return "", allWarnings, err
		}
		if len(ccv3Users) == 0 {
			return "", allWarnings, actionerror.UserNotFoundError{Username: userNameOrGUID, Origin: userOrigin}
		}
		if len(ccv3Users) > 1 {
			origins := []string{}
			for _, user := range ccv3Users {
				origins = append(origins, user.Origin)
			}
			return "", allWarnings, actionerror.AmbiguousUserError{Username: userNameOrGUID, Origins: origins}
		}
		userGUID = ccv3Users[0].GUID
	}
	return userGUID, allWarnings, nil
}

func (actor Actor) GetRoleGUID(queryKey ccv3.QueryKey, orgOrSpaceGUID string, userGUID string, roleType constant.RoleType) (string, Warnings, error) {
	ccv3Roles, _, warnings, err := actor.CloudControllerClient.GetRoles(
		ccv3.Query{Key: ccv3.UserGUIDFilter, Values: []string{userGUID}},
		ccv3.Query{Key: ccv3.RoleTypesFilter, Values: []string{string(roleType)}},
		ccv3.Query{Key: queryKey, Values: []string{orgOrSpaceGUID}},
		ccv3.Query{Key: ccv3.PerPage, Values: []string{"1"}},
		ccv3.Query{Key: ccv3.Page, Values: []string{"1"}},
	)

	if err != nil {
		return "", Warnings(warnings), err
	}

	if len(ccv3Roles) == 0 {
		return "", Warnings(warnings), nil
	}

	return ccv3Roles[0].GUID, Warnings(warnings), nil
}

func (actor Actor) GetOrgUsersByRoleType(orgGuid string) (map[constant.RoleType][]resources.User, Warnings, error) {
	return actor.getUsersByRoleType(orgGuid, ccv3.OrganizationGUIDFilter)
}

func (actor Actor) GetSpaceUsersByRoleType(spaceGuid string) (map[constant.RoleType][]resources.User, Warnings, error) {
	return actor.getUsersByRoleType(spaceGuid, ccv3.SpaceGUIDFilter)
}

func (actor Actor) getUsersByRoleType(guid string, filterKey ccv3.QueryKey) (map[constant.RoleType][]resources.User, Warnings, error) {
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
	usersByGuids := make(map[string]resources.User)
	for _, user := range includes.Users {
		usersByGuids[user.GUID] = user
	}
	usersByRoleType := make(map[constant.RoleType][]resources.User)
	for _, role := range ccv3Roles {
		user := resources.User(usersByGuids[role.UserGUID])
		usersByRoleType[role.Type] = append(usersByRoleType[role.Type], user)
	}
	return usersByRoleType, Warnings(ccWarnings), nil
}
