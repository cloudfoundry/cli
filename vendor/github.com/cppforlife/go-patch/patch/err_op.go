package patch

type ErrOp struct {
	Err error
}

func (op ErrOp) Apply(_ interface{}) (interface{}, error) {
	return nil, op.Err
}
