package requirements

type Requirement interface {
	Execute() (success bool)
}
