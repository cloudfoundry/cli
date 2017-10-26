package translatableerror

type NoHostnameAndSharedDomainError struct{}

func (NoHostnameAndSharedDomainError) Error() string {
	return "The route is invalid: a hostname is required for shared domains."
}

func (e NoHostnameAndSharedDomainError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error())
}
