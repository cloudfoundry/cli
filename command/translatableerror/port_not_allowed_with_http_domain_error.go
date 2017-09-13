package translatableerror

type PortNotAllowedWithHTTPDomainError struct {
	Domain string
}

func (PortNotAllowedWithHTTPDomainError) Error() string {
	return "Port not allowed in HTTP domain {{.Domain}}"
}

func (e PortNotAllowedWithHTTPDomainError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Domain": e.Domain,
	})
}
