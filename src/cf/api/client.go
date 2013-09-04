package api

import (
	term "cf/terminal"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"
	"regexp"
	"strings"
)

type AuthorizedRequest struct {
	*http.Request
}

type errorResponse struct {
	Code        int
	Description string
}

func newClient() *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return &http.Client{Transport: tr}
}

func NewAuthorizedRequest(method, path, accessToken string, body io.Reader) (authReq *AuthorizedRequest, err error) {
	request, err := http.NewRequest(method, path, body)
	if err != nil {
		return
	}
	request.Header.Set("Authorization", accessToken)
	request.Header.Set("accept", "application/json")

	authReq = &AuthorizedRequest{request}
	return
}

func PerformRequest(request *AuthorizedRequest) (errorCode int, err error) {
	_, errorCode, err = doRequest(request.Request)
	return
}

func PerformRequestAndParseResponse(request *AuthorizedRequest, response interface{}) (errorCode int, err error) {
	rawResponse, errorCode, err := doRequest(request.Request)
	if err != nil {
		return
	}

	jsonBytes, err := ioutil.ReadAll(rawResponse.Body)
	if err != nil {
		err = errors.New(fmt.Sprintf("Could not read response body: %s", err.Error()))
		return
	}

	err = json.Unmarshal(jsonBytes, &response)

	if err != nil {
		err = errors.New(fmt.Sprintf("Invalid JSON response from server: %s", err.Error()))
	}
	return
}

func SanitizeRequest(request string) string {
	re := regexp.MustCompile(`(?m)^Authorization: .*`)
	return re.ReplaceAllString(request, "Authorization: [PRIVATE DATA HIDDEN]")
}

func doRequest(request *http.Request) (response *http.Response, errorCode int, err error) {
	client := newClient()

	if traceEnabled() {
		dumpedRequest, err := httputil.DumpRequest(request, true)
		if err != nil {
			fmt.Println("Error dumping request")
		} else {
			fmt.Printf("\n%s\n%s\n", term.Cyan("REQUEST:"), SanitizeRequest(string(dumpedRequest)))
		}
	}

	response, err = client.Do(request)

	if traceEnabled() {
		dumpedResponse, err := httputil.DumpResponse(response, true)
		if err != nil {
			fmt.Println("Error dumping response")
		} else {
			fmt.Printf("\n%s\n%s\n", term.Cyan("RESPONSE:"), string(dumpedResponse))
		}
	}

	if err != nil {
		err = errors.New(fmt.Sprintf("Error performing request: %s", err.Error()))
		return
	}

	if response.StatusCode > 299 {
		errorResponse := getErrorResponse(response)
		errorCode = errorResponse.Code
		message := fmt.Sprintf("Server error, status code: %d, error code: %d, message: %s", response.StatusCode, errorCode, errorResponse.Description)
		err = errors.New(message)
	}

	return
}

func traceEnabled() bool {
	traceEnv := strings.ToLower(os.Getenv("CF_TRACE"))
	return traceEnv == "true" || traceEnv == "yes"
}

func getErrorResponse(response *http.Response) (eR errorResponse) {
	jsonBytes, _ := ioutil.ReadAll(response.Body)
	response.Body.Close()

	eR = errorResponse{}
	_ = json.Unmarshal(jsonBytes, &eR)
	return
}
