package api

import (
	"cf"
	term "cf/terminal"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

const PRIVATE_DATA_PLACEHOLDER = "[PRIVATE DATA HIDDEN]"

type Request struct {
	*http.Request
}

type ApiError struct {
	Message    string
	ErrorCode  string
	StatusCode int
}

func NewApiError(message string, errorCode string, statusCode int) (apiErr *ApiError) {
	return &ApiError{
		Message:    message,
		ErrorCode:  errorCode,
		StatusCode: statusCode,
	}
}

func NewApiErrorWithMessage(message string, a ...interface{}) (apiErr *ApiError) {
	return &ApiError{
		Message: fmt.Sprintf(message, a...),
	}
}

func NewApiErrorWithError(message string, err error) (apiErr *ApiError) {
	return &ApiError{
		Message: fmt.Sprintf("%s: %s", message, err.Error()),
	}
}

func (apiErr *ApiError) Error() string {
	return apiErr.Message
}

func NewRequest(method, path, accessToken string, body io.Reader) (authReq *Request, apiErr *ApiError) {
	request, err := http.NewRequest(method, path, body)
	if err != nil {
		apiErr = NewApiErrorWithError("Error building request", err)
		return
	}

	if accessToken != "" {
		request.Header.Set("Authorization", accessToken)
	}

	request.Header.Set("accept", "application/json")
	request.Header.Set("User-Agent", "go-cli "+cf.Version+" / "+runtime.GOOS)
	authReq = &Request{request}
	return
}

func PerformRequestAndParseResponse(request *Request, response interface{}) (apiErr *ApiError) {
	rawResponse, apiErr := doRequest(request.Request)
	if apiErr != nil {
		return
	}

	apiErr = parseResponse(rawResponse, response)
	return
}

func shouldRedirect(request *Request, response *http.Response) bool {
	return request.Method == "GET" && response.StatusCode == http.StatusTemporaryRedirect || response.StatusCode == http.StatusMovedPermanently
}

func newHttpClient() *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		Proxy:           http.ProxyFromEnvironment,
	}
	return &http.Client{Transport: tr}
}

type ApiClient struct {
	authenticator Authenticator
}

func NewApiClient(auth Authenticator) (client ApiClient) {
	client.authenticator = auth
	return
}

func (c ApiClient) PerformRequest(request *Request) (apiErr *ApiError) {
	_, apiErr = c.doRequestHandlingAuth(request)
	return
}

func (c ApiClient) PerformRequestAndParseResponse(request *Request, response interface{}) (apiErr *ApiError) {
	rawResponse, apiErr := c.doRequestHandlingAuth(request)
	if apiErr != nil {
		return
	}
	apiErr = parseResponse(rawResponse, response)
	return
}

func (c ApiClient) PerformRequestForTextResponse(request *Request) (response string, apiErr *ApiError) {
	rawResponse, apiErr := c.doRequestHandlingAuth(request)
	if apiErr != nil {
		return
	}

	textBytes, err := ioutil.ReadAll(rawResponse.Body)
	if err != nil {
		apiErr = NewApiErrorWithError("Error reading response body:", err)
		return
	}

	response = string(textBytes)
	return
}

func (c ApiClient) doRequestHandlingAuth(request *Request) (response *http.Response, apiErr *ApiError) {
	response, apiErr = doRequest(request.Request)

	if apiErr != nil && response == nil {
		return
	}

	if response.StatusCode == http.StatusUnauthorized && apiErr.ErrorCode == "1000" {
		newToken, apiErr := c.authenticator.RefreshAuthToken()
		if apiErr == nil {
			request.Header.Set("Authorization", newToken)
			return doRequest(request.Request)
		}
	}

	if shouldRedirect(request, response) {
		newRequest, apiErr := NewRequest("GET", response.Header.Get("location"), "", nil)
		if apiErr == nil {
			return doRequest(newRequest.Request)
		}
	}

	return
}

func doRequest(request *http.Request) (response *http.Response, apiError *ApiError) {
	var err error

	httpClient := newHttpClient()

	if traceEnabled() {
		dumpedRequest, err := httputil.DumpRequest(request, true)
		if err != nil {
			fmt.Println("Error dumping request")
		} else {
			fmt.Printf("\n%s\n%s\n", term.Cyan("REQUEST:"), Sanitize(string(dumpedRequest)))
		}
	}

	response, err = httpClient.Do(request)

	if err != nil {
		apiError = NewApiErrorWithError("Error performing request", err)
		return
	}

	if traceEnabled() {
		dumpedResponse, err := httputil.DumpResponse(response, true)
		if err != nil {
			fmt.Println("Error dumping response")
		} else {
			fmt.Printf("\n%s\n%s\n", term.Cyan("RESPONSE:"), Sanitize(string(dumpedResponse)))
		}
	}

	if response.StatusCode > 299 {
		errorResponse := getErrorResponse(response)
		message := fmt.Sprintf(
			"Server error, status code: %d, error code: %s, message: %s",
			response.StatusCode,
			errorResponse.Code,
			errorResponse.Description,
		)
		apiError = NewApiError(message, errorResponse.Code, response.StatusCode)
	}

	return
}

func Sanitize(input string) (sanitized string) {
	re := regexp.MustCompile(`(?m)^Authorization: .*`)
	sanitized = re.ReplaceAllString(input, "Authorization: "+PRIVATE_DATA_PLACEHOLDER)
	re = regexp.MustCompile(`password=[^&]*&`)
	sanitized = re.ReplaceAllString(sanitized, "password="+PRIVATE_DATA_PLACEHOLDER+"&")
	re = regexp.MustCompile(`"access_token":"[^"]*"`)
	sanitized = re.ReplaceAllString(sanitized, `"access_token":"`+PRIVATE_DATA_PLACEHOLDER+`"`)
	re = regexp.MustCompile(`"refresh_token":"[^"]*"`)
	sanitized = re.ReplaceAllString(sanitized, `"refresh_token":"`+PRIVATE_DATA_PLACEHOLDER+`"`)
	return
}

func parseResponse(rawResponse *http.Response, response interface{}) (apiError *ApiError) {
	jsonBytes, err := ioutil.ReadAll(rawResponse.Body)
	if err != nil {
		apiError = NewApiErrorWithError("Could not read response body", err)
		return
	}

	err = json.Unmarshal(jsonBytes, &response)

	if err != nil {
		apiError = NewApiErrorWithError("Invalid JSON response from server", err)
	}

	return
}

func traceEnabled() bool {
	traceEnv := strings.ToLower(os.Getenv("CF_TRACE"))
	return traceEnv == "true" || traceEnv == "yes"
}

type errorResponse struct {
	Code        string
	Description string
}

type uaaErrorResponse struct {
	Code        string `json:"error"`
	Description string `json:"error_description"`
}

type ccErrorResponse struct {
	Code        int
	Description string
}

func getErrorResponse(response *http.Response) errorResponse {
	jsonBytes, _ := ioutil.ReadAll(response.Body)
	response.Body.Close()

	ccResp := ccErrorResponse{}
	err := json.Unmarshal(jsonBytes, &ccResp)

	if err != nil || (ccResp == ccErrorResponse{}) {
		uaaResp := uaaErrorResponse{}
		json.Unmarshal(jsonBytes, &uaaResp)

		return errorResponse{Code: uaaResp.Code, Description: uaaResp.Description}
	}

	return errorResponse{Code: strconv.Itoa(ccResp.Code), Description: ccResp.Description}
}
