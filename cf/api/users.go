package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	neturl "net/url"
	"strings"

	"github.com/cloudfoundry/cli/cf/api/resources"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
)

var orgRoleToPathMap = map[string]string{
	models.ORG_USER:        "users",
	models.ORG_MANAGER:     "managers",
	models.BILLING_MANAGER: "billing_managers",
	models.ORG_AUDITOR:     "auditors",
}

var spaceRoleToPathMap = map[string]string{
	models.SPACE_MANAGER:   "managers",
	models.SPACE_DEVELOPER: "developers",
	models.SPACE_AUDITOR:   "auditors",
}

type UserRepository interface {
	FindByUsername(username string) (user models.UserFields, apiErr error)
	ListUsersInOrgForRole(orgGuid string, role string) ([]models.UserFields, error)
	ListUsersInOrgForRoleWithNoUAA(orgGuid string, role string) ([]models.UserFields, error)
	ListUsersInSpaceForRole(spaceGuid string, role string) ([]models.UserFields, error)
	ListUsersInSpaceForRoleWithNoUAA(spaceGuid string, role string) ([]models.UserFields, error)
	Create(username, password string) (apiErr error)
	Delete(userGuid string) (apiErr error)
	SetOrgRole(userGuid, orgGuid, role string) (apiErr error)
	UnsetOrgRole(userGuid, orgGuid, role string) (apiErr error)
	SetSpaceRole(userGuid, spaceGuid, orgGuid, role string) (apiErr error)
	UnsetSpaceRole(userGuid, spaceGuid, role string) (apiErr error)
}

type CloudControllerUserRepository struct {
	config     core_config.Reader
	uaaGateway net.Gateway
	ccGateway  net.Gateway
}

func NewCloudControllerUserRepository(config core_config.Reader, uaaGateway net.Gateway, ccGateway net.Gateway) (repo CloudControllerUserRepository) {
	repo.config = config
	repo.uaaGateway = uaaGateway
	repo.ccGateway = ccGateway
	return
}

func (repo CloudControllerUserRepository) FindByUsername(username string) (models.UserFields, error) {
	uaaEndpoint, apiErr := repo.getAuthEndpoint()
	var user models.UserFields
	if apiErr != nil {
		return user, apiErr
	}

	usernameFilter := neturl.QueryEscape(fmt.Sprintf(`userName Eq "%s"`, username))
	path := fmt.Sprintf("%s/Users?attributes=id,userName&filter=%s", uaaEndpoint, usernameFilter)
	users, apiErr := repo.updateOrFindUsersWithUAAPath([]models.UserFields{}, path)

	if apiErr != nil {
		errType, ok := apiErr.(errors.HttpError)
		if ok {
			if errType.StatusCode() == 403 {
				return user, errors.NewAccessDeniedError()
			}
		}
		return user, apiErr
	} else if len(users) == 0 {
		return user, errors.NewModelNotFoundError("User", username)
	}

	return users[0], apiErr
}

func (repo CloudControllerUserRepository) ListUsersInOrgForRole(orgGuid string, roleName string) (users []models.UserFields, apiErr error) {
	return repo.listUsersWithPath(fmt.Sprintf("/v2/organizations/%s/%s", orgGuid, orgRoleToPathMap[roleName]))
}

func (repo CloudControllerUserRepository) ListUsersInOrgForRoleWithNoUAA(orgGuid string, roleName string) (users []models.UserFields, apiErr error) {
	return repo.listUsersWithPathWithNoUAA(fmt.Sprintf("/v2/organizations/%s/%s", orgGuid, orgRoleToPathMap[roleName]))
}

func (repo CloudControllerUserRepository) ListUsersInSpaceForRole(spaceGuid string, roleName string) (users []models.UserFields, apiErr error) {
	return repo.listUsersWithPath(fmt.Sprintf("/v2/spaces/%s/%s", spaceGuid, spaceRoleToPathMap[roleName]))
}

func (repo CloudControllerUserRepository) ListUsersInSpaceForRoleWithNoUAA(spaceGuid string, roleName string) (users []models.UserFields, apiErr error) {
	return repo.listUsersWithPathWithNoUAA(fmt.Sprintf("/v2/spaces/%s/%s", spaceGuid, spaceRoleToPathMap[roleName]))
}

func (repo CloudControllerUserRepository) listUsersWithPathWithNoUAA(path string) (users []models.UserFields, apiErr error) {
	apiErr = repo.ccGateway.ListPaginatedResources(
		repo.config.ApiEndpoint(),
		path,
		resources.UserResource{},
		func(resource interface{}) bool {
			user := resource.(resources.UserResource).ToFields()
			users = append(users, user)
			return true
		})
	if apiErr != nil {
		return
	}

	return
}

func (repo CloudControllerUserRepository) listUsersWithPath(path string) (users []models.UserFields, apiErr error) {
	guidFilters := []string{}

	apiErr = repo.ccGateway.ListPaginatedResources(
		repo.config.ApiEndpoint(),
		path,
		resources.UserResource{},
		func(resource interface{}) bool {
			user := resource.(resources.UserResource).ToFields()
			users = append(users, user)
			guidFilters = append(guidFilters, fmt.Sprintf(`Id eq "%s"`, user.Guid))
			return true
		})
	if apiErr != nil {
		return
	}

	if len(guidFilters) == 0 {
		return
	}

	uaaEndpoint, apiErr := repo.getAuthEndpoint()
	if apiErr != nil {
		return
	}

	filter := strings.Join(guidFilters, " or ")
	usersURL := fmt.Sprintf("%s/Users?attributes=id,userName&filter=%s", uaaEndpoint, neturl.QueryEscape(filter))
	users, apiErr = repo.updateOrFindUsersWithUAAPath(users, usersURL)
	return
}

func (repo CloudControllerUserRepository) updateOrFindUsersWithUAAPath(ccUsers []models.UserFields, path string) (updatedUsers []models.UserFields, apiErr error) {
	uaaResponse := new(resources.UAAUserResources)
	apiErr = repo.uaaGateway.GetResource(path, uaaResponse)
	if apiErr != nil {
		return
	}

	for _, uaaResource := range uaaResponse.Resources {
		var ccUserFields models.UserFields

		for _, u := range ccUsers {
			if u.Guid == uaaResource.Id {
				ccUserFields = u
				break
			}
		}

		updatedUsers = append(updatedUsers, models.UserFields{
			Guid:     uaaResource.Id,
			Username: uaaResource.Username,
			IsAdmin:  ccUserFields.IsAdmin,
		})
	}
	return
}

func (repo CloudControllerUserRepository) Create(username, password string) (err error) {
	uaaEndpoint, err := repo.getAuthEndpoint()
	if err != nil {
		return
	}

	path := "/Users"
	body, err := json.Marshal(resources.NewUAAUserResource(username, password))

	if err != nil {
		return
	}

	createUserResponse := &resources.UAAUserFields{}
	err = repo.uaaGateway.CreateResource(uaaEndpoint, path, bytes.NewReader(body), createUserResponse)
	switch httpErr := err.(type) {
	case nil:
	case errors.HttpError:
		if httpErr.StatusCode() == http.StatusConflict {
			err = errors.NewModelAlreadyExistsError("user", username)
			return
		}
		return
	default:
		return
	}

	path = "/v2/users"
	body, err = json.Marshal(resources.Metadata{
		Guid: createUserResponse.Id,
	})

	if err != nil {
		return
	}

	return repo.ccGateway.CreateResource(repo.config.ApiEndpoint(), path, bytes.NewReader(body))
}

func (repo CloudControllerUserRepository) Delete(userGuid string) (apiErr error) {
	path := fmt.Sprintf("/v2/users/%s", userGuid)

	apiErr = repo.ccGateway.DeleteResource(repo.config.ApiEndpoint(), path)

	if httpErr, ok := apiErr.(errors.HttpError); ok && httpErr.ErrorCode() != errors.USER_NOT_FOUND {
		return
	}
	uaaEndpoint, apiErr := repo.getAuthEndpoint()
	if apiErr != nil {
		return
	}

	path = fmt.Sprintf("/Users/%s", userGuid)
	return repo.uaaGateway.DeleteResource(uaaEndpoint, path)
}

func (repo CloudControllerUserRepository) SetOrgRole(userGuid string, orgGuid string, role string) (apiErr error) {
	apiErr = repo.setOrUnsetOrgRole("PUT", userGuid, orgGuid, role)
	if apiErr != nil {
		return
	}
	return repo.addOrgUserRole(userGuid, orgGuid)
}

func (repo CloudControllerUserRepository) UnsetOrgRole(userGuid, orgGuid, role string) (apiErr error) {
	return repo.setOrUnsetOrgRole("DELETE", userGuid, orgGuid, role)
}

func (repo CloudControllerUserRepository) setOrUnsetOrgRole(verb, userGuid, orgGuid, role string) (apiErr error) {
	rolePath, found := orgRoleToPathMap[role]

	if !found {
		apiErr = errors.NewWithFmt(T("Invalid Role {{.Role}}",
			map[string]interface{}{"Role": role}))
		return
	}

	path := fmt.Sprintf("%s/v2/organizations/%s/%s/%s", repo.config.ApiEndpoint(), orgGuid, rolePath, userGuid)
	request, apiErr := repo.ccGateway.NewRequest(verb, path, repo.config.AccessToken(), nil)
	if apiErr != nil {
		return
	}

	_, apiErr = repo.ccGateway.PerformRequest(request)
	if apiErr != nil {
		return
	}
	return
}

func (repo CloudControllerUserRepository) SetSpaceRole(userGuid, spaceGuid, orgGuid, role string) (apiErr error) {
	rolePath, apiErr := repo.checkSpaceRole(userGuid, spaceGuid, role)
	if apiErr != nil {
		return
	}

	apiErr = repo.addOrgUserRole(userGuid, orgGuid)
	if apiErr != nil {
		return
	}

	return repo.ccGateway.UpdateResource(repo.config.ApiEndpoint(), rolePath, nil)
}

func (repo CloudControllerUserRepository) UnsetSpaceRole(userGuid, spaceGuid, role string) (apiErr error) {
	rolePath, apiErr := repo.checkSpaceRole(userGuid, spaceGuid, role)
	if apiErr != nil {
		return
	}
	return repo.ccGateway.DeleteResource(repo.config.ApiEndpoint(), rolePath)
}

func (repo CloudControllerUserRepository) checkSpaceRole(userGuid, spaceGuid, role string) (string, error) {
	var apiErr error

	rolePath, found := spaceRoleToPathMap[role]

	if !found {
		apiErr = errors.NewWithFmt(T("Invalid Role {{.Role}}",
			map[string]interface{}{"Role": role}))
	}

	apiPath := fmt.Sprintf("/v2/spaces/%s/%s/%s", spaceGuid, rolePath, userGuid)
	return apiPath, apiErr
}

func (repo CloudControllerUserRepository) addOrgUserRole(userGuid, orgGuid string) (apiErr error) {
	path := fmt.Sprintf("/v2/organizations/%s/users/%s", orgGuid, userGuid)
	return repo.ccGateway.UpdateResource(repo.config.ApiEndpoint(), path, nil)
}

func (repo CloudControllerUserRepository) getAuthEndpoint() (string, error) {
	uaaEndpoint := repo.config.UaaEndpoint()
	if uaaEndpoint == "" {
		return "", errors.New(T("UAA endpoint missing from config file"))
	}
	return uaaEndpoint, nil
}
