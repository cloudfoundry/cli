package common

import "fmt"

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

type NoTargetedOrgError struct {
	BinaryName string
}

func (e NoTargetedOrgError) Error() string {
	return "No org targeted, use '{{.Command}}' to target an org."
}

func (e NoTargetedOrgError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Command": fmt.Sprintf("%s target -o ORG", e.BinaryName),
	})
}

type NoTargetedSpaceError struct {
	BinaryName string
}

func (e NoTargetedSpaceError) Error() string {
	return "No space targeted, use '{{.Command}}' to target a space"
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
