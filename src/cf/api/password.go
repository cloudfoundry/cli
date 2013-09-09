package api

import (
	"cf/configuration"
	"fmt"
	"net/url"
	"strings"
)

type PasswordRepository interface {
	GetScore(password string) (string, error)
	UpdatePassword(old string, new string) error
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

func (repo CloudControllerPasswordRepository) GetScore(password string) (score string, err error) {

	infoResponse, err := repo.getTargetInfo()
	if err != nil {
		return
	}

	scorePath := fmt.Sprintf("%s/password/score", infoResponse.TokenEndpoint)
	scoreBody := url.Values{
		"password": []string{password},
	}

	scoreRequest, err := NewRequest("POST", scorePath, repo.config.AccessToken, strings.NewReader(scoreBody.Encode()))
	if err != nil {
		return
	}
	scoreRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	scoreResponse := ScoreResponse{}

	_, err = repo.apiClient.PerformRequestAndParseResponse(scoreRequest, &scoreResponse)
	if err != nil {
		return
	}

	score = translateScoreResponse(scoreResponse)
	return
}

func (repo CloudControllerPasswordRepository) UpdatePassword(old string, new string) (err error) {
	infoResponse, err := repo.getTargetInfo()
	if err != nil {
		return
	}

	path := fmt.Sprintf("%s/Users/%s/password", infoResponse.TokenEndpoint, infoResponse.UserGuid)
	body := fmt.Sprintf(`{"password":"%s","oldPassword":"%s"}`, new, old)
	request, err := NewRequest("PUT", path, repo.config.AccessToken, strings.NewReader(body))
	if err != nil {
		return
	}

	request.Header.Set("Content-Type", "application/json")

	_, err = repo.apiClient.PerformRequest(request)
	return
}

func (repo *CloudControllerPasswordRepository) getTargetInfo() (response InfoResponse, err error) {
	if repo.infoResponse.UserGuid == "" {
		path := fmt.Sprintf("%s/info", repo.config.Target)
		var request *Request
		request, err = NewRequest("GET", path, repo.config.AccessToken, nil)
		if err != nil {
			return
		}

		response = InfoResponse{}

		_, err = repo.apiClient.PerformRequestAndParseResponse(request, &response)

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
