package requirements

type Requirement interface {
	Execute() error
}
