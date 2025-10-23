package ccerror

import (
	"fmt"

	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3/constant"
)

// V3JobFailedError represents a failed Cloud Controller Job. It wraps the error
// returned back from the Cloud Controller.
type V3JobFailedError struct {
	JobGUID string

	// Code is a numeric code for this error.
	Code constant.JobErrorCode `json:"code"`
	// Detail is a verbose description of the error.
	Detail string `json:"detail"`
	// Title is a short description of the error.
	Title string `json:"title"`
}

func (e V3JobFailedError) Error() string {
	return fmt.Sprintf("Job (%s) failed: %s", e.JobGUID, e.Detail)
}
