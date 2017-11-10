package translatableerror

type HostAndPathNotAllowedWithTCPDomainError struct {
	Domain string
}

func (HostAndPathNotAllowedWithTCPDomainError) Error() string {
	return "Host and path not allowed in route with TCP domain {{.Domain}}"
}

func (e HostAndPathNotAllowedWithTCPDomainError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Domain": e.Domain,
	})
}
