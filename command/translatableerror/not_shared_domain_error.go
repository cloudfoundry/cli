package translatableerror

type NotSharedDomainError struct {
	DomainName string
}

func (NotSharedDomainError) Error() string {
	return "Domain '{{.DomainName}}' is a private domain, not a shared domain."
}

func (e NotSharedDomainError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"DomainName": e.DomainName,
	})
}
