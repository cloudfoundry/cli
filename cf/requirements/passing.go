package requirements

type Passing struct {
	Type string
}

func (r Passing) Execute() error {
	return nil
}
