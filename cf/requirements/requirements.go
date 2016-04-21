package requirements

type Requirements []Requirement

func (r Requirements) Execute() error {
	var err error

	for _, req := range r {
		err = req.Execute()
		if err != nil {
			return err
		}
	}

	return nil
}
