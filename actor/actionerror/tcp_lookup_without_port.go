package actionerror

type TCPLookupWithoutPort struct {
}

func (TCPLookupWithoutPort) Error() string {
	return "TCP route lookup must include a port"
}
