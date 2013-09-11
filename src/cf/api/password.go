package api

import (
	"cf/configuration"
	"fmt"
	"net/url"
	"strings"
)

type PasswordRepository interface {
	GetScore(password string) (string, *ApiError)
	UpdatePassword(old string, new string) *ApiError
}

type CloudControllerPasswordRepository struct {
	config    *configuration.Configuration
	apiClient ApiClient

	infoResponse InfoResponse
}

func NewCloudControllerPasswordRepository(config *configuration.Configuration, apiClient ApiClient) (repo CloudControllerPasswordRepository) {
	repo.config = config
	repo.apiClient = apiClient
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

func (repo CloudControllerPasswordRepository) GetScore(password string) (score string, apiErr *ApiError) {
	infoResponse, apiErr := repo.getTargetInfo()
	if apiErr != nil {
		return
	}

	scorePath := fmt.Sprintf("%s/password/score", infoResponse.TokenEndpoint)
	scoreBody := url.Values{
		"password": []string{password},
	}

	scoreRequest, apiErr := NewRequest("POST", scorePath, repo.config.AccessToken, strings.NewReader(scoreBody.Encode()))
	if apiErr != nil {
		return
	}
	scoreRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	scoreResponse := ScoreResponse{}

	apiErr = repo.apiClient.PerformRequestAndParseResponse(scoreRequest, &scoreResponse)
	if apiErr != nil {
		return
	}

	score = translateScoreResponse(scoreResponse)
	return
}

func (repo CloudControllerPasswordRepository) UpdatePassword(old string, new string) (apiErr *ApiError) {
	infoResponse, apiErr := repo.getTargetInfo()
	if apiErr != nil {
		return
	}

	path := fmt.Sprintf("%s/Users/%s/password", infoResponse.TokenEndpoint, repo.config.UserGuid())
	body := fmt.Sprintf(`{"password":"%s","oldPassword":"%s"}`, new, old)
	request, apiErr := NewRequest("PUT", path, repo.config.AccessToken, strings.NewReader(body))
	if apiErr != nil {
		return
	}

	request.Header.Set("Content-Type", "application/json")

	apiErr = repo.apiClient.PerformRequest(request)
	return
}

func (repo *CloudControllerPasswordRepository) getTargetInfo() (response InfoResponse, apiErr *ApiError) {
	if repo.infoResponse.UserGuid == "" {
		path := fmt.Sprintf("%s/info", repo.config.Target)
		var request *Request
		request, apiErr = NewRequest("GET", path, repo.config.AccessToken, nil)
		if apiErr != nil {
			return
		}

		response = InfoResponse{}

		apiErr = repo.apiClient.PerformRequestAndParseResponse(request, &response)

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
