package cloud

import (
	"fmt"
	"regexp"
)

const (
	VMNotFoundError       = "Bosh::Clouds::VMNotFound"
	DiskNotFoundError     = "Bosh::Clouds::DiskNotFound"
	StemcellNotFoundError = "Bosh::Clouds::StemcellNotFound"
	NotImplementedError   = "Bosh::Clouds::NotImplemented"
)

type Error interface {
	error
	Method() string
	Type() string
	Message() string
	OkToRetry() bool
}

type cpiError struct {
	method   string
	cmdError CmdError
}

func NewCPIError(method string, cmdError CmdError) Error {
	if mapsToNotImplementedError(method, cmdError) {
		cmdError = newNotImplementedCmdError(method, cmdError)
	}

	return cpiError{
		method:   method,
		cmdError: cmdError,
	}
}

func (e cpiError) Error() string {
	return fmt.Sprintf("CPI '%s' method responded with error: %s", e.method, e.cmdError)
}

func (e cpiError) Method() string {
	return e.method
}

func (e cpiError) Type() string {
	return e.cmdError.Type
}

func (e cpiError) Message() string {
	return e.cmdError.Message
}

func (e cpiError) OkToRetry() bool {
	return e.cmdError.OkToRetry
}

func mapsToNotImplementedError(method string, cmdError CmdError) bool {
	matched, _ := regexp.MatchString("^Invalid Method:", cmdError.Message)

	if cmdError.Type == "Bosh::Clouds::CloudError" && matched {
		return true
	}

	matched, _ = regexp.MatchString("^Method is not known, got", cmdError.Message)

	if cmdError.Type == "InvalidCall" && matched {
		return true
	}

	return false
}

func newNotImplementedCmdError(method string, cmdError CmdError) CmdError {
	return CmdError{
		NotImplementedError,
		fmt.Sprintf("CPI error '%s' with message '%s' in '%s' CPI method", cmdError.Type, cmdError.Message, method),
		cmdError.OkToRetry,
	}
}
