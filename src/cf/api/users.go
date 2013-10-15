package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"fmt"
	"strings"
)

type UserRepository interface {
	FindByUsername(username string) (user cf.User, apiResponse net.ApiResponse)
	Create(user cf.User) (apiResponse net.ApiResponse)
	Delete(user cf.User) (apiResponse net.ApiResponse)
	SetOrgRole(user cf.User, org cf.Organization, role string) (apiResponse net.ApiResponse)
	UnsetOrgRole(user cf.User, org cf.Organization, role string) (apiResponse net.ApiResponse)
}

type CloudControllerUserRepository struct {
	config       *configuration.Configuration
	uaaGateway   net.Gateway
	ccGateway    net.Gateway
	endpointRepo EndpointRepository
}

func NewCloudControllerUserRepository(config *configuration.Configuration, uaaGateway net.Gateway, ccGateway net.Gateway, endpointRepo EndpointRepository) (repo CloudControllerUserRepository) {
	repo.config = config
	repo.uaaGateway = uaaGateway
	repo.ccGateway = ccGateway
	repo.endpointRepo = endpointRepo
	return
}

func (repo CloudControllerUserRepository) FindByUsername(username string) (user cf.User, apiResponse net.ApiResponse) {
	uaaEndpoint, apiResponse := repo.endpointRepo.GetEndpoint(cf.UaaEndpointKey)
	if apiResponse.IsNotSuccessful() {
		return
	}

	path := fmt.Sprintf("%s/Users?attributes=id,userName&filter=userName%%20Eq%%20\"%s\"", uaaEndpoint, username)

	request, apiResponse := repo.uaaGateway.NewRequest("GET", path, repo.config.AccessToken, nil)
	if apiResponse.IsNotSuccessful() {
		return
	}

	type uaaUserResource struct {
		Id       string
		UserName string
	}
	type uaaUserResources struct {
		Resources []uaaUserResource
	}

	findUserResponse := uaaUserResources{}
	_, apiResponse = repo.uaaGateway.PerformRequestForJSONResponse(request, &findUserResponse)
	if apiResponse.IsNotSuccessful() {
		return
	}

	if len(findUserResponse.Resources) == 0 {
		apiResponse = net.NewNotFoundApiResponse("User %s not found", username)
		return
	}

	userResource := findUserResponse.Resources[0]
	user.Username = userResource.UserName
	user.Guid = userResource.Id

	return
}

func (repo CloudControllerUserRepository) Create(user cf.User) (apiResponse net.ApiResponse) {
	uaaEndpoint, apiResponse := repo.endpointRepo.GetEndpoint(cf.UaaEndpointKey)
	if apiResponse.IsNotSuccessful() {
		return
	}

	path := fmt.Sprintf("%s/Users", uaaEndpoint)
	body := fmt.Sprintf(`{
  "userName": "%s",
  "emails": [{"value":"%s"}],
  "password": "%s",
  "name": {"givenName":"%s", "familyName":"%s"}
}`,
		user.Username,
		user.Username,
		user.Password,
		user.Username,
		user.Username,
	)
	request, apiResponse := repo.uaaGateway.NewRequest("POST", path, repo.config.AccessToken, strings.NewReader(body))
	if apiResponse.IsNotSuccessful() {
		return
	}

	type uaaUser struct {
		Id string
	}
	createUserResponse := &uaaUser{}

	_, apiResponse = repo.uaaGateway.PerformRequestForJSONResponse(request, createUserResponse)
	if apiResponse.IsNotSuccessful() {
		return
	}

	path = fmt.Sprintf("%s/v2/users", repo.config.Target)
	body = fmt.Sprintf(`{"guid":"%s"}`, createUserResponse.Id)

	request, apiResponse = repo.ccGateway.NewRequest("POST", path, repo.config.AccessToken, strings.NewReader(body))
	if apiResponse.IsNotSuccessful() {
		return
	}

	apiResponse = repo.ccGateway.PerformRequest(request)
	return
}

func (repo CloudControllerUserRepository) Delete(user cf.User) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/users/%s", repo.config.Target, user.Guid)

	request, apiResponse := repo.ccGateway.NewRequest("DELETE", path, repo.config.AccessToken, nil)
	if apiResponse.IsNotSuccessful() {
		return
	}

	apiResponse = repo.ccGateway.PerformRequest(request)
	if apiResponse.IsNotSuccessful() && apiResponse.ErrorCode != cf.USER_NOT_FOUND {
		return
	}

	uaaEndpoint, apiResponse := repo.endpointRepo.GetEndpoint(cf.UaaEndpointKey)
	if apiResponse.IsNotSuccessful() {
		return
	}

	path = fmt.Sprintf("%s/Users/%s", uaaEndpoint, user.Guid)
	request, apiResponse = repo.uaaGateway.NewRequest("DELETE", path, repo.config.AccessToken, nil)
	if apiResponse.IsNotSuccessful() {
		return
	}

	apiResponse = repo.uaaGateway.PerformRequest(request)
	return
}

func (repo CloudControllerUserRepository) SetOrgRole(user cf.User, org cf.Organization, role string) (apiResponse net.ApiResponse) {
	return repo.setOrUnsetOrgRole("PUT", user, org, role)
}

func (repo CloudControllerUserRepository) UnsetOrgRole(user cf.User, org cf.Organization, role string) (apiResponse net.ApiResponse) {
	return repo.setOrUnsetOrgRole("DELETE", user, org, role)
}

func (repo CloudControllerUserRepository) setOrUnsetOrgRole(verb string, user cf.User, org cf.Organization, role string) (apiResponse net.ApiResponse) {
	roleToPathMap := map[string]string{
		"OrgManager":     "managers",
		"BillingManager": "billing_managers",
		"OrgAuditor":     "auditors",
	}

	rolePath, found := roleToPathMap[role]

	if !found {
		apiResponse = net.NewApiResponseWithMessage("Invalid Role %s", role)
		return
	}

	path := fmt.Sprintf("%s/v2/organizations/%s/%s/%s", repo.config.Target, org.Guid, rolePath, user.Guid)

	request, apiResponse := repo.ccGateway.NewRequest(verb, path, repo.config.AccessToken, nil)
	if apiResponse.IsNotSuccessful() {
		return
	}

	apiResponse = repo.ccGateway.PerformRequest(request)
	if apiResponse.IsNotSuccessful() {
		return
	}
	return
}
