package v7action

import (
	"bufio"
	"fmt"
	"net/http"
	"net/textproto"
	"os"
	"strings"

	"code.cloudfoundry.org/cli/v9/command/translatableerror"
)

func (actor Actor) MakeCurlRequest(
	method string,
	path string,
	customHeaders []string,
	data string,
	failOnHTTPError bool,
) ([]byte, *http.Response, error) {
	url := fmt.Sprintf("%s/%s", actor.Config.Target(), strings.TrimLeft(path, "/"))

	requestHeaders, err := buildRequestHeaders(customHeaders)
	if err != nil {
		return nil, nil, translatableerror.RequestCreationError{Err: err}
	}

	if method == "" && data != "" {
		method = "POST"
	}

	requestBodyBytes := []byte(data)

	trimmedData := strings.Trim(string(data), `"'`)
	if strings.HasPrefix(trimmedData, "@") {
		trimmedData = strings.Trim(trimmedData[1:], `"'`)
		requestBodyBytes, err = os.ReadFile(trimmedData)
		if err != nil {
			return nil, nil, translatableerror.RequestCreationError{Err: err}
		}
	}

	responseBody, httpResponse, err := actor.CloudControllerClient.MakeRequestSendReceiveRaw(
		method,
		url,
		requestHeaders,
		requestBodyBytes,
	)

	if err != nil {
		// A nil HTTP response means the request never reached the API (e.g. a token
		// refresh or other authentication failure in the request wrapper). There is no
		// status code to read and no response body to print, so surface the error
		// directly. This also avoids a nil-pointer dereference on httpResponse below
		// when the fail-on-http-error flag is set.
		if httpResponse == nil {
			return nil, nil, err
		}

		if failOnHTTPError {
			return nil, nil, translatableerror.CurlExit22Error{StatusCode: httpResponse.StatusCode}
		}
	}

	return responseBody, httpResponse, nil
}

func buildRequestHeaders(customHeaders []string) (http.Header, error) {
	headerString := strings.Join(customHeaders, "\n")
	headerString = strings.TrimSpace(headerString)
	headerString += "\n\n"

	headerReader := bufio.NewReader(strings.NewReader(headerString))
	parsedCustomHeaders, err := textproto.NewReader(headerReader).ReadMIMEHeader()
	if err != nil {
		return nil, err
	}

	return http.Header(parsedCustomHeaders), nil
}
