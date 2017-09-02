package translatableerror

type DomainNotFoundError struct {
	Name string
	GUID string
}

func (e DomainNotFoundError) Error() string {
	switch {
	case e.Name != "":
		return "Domain {{.DomainName}} not found"
	case e.GUID != "":
		return "Domain with GUID {{.DomainGUID}} not found"
	default:
		return "Domain not found"
	}
}

func (e DomainNotFoundError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"DomainName": e.Name,
		"DomainGUID": e.GUID,
	})
}
