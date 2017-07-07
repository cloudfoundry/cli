package translatableerror

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
