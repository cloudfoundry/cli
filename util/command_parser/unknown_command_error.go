package command_parser

type UnknownCommandError struct {
}

func (e UnknownCommandError) Error() string {
	return "Unknown Command"
}
