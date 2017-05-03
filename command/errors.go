package command

import (
	"fmt"
	"strings"
)

type APIRequestError struct {
	Err error
}

func (e APIRequestError) Error() string {
	return "Request error: {{.Error}}\nTIP: If you are behind a firewall and require an HTTP proxy, verify the https_proxy environment variable is correctly set. Else, check your network connection."
}

func (e APIRequestError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Error": e.Err,
	})
}

type InvalidSSLCertError struct {
	API string
}

func (e InvalidSSLCertError) Error() string {
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

func (e SSLCertErrorError) Error() string {
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

func (e NoAPISetError) Error() string {
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

func (e NotLoggedInError) Error() string {
	return "Not logged in. Use '{{.CFLoginCommand}}' to log in."
}

func (e NotLoggedInError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"CFLoginCommand": fmt.Sprintf("%s login", e.BinaryName),
	})
}

type NoTargetedOrganizationError struct {
	BinaryName string
}

func (e NoTargetedOrganizationError) Error() string {
	return "No org targeted, use '{{.Command}}' to target an org."
}

func (e NoTargetedOrganizationError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Command": fmt.Sprintf("%s target -o ORG", e.BinaryName),
	})
}

type NoTargetedSpaceError struct {
	BinaryName string
}

func (e NoTargetedSpaceError) Error() string {
	return "No space targeted, use '{{.Command}}' to target a space."
}

func (e NoTargetedSpaceError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Command": fmt.Sprintf("%s target -s SPACE", e.BinaryName),
	})
}

type ApplicationNotFoundError struct {
	Name string
}

func (e ApplicationNotFoundError) Error() string {
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

func (e ServiceInstanceNotFoundError) Error() string {
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

func (e APINotFoundError) Error() string {
	return "API endpoint not found at '{{.URL}}'"
}

func (e APINotFoundError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"URL": e.URL,
	})
}

type ParseArgumentError struct {
	ArgumentName string
	ExpectedType string
}

func (e ParseArgumentError) Error() string {
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

func (e RequiredArgumentError) Error() string {
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

func (e ThreeRequiredArgumentsError) Error() string {
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

func (e MinimumAPIVersionNotMetError) Error() string {
	return "This command requires CF API version {{.MinimumVersion}}. Your target is {{.CurrentVersion}}."
}

func (e MinimumAPIVersionNotMetError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"CurrentVersion": e.CurrentVersion,
		"MinimumVersion": e.MinimumVersion,
	})
}

type HealthCheckTypeUnsupportedError struct {
	SupportedTypes []string
}

func (e HealthCheckTypeUnsupportedError) Error() string {
	return "Your target CF API version only supports health check type values {{.SupportedTypes}} and {{.LastSupportedType}}."
}

func (e HealthCheckTypeUnsupportedError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"SupportedTypes":    strings.Join(e.SupportedTypes[:len(e.SupportedTypes)-1], ", "),
		"LastSupportedType": e.SupportedTypes[len(e.SupportedTypes)-1],
	})
}
