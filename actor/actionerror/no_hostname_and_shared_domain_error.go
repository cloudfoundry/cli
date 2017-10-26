package actionerror

type NoHostnameAndSharedDomainError struct{}

func (NoHostnameAndSharedDomainError) Error() string {
	return "hostname required for shared domains"
}
