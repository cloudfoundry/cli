package pluginerror

import "fmt"

// SSLValidationHostnameError replaces x509.HostnameError when the server has
// SSL certificate that does not match the hostname.
type SSLValidationHostnameError struct {
	Message string
}

func (e SSLValidationHostnameError) Error() string {
	return fmt.Sprintf("Hostname does not match SSL Certificate (%s)", e.Message)
}
