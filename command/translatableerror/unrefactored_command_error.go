package translatableerror

type UnrefactoredCommandError struct{}

func (UnrefactoredCommandError) LegacyMain() {}

func (e UnrefactoredCommandError) Error() string {
	return ""
}
