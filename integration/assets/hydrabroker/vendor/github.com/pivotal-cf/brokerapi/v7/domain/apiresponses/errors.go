package apiresponses

import (
	"errors"
	"net/http"
)

const (
	instanceExistsMsg             = "instance already exists"
	instanceDoesntExistMsg        = "instance does not exist"
	serviceLimitReachedMsg        = "instance limit for this service has been reached"
	servicePlanQuotaExceededMsg   = "The quota for this service plan has been exceeded. Please contact your Operator for help."
	serviceQuotaExceededMsg       = "The quota for this service has been exceeded. Please contact your Operator for help."
	bindingExistsMsg              = "binding already exists"
	bindingDoesntExistMsg         = "binding does not exist"
	bindingNotFoundMsg            = "binding cannot be fetched"
	asyncRequiredMsg              = "This service plan requires client support for asynchronous service operations."
	planChangeUnsupportedMsg      = "The requested plan migration cannot be performed"
	rawInvalidParamsMsg           = "The format of the parameters is not valid JSON"
	appGuidMissingMsg             = "app_guid is a required field but was not provided"
	concurrentInstanceAccessMsg   = "instance is being updated and cannot be retrieved"
	maintenanceInfoConflictMsg    = "passed maintenance_info does not match the catalog maintenance_info"
	maintenanceInfoNilConflictMsg = "maintenance_info was passed, but the broker catalog contains no maintenance_info"

	instanceLimitReachedErrorKey  = "instance-limit-reached"
	instanceAlreadyExistsErrorKey = "instance-already-exists"
	bindingAlreadyExistsErrorKey  = "binding-already-exists"
	instanceMissingErrorKey       = "instance-missing"
	bindingMissingErrorKey        = "binding-missing"
	bindingNotFoundErrorKey       = "binding-not-found"
	asyncRequiredKey              = "async-required"
	planChangeNotSupportedKey     = "plan-change-not-supported"
	invalidRawParamsKey           = "invalid-raw-params"
	appGuidNotProvidedErrorKey    = "app-guid-not-provided"
	concurrentAccessKey           = "get-instance-during-update"
	maintenanceInfoConflictKey    = "maintenance-info-conflict"
)

var (
	ErrInstanceAlreadyExists = NewFailureResponseBuilder(
		errors.New(instanceExistsMsg), http.StatusConflict, instanceAlreadyExistsErrorKey,
	).WithEmptyResponse().Build()

	ErrInstanceDoesNotExist = NewFailureResponseBuilder(
		errors.New(instanceDoesntExistMsg), http.StatusGone, instanceMissingErrorKey,
	).WithEmptyResponse().Build()

	ErrInstanceLimitMet = NewFailureResponse(
		errors.New(serviceLimitReachedMsg), http.StatusInternalServerError, instanceLimitReachedErrorKey,
	)

	ErrBindingAlreadyExists = NewFailureResponse(
		errors.New(bindingExistsMsg), http.StatusConflict, bindingAlreadyExistsErrorKey,
	)

	ErrBindingDoesNotExist = NewFailureResponseBuilder(
		errors.New(bindingDoesntExistMsg), http.StatusGone, bindingMissingErrorKey,
	).WithEmptyResponse().Build()

	ErrBindingNotFound = NewFailureResponseBuilder(
		errors.New(bindingNotFoundMsg), http.StatusNotFound, bindingNotFoundErrorKey,
	).WithEmptyResponse().Build()

	ErrAsyncRequired = NewFailureResponseBuilder(
		errors.New(asyncRequiredMsg), http.StatusUnprocessableEntity, asyncRequiredKey,
	).WithErrorKey("AsyncRequired").Build()

	ErrPlanChangeNotSupported = NewFailureResponseBuilder(
		errors.New(planChangeUnsupportedMsg), http.StatusUnprocessableEntity, planChangeNotSupportedKey,
	).WithErrorKey("PlanChangeNotSupported").Build()

	ErrRawParamsInvalid = NewFailureResponse(
		errors.New(rawInvalidParamsMsg), http.StatusUnprocessableEntity, invalidRawParamsKey,
	)

	ErrAppGuidNotProvided = NewFailureResponse(
		errors.New(appGuidMissingMsg), http.StatusUnprocessableEntity, appGuidNotProvidedErrorKey,
	)

	ErrPlanQuotaExceeded    = errors.New(servicePlanQuotaExceededMsg)
	ErrServiceQuotaExceeded = errors.New(serviceQuotaExceededMsg)

	ErrConcurrentInstanceAccess = NewFailureResponseBuilder(
		errors.New(concurrentInstanceAccessMsg), http.StatusUnprocessableEntity, concurrentAccessKey,
	).WithErrorKey("ConcurrencyError").Build()

	ErrMaintenanceInfoConflict = NewFailureResponseBuilder(
		errors.New(maintenanceInfoConflictMsg), http.StatusUnprocessableEntity, maintenanceInfoConflictKey,
	).WithErrorKey("MaintenanceInfoConflict").Build()

	ErrMaintenanceInfoNilConflict = NewFailureResponseBuilder(
		errors.New(maintenanceInfoNilConflictMsg), http.StatusUnprocessableEntity, maintenanceInfoConflictKey,
	).WithErrorKey("MaintenanceInfoConflict").Build()
)
