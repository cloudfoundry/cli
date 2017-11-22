package authenticators

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"code.cloudfoundry.org/lager"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/ssh"
)

type CFAuthenticator struct {
	logger             lager.Logger
	httpClient         *http.Client
	ccURL              string
	uaaTokenURL        string
	uaaPassword        string
	uaaUsername        string
	permissionsBuilder PermissionsBuilder
}

type AppSSHResponse struct {
	ProcessGuid string `json:"process_guid"`
}

type UAAAuthTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

var CFUserRegex *regexp.Regexp = regexp.MustCompile(`cf:([0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12})/(\d+)`)

func NewCFAuthenticator(
	logger lager.Logger,
	httpClient *http.Client,
	ccURL string,
	uaaTokenURL string,
	uaaUsername string,
	uaaPassword string,
	permissionsBuilder PermissionsBuilder,
) *CFAuthenticator {
	return &CFAuthenticator{
		logger:             logger,
		httpClient:         httpClient,
		ccURL:              ccURL,
		uaaTokenURL:        uaaTokenURL,
		uaaUsername:        uaaUsername,
		uaaPassword:        uaaPassword,
		permissionsBuilder: permissionsBuilder,
	}
}

func (cfa *CFAuthenticator) UserRegexp() *regexp.Regexp {
	return CFUserRegex
}

func (cfa *CFAuthenticator) Authenticate(metadata ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
	logger := cfa.logger.Session("cf-authenticate")
	logger.Info("authenticate-starting")
	defer logger.Info("authenticate-finished")

	if !CFUserRegex.MatchString(metadata.User()) {
		logger.Error("regex-match-fail", InvalidCredentialsErr)
		return nil, InvalidCredentialsErr
	}

	guidAndIndex := CFUserRegex.FindStringSubmatch(metadata.User())

	appGuid := guidAndIndex[1]

	index, err := strconv.Atoi(guidAndIndex[2])
	if err != nil {
		logger.Error("atoi-failed", err)
		return nil, InvalidCredentialsErr
	}

	cred, err := cfa.exchangeAccessCodeForToken(logger, string(password))
	if err != nil {
		return nil, err
	}

	parts := strings.Split(cred, " ")
	if len(parts) != 2 {
		return nil, AuthenticationFailedErr
	}
	tokenString := parts[1]
	// When parsing the certificate validating the signature is not required and we don't readily have the
	// certificate to validate the signature.  This is just to parse the second information part of the token anyway.
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte("Doesntmatter"), nil
	})

	username, ok := token.Claims["user_name"].(string)
	if !ok {
		username = "unknown"
	}
	principal, ok := token.Claims["user_id"].(string)
	if !ok {
		principal = "unknown"
	}

	logger = logger.WithData(lager.Data{
		"app":       fmt.Sprintf("%s/%d", appGuid, index),
		"principal": principal,
		"username":  username,
	})

	processGuid, err := cfa.checkAccess(logger, appGuid, index, string(cred))
	if err != nil {
		return nil, err
	}

	permissions, err := cfa.permissionsBuilder.Build(logger, processGuid, index, metadata)
	if err != nil {
		logger.Error("building-ssh-permissions-failed", err)
	}

	logger.Info("app-access-success")

	return permissions, err
}

func (cfa *CFAuthenticator) exchangeAccessCodeForToken(logger lager.Logger, code string) (string, error) {
	logger = logger.Session("exchange-access-code-for-token")

	formValues := make(url.Values)
	formValues.Set("grant_type", "authorization_code")
	formValues.Set("code", code)

	req, err := http.NewRequest("POST", cfa.uaaTokenURL, strings.NewReader(formValues.Encode()))
	if err != nil {
		return "", err
	}

	req.SetBasicAuth(cfa.uaaUsername, cfa.uaaPassword)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := cfa.httpClient.Do(req)
	if err != nil {
		logger.Error("request-failed", err)
		return "", AuthenticationFailedErr
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Error("response-status-not-ok", AuthenticationFailedErr, lager.Data{
			"status-code": resp.StatusCode,
		})
		return "", AuthenticationFailedErr
	}

	var tokenResponse UAAAuthTokenResponse
	err = json.NewDecoder(resp.Body).Decode(&tokenResponse)
	if err != nil {
		logger.Error("decode-token-response-failed", err)
		return "", AuthenticationFailedErr
	}

	return fmt.Sprintf("%s %s", tokenResponse.TokenType, tokenResponse.AccessToken), nil
}

func (cfa *CFAuthenticator) checkAccess(logger lager.Logger, appGuid string, index int, token string) (string, error) {
	path := fmt.Sprintf("%s/internal/apps/%s/ssh_access/%d", cfa.ccURL, appGuid, index)

	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		logger.Error("creating-request-failed", InvalidRequestErr)
		return "", InvalidRequestErr
	}
	req.Header.Add("Authorization", token)

	resp, err := cfa.httpClient.Do(req)
	if err != nil {
		logger.Error("fetching-app-failed", err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Error("fetching-app-failed", FetchAppFailedErr, lager.Data{
			"StatusCode":   resp.Status,
			"ResponseBody": resp.Body,
		})
		return "", FetchAppFailedErr
	}

	var app AppSSHResponse
	err = json.NewDecoder(resp.Body).Decode(&app)
	if err != nil {
		logger.Error("invalid-cc-response", err)
		return "", InvalidCCResponse
	}

	return app.ProcessGuid, nil
}
