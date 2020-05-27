package ccerror

type InvalidStateError struct {
}

func (e InvalidStateError) Error() string {
	return "Cannot stage package unless its state is 'READY'."
}
