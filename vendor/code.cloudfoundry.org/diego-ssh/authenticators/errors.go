package authenticators

import "errors"

var AuthenticationFailedErr = errors.New("Authentication failed")
var FetchAppFailedErr = errors.New("Fetching application data failed")
var InvalidCCResponse = errors.New("Invalid response from Cloud Controller")
var InvalidCredentialsErr error = errors.New("Invalid credentials")
var InvalidDomainErr error = errors.New("Invalid authentication domain")
var InvalidRequestErr = errors.New("CloudController URL Invalid")
var InvalidUserFormatErr = errors.New("Invalid user format")
var NotDiegoErr = errors.New("Diego Not Enabled")
var RouteNotFoundErr error = errors.New("SSH routing info not found")
var SSHDisabledErr = errors.New("SSH Disabled")
