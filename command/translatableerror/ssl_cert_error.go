package translatableerror

type SSLCertError struct {
	Message string
}

func (SSLCertError) Error() string {
	return "SSL Certificate Error {{.Message}}\nTIP: Use 'cf api --skip-ssl-validation' to continue with an insecure API endpoint"
}

func (e SSLCertError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Message": e.Message,
	})
}
