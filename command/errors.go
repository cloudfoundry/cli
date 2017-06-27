package command

import (
	"fmt"
	"strings"
)

type APIRequestError struct {
	Err error
}

func (_ APIRequestError) Error() string {
	return "Request error: {{.Error}}\nTIP: If you are behind a firewall and require an HTTP proxy, verify the https_proxy environment variable is correctly set. Else, check your network connection."
}

func (e APIRequestError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Error": e.Err,
	})
}

type BadCredentialsError struct{}

func (_ BadCredentialsError) Error() string {
	return "Credentials were rejected, please try again."
}

func (e BadCredentialsError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{})
}

type InvalidSSLCertError struct {
	API string
}

func (_ InvalidSSLCertError) Error() string {
	return "Invalid SSL Cert for {{.API}}\nTIP: Use 'cf api --skip-ssl-validation' to continue with an insecure API endpoint"
}

func (e InvalidSSLCertError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"API": e.API,
	})
}

type SSLCertErrorError struct {
	Message string
}

func (_ SSLCertErrorError) Error() string {
	return "SSL Certificate Error {{.Message}}\nTIP: Use 'cf api --skip-ssl-validation' to continue with an insecure API endpoint"
}

func (e SSLCertErrorError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Message": e.Message,
	})
}

type NoAPISetError struct {
	BinaryName string
}

func (_ NoAPISetError) Error() string {
	return "No API endpoint set. Use '{{.LoginTip}}' or '{{.APITip}}' to target an endpoint."
}

func (e NoAPISetError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"LoginTip": fmt.Sprintf("%s login", e.BinaryName),
		"APITip":   fmt.Sprintf("%s api", e.BinaryName),
	})
}

type NotLoggedInError struct {
	BinaryName string
}

func (_ NotLoggedInError) Error() string {
	return "Not logged in. Use '{{.CFLoginCommand}}' to log in."
}

func (e NotLoggedInError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"CFLoginCommand": fmt.Sprintf("%s login", e.BinaryName),
	})
}

type NoOrganizationTargetedError struct {
	BinaryName string
}

func (_ NoOrganizationTargetedError) Error() string {
	return "No org targeted, use '{{.Command}}' to target an org."
}

func (e NoOrganizationTargetedError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Command": fmt.Sprintf("%s target -o ORG", e.BinaryName),
	})
}

type NoSpaceTargetedError struct {
	BinaryName string
}

func (_ NoSpaceTargetedError) Error() string {
	return "No space targeted, use '{{.Command}}' to target a space."
}

func (e NoSpaceTargetedError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Command": fmt.Sprintf("%s target -s SPACE", e.BinaryName),
	})
}

type ApplicationNotFoundError struct {
	Name string
}

func (_ ApplicationNotFoundError) Error() string {
	return "App {{.AppName}} not found"
}

func (e ApplicationNotFoundError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"AppName": e.Name,
	})
}

type ServiceInstanceNotFoundError struct {
	Name string
}

func (_ ServiceInstanceNotFoundError) Error() string {
	return "Service instance {{.ServiceInstance}} not found"
}

func (e ServiceInstanceNotFoundError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"ServiceInstance": e.Name,
	})
}

type APINotFoundError struct {
	URL string
}

func (_ APINotFoundError) Error() string {
	return "API endpoint not found at '{{.URL}}'"
}

func (e APINotFoundError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"URL": e.URL,
	})
}

// ArgumentCombinationError represent an error caused by using two command line
// arguments that cannot be used together.
type ArgumentCombinationError struct {
	Arg1 string
	Arg2 string
}

func (_ ArgumentCombinationError) Error() string {
	return "Incorrect Usage: '{{.Arg1}}' and '{{.Arg2}}' cannot be used together."
}

func (e ArgumentCombinationError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Arg1": e.Arg1,
		"Arg2": e.Arg2,
	})
}

type ParseArgumentError struct {
	ArgumentName string
	ExpectedType string
}

func (_ ParseArgumentError) Error() string {
	return "Incorrect usage: Value for {{.ArgumentName}} must be {{.ExpectedType}}"
}

func (e ParseArgumentError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"ArgumentName": e.ArgumentName,
		"ExpectedType": e.ExpectedType,
	})
}

type RequiredArgumentError struct {
	ArgumentName string
}

func (_ RequiredArgumentError) Error() string {
	return "Incorrect Usage: the required argument `{{.ArgumentName}}` was not provided"
}

func (e RequiredArgumentError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"ArgumentName": e.ArgumentName,
	})
}

type ThreeRequiredArgumentsError struct {
	ArgumentName1 string
	ArgumentName2 string
	ArgumentName3 string
}

func (_ ThreeRequiredArgumentsError) Error() string {
	return "Incorrect Usage: the required arguments `{{.ArgumentName1}}`, `{{.ArgumentName2}}`, and `{{.ArgumentName3}}` were not provided"
}

func (e ThreeRequiredArgumentsError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"ArgumentName1": e.ArgumentName1,
		"ArgumentName2": e.ArgumentName2,
		"ArgumentName3": e.ArgumentName3,
	})
}

type MinimumAPIVersionNotMetError struct {
	CurrentVersion string
	MinimumVersion string
}

func (_ MinimumAPIVersionNotMetError) Error() string {
	return "This command requires CF API version {{.MinimumVersion}} or higher. Your target is {{.CurrentVersion}}."
}

func (e MinimumAPIVersionNotMetError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"CurrentVersion": e.CurrentVersion,
		"MinimumVersion": e.MinimumVersion,
	})
}

type LifecycleMinimumAPIVersionNotMetError struct {
	CurrentVersion string
	MinimumVersion string
}

func (_ LifecycleMinimumAPIVersionNotMetError) Error() string {
	return "Lifecycle value 'staging' requires CF API version {{.MinimumVersion}} or higher. Your target is {{.CurrentVersion}}."
}

func (e LifecycleMinimumAPIVersionNotMetError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"CurrentVersion": e.CurrentVersion,
		"MinimumVersion": e.MinimumVersion,
	})
}

type HealthCheckTypeUnsupportedError struct {
	SupportedTypes []string
}

func (_ HealthCheckTypeUnsupportedError) Error() string {
	return "Your target CF API version only supports health check type values {{.SupportedTypes}} and {{.LastSupportedType}}."
}

func (e HealthCheckTypeUnsupportedError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"SupportedTypes":    strings.Join(e.SupportedTypes[:len(e.SupportedTypes)-1], ", "),
		"LastSupportedType": e.SupportedTypes[len(e.SupportedTypes)-1],
	})
}

type UnsupportedURLSchemeError struct {
	UnsupportedURL string
}

func (e UnsupportedURLSchemeError) Error() string {
	return "This command does not support the URL scheme in {{.UnsupportedURL}}."
}

func (e UnsupportedURLSchemeError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"UnsupportedURL": e.UnsupportedURL,
	})
}
