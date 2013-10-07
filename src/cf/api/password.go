package api

import (
	"cf/configuration"
	"cf/net"
	"fmt"
	"net/url"
	"strings"
)

type PasswordRepository interface {
	GetScore(password string) (string, net.ApiStatus)
	UpdatePassword(old string, new string) net.ApiStatus
}

type CloudControllerPasswordRepository struct {
	config  *configuration.Configuration
	gateway net.Gateway

	infoResponse InfoResponse
}

func NewCloudControllerPasswordRepository(config *configuration.Configuration, gateway net.Gateway) (repo CloudControllerPasswordRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

type ScoreResponse struct {
	Score         int
	RequiredScore int
}

type InfoResponse struct {
	TokenEndpoint string `json:"token_endpoint"`
	UserGuid      string `json:"user"`
}

func (repo CloudControllerPasswordRepository) GetScore(password string) (score string, apiStatus net.ApiStatus) {
	infoResponse, apiStatus := repo.getTargetInfo()
	if apiStatus.IsNotSuccessful() {
		return
	}

	scorePath := fmt.Sprintf("%s/password/score", infoResponse.TokenEndpoint)
	scoreBody := url.Values{
		"password": []string{password},
	}

	scoreRequest, apiStatus := repo.gateway.NewRequest("POST", scorePath, repo.config.AccessToken, strings.NewReader(scoreBody.Encode()))
	if apiStatus.IsNotSuccessful() {
		return
	}
	scoreRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	scoreResponse := ScoreResponse{}

	_, apiStatus = repo.gateway.PerformRequestForJSONResponse(scoreRequest, &scoreResponse)
	if apiStatus.IsNotSuccessful() {
		return
	}

	score = translateScoreResponse(scoreResponse)
	return
}

func (repo CloudControllerPasswordRepository) UpdatePassword(old string, new string) (apiStatus net.ApiStatus) {
	infoResponse, apiStatus := repo.getTargetInfo()
	if apiStatus.IsNotSuccessful() {
		return
	}

	path := fmt.Sprintf("%s/Users/%s/password", infoResponse.TokenEndpoint, repo.config.UserGuid())
	body := fmt.Sprintf(`{"password":"%s","oldPassword":"%s"}`, new, old)
	request, apiStatus := repo.gateway.NewRequest("PUT", path, repo.config.AccessToken, strings.NewReader(body))
	if apiStatus.IsNotSuccessful() {
		return
	}

	request.Header.Set("Content-Type", "application/json")

	apiStatus = repo.gateway.PerformRequest(request)
	return
}

func (repo *CloudControllerPasswordRepository) getTargetInfo() (response InfoResponse, apiStatus net.ApiStatus) {
	if repo.infoResponse.UserGuid == "" {
		path := fmt.Sprintf("%s/info", repo.config.Target)
		var request *net.Request
		request, apiStatus = repo.gateway.NewRequest("GET", path, repo.config.AccessToken, nil)
		if apiStatus.IsNotSuccessful() {
			return
		}

		response = InfoResponse{}

		_, apiStatus = repo.gateway.PerformRequestForJSONResponse(request, &response)

		repo.infoResponse = response
	}

	response = repo.infoResponse

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
