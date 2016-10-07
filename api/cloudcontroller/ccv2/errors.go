package ccv2

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type CCErrorResponse struct {
	Code        int    `json:"code"`
	Description string `json:"description"`
	ErrorCode   string `json:"error_code"`
}

type UnauthorizedError struct {
	Message string
}

func (e UnauthorizedError) Error() string {
	return e.Message
}

type InvalidAuthTokenError struct {
	Message string
}

func (e InvalidAuthTokenError) Error() string {
	return e.Message
}

type ForbiddenError struct {
	Message string
}

func (e ForbiddenError) Error() string {
	return e.Message
}

type ResourceNotFoundError struct {
	Message string
}

func (e ResourceNotFoundError) Error() string {
	return e.Message
}

type UnexpectedResponseError struct {
	ResponseCode int
	CCErrorResponse
}

func (e UnexpectedResponseError) Error() string {
	return fmt.Sprintf("Unexpected Response\nResponse Code: %s\nCC Code: %i\nCC ErrorCode: %s\nDescription: %s", e.ResponseCode, e.Code, e.ErrorCode, e.Description)
}

type errorWrapper struct {
	connection Connection
}

func newErrorWrapper() *errorWrapper {
	return &errorWrapper{}
}

func (e *errorWrapper) Wrap(innerconnection Connection) Connection {
	e.connection = innerconnection
	return e
}

func (e *errorWrapper) Make(passedRequest Request, passedResponse *Response) error {
	err := e.connection.Make(passedRequest, passedResponse)

	if rawErr, ok := err.(RawCCError); ok {
		return e.convert(rawErr)
	}
	return err
}

func (e errorWrapper) convert(rawErr RawCCError) error {
	var errorResponse CCErrorResponse
	err := json.Unmarshal(rawErr.RawResponse, &errorResponse)
	if err != nil {
		return err
	}

	switch rawErr.StatusCode {
	case http.StatusUnauthorized:
		if errorResponse.ErrorCode == "CF-InvalidAuthToken" {
			return InvalidAuthTokenError{Message: errorResponse.Description}
		}
		return UnauthorizedError{Message: errorResponse.Description}
	case http.StatusForbidden:
		return ForbiddenError{Message: errorResponse.Description}
	case http.StatusNotFound:
		return ResourceNotFoundError{Message: errorResponse.Description}
	default:
		return UnexpectedResponseError{
			ResponseCode:    rawErr.StatusCode,
			CCErrorResponse: errorResponse,
		}
	}
	return nil
}
