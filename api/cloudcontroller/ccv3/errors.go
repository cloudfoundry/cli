package ccv3

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
)

const taskWorkersUnavailable = "CF-TaskWorkersUnavailable"

// errorWrapper is the wrapper that converts responses with 4xx and 5xx status
// codes to an error.
type errorWrapper struct {
	connection cloudcontroller.Connection
}

func newErrorWrapper() *errorWrapper {
	return new(errorWrapper)
}

// Make creates a connection in the wrapped connection and handles errors
// that it returns.
func (e *errorWrapper) Make(request *cloudcontroller.Request, passedResponse *cloudcontroller.Response) error {
	err := e.connection.Make(request, passedResponse)

	if rawHTTPStatusErr, ok := err.(ccerror.RawHTTPStatusError); ok {
		if rawHTTPStatusErr.StatusCode >= http.StatusInternalServerError {
			return convert500(rawHTTPStatusErr)
		}
		return convert400(rawHTTPStatusErr, request)
	}
	return err
}

// Wrap wraps a Cloud Controller connection in this error handling wrapper.
func (e *errorWrapper) Wrap(innerconnection cloudcontroller.Connection) cloudcontroller.Connection {
	e.connection = innerconnection
	return e
}

func convert400(rawHTTPStatusErr ccerror.RawHTTPStatusError, request *cloudcontroller.Request) error {
	firstErr, errorResponse, err := unmarshalFirstV3Error(rawHTTPStatusErr)
	if err != nil {
		return err
	}

	if len(errorResponse.Errors) > 1 {
		return ccerror.MultiError{Errors: errorResponse.Errors, ResponseCode: rawHTTPStatusErr.StatusCode}
	}

	switch rawHTTPStatusErr.StatusCode {
	case http.StatusBadRequest: // 400
		return handleBadRequest(firstErr, request)
	case http.StatusUnauthorized: // 401
		if firstErr.Title == "CF-InvalidAuthToken" {
			return ccerror.InvalidAuthTokenError{Message: firstErr.Detail}
		}
		return ccerror.UnauthorizedError{Message: firstErr.Detail}
	case http.StatusForbidden: // 403
		return ccerror.ForbiddenError{Message: firstErr.Detail}
	case http.StatusNotFound: // 404
		return handleNotFound(firstErr, request)
	case http.StatusUnprocessableEntity: // 422
		return handleUnprocessableEntity(firstErr)
	case http.StatusServiceUnavailable: // 503
		if firstErr.Title == taskWorkersUnavailable {
			return ccerror.TaskWorkersUnavailableError{Message: firstErr.Detail}
		}
		return ccerror.ServiceUnavailableError{Message: firstErr.Detail}
	default:
		return ccerror.V3UnexpectedResponseError{
			ResponseCode:    rawHTTPStatusErr.StatusCode,
			RequestIDs:      rawHTTPStatusErr.RequestIDs,
			V3ErrorResponse: errorResponse,
		}
	}
}

func convert500(rawHTTPStatusErr ccerror.RawHTTPStatusError) error {
	switch rawHTTPStatusErr.StatusCode {
	case http.StatusServiceUnavailable: // 503
		firstErr, _, err := unmarshalFirstV3Error(rawHTTPStatusErr)
		if err != nil {
			return err
		}
		if firstErr.Title == taskWorkersUnavailable {
			return ccerror.TaskWorkersUnavailableError{Message: firstErr.Detail}
		}
		return ccerror.ServiceUnavailableError{Message: firstErr.Detail}
	default:
		return ccerror.V3UnexpectedResponseError{
			ResponseCode: rawHTTPStatusErr.StatusCode,
			RequestIDs:   rawHTTPStatusErr.RequestIDs,
			V3ErrorResponse: ccerror.V3ErrorResponse{
				Errors: []ccerror.V3Error{{
					Detail: string(rawHTTPStatusErr.RawResponse),
				}},
			},
		}
	}
}

func handleBadRequest(errorResponse ccerror.V3Error, _ *cloudcontroller.Request) error {
	// Currently the CF-BadQueryParameter is the only 400 BadRequest error returned from v3
	return ccerror.BadRequestError{Message: errorResponse.Detail}
}

func handleNotFound(errorResponse ccerror.V3Error, request *cloudcontroller.Request) error {
	switch errorResponse.Detail {
	case "App not found":
		return ccerror.ApplicationNotFoundError{}
	case "Droplet not found":
		return ccerror.DropletNotFoundError{}
	case "Deployment not found":
		return ccerror.DeploymentNotFoundError{}
	case "Feature flag not found":
		return ccerror.FeatureFlagNotFoundError{}
	case "Instance not found":
		return ccerror.InstanceNotFoundError{}
	case "Process not found":
		return ccerror.ProcessNotFoundError{}
	case "Role not found":
		return ccerror.RoleNotFoundError{}
	case "Unknown request":
		return ccerror.APINotFoundError{URL: request.URL.String()}
	default:
		return ccerror.ResourceNotFoundError{Message: errorResponse.Detail}
	}
}

func handleUnprocessableEntity(errorResponse ccerror.V3Error) error {
	//idea to make route already exist error flexible for all relevant error cases
	errorString := errorResponse.Detail
	err := ccerror.UnprocessableEntityError{Message: errorResponse.Detail}
	appNameTakenRegexp := regexp.MustCompile(`App with the name '.*' already exists\.`)
	orgNameTakenRegexp := regexp.MustCompile(`Organization '.*' already exists\.`)
	roleExistsRegexp := regexp.MustCompile(`User '.*' already has '.*' role.*`)

	// boolean switch case with partial/regex string matchers
	switch {
	case appNameTakenRegexp.MatchString(errorString) || strings.Contains(errorString, "name must be unique in space"):
		return ccerror.NameNotUniqueInSpaceError{UnprocessableEntityError: err}
	case strings.Contains(errorString,
		"Name must be unique per organization"):
		return ccerror.NameNotUniqueInOrgError{}
	case strings.Contains(errorString,
		"Route already exists"):
		return ccerror.RouteNotUniqueError{UnprocessableEntityError: err}
	case strings.Contains(errorString,
		"Buildpack must be an existing admin buildpack or a valid git URI"):
		return ccerror.InvalidBuildpackError{}
	case strings.Contains(errorString,
		"Assign a droplet before starting this app."):
		return ccerror.InvalidStartError{}
	case orgNameTakenRegexp.MatchString(errorString):
		return ccerror.OrganizationNameTakenError{UnprocessableEntityError: err}
	case roleExistsRegexp.MatchString(errorString):
		return ccerror.RoleAlreadyExistsError{UnprocessableEntityError: err}
	default:
		return err
	}
}

func unmarshalFirstV3Error(rawHTTPStatusErr ccerror.RawHTTPStatusError) (ccerror.V3Error, ccerror.V3ErrorResponse, error) {
	// Try to unmarshal the raw error into a CC error. If unmarshaling fails,
	// return the raw error.
	var errorResponse ccerror.V3ErrorResponse
	err := json.Unmarshal(rawHTTPStatusErr.RawResponse, &errorResponse)
	if err != nil {
		return ccerror.V3Error{}, errorResponse, ccerror.UnknownHTTPSourceError{
			StatusCode:  rawHTTPStatusErr.StatusCode,
			RawResponse: rawHTTPStatusErr.RawResponse,
		}
	}

	errors := errorResponse.Errors
	if len(errors) == 0 {
		return ccerror.V3Error{}, errorResponse, ccerror.V3UnexpectedResponseError{
			ResponseCode:    rawHTTPStatusErr.StatusCode,
			V3ErrorResponse: errorResponse,
		}
	}

	return errors[0], errorResponse, nil
}
