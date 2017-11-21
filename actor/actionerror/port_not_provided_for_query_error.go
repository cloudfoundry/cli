package actionerror

type PortNotProvidedForQueryError struct {
}

// PortNotProvidedForQueryError is returned when trying to GetRouteByComponents, querying by a tcp domain, but omitting a port
func (PortNotProvidedForQueryError) Error() string {
	return "Ambiguous TCP Route lookup. Port must be set with TCP domains."
}
