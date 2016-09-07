package ui

import "github.com/nicksnyder/go-i18n/i18n"

func (e APIRequestError) Error() string {
	return e.translate(
		"Request error: {{.Error}}\nTIP: If you are behind a firewall and require an HTTP proxy, verify the https_proxy environment variable is correctly set. Else, check your network connection.",
		map[string]interface{}{
			"Error": e.Err,
		})
}

func (e APIRequestError) SetTranslation(t i18n.TranslateFunc) error {
	e.translate = t
	return e
}

type InvalidSSLCertError struct {
	API       string
	translate i18n.TranslateFunc
}

func (e InvalidSSLCertError) Error() string {
	return e.translate(
		"Invalid SSL Cert for {{.API}}\nTIP: Use 'cf api --skip-ssl-validation' to continue with an insecure API endpoint",
		map[string]interface{}{
			"API": e.API,
		})
}

func (e InvalidSSLCertError) SetTranslation(t i18n.TranslateFunc) error {
	e.translate = t
	return e
}

type APIRequestError struct {
	Err       error
	translate i18n.TranslateFunc
}
