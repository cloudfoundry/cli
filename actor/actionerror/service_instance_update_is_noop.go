package actionerror

type ServiceInstanceUpdateIsNoop struct{}

func (ServiceInstanceUpdateIsNoop) Error() string {
	return "ServiceInstanceUpdateIsNoop"
}
