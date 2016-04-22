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
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
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

//go:generate counterfeiter . UserRepository

type UserRepository interface {
	FindByUsername(username string) (user models.UserFields, apiErr error)
	ListUsersInOrgForRole(orgGUID string, role string) ([]models.UserFields, error)
	ListUsersInOrgForRoleWithNoUAA(orgGUID string, role string) ([]models.UserFields, error)
	ListUsersInSpaceForRole(spaceGUID string, role string) ([]models.UserFields, error)
	ListUsersInSpaceForRoleWithNoUAA(spaceGUID string, role string) ([]models.UserFields, error)
	Create(username, password string) (apiErr error)
	CreateExternal(username, origin, externalID string) (apiErr error)
	Delete(userGUID string) (apiErr error)
	SetOrgRoleByGUID(userGUID, orgGUID, role string) (apiErr error)
	SetOrgRoleByUsername(username, orgGUID, role string) (apiErr error)
	UnsetOrgRoleByGUID(userGUID, orgGUID, role string) (apiErr error)
	UnsetOrgRoleByUsername(username, orgGUID, role string) (apiErr error)
	SetSpaceRoleByGUID(userGUID, spaceGUID, orgGUID, role string) (apiErr error)
	SetSpaceRoleByUsername(username, spaceGUID, orgGUID, role string) (apiErr error)
	UnsetSpaceRoleByGUID(userGUID, spaceGUID, role string) (apiErr error)
	UnsetSpaceRoleByUsername(userGUID, spaceGUID, role string) (apiErr error)
}

type CloudControllerUserRepository struct {
	config     coreconfig.Reader
	uaaGateway net.Gateway
	ccGateway  net.Gateway
}

func NewCloudControllerUserRepository(config coreconfig.Reader, uaaGateway net.Gateway, ccGateway net.Gateway) (repo CloudControllerUserRepository) {
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
		errType, ok := apiErr.(errors.HTTPError)
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

func (repo CloudControllerUserRepository) ListUsersInOrgForRole(orgGUID string, roleName string) (users []models.UserFields, apiErr error) {
	return repo.listUsersWithPath(fmt.Sprintf("/v2/organizations/%s/%s", orgGUID, orgRoleToPathMap[roleName]))
}

func (repo CloudControllerUserRepository) ListUsersInOrgForRoleWithNoUAA(orgGUID string, roleName string) (users []models.UserFields, apiErr error) {
	return repo.listUsersWithPathWithNoUAA(fmt.Sprintf("/v2/organizations/%s/%s", orgGUID, orgRoleToPathMap[roleName]))
}

func (repo CloudControllerUserRepository) ListUsersInSpaceForRole(spaceGUID string, roleName string) (users []models.UserFields, apiErr error) {
	return repo.listUsersWithPath(fmt.Sprintf("/v2/spaces/%s/%s", spaceGUID, spaceRoleToPathMap[roleName]))
}

func (repo CloudControllerUserRepository) ListUsersInSpaceForRoleWithNoUAA(spaceGUID string, roleName string) (users []models.UserFields, apiErr error) {
	return repo.listUsersWithPathWithNoUAA(fmt.Sprintf("/v2/spaces/%s/%s", spaceGUID, spaceRoleToPathMap[roleName]))
}

func (repo CloudControllerUserRepository) listUsersWithPathWithNoUAA(path string) (users []models.UserFields, apiErr error) {
	apiErr = repo.ccGateway.ListPaginatedResources(
		repo.config.APIEndpoint(),
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
		repo.config.APIEndpoint(),
		path,
		resources.UserResource{},
		func(resource interface{}) bool {
			user := resource.(resources.UserResource).ToFields()
			users = append(users, user)
			guidFilters = append(guidFilters, fmt.Sprintf(`ID eq "%s"`, user.GUID))
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
			if u.GUID == uaaResource.ID {
				ccUserFields = u
				break
			}
		}

		updatedUsers = append(updatedUsers, models.UserFields{
			GUID:     uaaResource.ID,
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
	case errors.HTTPError:
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
		GUID: createUserResponse.ID,
	})

	if err != nil {
		return
	}

	return repo.ccGateway.CreateResource(repo.config.APIEndpoint(), path, bytes.NewReader(body))
}

func (repo CloudControllerUserRepository) CreateExternal(username, origin, externalID string) (err error) {
	uaaEndpoint, err := repo.getAuthEndpoint()
	if err != nil {
		return
	}

	path := "/Users"
	body, err := json.Marshal(resources.NewExternalUserResource(username, origin, externalID))

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

func (repo CloudControllerUserRepository) Delete(userGUID string) (apiErr error) {
	path := fmt.Sprintf("/v2/users/%s", userGUID)

	apiErr = repo.ccGateway.DeleteResource(repo.config.APIEndpoint(), path)

	if httpErr, ok := apiErr.(errors.HTTPError); ok && httpErr.ErrorCode() != errors.UserNotFound {
		return
	}
	uaaEndpoint, apiErr := repo.getAuthEndpoint()
	if apiErr != nil {
		return
	}

	path = fmt.Sprintf("/Users/%s", userGUID)
	return repo.uaaGateway.DeleteResource(uaaEndpoint, path)
}

func (repo CloudControllerUserRepository) SetOrgRoleByGUID(userGUID string, orgGUID string, role string) (err error) {
	path, err := userGUIDPath(repo.config.APIEndpoint(), userGUID, orgGUID, role)
	if err != nil {
		return
	}
	err = repo.callApi("PUT", path, nil)
	if err != nil {
		return
	}
	return repo.assocUserWithOrgByUserGUID(userGUID, orgGUID)
}

func (repo CloudControllerUserRepository) UnsetOrgRoleByGUID(userGUID, orgGUID, role string) (err error) {
	path, err := userGUIDPath(repo.config.APIEndpoint(), userGUID, orgGUID, role)
	if err != nil {
		return
	}
	return repo.callApi("DELETE", path, nil)
}

func (repo CloudControllerUserRepository) UnsetOrgRoleByUsername(username, orgGUID, role string) error {
	rolePath, err := rolePath(role)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("%s/v2/organizations/%s/%s", repo.config.APIEndpoint(), orgGUID, rolePath)

	return repo.callApi("DELETE", path, usernamePayload(username))
}

func (repo CloudControllerUserRepository) UnsetSpaceRoleByUsername(username, spaceGUID, role string) error {
	rolePath := spaceRoleToPathMap[role]
	path := fmt.Sprintf("%s/v2/spaces/%s/%s", repo.config.APIEndpoint(), spaceGUID, rolePath)

	return repo.callApi("DELETE", path, usernamePayload(username))
}

func (repo CloudControllerUserRepository) SetOrgRoleByUsername(username string, orgGUID string, role string) error {
	rolePath, err := rolePath(role)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("%s/v2/organizations/%s/%s", repo.config.APIEndpoint(), orgGUID, rolePath)
	err = repo.callApi("PUT", path, usernamePayload(username))
	if err != nil {
		return err
	}
	return repo.assocUserWithOrgByUsername(username, orgGUID, nil)
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

func userGUIDPath(apiEndpoint, userGUID, orgGUID, role string) (string, error) {
	rolePath, err := rolePath(role)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/v2/organizations/%s/%s/%s", apiEndpoint, orgGUID, rolePath, userGUID), nil
}

func (repo CloudControllerUserRepository) SetSpaceRoleByGUID(userGUID, spaceGUID, orgGUID, role string) error {
	rolePath, found := spaceRoleToPathMap[role]
	if !found {
		return fmt.Errorf(T("Invalid Role {{.Role}}", map[string]interface{}{"Role": role}))
	}

	err := repo.assocUserWithOrgByUserGUID(userGUID, orgGUID)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/v2/spaces/%s/%s/%s", spaceGUID, rolePath, userGUID)

	return repo.ccGateway.UpdateResource(repo.config.APIEndpoint(), path, nil)
}

func (repo CloudControllerUserRepository) SetSpaceRoleByUsername(username, spaceGUID, orgGUID, role string) (apiErr error) {
	rolePath, apiErr := repo.checkSpaceRole(spaceGUID, role)
	if apiErr != nil {
		return
	}

	setOrgRoleErr := apiErrResponse{}
	apiErr = repo.assocUserWithOrgByUsername(username, orgGUID, &setOrgRoleErr)
	if setOrgRoleErr.Code == 10003 {
		//operator lacking the privilege to set org role
		//user might already be in org, so ignoring error and attempt to set space role
	} else if apiErr != nil {
		return
	}

	setSpaceRoleErr := apiErrResponse{}
	apiErr = repo.ccGateway.UpdateResourceSync(repo.config.APIEndpoint(), rolePath, usernamePayload(username), &setSpaceRoleErr)
	if setSpaceRoleErr.Code == 1002 {
		return errors.New(T("Server error, error code: 1002, message: cannot set space role because user is not part of the org"))
	}

	return apiErr
}

func (repo CloudControllerUserRepository) UnsetSpaceRoleByGUID(userGUID, spaceGUID, role string) error {
	rolePath, found := spaceRoleToPathMap[role]
	if !found {
		return fmt.Errorf(T("Invalid Role {{.Role}}", map[string]interface{}{"Role": role}))
	}

	return repo.ccGateway.DeleteResource(repo.config.APIEndpoint(), rolePath)
}

func (repo CloudControllerUserRepository) checkSpaceRole(spaceGUID, role string) (string, error) {
	var apiErr error

	rolePath, found := spaceRoleToPathMap[role]

	if !found {
		apiErr = fmt.Errorf(T("Invalid Role {{.Role}}",
			map[string]interface{}{"Role": role}))
	}

	apiPath := fmt.Sprintf("/v2/spaces/%s/%s", spaceGUID, rolePath)
	return apiPath, apiErr
}

func (repo CloudControllerUserRepository) assocUserWithOrgByUsername(username, orgGUID string, resource interface{}) (apiErr error) {
	path := fmt.Sprintf("/v2/organizations/%s/users", orgGUID)
	return repo.ccGateway.UpdateResourceSync(repo.config.APIEndpoint(), path, usernamePayload(username), resource)
}

func (repo CloudControllerUserRepository) assocUserWithOrgByUserGUID(userGUID, orgGUID string) (apiErr error) {
	path := fmt.Sprintf("/v2/organizations/%s/users/%s", orgGUID, userGUID)
	return repo.ccGateway.UpdateResource(repo.config.APIEndpoint(), path, nil)
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
