package v2

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
