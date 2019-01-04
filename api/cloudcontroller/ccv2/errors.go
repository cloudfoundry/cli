package ccv2

import (
	"encoding/json"
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
)

// errorWrapper is the wrapper that converts responses with 4xx and 5xx status
// codes to an error.
type errorWrapper struct {
	connection cloudcontroller.Connection
}

func newErrorWrapper() *errorWrapper {
	return new(errorWrapper)
}

// Make converts RawHTTPStatusError, which represents responses with 4xx and
// 5xx status codes, to specific errors.
func (e *errorWrapper) Make(request *cloudcontroller.Request, passedResponse *cloudcontroller.Response) error {
	err := e.connection.Make(request, passedResponse)

	if rawHTTPStatusErr, ok := err.(ccerror.RawHTTPStatusError); ok {
		if passedResponse.HTTPResponse.StatusCode >= http.StatusInternalServerError {
			return convert500(rawHTTPStatusErr)
		}

		return convert400(rawHTTPStatusErr)
	}
	return err
}

// Wrap wraps a Cloud Controller connection in this error handling wrapper.
func (e *errorWrapper) Wrap(innerconnection cloudcontroller.Connection) cloudcontroller.Connection {
	e.connection = innerconnection
	return e
}

func convert400(rawHTTPStatusErr ccerror.RawHTTPStatusError) error {
	// Try to unmarshal the raw error into a CC error. If unmarshaling fails,
	// either we're not talking to a CC, or the CC returned invalid json.
	errorResponse, err := unmarshalRawHTTPErr(rawHTTPStatusErr)
	if err != nil {
		return err
	}

	switch rawHTTPStatusErr.StatusCode {
	case http.StatusBadRequest: // 400
		return handleBadRequest(errorResponse)
	case http.StatusUnauthorized: // 401
		return handleUnauthorized(errorResponse)
	case http.StatusForbidden: // 403
		return ccerror.ForbiddenError{Message: errorResponse.Description}
	case http.StatusNotFound: // 404
		return ccerror.ResourceNotFoundError{Message: errorResponse.Description}
	case http.StatusUnprocessableEntity: // 422
		return handleUnprocessableEntity(errorResponse)
	default:
		return ccerror.V2UnexpectedResponseError{
			RequestIDs:      rawHTTPStatusErr.RequestIDs,
			ResponseCode:    rawHTTPStatusErr.StatusCode,
			V2ErrorResponse: errorResponse,
		}
	}
}

func convert500(rawHTTPStatusErr ccerror.RawHTTPStatusError) error {
	errorResponse, err := unmarshalRawHTTPErr(rawHTTPStatusErr)
	if err != nil {
		return err
	}

	switch rawHTTPStatusErr.StatusCode {
	case http.StatusBadGateway: // 502
		return handleBadGateway(errorResponse, rawHTTPStatusErr)
	default:
		return v2UnexpectedResponseError(rawHTTPStatusErr)
	}
}

func handleBadGateway(errorResponse ccerror.V2ErrorResponse, rawHTTPStatusErr ccerror.RawHTTPStatusError) error {
	switch errorResponse.ErrorCode {
	case "CF-ServiceBrokerCatalogInvalid":
		return ccerror.ServiceBrokerCatalogInvalidError{Message: errorResponse.Description}
	case "CF-ServiceBrokerRequestRejected":
		return ccerror.ServiceBrokerRequestRejectedError{Message: errorResponse.Description}
	case "CF-ServiceBrokerBadResponse":
		return ccerror.ServiceBrokerBadResponseError{Message: errorResponse.Description}
	default:
		return v2UnexpectedResponseError(rawHTTPStatusErr)
	}
}

func handleBadRequest(errorResponse ccerror.V2ErrorResponse) error {
	switch errorResponse.ErrorCode {
	case "CF-AppStoppedStatsError":
		return ccerror.ApplicationStoppedStatsError{Message: errorResponse.Description}
	case "CF-BuildpackInvalid":
		return ccerror.BuildpackAlreadyExistsWithoutStackError{Message: errorResponse.Description}
	case "CF-BuildpackNameTaken":
		return ccerror.BuildpackNameTakenError{Message: errorResponse.Description}
	case "CF-InstancesError":
		return ccerror.InstancesError{Message: errorResponse.Description}
	case "CF-InvalidRelation":
		return ccerror.InvalidRelationError{Message: errorResponse.Description}
	case "CF-NotStaged":
		return ccerror.NotStagedError{Message: errorResponse.Description}
	case "CF-ServiceBindingAppServiceTaken":
		return ccerror.ServiceBindingTakenError{Message: errorResponse.Description}
	case "CF-ServiceKeyNameTaken":
		return ccerror.ServiceKeyTakenError{Message: errorResponse.Description}
	case "CF-OrganizationNameTaken":
		return ccerror.OrganizationNameTakenError{Message: errorResponse.Description}
	case "CF-SpaceNameTaken":
		return ccerror.SpaceNameTakenError{Message: errorResponse.Description}
	case "CF-ServiceInstanceNameTaken":
		return ccerror.ServiceInstanceNameTakenError{Message: errorResponse.Description}
	default:
		return ccerror.BadRequestError{Message: errorResponse.Description}
	}
}

func handleUnauthorized(errorResponse ccerror.V2ErrorResponse) error {
	if errorResponse.ErrorCode == "CF-InvalidAuthToken" {
		return ccerror.InvalidAuthTokenError{Message: errorResponse.Description}
	}

	return ccerror.UnauthorizedError{Message: errorResponse.Description}
}

func handleUnprocessableEntity(errorResponse ccerror.V2ErrorResponse) error {
	if errorResponse.ErrorCode == "CF-BuildpackNameStackTaken" {
		return ccerror.BuildpackAlreadyExistsForStackError{Message: errorResponse.Description}
	}
	return ccerror.UnprocessableEntityError{Message: errorResponse.Description}
}

func unmarshalRawHTTPErr(rawHTTPStatusErr ccerror.RawHTTPStatusError) (ccerror.V2ErrorResponse, error) {
	var errorResponse ccerror.V2ErrorResponse
	err := json.Unmarshal(rawHTTPStatusErr.RawResponse, &errorResponse)
	if err != nil {
		// ccv2/info.go converts this error to an APINotFoundError.
		return ccerror.V2ErrorResponse{}, ccerror.UnknownHTTPSourceError{StatusCode: rawHTTPStatusErr.StatusCode, RawResponse: rawHTTPStatusErr.RawResponse}
	}
	return errorResponse, nil
}

func v2UnexpectedResponseError(rawHTTPStatusErr ccerror.RawHTTPStatusError) ccerror.V2UnexpectedResponseError {
	return ccerror.V2UnexpectedResponseError{
		ResponseCode: rawHTTPStatusErr.StatusCode,
		RequestIDs:   rawHTTPStatusErr.RequestIDs,
		V2ErrorResponse: ccerror.V2ErrorResponse{
			Description: string(rawHTTPStatusErr.RawResponse),
		},
	}
}
