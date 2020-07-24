package v6manifestparser

type validatorFunc func(Parser) error

func (parser Parser) Validate() error {
	var err error

	for _, validator := range parser.validators {
		err = validator(parser)
		if err != nil {
			return err
		}
	}

	return nil
}
