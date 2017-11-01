package translatableerror

type HostnameWithTCPDomainError struct{}

func (HostnameWithTCPDomainError) Error() string {
	return "The route is invalid: a hostname cannot be used with a TCP domain."
}

func (e HostnameWithTCPDomainError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error())
}
