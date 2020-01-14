package ccerror

type OrgQuotaAlreadyExists struct {
	Message string
}

func (e OrgQuotaAlreadyExists) Error() string {
	return e.Message
}
