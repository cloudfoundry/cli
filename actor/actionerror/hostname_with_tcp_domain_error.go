package actionerror

type HostnameWithTCPDomainError struct {
}

func (HostnameWithTCPDomainError) Error() string {
	return "cannot use provided hostname with a TCP domain"
}
