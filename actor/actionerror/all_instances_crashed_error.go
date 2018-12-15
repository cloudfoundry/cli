package actionerror

type AllInstancesCrashedError struct {
}

func (e AllInstancesCrashedError) Error() string {
	return "All instances crashed"
}
