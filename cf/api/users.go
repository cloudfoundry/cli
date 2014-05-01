package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/cloudfoundry/cli/cf/api/resources"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
	"net/http"
	neturl "net/url"
	"strings"
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
	ListUsersInSpaceForRole(spaceGuid string, role string) ([]models.UserFields, error)
	Create(username, password string) (apiErr error)
	Delete(userGuid string) (apiErr error)
	SetOrgRole(userGuid, orgGuid, role string) (apiErr error)
	UnsetOrgRole(userGuid, orgGuid, role string) (apiErr error)
	SetSpaceRole(userGuid, spaceGuid, orgGuid, role string) (apiErr error)
	UnsetSpaceRole(userGuid, spaceGuid, role string) (apiErr error)
}

type CloudControllerUserRepository struct {
	config     configuration.Reader
	uaaGateway net.Gateway
	ccGateway  net.Gateway
}

func NewCloudControllerUserRepository(config configuration.Reader, uaaGateway net.Gateway, ccGateway net.Gateway) (repo CloudControllerUserRepository) {
	repo.config = config
	repo.uaaGateway = uaaGateway
	repo.ccGateway = ccGateway
	return
}

func (repo CloudControllerUserRepository) FindByUsername(username string) (user models.UserFields, apiErr error) {
	uaaEndpoint, apiErr := repo.getAuthEndpoint()
	if apiErr != nil {
		return
	}

	usernameFilter := neturl.QueryEscape(fmt.Sprintf(`userName Eq "%s"`, username))
	path := fmt.Sprintf("%s/Users?attributes=id,userName&filter=%s", uaaEndpoint, usernameFilter)

	users, apiErr := repo.updateOrFindUsersWithUAAPath([]models.UserFields{}, path)
	if len(users) == 0 {
		apiErr = errors.NewModelNotFoundError("User", username)
		return
	}

	user = users[0]
	return
}

func (repo CloudControllerUserRepository) ListUsersInOrgForRole(orgGuid string, roleName string) (users []models.UserFields, apiErr error) {
	return repo.listUsersWithPath(fmt.Sprintf("/v2/organizations/%s/%s", orgGuid, orgRoleToPathMap[roleName]))
}

func (repo CloudControllerUserRepository) ListUsersInSpaceForRole(spaceGuid string, roleName string) (users []models.UserFields, apiErr error) {
	return repo.listUsersWithPath(fmt.Sprintf("/v2/spaces/%s/%s", spaceGuid, spaceRoleToPathMap[roleName]))
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

	path := fmt.Sprintf("%s/Users", uaaEndpoint)
	body, err := json.Marshal(resources.NewUAAUserResource(username, password))

	if err != nil {
		return
	}

	createUserResponse := &resources.UAAUserFields{}
	err = repo.uaaGateway.CreateResource(path, bytes.NewReader(body), createUserResponse)
	switch httpErr := err.(type) {
	case nil:
	case errors.HttpError:
		if httpErr.StatusCode() == http.StatusConflict {
			err = errors.NewModelAlreadyExistsError("user", username)
			return
		}
	default:
		return
	}

	path = fmt.Sprintf("%s/v2/users", repo.config.ApiEndpoint())
	body, err = json.Marshal(resources.Metadata{
		Guid: createUserResponse.Id,
	})

	if err != nil {
		return
	}

	return repo.ccGateway.CreateResource(path, bytes.NewReader(body))
}

func (repo CloudControllerUserRepository) Delete(userGuid string) (apiErr error) {
	path := fmt.Sprintf("%s/v2/users/%s", repo.config.ApiEndpoint(), userGuid)

	apiErr = repo.ccGateway.DeleteResource(path)

	if httpErr, ok := apiErr.(errors.HttpError); ok && httpErr.ErrorCode() != errors.USER_NOT_FOUND {
		return
	}
	uaaEndpoint, apiErr := repo.getAuthEndpoint()
	if apiErr != nil {
		return
	}

	path = fmt.Sprintf("%s/Users/%s", uaaEndpoint, userGuid)
	return repo.uaaGateway.DeleteResource(path)
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
		apiErr = errors.NewWithFmt("Invalid Role %s", role)
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

	return repo.ccGateway.UpdateResource(rolePath, nil)
}

func (repo CloudControllerUserRepository) UnsetSpaceRole(userGuid, spaceGuid, role string) (apiErr error) {
	rolePath, apiErr := repo.checkSpaceRole(userGuid, spaceGuid, role)
	if apiErr != nil {
		return
	}
	return repo.ccGateway.DeleteResource(rolePath)
}

func (repo CloudControllerUserRepository) checkSpaceRole(userGuid, spaceGuid, role string) (fullPath string, apiErr error) {
	rolePath, found := spaceRoleToPathMap[role]

	if !found {
		apiErr = errors.NewWithFmt("Invalid Role %s", role)
	}

	fullPath = fmt.Sprintf("%s/v2/spaces/%s/%s/%s", repo.config.ApiEndpoint(), spaceGuid, rolePath, userGuid)
	return
}

func (repo CloudControllerUserRepository) addOrgUserRole(userGuid, orgGuid string) (apiErr error) {
	path := fmt.Sprintf("%s/v2/organizations/%s/users/%s", repo.config.ApiEndpoint(), orgGuid, userGuid)
	return repo.ccGateway.UpdateResource(path, nil)
}

func (repo CloudControllerUserRepository) getAuthEndpoint() (string, error) {
	uaaEndpoint := repo.config.UaaEndpoint()
	if uaaEndpoint == "" {
		return "", errors.New("UAA endpoint missing from config file")
	}
	return uaaEndpoint, nil
}
