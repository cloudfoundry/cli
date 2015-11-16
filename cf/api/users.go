package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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

type apiErrResponse struct {
	Code        int    `json:"code,omitempty"`
	ErrorCode   string `json:"error_code,omitempty"`
	Description string `json:"description,omitempty"`
}

//go:generate counterfeiter -o fakes/fake_user_repo.go . UserRepository
type UserRepository interface {
	FindByUsername(username string) (user models.UserFields, apiErr error)
	ListUsersInOrgForRole(orgGuid string, role string) ([]models.UserFields, error)
	ListUsersInOrgForRoleWithNoUAA(orgGuid string, role string) ([]models.UserFields, error)
	ListUsersInSpaceForRole(spaceGuid string, role string) ([]models.UserFields, error)
	ListUsersInSpaceForRoleWithNoUAA(spaceGuid string, role string) ([]models.UserFields, error)
	Create(username, password string) (apiErr error)
	Delete(userGuid string) (apiErr error)
	SetOrgRoleByGuid(userGuid, orgGuid, role string) (apiErr error)
	SetOrgRoleByUsername(username, orgGuid, role string) (apiErr error)
	UnsetOrgRoleByGuid(userGuid, orgGuid, role string) (apiErr error)
	UnsetOrgRoleByUsername(username, orgGuid, role string) (apiErr error)
	SetSpaceRoleByGuid(userGuid, spaceGuid, orgGuid, role string) (apiErr error)
	SetSpaceRoleByUsername(username, spaceGuid, orgGuid, role string) (apiErr error)
	UnsetSpaceRoleByGuid(userGuid, spaceGuid, role string) (apiErr error)
	UnsetSpaceRoleByUsername(userGuid, spaceGuid, role string) (apiErr error)
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

func (repo CloudControllerUserRepository) SetOrgRoleByGuid(userGuid string, orgGuid string, role string) (err error) {
	path, err := userGuidPath(repo.config.ApiEndpoint(), userGuid, orgGuid, role)
	if err != nil {
		return
	}
	err = repo.callApi("PUT", path, nil)
	if err != nil {
		return
	}
	return repo.assocUserWithOrgByUserGuid(userGuid, orgGuid)
}

func (repo CloudControllerUserRepository) UnsetOrgRoleByGuid(userGuid, orgGuid, role string) (err error) {
	path, err := userGuidPath(repo.config.ApiEndpoint(), userGuid, orgGuid, role)
	if err != nil {
		return
	}
	return repo.callApi("DELETE", path, nil)
}

func (repo CloudControllerUserRepository) UnsetOrgRoleByUsername(username, orgGuid, role string) error {
	rolePath, err := rolePath(role)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("%s/v2/organizations/%s/%s", repo.config.ApiEndpoint(), orgGuid, rolePath)

	return repo.callApi("DELETE", path, usernamePayload(username))
}

func (repo CloudControllerUserRepository) UnsetSpaceRoleByUsername(username, spaceGuid, role string) error {
	rolePath := spaceRoleToPathMap[role]
	path := fmt.Sprintf("%s/v2/spaces/%s/%s", repo.config.ApiEndpoint(), spaceGuid, rolePath)

	return repo.callApi("DELETE", path, usernamePayload(username))
}

func (repo CloudControllerUserRepository) SetOrgRoleByUsername(username string, orgGuid string, role string) error {
	rolePath, err := rolePath(role)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("%s/v2/organizations/%s/%s", repo.config.ApiEndpoint(), orgGuid, rolePath)
	err = repo.callApi("PUT", path, usernamePayload(username))
	if err != nil {
		return err
	}
	return repo.assocUserWithOrgByUsername(username, orgGuid, nil)
}

func (repo CloudControllerUserRepository) callApi(verb, path string, body io.ReadSeeker) (err error) {
	request, err := repo.ccGateway.NewRequest(verb, path, repo.config.AccessToken(), body)
	if err != nil {
		return
	}
	_, err = repo.ccGateway.PerformRequest(request)
	if err != nil {
		return
	}
	return
}

func userGuidPath(apiEndpoint, userGuid, orgGuid, role string) (string, error) {
	rolePath, err := rolePath(role)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/v2/organizations/%s/%s/%s", apiEndpoint, orgGuid, rolePath, userGuid), nil
}

func (repo CloudControllerUserRepository) SetSpaceRoleByGuid(userGuid, spaceGuid, orgGuid, role string) error {
	rolePath, found := spaceRoleToPathMap[role]
	if !found {
		return fmt.Errorf(T("Invalid Role {{.Role}}", map[string]interface{}{"Role": role}))
	}

	err := repo.assocUserWithOrgByUserGuid(userGuid, orgGuid)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/v2/spaces/%s/%s/%s", spaceGuid, rolePath, userGuid)

	return repo.ccGateway.UpdateResource(repo.config.ApiEndpoint(), path, nil)
}

func (repo CloudControllerUserRepository) SetSpaceRoleByUsername(username, spaceGuid, orgGuid, role string) (apiErr error) {
	rolePath, apiErr := repo.checkSpaceRole(spaceGuid, role)
	if apiErr != nil {
		return
	}

	setOrgRoleErr := apiErrResponse{}
	apiErr = repo.assocUserWithOrgByUsername(username, orgGuid, &setOrgRoleErr)
	if setOrgRoleErr.Code == 10003 {
		//operator lacking the privilege to set org role
		//user might already be in org, so ignoring error and attempt to set space role
	} else if apiErr != nil {
		return
	}

	setSpaceRoleErr := apiErrResponse{}
	apiErr = repo.ccGateway.UpdateResourceSync(repo.config.ApiEndpoint(), rolePath, usernamePayload(username), &setSpaceRoleErr)
	if setSpaceRoleErr.Code == 1002 {
		return errors.New(T("Server error, error code: 1002, message: cannot set space role because user is not part of the org"))
	}

	return apiErr
}

func (repo CloudControllerUserRepository) UnsetSpaceRoleByGuid(userGuid, spaceGuid, role string) error {
	rolePath, found := spaceRoleToPathMap[role]
	if !found {
		return fmt.Errorf(T("Invalid Role {{.Role}}", map[string]interface{}{"Role": role}))
	}

	return repo.ccGateway.DeleteResource(repo.config.ApiEndpoint(), rolePath)
}

func (repo CloudControllerUserRepository) checkSpaceRole(spaceGuid, role string) (string, error) {
	var apiErr error

	rolePath, found := spaceRoleToPathMap[role]

	if !found {
		apiErr = fmt.Errorf(T("Invalid Role {{.Role}}",
			map[string]interface{}{"Role": role}))
	}

	apiPath := fmt.Sprintf("/v2/spaces/%s/%s", spaceGuid, rolePath)
	return apiPath, apiErr
}

func (repo CloudControllerUserRepository) assocUserWithOrgByUsername(username, orgGuid string, resource interface{}) (apiErr error) {
	path := fmt.Sprintf("/v2/organizations/%s/users", orgGuid)
	return repo.ccGateway.UpdateResourceSync(repo.config.ApiEndpoint(), path, usernamePayload(username), resource)
}

func (repo CloudControllerUserRepository) assocUserWithOrgByUserGuid(userGuid, orgGuid string) (apiErr error) {
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

func rolePath(role string) (string, error) {
	path, found := orgRoleToPathMap[role]

	if !found {
		return "", fmt.Errorf(T("Invalid Role {{.Role}}",
			map[string]interface{}{"Role": role}))
	}
	return path, nil
}

func usernamePayload(username string) *strings.Reader {
	return strings.NewReader(`{"username": "` + username + `"}`)
}
