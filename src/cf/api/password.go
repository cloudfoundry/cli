package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"fmt"
	"net/url"
	"strings"
)

type PasswordRepository interface {
	GetScore(password string) (string, net.ApiResponse)
	UpdatePassword(old string, new string) net.ApiResponse
}

type CloudControllerPasswordRepository struct {
	config       *configuration.Configuration
	gateway      net.Gateway
	endpointRepo EndpointRepository
}

func NewCloudControllerPasswordRepository(config *configuration.Configuration, gateway net.Gateway, endpointRepo EndpointRepository) (repo CloudControllerPasswordRepository) {
	repo.config = config
	repo.gateway = gateway
	repo.endpointRepo = endpointRepo
	return
}

type ScoreResponse struct {
	Score         int
	RequiredScore int
}

func (repo CloudControllerPasswordRepository) GetScore(password string) (score string, apiResponse net.ApiResponse) {
	uaaEndpoint, apiResponse := repo.endpointRepo.GetEndpoint(cf.UaaEndpointKey)
	if apiResponse.IsNotSuccessful() {
		return
	}

	scorePath := fmt.Sprintf("%s/password/score", uaaEndpoint)
	scoreBody := url.Values{
		"password": []string{password},
	}

	scoreRequest, apiResponse := repo.gateway.NewRequest("POST", scorePath, repo.config.AccessToken, strings.NewReader(scoreBody.Encode()))
	if apiResponse.IsNotSuccessful() {
		return
	}
	scoreRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	scoreResponse := ScoreResponse{}

	_, apiResponse = repo.gateway.PerformRequestForJSONResponse(scoreRequest, &scoreResponse)
	if apiResponse.IsNotSuccessful() {
		return
	}

	score = translateScoreResponse(scoreResponse)
	return
}

func (repo CloudControllerPasswordRepository) UpdatePassword(old string, new string) (apiResponse net.ApiResponse) {
	uaaEndpoint, apiResponse := repo.endpointRepo.GetEndpoint(cf.UaaEndpointKey)
	if apiResponse.IsNotSuccessful() {
		return
	}

	path := fmt.Sprintf("%s/Users/%s/password", uaaEndpoint, repo.config.UserGuid())
	body := fmt.Sprintf(`{"password":"%s","oldPassword":"%s"}`, new, old)
	request, apiResponse := repo.gateway.NewRequest("PUT", path, repo.config.AccessToken, strings.NewReader(body))
	if apiResponse.IsNotSuccessful() {
		return
	}

	request.Header.Set("Content-Type", "application/json")

	apiResponse = repo.gateway.PerformRequest(request)
	return
}

func translateScoreResponse(response ScoreResponse) string {
	if response.Score == 10 {
		return "strong"
	}

	if response.Score >= response.RequiredScore {
		return "good"
	}

	return "weak"
}
