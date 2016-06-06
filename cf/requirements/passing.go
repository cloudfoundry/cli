package requirements

type Passing struct{}

func (r Passing) Execute() error {
	return nil
}
