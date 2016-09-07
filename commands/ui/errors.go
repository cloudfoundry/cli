package ui

import "github.com/nicksnyder/go-i18n/i18n"

type InvalidSSLCertError struct {
	API       string
	Translate i18n.TranslateFunc
}

func (e InvalidSSLCertError) Error() string {
	return e.Translate(
		"Invalid SSL Cert for {{.API}}\nTIP: Use 'cf api --skip-ssl-validation' to continue with an insecure API endpoint",
		map[string]interface{}{
			"API": e.API,
		})
}
